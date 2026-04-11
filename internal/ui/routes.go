// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"net/url"
	"strings"
)

// isStaticAssetRoute checks if the request path corresponds to a static
// asset route that does not require a web session.
func isStaticAssetRoute(r *http.Request) bool {
	path := r.URL.Path

	switch path {
	case "/favicon.ico", "/robots.txt":
		return true
	}

	return strings.HasPrefix(path, "/stylesheets/") ||
		strings.HasPrefix(path, "/js/") ||
		strings.HasPrefix(path, "/icon/") ||
		strings.HasPrefix(path, "/feed-icon/")
}

// isPublicRoute checks if the request path corresponds to a route that
// does not require authentication. The path is expected to have the base
// path already stripped.
func isPublicRoute(r *http.Request) bool {
	if isStaticAssetRoute(r) {
		return true
	}

	path := r.URL.Path

	switch path {
	case "/", "/login", "/manifest.json",
		"/healthcheck", "/offline",
		"/webauthn/login/begin", "/webauthn/login/finish":
		return true
	}

	return strings.HasPrefix(path, "/oauth2/") && (strings.HasSuffix(path, "/redirect") || strings.HasSuffix(path, "/callback")) ||
		strings.HasPrefix(path, "/share/") ||
		strings.HasPrefix(path, "/proxy/")
}

// loginRedirectURL builds the login page URL with the given request URI
// stored in the redirect_url query parameter.
func loginRedirectURL(basePath, requestURI string) string {
	loginURL, _ := url.Parse(basePath + "/")
	values := loginURL.Query()
	values.Set("redirect_url", requestURI)
	loginURL.RawQuery = values.Encode()
	return loginURL.String()
}
