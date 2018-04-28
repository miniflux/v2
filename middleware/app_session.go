// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package middleware

import (
	"context"
	"net/http"

	"github.com/miniflux/miniflux/http/cookie"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/model"
)

// AppSession handles application session middleware.
func (m *Middleware) AppSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		session := m.getSessionValueFromCookie(r)

		if session == nil {
			logger.Debug("[Middleware:Session] Session not found")
			session, err = m.store.CreateSession()
			if err != nil {
				logger.Error("[Middleware:Session] %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, cookie.New(cookie.CookieSessionID, session.ID, m.cfg.IsHTTPS, m.cfg.BasePath()))
		} else {
			logger.Debug("[Middleware:Session] %s", session)
		}

		if r.Method == "POST" {
			formValue := r.FormValue("csrf")
			headerValue := r.Header.Get("X-Csrf-Token")

			if session.Data.CSRF != formValue && session.Data.CSRF != headerValue {
				logger.Error(`[Middleware:Session] Invalid or missing CSRF token: Form="%s", Header="%s"`, formValue, headerValue)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Invalid or missing CSRF session!"))
				return
			}
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, SessionIDContextKey, session.ID)
		ctx = context.WithValue(ctx, CSRFContextKey, session.Data.CSRF)
		ctx = context.WithValue(ctx, OAuth2StateContextKey, session.Data.OAuth2State)
		ctx = context.WithValue(ctx, FlashMessageContextKey, session.Data.FlashMessage)
		ctx = context.WithValue(ctx, FlashErrorMessageContextKey, session.Data.FlashErrorMessage)
		ctx = context.WithValue(ctx, UserLanguageContextKey, session.Data.Language)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) getSessionValueFromCookie(r *http.Request) *model.Session {
	sessionCookie, err := r.Cookie(cookie.CookieSessionID)
	if err == http.ErrNoCookie {
		return nil
	}

	session, err := m.store.Session(sessionCookie.Value)
	if err != nil {
		logger.Error("[Middleware:Session] %v", err)
		return nil
	}

	return session
}
