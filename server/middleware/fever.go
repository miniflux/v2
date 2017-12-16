// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package middleware

import (
	"context"
	"net/http"

	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/storage"
)

// FeverMiddleware is the middleware that handles Fever API.
type FeverMiddleware struct {
	store *storage.Storage
}

// Handler executes the middleware.
func (f *FeverMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("[Middleware:Fever]")

		apiKey := r.FormValue("api_key")
		user, err := f.store.UserByFeverToken(apiKey)
		if err != nil {
			logger.Error("[Fever] %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"api_version": 3, "auth": 0}`))
			return
		}

		if user == nil {
			logger.Info("[Middleware:Fever] Fever authentication failure")
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"api_version": 3, "auth": 0}`))
			return
		}

		logger.Info("[Middleware:Fever] User #%d is authenticated", user.ID)
		f.store.SetLastLogin(user.ID)

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDContextKey, user.ID)
		ctx = context.WithValue(ctx, UserTimezoneContextKey, user.Timezone)
		ctx = context.WithValue(ctx, IsAdminUserContextKey, user.IsAdmin)
		ctx = context.WithValue(ctx, IsAuthenticatedContextKey, true)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// NewFeverMiddleware returns a new FeverMiddleware.
func NewFeverMiddleware(s *storage.Storage) *FeverMiddleware {
	return &FeverMiddleware{store: s}
}
