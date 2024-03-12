// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

// SubscriptionDiscoveryRequest represents a request to discover subscriptions.
type SubscriptionDiscoveryRequest struct {
	URL                         string `json:"url"`
	UserAgent                   string `json:"user_agent"`
	Cookie                      string `json:"cookie"`
	Username                    string `json:"username"`
	Password                    string `json:"password"`
	FetchViaProxy               bool   `json:"fetch_via_proxy"`
	AllowSelfSignedCertificates bool   `json:"allow_self_signed_certificates"`
	DisableHTTP2                bool   `json:"disable_http2"`
}
