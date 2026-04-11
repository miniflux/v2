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
		if r.Method == http.MethodPost && !m.validate(w, r) {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *csrfMiddleware) validate(w http.ResponseWriter, r *http.Request) bool {
	csrfToken := request.WebSession(r).CSRF()
	formValue := r.FormValue("csrf")
	headerValue := r.Header.Get("X-Csrf-Token")

	if crypto.ConstantTimeCmp(csrfToken, formValue) || crypto.ConstantTimeCmp(csrfToken, headerValue) {
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
