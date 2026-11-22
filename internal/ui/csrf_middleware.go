// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"errors"
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
)

type csrfMiddleware struct {
	basePath string
}

func newCSRFMiddleware(basePath string) *csrfMiddleware {
	return &csrfMiddleware{basePath: basePath}
}

// handle validates the CSRF token on state-changing requests. It must be
// chained inside handleWebSession so that the session is already present
// in the request context.
func (m *csrfMiddleware) handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The request passed down is a copy made by WithContext, so net/http
		// won't remove the multipart temporary files on its own.
		defer func() {
			if r.MultipartForm != nil {
				r.MultipartForm.RemoveAll()
			}
		}()

		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			// Safe methods don't require CSRF validation.
		default:
			if !m.validate(w, r) {
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (m *csrfMiddleware) validate(w http.ResponseWriter, r *http.Request) bool {
	session := request.WebSession(r)
	if session == nil {
		response.HTMLBadRequest(w, r, errors.New("invalid or missing CSRF"))
		return false
	}

	csrfToken := session.CSRF()
	if crypto.ConstantTimeCmp(csrfToken, r.Header.Get("X-Csrf-Token")) || crypto.ConstantTimeCmp(csrfToken, csrfFormValue(r)) {
		return true
	}

	slog.Warn("Invalid or missing CSRF token",
		slog.String("url", r.RequestURI),
	)

	if r.URL.Path == "/login" {
		response.HTMLRedirect(w, r, m.basePath+"/")
	} else {
		response.HTMLBadRequest(w, r, errors.New("invalid or missing CSRF"))
	}

	return false
}

// csrfFormValue returns the CSRF token from the request body, parsing multipart forms only for authenticated users.
func csrfFormValue(r *http.Request) string {
	if request.IsAuthenticated(r) {
		return r.PostFormValue("csrf")
	}

	if err := r.ParseForm(); err != nil {
		return ""
	}
	return r.PostForm.Get("csrf")
}
