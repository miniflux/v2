// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/miniflux/miniflux/scheduler"
	"golang.org/x/crypto/acme/autocert"

	"github.com/miniflux/miniflux/config"
	"github.com/miniflux/miniflux/reader/feed"
	"github.com/miniflux/miniflux/storage"
)

// NewServer returns a new HTTP server.
func NewServer(cfg *config.Config, store *storage.Storage, pool *scheduler.WorkerPool, feedHandler *feed.Handler) *http.Server {
	return startServer(cfg, getRoutes(cfg, store, feedHandler, pool))
}

func startServer(cfg *config.Config, handler *mux.Router) *http.Server {
	certFile := cfg.Get("CERT_FILE", config.DefaultCertFile)
	keyFile := cfg.Get("KEY_FILE", config.DefaultKeyFile)
	certDomain := cfg.Get("CERT_DOMAIN", config.DefaultCertDomain)
	certCache := cfg.Get("CERT_CACHE", config.DefaultCertCache)
	server := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Addr:         cfg.Get("LISTEN_ADDR", config.DefaultListenAddr),
		Handler:      handler,
	}

	if certDomain != "" && certCache != "" {
		server.Addr = ":https"
		certManager := autocert.Manager{
			Cache:      autocert.DirCache(certCache),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(certDomain),
		}

		go func() {
			log.Printf(`Listening on "%s" by using auto-configured certificate for "%s"`, server.Addr, certDomain)
			log.Fatalln(server.Serve(certManager.Listener()))
		}()
	} else if certFile != "" && keyFile != "" {
		server.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}

		go func() {
			log.Printf(`Listening on "%s" by using certificate "%s" and key "%s"`, server.Addr, certFile, keyFile)
			log.Fatalln(server.ListenAndServeTLS(certFile, keyFile))
		}()
	} else {
		go func() {
			log.Printf(`Listening on "%s" without TLS`, server.Addr)
			log.Fatalln(server.ListenAndServe())
		}()
	}

	return server
}
