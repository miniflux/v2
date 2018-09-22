// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package daemon // import "miniflux.app/daemon"

import (
	"crypto/tls"
	"net/http"
	"time"

	"miniflux.app/config"
	"miniflux.app/logger"
	"miniflux.app/reader/feed"
	"miniflux.app/scheduler"
	"miniflux.app/storage"

	"golang.org/x/crypto/acme/autocert"
)

func newServer(cfg *config.Config, store *storage.Storage, pool *scheduler.WorkerPool, feedHandler *feed.Handler) *http.Server {
	certFile := cfg.CertFile()
	keyFile := cfg.KeyFile()
	certDomain := cfg.CertDomain()
	certCache := cfg.CertCache()
	server := &http.Server{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		Addr:         cfg.ListenAddr(),
		Handler:      routes(cfg, store, feedHandler, pool),
	}

	if certDomain != "" && certCache != "" {
		cfg.IsHTTPS = true
		server.Addr = ":https"
		certManager := autocert.Manager{
			Cache:      autocert.DirCache(certCache),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(certDomain),
		}

		// Handle http-01 challenge.
		s := &http.Server{
			Handler: certManager.HTTPHandler(nil),
			Addr:    ":http",
		}
		go s.ListenAndServe()

		go func() {
			logger.Info(`Listening on "%s" by using auto-configured certificate for "%s"`, server.Addr, certDomain)
			if err := server.Serve(certManager.Listener()); err != http.ErrServerClosed {
				logger.Fatal(`Server failed to start: %v`, err)
			}
		}()
	} else if certFile != "" && keyFile != "" {
		cfg.IsHTTPS = true

		// See https://blog.cloudflare.com/exposing-go-on-the-internet/
		// And https://wiki.mozilla.org/Security/Server_Side_TLS
		server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}

		go func() {
			logger.Info(`Listening on "%s" by using certificate "%s" and key "%s"`, server.Addr, certFile, keyFile)
			if err := server.ListenAndServeTLS(certFile, keyFile); err != http.ErrServerClosed {
				logger.Fatal(`Server failed to start: %v`, err)
			}
		}()
	} else {
		go func() {
			logger.Info(`Listening on "%s" without TLS`, server.Addr)
			if err := server.ListenAndServe(); err != http.ErrServerClosed {
				logger.Fatal(`Server failed to start: %v`, err)
			}
		}()
	}

	return server
}
