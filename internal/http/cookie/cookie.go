// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cookie // import "miniflux.app/v2/internal/http/cookie"

import (
	"net/http"
	"time"

	"miniflux.app/v2/internal/config"
)

// Cookie names.
const (
	CookieAppSessionID  = "MinifluxAppSessionID"
	CookieUserSessionID = "MinifluxUserSessionID"
)

// New creates a new cookie.
func New(name, value string, isHTTPS bool, path string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     basePath(path),
		Secure:   isHTTPS,
		HttpOnly: true,
		Expires:  time.Now().Add(time.Duration(config.Opts.CleanupRemoveSessionsDays()) * 24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
	}
}

// Expired returns an expired cookie.
func Expired(name string, isHTTPS bool, path string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     basePath(path),
		Secure:   isHTTPS,
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		SameSite: http.SameSiteLaxMode,
	}
}

func basePath(path string) string {
	if path == "" {
		return "/"
	}
	return path
}
