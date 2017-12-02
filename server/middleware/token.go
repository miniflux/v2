// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/storage"
)

// TokenMiddleware represents a token middleware.
type TokenMiddleware struct {
	store *storage.Storage
}

// Handler execute the middleware.
func (t *TokenMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		token := t.getTokenValueFromCookie(r)

		if token == nil {
			log.Println("[Middleware:Token] Token not found")
			token, err = t.store.CreateToken()
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			cookie := &http.Cookie{
				Name:     "tokenID",
				Value:    token.ID,
				Path:     "/",
				Secure:   r.URL.Scheme == "https",
				HttpOnly: true,
			}

			http.SetCookie(w, cookie)
		} else {
			log.Println("[Middleware:Token]", token)
		}

		isTokenValid := token.Value == r.FormValue("csrf") || token.Value == r.Header.Get("X-Csrf-Token")

		if r.Method == "POST" && !isTokenValid {
			log.Println("[Middleware:CSRF] Invalid or missing CSRF token!")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid or missing CSRF token!"))
		} else {
			ctx := r.Context()
			ctx = context.WithValue(ctx, TokenContextKey, token.Value)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func (t *TokenMiddleware) getTokenValueFromCookie(r *http.Request) *model.Token {
	tokenCookie, err := r.Cookie("tokenID")
	if err == http.ErrNoCookie {
		return nil
	}

	token, err := t.store.Token(tokenCookie.Value)
	if err != nil {
		log.Println(err)
		return nil
	}

	return token
}

// NewTokenMiddleware returns a new TokenMiddleware.
func NewTokenMiddleware(s *storage.Storage) *TokenMiddleware {
	return &TokenMiddleware{store: s}
}
