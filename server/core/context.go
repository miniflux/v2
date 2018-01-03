// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package core

import (
	"net/http"

	"github.com/miniflux/miniflux/crypto"
	"github.com/miniflux/miniflux/locale"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/server/middleware"
	"github.com/miniflux/miniflux/server/route"
	"github.com/miniflux/miniflux/storage"

	"github.com/gorilla/mux"
)

// Context contains helper functions related to the current request.
type Context struct {
	writer     http.ResponseWriter
	request    *http.Request
	store      *storage.Storage
	router     *mux.Router
	user       *model.User
	translator *locale.Translator
}

// IsAdminUser checks if the logged user is administrator.
func (c *Context) IsAdminUser() bool {
	if v := c.request.Context().Value(middleware.IsAdminUserContextKey); v != nil {
		return v.(bool)
	}
	return false
}

// UserTimezone returns the timezone used by the logged user.
func (c *Context) UserTimezone() string {
	value := c.getContextStringValue(middleware.UserTimezoneContextKey)
	if value == "" {
		value = "UTC"
	}
	return value
}

// IsAuthenticated returns a boolean if the user is authenticated.
func (c *Context) IsAuthenticated() bool {
	if v := c.request.Context().Value(middleware.IsAuthenticatedContextKey); v != nil {
		return v.(bool)
	}
	return false
}

// UserID returns the UserID of the logged user.
func (c *Context) UserID() int64 {
	if v := c.request.Context().Value(middleware.UserIDContextKey); v != nil {
		return v.(int64)
	}
	return 0
}

// LoggedUser returns all properties related to the logged user.
func (c *Context) LoggedUser() *model.User {
	if c.user == nil {
		var err error
		c.user, err = c.store.UserByID(c.UserID())
		if err != nil {
			logger.Fatal("[Context] %v", err)
		}

		if c.user == nil {
			logger.Fatal("Unable to find user from context")
		}
	}

	return c.user
}

// UserLanguage get the locale used by the current logged user.
func (c *Context) UserLanguage() string {
	user := c.LoggedUser()
	return user.Language
}

// Translate translates a message in the current language.
func (c *Context) Translate(message string, args ...interface{}) string {
	return c.translator.GetLanguage(c.UserLanguage()).Get(message, args...)
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

// GenerateOAuth2State generate a new OAuth2 state.
func (c *Context) GenerateOAuth2State() string {
	state := crypto.GenerateRandomString(32)
	c.store.UpdateSessionField(c.SessionID(), "oauth2_state", state)
	return state
}

// SetFlashMessage defines a new flash message.
func (c *Context) SetFlashMessage(message string) {
	c.store.UpdateSessionField(c.SessionID(), "flash_message", message)
}

// FlashMessage returns the flash message and remove it.
func (c *Context) FlashMessage() string {
	message := c.getContextStringValue(middleware.FlashMessageContextKey)
	c.store.UpdateSessionField(c.SessionID(), "flash_message", "")
	return message
}

// SetFlashErrorMessage defines a new flash error message.
func (c *Context) SetFlashErrorMessage(message string) {
	c.store.UpdateSessionField(c.SessionID(), "flash_error_message", message)
}

// FlashErrorMessage returns the error flash message and remove it.
func (c *Context) FlashErrorMessage() string {
	message := c.getContextStringValue(middleware.FlashErrorMessageContextKey)
	c.store.UpdateSessionField(c.SessionID(), "flash_error_message", "")
	return message
}

func (c *Context) getContextStringValue(key *middleware.ContextKey) string {
	if v := c.request.Context().Value(key); v != nil {
		return v.(string)
	}

	logger.Error("[Core:Context] Missing key: %s", key)
	return ""
}

// Route returns the path for the given arguments.
func (c *Context) Route(name string, args ...interface{}) string {
	return route.Path(c.router, name, args...)
}

// NewContext creates a new Context.
func NewContext(w http.ResponseWriter, r *http.Request, store *storage.Storage, router *mux.Router, translator *locale.Translator) *Context {
	return &Context{writer: w, request: r, store: store, router: router, translator: translator}
}
