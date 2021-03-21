// Copyright 2020 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

// SubscriptionDiscoveryRequest represents a request to discover subscriptions.
type SubscriptionDiscoveryRequest struct {
	URL                         string `json:"url"`
	UserAgent                   string `json:"user_agent"`
	Cookie                      string `json:"cookie"`
	Username                    string `json:"username"`
	Password                    string `json:"password"`
	FetchViaProxy               bool   `json:"fetch_via_proxy"`
	AllowSelfSignedCertificates bool   `json:"allow_self_signed_certificates"`
}
