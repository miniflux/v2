// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package cookie

import (
	"net/http"
	"time"
)

// Cookie names.
const (
	CookieSessionID     = "sessionID"
	CookieUserSessionID = "userSessionID"

	// Cookie duration in days.
	cookieDuration = 30
)

// New creates a new cookie.
func New(name, value string, isHTTPS bool, path string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     basePath(path),
		Secure:   isHTTPS,
		HttpOnly: true,
		Expires:  time.Now().Add(cookieDuration * 24 * time.Hour),
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
	}
}

func basePath(path string) string {
	if path == "" {
		return "/"
	}
	return path
}
