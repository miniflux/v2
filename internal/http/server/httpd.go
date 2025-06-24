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
	"time"

	"miniflux.app/v2/internal/api"
	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/fever"
	"miniflux.app/v2/internal/googlereader"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/ui"
	"miniflux.app/v2/internal/version"
	"miniflux.app/v2/internal/worker"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
		config.Opts.HTTPS = true
		httpServers = append(httpServers, challengeServer)
	}

	for i, listenAddr := range listenAddresses {
		server := &http.Server{
			ReadTimeout:  time.Duration(config.Opts.HTTPServerTimeout()) * time.Second,
			WriteTimeout: time.Duration(config.Opts.HTTPServerTimeout()) * time.Second,
			IdleTimeout:  time.Duration(config.Opts.HTTPServerTimeout()) * time.Second,
			Handler:      setupHandler(store, pool),
		}

		if !strings.HasPrefix(listenAddr, "/") && os.Getenv("LISTEN_PID") != strconv.Itoa(os.Getpid()) {
			server.Addr = listenAddr
		}

		shouldAddServer := true

		switch {
		case os.Getenv("LISTEN_PID") == strconv.Itoa(os.Getpid()):
			if i == 0 {
				slog.Info("Starting server using systemd socket for the first listen address", slog.String("address_info", listenAddr))
				startSystemdSocketServer(server)
			} else {
				slog.Warn("Systemd socket activation: Only the first listen address is used by systemd. Other addresses ignored.", slog.String("skipped_address", listenAddr))
				shouldAddServer = false
			}
		case strings.HasPrefix(listenAddr, "/"): // Unix socket
			startUnixSocketServer(server, listenAddr)
		case certDomain != "" && (listenAddr == ":https" || (i == 0 && strings.Contains(listenAddr, ":"))):
			server.Addr = listenAddr
			startAutoCertTLSServer(server, sharedAutocertTLSConfig)
		case certFile != "" && keyFile != "":
			server.Addr = listenAddr
			startTLSServer(server, certFile, keyFile)
			config.Opts.HTTPS = true
		default:
			server.Addr = listenAddr
			startHTTPServer(server)
		}

		if shouldAddServer {
			httpServers = append(httpServers, server)
		}
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
			config.Opts.HTTPS = true
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

func setupHandler(store *storage.Storage, pool *worker.Pool) *mux.Router {
	livenessProbe := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
	readinessProbe := func(w http.ResponseWriter, r *http.Request) {
		if err := store.Ping(); err != nil {
			http.Error(w, fmt.Sprintf("Database Connection Error: %q", err), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	router := mux.NewRouter()

	// These routes do not take the base path into consideration and are always available at the root of the server.
	router.HandleFunc("/liveness", livenessProbe).Name("liveness")
	router.HandleFunc("/healthz", livenessProbe).Name("healthz")
	router.HandleFunc("/readiness", readinessProbe).Name("readiness")
	router.HandleFunc("/readyz", readinessProbe).Name("readyz")

	var subrouter *mux.Router
	if config.Opts.BasePath() != "" {
		subrouter = router.PathPrefix(config.Opts.BasePath()).Subrouter()
	} else {
		subrouter = router.NewRoute().Subrouter()
	}

	if config.Opts.HasMaintenanceMode() {
		subrouter.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(config.Opts.MaintenanceMessage()))
			})
		})
	}

	subrouter.Use(middleware)

	fever.Serve(subrouter, store)
	googlereader.Serve(subrouter, store)
	api.Serve(subrouter, store, pool)
	ui.Serve(subrouter, store, pool)

	subrouter.HandleFunc("/healthcheck", readinessProbe).Name("healthcheck")

	subrouter.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(version.Version))
	}).Name("version")

	if config.Opts.HasMetricsCollector() {
		subrouter.Handle("/metrics", promhttp.Handler()).Name("metrics")
		subrouter.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				route := mux.CurrentRoute(r)

				// Returns a 404 if the client is not authorized to access the metrics endpoint.
				if route.GetName() == "metrics" && !isAllowedToAccessMetricsEndpoint(r) {
					slog.Warn("Authentication failed while accessing the metrics endpoint",
						slog.String("client_ip", request.ClientIP(r)),
						slog.String("client_user_agent", r.UserAgent()),
						slog.String("client_remote_addr", r.RemoteAddr),
					)
					http.NotFound(w, r)
					return
				}

				next.ServeHTTP(w, r)
			})
		})
	}

	return router
}

func isAllowedToAccessMetricsEndpoint(r *http.Request) bool {
	clientIP := request.ClientIP(r)

	if config.Opts.MetricsUsername() != "" && config.Opts.MetricsPassword() != "" {
		username, password, authOK := r.BasicAuth()
		if !authOK {
			slog.Warn("Metrics endpoint accessed without authentication header",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("client_user_agent", r.UserAgent()),
				slog.String("client_remote_addr", r.RemoteAddr),
			)
			return false
		}

		if username == "" || password == "" {
			slog.Warn("Metrics endpoint accessed with empty username or password",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("client_user_agent", r.UserAgent()),
				slog.String("client_remote_addr", r.RemoteAddr),
			)
			return false
		}

		if username != config.Opts.MetricsUsername() || password != config.Opts.MetricsPassword() {
			slog.Warn("Metrics endpoint accessed with invalid username or password",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("client_user_agent", r.UserAgent()),
				slog.String("client_remote_addr", r.RemoteAddr),
			)
			return false
		}
	}

	remoteIP := request.FindRemoteIP(r)
	if remoteIP == "@" {
		// This indicates a request sent via a Unix socket, always consider these trusted.
		return true
	}

	for _, cidr := range config.Opts.MetricsAllowedNetworks() {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			slog.Error("Metrics endpoint accessed with invalid CIDR",
				slog.Bool("authentication_failed", true),
				slog.String("client_ip", clientIP),
				slog.String("client_user_agent", r.UserAgent()),
				slog.String("client_remote_addr", r.RemoteAddr),
				slog.String("cidr", cidr),
			)
			return false
		}

		// We use r.RemoteAddr in this case because HTTP headers like X-Forwarded-For can be easily spoofed.
		// The recommendation is to use HTTP Basic authentication.
		if network.Contains(net.ParseIP(remoteIP)) {
			return true
		}
	}

	return false
}

func printErrorAndExit(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	slog.Error(message)
	fmt.Fprintf(os.Stderr, "%v\n", message)
	os.Exit(1)
}
