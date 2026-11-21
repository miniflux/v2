// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/reader/fetcher"
)

// ValidateTLSCertificates checks optional TLS credentials for a feed or discovery request.
func ValidateTLSCertificates(clientCertificate, clientKey, caCertificate string) *locale.LocalizedError {
	if err := fetcher.ValidateTLSCertificates(clientCertificate, clientKey, caCertificate); err != nil {
		return locale.NewLocalizedError("error.invalid_tls_certificates", err)
	}
	return nil
}
