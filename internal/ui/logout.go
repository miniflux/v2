// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
)

func (h *handler) logout(w http.ResponseWriter, r *http.Request) {
	if s := request.WebSession(r); s != nil {
		if err := h.store.RemoveUserWebSession(request.UserID(r), s.ID); err != nil {
			response.HTMLServerError(w, r, err)
			return
		}
	}

	clearSessionCookie(w)
	response.HTMLRedirect(w, r, h.routePath("/"))
}

// clearSessionCookie expires the session cookie on the client.
func clearSessionCookie(w http.ResponseWriter) {
	path := config.Opts.BasePath()
	if path == "" {
		path = "/"
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     path,
		Secure:   config.Opts.HTTPS(),
		HttpOnly: true,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})
}
