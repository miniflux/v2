// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server // import "miniflux.app/v2/internal/http/server"

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/worker"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

func StartWebServer(store *storage.Storage, pool *worker.Pool) []*http.Server {
	var servers []*http.Server

	autocertTLSConfig, challengeServer := setupAutocert(store)
	if challengeServer != nil {
		servers = append(servers, challengeServer)
	}

	certFile := config.Opts.CertFile()
	keyFile := config.Opts.CertKeyFile()
	certDomain := config.Opts.CertDomain()

	targets := determineListenTargets(config.Opts.ListenAddr(), certDomain, certFile, keyFile)

	if autocertTLSConfig != nil || anyTLS(targets) {
		config.Opts.SetHTTPSValue(true)
	}

	for _, t := range targets {
		srv := &http.Server{
			Addr:         t.address,
			ReadTimeout:  config.Opts.HTTPServerTimeout(),
			WriteTimeout: config.Opts.HTTPServerTimeout(),
			IdleTimeout:  config.Opts.HTTPServerTimeout(),
			Handler:      newRouter(store, pool),
		}

		switch t.mode {
		case modeSystemd:
			startSystemdSocketServer(srv)
		case modeUnixSocket:
			startUnixSocketServer(srv, t.address)
		case modeUnixSocketTLS:
			startUnixSocketTLSServer(srv, t.address, t.certFile, t.keyFile)
		case modeAutocertTLS:
			startAutoCertTLSServer(srv, autocertTLSConfig)
		case modeTLS:
			startTLSServer(srv, t.certFile, t.keyFile)
		default:
			startHTTPServer(srv)
		}

		servers = append(servers, srv)
	}

	return servers
}

type listenerMode int

const (
	modeHTTP listenerMode = iota
	modeTLS
	modeAutocertTLS
	modeUnixSocket
	modeUnixSocketTLS
	modeSystemd
)

type listenTarget struct {
	address  string
	mode     listenerMode
	certFile string
	keyFile  string
}

func determineListenTargets(addresses []string, certDomain, certFile, keyFile string) []listenTarget {
	isSystemd := os.Getenv("LISTEN_PID") == strconv.Itoa(os.Getpid())
	hasCertFiles := certFile != "" && keyFile != ""
	hasAutocert := certDomain != ""

	var targets []listenTarget

	for i, addr := range addresses {
		if isSystemd {
			if i == 0 {
				targets = append(targets, listenTarget{address: addr, mode: modeSystemd})
			} else {
				slog.Warn("Systemd socket activation: only the first listen address is used, others are ignored",
					slog.String("skipped_address", addr),
				)
			}
			continue
		}

		isUnix := strings.HasPrefix(addr, "/")

		switch {
		case isUnix && hasCertFiles:
			targets = append(targets, listenTarget{address: addr, mode: modeUnixSocketTLS, certFile: certFile, keyFile: keyFile})
		case isUnix:
			targets = append(targets, listenTarget{address: addr, mode: modeUnixSocket})
		case hasAutocert && (addr == ":https" || (i == 0 && strings.Contains(addr, ":"))):
			targets = append(targets, listenTarget{address: addr, mode: modeAutocertTLS})
		case hasCertFiles:
			targets = append(targets, listenTarget{address: addr, mode: modeTLS, certFile: certFile, keyFile: keyFile})
		default:
			targets = append(targets, listenTarget{address: addr, mode: modeHTTP})
		}
	}

	return targets
}

func anyTLS(targets []listenTarget) bool {
	for _, t := range targets {
		switch t.mode {
		case modeTLS, modeAutocertTLS, modeUnixSocketTLS:
			return true
		}
	}
	return false
}

func setupAutocert(store *storage.Storage) (*tls.Config, *http.Server) {
	certDomain := config.Opts.CertDomain()
	if certDomain == "" {
		return nil, nil
	}

	slog.Debug("Configuring autocert manager", slog.String("domain", certDomain))
	certManager := autocert.Manager{
		Cache:      storage.NewCertificateCache(store),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(certDomain),
	}

	tlsConfig := &tls.Config{
		NextProtos: []string{"h2", "http/1.1", acme.ALPNProto},
	}
	tlsConfig.GetCertificate = certManager.GetCertificate

	challengeServer := &http.Server{
		Handler: certManager.HTTPHandler(nil),
		Addr:    ":http",
	}

	slog.Info("Starting ACME HTTP challenge server", slog.String("address", challengeServer.Addr))
	go func() {
		if err := challengeServer.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("ACME HTTP challenge server failed", slog.Any("error", err))
		}
	}()

	return tlsConfig, challengeServer
}

func startSystemdSocketServer(server *http.Server) {
	go func() {
		f := os.NewFile(3, "systemd socket")
		listener, err := net.FileListener(f)
		if err != nil {
			printErrorAndExit(`Unable to create listener from systemd socket: %v`, err)
		}

		slog.Info(`Starting server using systemd socket`)
		if err := server.Serve(listener); err != http.ErrServerClosed {
			printErrorAndExit(`Systemd socket server failed to start: %v`, err)
		}
	}()
}

func startUnixSocketServer(server *http.Server, socketFile string) {
	listener := createUnixSocketListener(socketFile)

	go func() {
		slog.Info("Starting server using a Unix socket", slog.String("socket", socketFile))
		if err := server.Serve(listener); err != http.ErrServerClosed {
			printErrorAndExit("Unix socket server failed to start on %s: %v", socketFile, err)
		}
	}()
}

func startUnixSocketTLSServer(server *http.Server, socketFile, certFile, keyFile string) {
	listener := createUnixSocketListener(socketFile)

	go func() {
		slog.Info("Starting TLS server using a Unix socket",
			slog.String("socket", socketFile),
			slog.String("cert_file", certFile),
			slog.String("key_file", keyFile),
		)
		if err := server.ServeTLS(listener, certFile, keyFile); err != http.ErrServerClosed {
			printErrorAndExit("TLS Unix socket server failed to start on %s: %v", socketFile, err)
		}
	}()
}

func createUnixSocketListener(socketFile string) net.Listener {
	if err := os.Remove(socketFile); err != nil && !os.IsNotExist(err) {
		printErrorAndExit("Unable to remove existing Unix socket %s: %v", socketFile, err)
	}
	listener, err := net.Listen("unix", socketFile)
	if err != nil {
		printErrorAndExit(`Server failed to listen on Unix socket %s: %v`, socketFile, err)
	}

	if err := os.Chmod(socketFile, 0666); err != nil {
		printErrorAndExit(`Unable to change socket permission for %s: %v`, socketFile, err)
	}

	return listener
}

func startAutoCertTLSServer(server *http.Server, autoTLSConfig *tls.Config) {
	if server.TLSConfig == nil {
		server.TLSConfig = &tls.Config{}
	}
	server.TLSConfig.GetCertificate = autoTLSConfig.GetCertificate
	server.TLSConfig.NextProtos = autoTLSConfig.NextProtos

	go func() {
		slog.Info("Starting TLS server using automatic certificate management",
			slog.String("listen_address", server.Addr),
		)
		if err := server.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			printErrorAndExit("Autocert server failed to start on %s: %v", server.Addr, err)
		}
	}()
}

func startTLSServer(server *http.Server, certFile, keyFile string) {
	go func() {
		slog.Info("Starting TLS server using a certificate",
			slog.String("listen_address", server.Addr),
			slog.String("cert_file", certFile),
			slog.String("key_file", keyFile),
		)
		if err := server.ListenAndServeTLS(certFile, keyFile); err != http.ErrServerClosed {
			printErrorAndExit("TLS server failed to start on %s: %v", server.Addr, err)
		}
	}()
}

func startHTTPServer(server *http.Server) {
	go func() {
		slog.Info("Starting HTTP server",
			slog.String("listen_address", server.Addr),
		)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			printErrorAndExit("HTTP server failed to start on %s: %v", server.Addr, err)
		}
	}()
}

func printErrorAndExit(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	slog.Error(message)
	fmt.Fprintf(os.Stderr, "%v\n", message)
	os.Exit(1)
}
