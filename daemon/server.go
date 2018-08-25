// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package daemon // import "miniflux.app/daemon"

import (
	"crypto/tls"
	"net/http"
	"time"

	"miniflux.app/config"
	"miniflux.app/locale"
	"miniflux.app/logger"
	"miniflux.app/reader/feed"
	"miniflux.app/scheduler"
	"miniflux.app/storage"

	"golang.org/x/crypto/acme/autocert"
)

func newServer(cfg *config.Config, store *storage.Storage, pool *scheduler.WorkerPool, feedHandler *feed.Handler, translator *locale.Translator) *http.Server {
	certFile := cfg.CertFile()
	keyFile := cfg.KeyFile()
	certDomain := cfg.CertDomain()
	certCache := cfg.CertCache()
	server := &http.Server{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		Addr:         cfg.ListenAddr(),
		Handler:      routes(cfg, store, feedHandler, pool, translator),
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
			logger.Fatal(server.Serve(certManager.Listener()).Error())
		}()
	} else if certFile != "" && keyFile != "" {
		server.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		cfg.IsHTTPS = true

		go func() {
			logger.Info(`Listening on "%s" by using certificate "%s" and key "%s"`, server.Addr, certFile, keyFile)
			logger.Fatal(server.ListenAndServeTLS(certFile, keyFile).Error())
		}()
	} else {
		go func() {
			logger.Info(`Listening on "%s" without TLS`, server.Addr)
			logger.Fatal(server.ListenAndServe().Error())
		}()
	}

	return server
}
