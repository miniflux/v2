// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fetcher // import "miniflux.app/v2/internal/reader/fetcher"

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

// ValidateTLSCertificates checks optional PEM credentials without making an HTTP request.
func ValidateTLSCertificates(clientCertificate, clientKey, caCertificate string) error {
	_, _, err := parseTLSCertificates(clientCertificate, clientKey, caCertificate)
	return err
}

func parseTLSCertificates(clientCertificate, clientKey, caCertificate string) (*tls.Certificate, []*x509.Certificate, error) {
	var clientPair *tls.Certificate
	if (clientCertificate == "") != (clientKey == "") {
		return nil, nil, errors.New("the client certificate and private key must be provided together")
	}
	if clientCertificate != "" {
		pair, err := tls.X509KeyPair([]byte(clientCertificate), []byte(clientKey))
		if err != nil {
			return nil, nil, fmt.Errorf("unable to load the client certificate and private key: %w", err)
		}
		for _, certificate := range pair.Certificate[1:] {
			if _, err := x509.ParseCertificate(certificate); err != nil {
				return nil, nil, fmt.Errorf("unable to parse the client certificate chain: %w", err)
			}
		}
		clientPair = &pair
	}

	var caCertificates []*x509.Certificate
	for rest := bytes.TrimSpace([]byte(caCertificate)); len(rest) > 0; rest = bytes.TrimSpace(rest) {
		block, remaining := pem.Decode(rest)
		if block == nil || block.Type != "CERTIFICATE" || !bytes.HasPrefix(rest, []byte("-----BEGIN CERTIFICATE-----")) {
			return nil, nil, errors.New("the CA certificate must contain PEM-encoded certificates")
		}
		certificate, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to parse the CA certificate: %w", err)
		}
		caCertificates = append(caCertificates, certificate)
		rest = remaining
	}
	if caCertificate != "" && len(caCertificates) == 0 {
		return nil, nil, errors.New("the CA certificate must contain PEM-encoded certificates")
	}
	return clientPair, caCertificates, nil
}

func systemRootsWithCustomCAs(certificates []*x509.Certificate) (*x509.CertPool, error) {
	roots, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("unable to load system root certificates: %w", err)
	}
	for _, certificate := range certificates {
		roots.AddCert(certificate)
	}
	return roots, nil
}
