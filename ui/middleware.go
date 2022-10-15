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
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/storage"
	"miniflux.app/ui/session"

	"github.com/gorilla/mux"
)

type middleware struct {
	router *mux.Router
	store  *storage.Storage
}

func newMiddleware(router *mux.Router, store *storage.Storage) *middleware {
	return &middleware{router, store}
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
			if request.IsAuthenticated(r) {
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

			http.SetCookie(w, cookie.New(cookie.CookieAppSessionID, session.ID, config.Opts.HTTPS, config.Opts.BasePath()))
		} else {
			logger.Debug("[UI:AppSession] %s", session)
		}

		if r.Method == http.MethodPost {
			formValue := r.FormValue("csrf")
			headerValue := r.Header.Get("X-Csrf-Token")

			if session.Data.CSRF != formValue && session.Data.CSRF != headerValue {
				logger.Error(`[UI:AppSession] Invalid or missing CSRF token: Form="%s", Header="%s"`, formValue, headerValue)

				if mux.CurrentRoute(r).GetName() == "checkLogin" {
					html.Redirect(w, r, route.Path(m.router, "login"))
					return
				}

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
		"sharedEntry",
		"healthcheck",
		"offline",
		"proxy":
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

func (m *middleware) handleAuthProxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if request.IsAuthenticated(r) || config.Opts.AuthProxyHeader() == "" {
			next.ServeHTTP(w, r)
			return
		}

		username := r.Header.Get(config.Opts.AuthProxyHeader())
		if username == "" {
			next.ServeHTTP(w, r)
			return
		}

		clientIP := request.ClientIP(r)
		logger.Info("[AuthProxy] [ClientIP=%s] Received authenticated requested for %q", clientIP, username)

		user, err := m.store.UserByUsername(username)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}

		if user == nil {
			logger.Error("[AuthProxy] [ClientIP=%s] %q doesn't exist", clientIP, username)

			if !config.Opts.IsAuthProxyUserCreationAllowed() {
				html.Forbidden(w, r)
				return
			}

			if user, err = m.store.CreateUser(&model.UserCreationRequest{Username: username}); err != nil {
				html.ServerError(w, r, err)
				return
			}
		}

		sessionToken, _, err := m.store.CreateUserSessionFromUsername(user.Username, r.UserAgent(), clientIP)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}

		logger.Info("[AuthProxy] [ClientIP=%s] username=%s just logged in", clientIP, user.Username)

		m.store.SetLastLogin(user.ID)

		sess := session.New(m.store, request.SessionID(r))
		sess.SetLanguage(user.Language)
		sess.SetTheme(user.Theme)

		http.SetCookie(w, cookie.New(
			cookie.CookieUserSessionID,
			sessionToken,
			config.Opts.HTTPS,
			config.Opts.BasePath(),
		))

		html.Redirect(w, r, route.Path(m.router, "unread"))
	})
}
