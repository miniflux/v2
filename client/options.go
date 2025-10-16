// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package client // import "miniflux.app/v2/client"

import "net/http"

type Option func(*request)

// WithAPIKey sets the API key for the client.
func WithAPIKey(apiKey string) Option {
	return func(r *request) {
		r.apiKey = apiKey
	}
}

// WithCredentials sets the username and password for the client.
func WithCredentials(username, password string) Option {
	return func(r *request) {
		r.username = username
		r.password = password
	}
}

// WithHTTPClient sets the HTTP client for the client.
func WithHTTPClient(client *http.Client) Option {
	return func(r *request) {
		r.client = client
	}
}
