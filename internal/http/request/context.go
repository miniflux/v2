// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package request // import "miniflux.app/v2/internal/http/request"

import (
	"net/http"
	"strconv"

	"miniflux.app/v2/internal/model"
)

// ContextKey represents a context key.
type ContextKey int

// List of context keys.
const (
	UserIDContextKey ContextKey = iota
	UserTimezoneContextKey
	IsAdminUserContextKey
	IsAuthenticatedContextKey
	UserSessionTokenContextKey
	UserLanguageContextKey
	UserThemeContextKey
	SessionIDContextKey
	CSRFContextKey
	OAuth2StateContextKey
	OAuth2CodeVerifierContextKey
	FlashMessageContextKey
	FlashErrorMessageContextKey
	PocketRequestTokenContextKey
	LastForceRefreshContextKey
	ClientIPContextKey
	GoogleReaderToken
	WebAuthnDataContextKey
)

func WebAuthnSessionData(r *http.Request) *model.WebAuthnSession {
	if v := r.Context().Value(WebAuthnDataContextKey); v != nil {
		value, valid := v.(model.WebAuthnSession)
		if !valid {
			return nil
		}

		return &value
	}

	return nil
}

// GoolgeReaderToken returns the google reader token if it exists.
func GoolgeReaderToken(r *http.Request) string {
	return getContextStringValue(r, GoogleReaderToken)
}

// IsAdminUser checks if the logged user is administrator.
func IsAdminUser(r *http.Request) bool {
	return getContextBoolValue(r, IsAdminUserContextKey)
}

// IsAuthenticated returns a boolean if the user is authenticated.
func IsAuthenticated(r *http.Request) bool {
	return getContextBoolValue(r, IsAuthenticatedContextKey)
}

// UserID returns the UserID of the logged user.
func UserID(r *http.Request) int64 {
	return getContextInt64Value(r, UserIDContextKey)
}

// UserTimezone returns the timezone used by the logged user.
func UserTimezone(r *http.Request) string {
	value := getContextStringValue(r, UserTimezoneContextKey)
	if value == "" {
		value = "UTC"
	}
	return value
}

// UserLanguage get the locale used by the current logged user.
func UserLanguage(r *http.Request) string {
	language := getContextStringValue(r, UserLanguageContextKey)
	if language == "" {
		language = "en_US"
	}
	return language
}

// UserTheme get the theme used by the current logged user.
func UserTheme(r *http.Request) string {
	theme := getContextStringValue(r, UserThemeContextKey)
	if theme == "" {
		theme = "system_serif"
	}
	return theme
}

// CSRF returns the current CSRF token.
func CSRF(r *http.Request) string {
	return getContextStringValue(r, CSRFContextKey)
}

// SessionID returns the current session ID.
func SessionID(r *http.Request) string {
	return getContextStringValue(r, SessionIDContextKey)
}

// UserSessionToken returns the current user session token.
func UserSessionToken(r *http.Request) string {
	return getContextStringValue(r, UserSessionTokenContextKey)
}

// OAuth2State returns the current OAuth2 state.
func OAuth2State(r *http.Request) string {
	return getContextStringValue(r, OAuth2StateContextKey)
}

func OAuth2CodeVerifier(r *http.Request) string {
	return getContextStringValue(r, OAuth2CodeVerifierContextKey)
}

// FlashMessage returns the message message if any.
func FlashMessage(r *http.Request) string {
	return getContextStringValue(r, FlashMessageContextKey)
}

// FlashErrorMessage returns the message error message if any.
func FlashErrorMessage(r *http.Request) string {
	return getContextStringValue(r, FlashErrorMessageContextKey)
}

// PocketRequestToken returns the Pocket Request Token if any.
func PocketRequestToken(r *http.Request) string {
	return getContextStringValue(r, PocketRequestTokenContextKey)
}

// LastForceRefresh returns the last force refresh timestamp.
func LastForceRefresh(r *http.Request) int64 {
	jsonStringValue := getContextStringValue(r, LastForceRefreshContextKey)
	timestamp, err := strconv.ParseInt(jsonStringValue, 10, 64)
	if err != nil {
		return 0
	}
	return timestamp
}

// ClientIP returns the client IP address stored in the context.
func ClientIP(r *http.Request) string {
	return getContextStringValue(r, ClientIPContextKey)
}

func getContextStringValue(r *http.Request, key ContextKey) string {
	if v := r.Context().Value(key); v != nil {
		value, valid := v.(string)
		if !valid {
			return ""
		}

		return value
	}

	return ""
}

func getContextBoolValue(r *http.Request, key ContextKey) bool {
	if v := r.Context().Value(key); v != nil {
		value, valid := v.(bool)
		if !valid {
			return false
		}

		return value
	}

	return false
}

func getContextInt64Value(r *http.Request, key ContextKey) int64 {
	if v := r.Context().Value(key); v != nil {
		value, valid := v.(int64)
		if !valid {
			return 0
		}

		return value
	}

	return 0
}
