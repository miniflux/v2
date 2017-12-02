// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package middleware

type contextKey struct {
	name string
}

var (
	// UserIDContextKey is the context key used to store the user ID.
	UserIDContextKey = &contextKey{"UserID"}

	// UserTimezoneContextKey is the context key used to store the user timezone.
	UserTimezoneContextKey = &contextKey{"UserTimezone"}

	// IsAdminUserContextKey is the context key used to store the user role.
	IsAdminUserContextKey = &contextKey{"IsAdminUser"}

	// IsAuthenticatedContextKey is the context key used to store the authentication flag.
	IsAuthenticatedContextKey = &contextKey{"IsAuthenticated"}

	// TokenContextKey is the context key used to store CSRF token.
	TokenContextKey = &contextKey{"CSRF"}
)
