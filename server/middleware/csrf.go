// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/miniflux/miniflux2/helper"
)

// Csrf is a middleware that handle CSRF tokens.
func Csrf(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var csrfToken string

		csrfCookie, err := r.Cookie("csrfToken")
		if err == http.ErrNoCookie || csrfCookie.Value == "" {
			csrfToken = helper.GenerateRandomString(64)
			cookie := &http.Cookie{
				Name:     "csrfToken",
				Value:    csrfToken,
				Path:     "/",
				Secure:   r.URL.Scheme == "https",
				HttpOnly: true,
			}

			http.SetCookie(w, cookie)
		} else {
			csrfToken = csrfCookie.Value
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, CsrfContextKey, csrfToken)

		w.Header().Add("Vary", "Cookie")
		isTokenValid := csrfToken == r.FormValue("csrf") || csrfToken == r.Header.Get("X-Csrf-Token")

		if r.Method == "POST" && !isTokenValid {
			log.Println("[Middleware:CSRF] Invalid or missing CSRF token!")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid or missing CSRF token!"))
		} else {
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}
