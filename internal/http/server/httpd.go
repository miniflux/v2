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
	listenAddresses := config.Opts.ListenAddr()
	var httpServers []*http.Server

	certFile := config.Opts.CertFile()
	keyFile := config.Opts.CertKeyFile()
	certDomain := config.Opts.CertDomain()
	var sharedAutocertTLSConfig *tls.Config

	if certDomain != "" {
		slog.Debug("Configuring autocert manager and shared TLS config", slog.String("domain", certDomain))
		certManager := autocert.Manager{
			Cache:      storage.NewCertificateCache(store),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(certDomain),
		}

		sharedAutocertTLSConfig = &tls.Config{}
		sharedAutocertTLSConfig.GetCertificate = certManager.GetCertificate
		sharedAutocertTLSConfig.NextProtos = []string{"h2", "http/1.1", acme.ALPNProto}

		challengeServer := &http.Server{
			Handler: certManager.HTTPHandler(nil),
			Addr:    ":http",
		}
		slog.Info("Starting ACME HTTP challenge server for autocert", slog.String("address", challengeServer.Addr))
		go func() {
			if err := challengeServer.ListenAndServe(); err != http.ErrServerClosed {
				slog.Error("ACME HTTP challenge server failed", slog.Any("error", err))
			}
		}()
		config.Opts.SetHTTPSValue(true)
		httpServers = append(httpServers, challengeServer)
	}

	for i, listenAddr := range listenAddresses {
		server := &http.Server{
			ReadTimeout:  config.Opts.HTTPServerTimeout(),
			WriteTimeout: config.Opts.HTTPServerTimeout(),
			IdleTimeout:  config.Opts.HTTPServerTimeout(),
			Handler:      newRouter(store, pool),
		}

		isUNIXSocket := strings.HasPrefix(listenAddr, "/")
		isListenPID := os.Getenv("LISTEN_PID") == strconv.Itoa(os.Getpid())

		if !isUNIXSocket && !isListenPID {
			server.Addr = listenAddr
		}

		switch {
		case isListenPID:
			if i == 0 {
				slog.Info("Starting server using systemd socket for the first listen address", slog.String("address_info", listenAddr))
				startSystemdSocketServer(server)
			} else {
				slog.Warn("Systemd socket activation: Only the first listen address is used by systemd. Other addresses are ignored.", slog.String("skipped_address", listenAddr))
				continue
			}
		case isUNIXSocket:
			startUnixSocketServer(server, listenAddr)
		case certDomain != "" && (listenAddr == ":https" || (i == 0 && strings.Contains(listenAddr, ":"))):
			server.Addr = listenAddr
			startAutoCertTLSServer(server, sharedAutocertTLSConfig)
		case certFile != "" && keyFile != "":
			server.Addr = listenAddr
			startTLSServer(server, certFile, keyFile)
			config.Opts.SetHTTPSValue(true)
		default:
			server.Addr = listenAddr
			startHTTPServer(server)
		}

		httpServers = append(httpServers, server)
	}

	return httpServers
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

	go func() {
		certFile := config.Opts.CertFile()
		keyFile := config.Opts.CertKeyFile()

		if certFile != "" && keyFile != "" {
			slog.Info("Starting TLS server using a Unix socket",
				slog.String("socket", socketFile),
				slog.String("cert_file", certFile),
				slog.String("key_file", keyFile),
			)
			// Ensure HTTPS is marked as true if any listener uses TLS
			config.Opts.SetHTTPSValue(true)
			if err := server.ServeTLS(listener, certFile, keyFile); err != http.ErrServerClosed {
				printErrorAndExit("TLS Unix socket server failed to start on %s: %v", socketFile, err)
			}
		} else {
			slog.Info("Starting server using a Unix socket", slog.String("socket", socketFile))
			if err := server.Serve(listener); err != http.ErrServerClosed {
				printErrorAndExit("Unix socket server failed to start on %s: %v", socketFile, err)
			}
		}
	}()
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
