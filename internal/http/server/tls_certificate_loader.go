// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"crypto/tls"
	"log/slog"
	"path/filepath"
	"sync"
)

// certificateLoader loads and caches a TLS certificate from disk, and
// provides a reload method that can be triggered on SIGHUP.
type certificateLoader struct {
	mu       sync.RWMutex
	cert     *tls.Certificate
	certFile string
	keyFile  string
}

func newCertificateLoader(certFile, keyFile string) (*certificateLoader, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	loader := &certificateLoader{
		cert:     &cert,
		certFile: filepath.Clean(certFile),
		keyFile:  filepath.Clean(keyFile),
	}

	slog.Info("TLS certificate loaded",
		slog.String("cert_file", loader.certFile),
		slog.String("key_file", loader.keyFile),
	)

	return loader, nil
}

// getCertificate returns the currently cached TLS certificate. It satisfies
// the tls.Config.GetCertificate callback and is called by the TLS layer on
// every handshake.
func (cl *certificateLoader) getCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return cl.cert, nil
}

// Reload loads the certificate and key from disk and replaces the cached
// copy. If loading fails, the existing certificate is kept and the error
// is logged.
func (cl *certificateLoader) Reload() {
	cert, err := tls.LoadX509KeyPair(cl.certFile, cl.keyFile)
	if err != nil {
		slog.Error("Unable to reload TLS certificate",
			slog.String("cert_file", cl.certFile),
			slog.String("key_file", cl.keyFile),
			slog.Any("error", err),
		)
		return
	}

	cl.mu.Lock()
	cl.cert = &cert
	cl.mu.Unlock()

	slog.Info("TLS certificate reloaded successfully",
		slog.String("cert_file", cl.certFile),
		slog.String("key_file", cl.keyFile),
	)
}
