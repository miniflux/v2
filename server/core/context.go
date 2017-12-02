// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package core

import (
	"log"
	"net/http"

	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/server/middleware"
	"github.com/miniflux/miniflux2/server/route"
	"github.com/miniflux/miniflux2/storage"

	"github.com/gorilla/mux"
)

// Context contains helper functions related to the current request.
type Context struct {
	writer  http.ResponseWriter
	request *http.Request
	store   *storage.Storage
	router  *mux.Router
	user    *model.User
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
	if v := c.request.Context().Value(middleware.UserTimezoneContextKey); v != nil {
		return v.(string)
	}
	return "UTC"
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
			log.Fatalln(err)
		}

		if c.user == nil {
			log.Fatalln("Unable to find user from context")
		}
	}

	return c.user
}

// UserLanguage get the locale used by the current logged user.
func (c *Context) UserLanguage() string {
	user := c.LoggedUser()
	return user.Language
}

// CsrfToken returns the current CSRF token.
func (c *Context) CsrfToken() string {
	if v := c.request.Context().Value(middleware.TokenContextKey); v != nil {
		return v.(string)
	}

	log.Println("No CSRF token in context!")
	return ""
}

// Route returns the path for the given arguments.
func (c *Context) Route(name string, args ...interface{}) string {
	return route.Path(c.router, name, args...)
}

// NewContext creates a new Context.
func NewContext(w http.ResponseWriter, r *http.Request, store *storage.Storage, router *mux.Router) *Context {
	return &Context{writer: w, request: r, store: store, router: router}
}
