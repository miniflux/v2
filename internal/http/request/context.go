// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package request // import "miniflux.app/v2/internal/http/request"

import (
	"net/http"

	"miniflux.app/v2/internal/model"
)

// ContextKey represents a context key.
type ContextKey int

// List of context keys.
const (
	UserIDContextKey ContextKey = iota
	UserNameContextKey
	UserTimezoneContextKey
	IsAdminUserContextKey
	IsAuthenticatedContextKey
	WebSessionContextKey
	ClientIPContextKey
	GoogleReaderTokenKey
)

// WebSession returns the current web session from the request context, if present.
func WebSession(r *http.Request) *model.WebSession {
	if v := r.Context().Value(WebSessionContextKey); v != nil {
		if value, valid := v.(*model.WebSession); valid {
			return value
		}
	}
	return nil
}

// GoogleReaderToken returns the Google Reader token from the request context, if present.
func GoogleReaderToken(r *http.Request) string {
	return getContextStringValue(r, GoogleReaderTokenKey)
}

// IsAdminUser reports whether the logged-in user is an administrator.
func IsAdminUser(r *http.Request) bool {
	return getContextBoolValue(r, IsAdminUserContextKey)
}

// IsAuthenticated reports whether the user is authenticated.
func IsAuthenticated(r *http.Request) bool {
	if getContextBoolValue(r, IsAuthenticatedContextKey) {
		return true
	}

	if session := WebSession(r); session != nil {
		return session.IsAuthenticated()
	}

	return false
}

// UserID returns the logged-in user's ID from the request context.
func UserID(r *http.Request) int64 {
	if userID := getContextInt64Value(r, UserIDContextKey); userID != 0 {
		return userID
	}

	if session := WebSession(r); session != nil {
		if id, ok := session.UserID(); ok {
			return id
		}
	}

	return 0
}

// UserName returns the logged-in user's username, or "unknown" when unset.
func UserName(r *http.Request) string {
	value := getContextStringValue(r, UserNameContextKey)
	if value == "" {
		value = "unknown"
	}
	return value
}

// UserTimezone returns the user's timezone, defaulting to "UTC" when unset.
func UserTimezone(r *http.Request) string {
	value := getContextStringValue(r, UserTimezoneContextKey)
	if value == "" {
		value = "UTC"
	}
	return value
}

// ClientIP returns the client IP address stored in the request context.
func ClientIP(r *http.Request) string {
	return getContextStringValue(r, ClientIPContextKey)
}

func getContextStringValue(r *http.Request, key ContextKey) string {
	if v := r.Context().Value(key); v != nil {
		if value, valid := v.(string); valid {
			return value
		}
	}
	return ""
}

func getContextBoolValue(r *http.Request, key ContextKey) bool {
	if v := r.Context().Value(key); v != nil {
		if value, valid := v.(bool); valid {
			return value
		}
	}
	return false
}

func getContextInt64Value(r *http.Request, key ContextKey) int64 {
	if v := r.Context().Value(key); v != nil {
		if value, valid := v.(int64); valid {
			return value
		}
	}
	return 0
}
