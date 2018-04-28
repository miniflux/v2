// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package middleware

// ContextKey represents a context key.
type ContextKey struct {
	name string
}

func (c ContextKey) String() string {
	return c.name
}

var (
	// UserIDContextKey is the context key used to store the user ID.
	UserIDContextKey = &ContextKey{"UserID"}

	// UserTimezoneContextKey is the context key used to store the user timezone.
	UserTimezoneContextKey = &ContextKey{"UserTimezone"}

	// IsAdminUserContextKey is the context key used to store the user role.
	IsAdminUserContextKey = &ContextKey{"IsAdminUser"}

	// IsAuthenticatedContextKey is the context key used to store the authentication flag.
	IsAuthenticatedContextKey = &ContextKey{"IsAuthenticated"}

	// UserSessionTokenContextKey is the context key used to store the user session ID.
	UserSessionTokenContextKey = &ContextKey{"UserSessionToken"}

	// UserLanguageContextKey is the context key to store user language.
	UserLanguageContextKey = &ContextKey{"UserLanguageContextKey"}

	// SessionIDContextKey is the context key used to store the session ID.
	SessionIDContextKey = &ContextKey{"SessionID"}

	// CSRFContextKey is the context key used to store CSRF token.
	CSRFContextKey = &ContextKey{"CSRF"}

	// OAuth2StateContextKey is the context key used to store OAuth2 state.
	OAuth2StateContextKey = &ContextKey{"OAuth2State"}

	// FlashMessageContextKey is the context key used to store a flash message.
	FlashMessageContextKey = &ContextKey{"FlashMessage"}

	// FlashErrorMessageContextKey is the context key used to store a flash error message.
	FlashErrorMessageContextKey = &ContextKey{"FlashErrorMessage"}
)
