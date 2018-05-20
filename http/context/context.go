// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package context

import (
	"net/http"

	"github.com/miniflux/miniflux/middleware"
)

// Context contains helper functions related to the current request.
type Context struct {
	request *http.Request
}

// IsAdminUser checks if the logged user is administrator.
func (c *Context) IsAdminUser() bool {
	return c.getContextBoolValue(middleware.IsAdminUserContextKey)
}

// IsAuthenticated returns a boolean if the user is authenticated.
func (c *Context) IsAuthenticated() bool {
	return c.getContextBoolValue(middleware.IsAuthenticatedContextKey)
}

// UserID returns the UserID of the logged user.
func (c *Context) UserID() int64 {
	return c.getContextIntValue(middleware.UserIDContextKey)
}

// UserTimezone returns the timezone used by the logged user.
func (c *Context) UserTimezone() string {
	value := c.getContextStringValue(middleware.UserTimezoneContextKey)
	if value == "" {
		value = "UTC"
	}
	return value
}

// UserLanguage get the locale used by the current logged user.
func (c *Context) UserLanguage() string {
	language := c.getContextStringValue(middleware.UserLanguageContextKey)
	if language == "" {
		language = "en_US"
	}
	return language
}

// CSRF returns the current CSRF token.
func (c *Context) CSRF() string {
	return c.getContextStringValue(middleware.CSRFContextKey)
}

// SessionID returns the current session ID.
func (c *Context) SessionID() string {
	return c.getContextStringValue(middleware.SessionIDContextKey)
}

// UserSessionToken returns the current user session token.
func (c *Context) UserSessionToken() string {
	return c.getContextStringValue(middleware.UserSessionTokenContextKey)
}

// OAuth2State returns the current OAuth2 state.
func (c *Context) OAuth2State() string {
	return c.getContextStringValue(middleware.OAuth2StateContextKey)
}

// FlashMessage returns the message message if any.
func (c *Context) FlashMessage() string {
	return c.getContextStringValue(middleware.FlashMessageContextKey)
}

// FlashErrorMessage returns the message error message if any.
func (c *Context) FlashErrorMessage() string {
	return c.getContextStringValue(middleware.FlashErrorMessageContextKey)
}

// PocketRequestToken returns the Pocket Request Token if any.
func (c *Context) PocketRequestToken() string {
	return c.getContextStringValue(middleware.PocketRequestTokenContextKey)
}

func (c *Context) getContextStringValue(key *middleware.ContextKey) string {
	if v := c.request.Context().Value(key); v != nil {
		return v.(string)
	}

	return ""
}

func (c *Context) getContextBoolValue(key *middleware.ContextKey) bool {
	if v := c.request.Context().Value(key); v != nil {
		return v.(bool)
	}

	return false
}

func (c *Context) getContextIntValue(key *middleware.ContextKey) int64 {
	if v := c.request.Context().Value(key); v != nil {
		return v.(int64)
	}

	return 0
}

// New creates a new Context.
func New(r *http.Request) *Context {
	return &Context{r}
}
