// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package middleware // import "miniflux.app/middleware"

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
	FlashMessageContextKey
	FlashErrorMessageContextKey
	PocketRequestTokenContextKey
)
