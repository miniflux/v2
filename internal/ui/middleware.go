// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/http/cookie"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/ui/session"

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
				slog.Debug("Redirecting to login page because no user session has been found",
					slog.Any("url", r.RequestURI),
				)
				html.Redirect(w, r, route.Path(m.router, "login"))
			}
		} else {
			slog.Debug("User session found",
				slog.Any("url", r.RequestURI),
				slog.Int64("user_id", session.UserID),
				slog.Int64("user_session_id", session.ID),
			)

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
				slog.Debug("Cookie expired but user is logged: creating a new app session",
					slog.Int64("user_id", userID),
				)
				session, err = m.store.CreateAppSessionWithUserPrefs(userID)
				if err != nil {
					html.ServerError(w, r, err)
					return
				}
			} else {
				slog.Debug("App session not found, creating a new one")
				session, err = m.store.CreateAppSession()
				if err != nil {
					html.ServerError(w, r, err)
					return
				}
			}

			http.SetCookie(w, cookie.New(cookie.CookieAppSessionID, session.ID, config.Opts.HTTPS, config.Opts.BasePath()))
		}

		if r.Method == http.MethodPost {
			formValue := r.FormValue("csrf")
			headerValue := r.Header.Get("X-Csrf-Token")

			if !crypto.ConstantTimeCmp(session.Data.CSRF, formValue) && !crypto.ConstantTimeCmp(session.Data.CSRF, headerValue) {
				slog.Warn("Invalid or missing CSRF token",
					slog.Any("url", r.RequestURI),
					slog.String("form_csrf", formValue),
					slog.String("header_csrf", headerValue),
				)

				if mux.CurrentRoute(r).GetName() == "checkLogin" {
					html.Redirect(w, r, route.Path(m.router, "login"))
					return
				}

				html.BadRequest(w, r, errors.New("invalid or missing CSRF"))
				return
			}
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, request.SessionIDContextKey, session.ID)
		ctx = context.WithValue(ctx, request.CSRFContextKey, session.Data.CSRF)
		ctx = context.WithValue(ctx, request.OAuth2StateContextKey, session.Data.OAuth2State)
		ctx = context.WithValue(ctx, request.OAuth2CodeVerifierContextKey, session.Data.OAuth2CodeVerifier)
		ctx = context.WithValue(ctx, request.FlashMessageContextKey, session.Data.FlashMessage)
		ctx = context.WithValue(ctx, request.FlashErrorMessageContextKey, session.Data.FlashErrorMessage)
		ctx = context.WithValue(ctx, request.UserLanguageContextKey, session.Data.Language)
		ctx = context.WithValue(ctx, request.UserThemeContextKey, session.Data.Theme)
		ctx = context.WithValue(ctx, request.PocketRequestTokenContextKey, session.Data.PocketRequestToken)
		ctx = context.WithValue(ctx, request.LastForceRefreshContextKey, session.Data.LastForceRefresh)
		ctx = context.WithValue(ctx, request.WebAuthnDataContextKey, session.Data.WebAuthnSessionData)
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
		slog.Debug("Unable to fetch app session from the database; another session will be created",
			slog.Any("cookie_value", cookieValue),
			slog.Any("error", err),
		)
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
		"proxy",
		"webauthnLoginBegin",
		"webauthnLoginFinish":
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
		slog.Error("Unable to fetch user session from the database",
			slog.Any("cookie_value", cookieValue),
			slog.Any("error", err),
		)
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
		slog.Debug("[AuthProxy] Received authenticated requested",
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
			slog.String("username", username),
		)

		user, err := m.store.UserByUsername(username)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}

		if user == nil {
			if !config.Opts.IsAuthProxyUserCreationAllowed() {
				slog.Debug("[AuthProxy] User doesn't exist and user creation is not allowed",
					slog.Bool("authentication_failed", true),
					slog.String("client_ip", clientIP),
					slog.String("user_agent", r.UserAgent()),
					slog.String("username", username),
				)
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

		slog.Info("[AuthProxy] User authenticated successfully",
			slog.Bool("authentication_successful", true),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
			slog.Int64("user_id", user.ID),
			slog.String("username", user.Username),
		)

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

		html.Redirect(w, r, route.Path(m.router, user.DefaultHomePage))
	})
}
