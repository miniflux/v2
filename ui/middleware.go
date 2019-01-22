// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"context"
	"errors"
	"net/http"

	"miniflux.app/config"
	"miniflux.app/http/cookie"
	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/storage"
	"miniflux.app/logger"
	"miniflux.app/model"

	"github.com/gorilla/mux"
)

type middleware struct {
	router *mux.Router
	cfg *config.Config
	store *storage.Storage
}

func newMiddleware(router *mux.Router, cfg *config.Config, store *storage.Storage) *middleware {
	return &middleware{router, cfg, store}
}

func (m *middleware) handleUserSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := m.getUserSessionFromCookie(r)

		if session == nil {
			if m.isPublicRoute(r) {
				next.ServeHTTP(w, r)
			} else {
				logger.Debug("[UI:UserSession] Session not found, redirect to login page")
				html.Redirect(w, r, route.Path(m.router, "login"))
			}
		} else {
			logger.Debug("[UI:UserSession] %s", session)

			ctx := r.Context()
			ctx = context.WithValue(ctx, request.UserIDContextKey, session.UserID)
			ctx = context.WithValue(ctx, request.IsAuthenticatedContextKey, true)
			ctx = context.WithValue(ctx, request.UserSessionTokenContextKey, session.Token)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func (m *middleware) handleAppSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		session := m.getAppSessionValueFromCookie(r)

		if session == nil {
			if (request.IsAuthenticated(r)) {
				userID := request.UserID(r)
				logger.Debug("[UI:AppSession] Cookie expired but user #%d is logged: creating a new session", userID)
				session, err = m.store.CreateAppSessionWithUserPrefs(userID)
				if err != nil {
					html.ServerError(w, r, err)
					return
				}
			} else {
				logger.Debug("[UI:AppSession] Session not found, creating a new one")
				session, err = m.store.CreateAppSession()
				if err != nil {
					html.ServerError(w, r, err)
					return
				}
			}

			http.SetCookie(w, cookie.New(cookie.CookieAppSessionID, session.ID, m.cfg.IsHTTPS, m.cfg.BasePath()))
		} else {
			logger.Debug("[UI:AppSession] %s", session)
		}

		if r.Method == "POST" {
			formValue := r.FormValue("csrf")
			headerValue := r.Header.Get("X-Csrf-Token")

			if session.Data.CSRF != formValue && session.Data.CSRF != headerValue {
				logger.Error(`[UI:AppSession] Invalid or missing CSRF token: Form="%s", Header="%s"`, formValue, headerValue)
				html.BadRequest(w, r, errors.New("Invalid or missing CSRF"))
				return
			}
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, request.SessionIDContextKey, session.ID)
		ctx = context.WithValue(ctx, request.CSRFContextKey, session.Data.CSRF)
		ctx = context.WithValue(ctx, request.OAuth2StateContextKey, session.Data.OAuth2State)
		ctx = context.WithValue(ctx, request.FlashMessageContextKey, session.Data.FlashMessage)
		ctx = context.WithValue(ctx, request.FlashErrorMessageContextKey, session.Data.FlashErrorMessage)
		ctx = context.WithValue(ctx, request.UserLanguageContextKey, session.Data.Language)
		ctx = context.WithValue(ctx, request.UserThemeContextKey, session.Data.Theme)
		ctx = context.WithValue(ctx, request.PocketRequestTokenContextKey, session.Data.PocketRequestToken)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *middleware) getAppSessionValueFromCookie(r *http.Request) *model.Session {
	cookieValue := request.CookieValue(r, cookie.CookieAppSessionID)
	if cookieValue == "" {
		return nil
	}

	session, err := m.store.AppSession(cookieValue)
	if err != nil {
		logger.Error("[UI:AppSession] %v", err)
		return nil
	}

	return session
}

func (m *middleware) isPublicRoute(r *http.Request) bool {
	route := mux.CurrentRoute(r)
	switch route.GetName() {
	case "login",
		"checkLogin",
		"stylesheet",
		"javascript",
		"oauth2Redirect",
		"oauth2Callback",
		"appIcon",
		"favicon",
		"webManifest",
		"robots",
		"healthcheck":
		return true
	default:
		return false
	}
}

func (m *middleware) getUserSessionFromCookie(r *http.Request) *model.UserSession {
	cookieValue := request.CookieValue(r, cookie.CookieUserSessionID)
	if cookieValue == "" {
		return nil
	}

	session, err := m.store.UserSessionByToken(cookieValue)
	if err != nil {
		logger.Error("[UI:UserSession] %v", err)
		return nil
	}

	return session
}
