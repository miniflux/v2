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
)

// New create a new cookie.
func New(name, value string, isHTTPS bool) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Secure:   isHTTPS,
		HttpOnly: true,
	}
}

// Expired returns an expired cookie.
func Expired(name string, isHTTPS bool) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Secure:   isHTTPS,
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}
