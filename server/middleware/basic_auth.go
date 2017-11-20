// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package middleware

import (
	"context"
	"github.com/miniflux/miniflux2/storage"
	"log"
	"net/http"
)

type BasicAuthMiddleware struct {
	store *storage.Storage
}

func (b *BasicAuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		errorResponse := `{"error_message": "Not Authorized"}`

		username, password, authOK := r.BasicAuth()
		if !authOK {
			log.Println("[Middleware:BasicAuth] No authentication headers sent")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(errorResponse))
			return
		}

		if err := b.store.CheckPassword(username, password); err != nil {
			log.Println("[Middleware:BasicAuth] Invalid username or password:", username)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(errorResponse))
			return
		}

		user, err := b.store.GetUserByUsername(username)
		if err != nil || user == nil {
			log.Println("[Middleware:BasicAuth] User not found:", username)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(errorResponse))
			return
		}

		log.Println("[Middleware:BasicAuth] User authenticated:", username)
		b.store.SetLastLogin(user.ID)

		ctx := r.Context()
		ctx = context.WithValue(ctx, "UserId", user.ID)
		ctx = context.WithValue(ctx, "UserTimezone", user.Timezone)
		ctx = context.WithValue(ctx, "IsAdminUser", user.IsAdmin)
		ctx = context.WithValue(ctx, "IsAuthenticated", true)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewBasicAuthMiddleware(s *storage.Storage) *BasicAuthMiddleware {
	return &BasicAuthMiddleware{store: s}
}
