// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
)

type webSessionMiddleware struct {
	basePath string
	store    *storage.Storage
}

func newWebSessionMiddleware(basePath string, store *storage.Storage) *webSessionMiddleware {
	return &webSessionMiddleware{basePath: basePath, store: store}
}

func (m *webSessionMiddleware) handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isStaticAssetRoute(r) {
			next.ServeHTTP(w, r)
			return
		}

		session, err := m.loadWebSessionFromCookie(r)
		if err != nil {
			response.HTMLServerError(w, r, err)
			return
		}
		if session == nil {
			var secret string
			session, secret = model.NewWebSession(r.UserAgent(), request.ClientIP(r))
			if err := m.store.CreateWebSession(session); err != nil {
				response.HTMLServerError(w, r, err)
				return
			}
			setSessionCookie(w, session, secret)
		}

		ctx := context.WithValue(r.Context(), request.WebSessionContextKey, session)
		r = r.WithContext(ctx)

		if !request.IsAuthenticated(r) && !isPublicRoute(r) {
			response.HTMLRedirect(w, r, loginRedirectURL(m.basePath, r.RequestURI))
			return
		}

		next.ServeHTTP(w, r)

		if session.IsDirty() {
			if err := m.store.UpdateWebSession(session); err != nil {
				slog.Error("Unable to persist web session changes",
					slog.String("session_id", session.ID),
					slog.Any("error", err),
				)
			}
		}
	})
}

func (m *webSessionMiddleware) loadWebSessionFromCookie(r *http.Request) (*model.WebSession, error) {
	cookieValue := request.CookieValue(r, sessionCookieName)
	if cookieValue == "" {
		return nil, nil
	}

	sessionID, secret, ok := strings.Cut(cookieValue, ".")
	if !ok || sessionID == "" || secret == "" {
		return nil, nil
	}

	session, err := m.store.WebSessionByID(sessionID)
	if err != nil {
		return nil, err
	}

	if session == nil || !session.VerifySecret(secret) {
		return nil, nil
	}

	return session, nil
}
