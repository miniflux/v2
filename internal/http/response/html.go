// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package response // import "miniflux.app/v2/internal/http/response"

import (
	"fmt"
	"html"
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/urllib"
)

// HTML creates a new HTML response with a 200 status code.
func HTML[T []byte | string](w http.ResponseWriter, r *http.Request, body T) {
	builder := NewBuilder(w, r)
	builder.WithHeader("Content-Type", "text/html; charset=utf-8")
	builder.WithHeader("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")
	switch v := any(body).(type) {
	case []byte:
		builder.WithBodyAsBytes(v)
	case string:
		builder.WithBodyAsString(v)
	}
	builder.Write()
}

// HTMLServerError sends an internal error to the client.
func HTMLServerError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error(http.StatusText(http.StatusInternalServerError),
		slog.Any("error", err),
		slog.String("client_ip", request.ClientIP(r)),
		slog.Group("request",
			slog.String("method", r.Method),
			slog.String("uri", r.RequestURI),
			slog.String("user_agent", r.UserAgent()),
		),
		slog.Group("response",
			slog.Int("status_code", http.StatusInternalServerError),
		),
	)

	builder := NewBuilder(w, r)
	builder.WithStatus(http.StatusInternalServerError)
	builder.WithHeader("Content-Security-Policy", ContentSecurityPolicyForUntrustedContent)
	builder.WithHeader("Content-Type", "text/plain; charset=utf-8")
	builder.WithHeader("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")
	builder.WithBodyAsString(html.EscapeString(err.Error()))
	builder.Write()
}

// HTMLBadRequest sends a bad request error to the client.
func HTMLBadRequest(w http.ResponseWriter, r *http.Request, err error) {
	slog.Warn(http.StatusText(http.StatusBadRequest),
		slog.Any("error", err),
		slog.String("client_ip", request.ClientIP(r)),
		slog.Group("request",
			slog.String("method", r.Method),
			slog.String("uri", r.RequestURI),
			slog.String("user_agent", r.UserAgent()),
		),
		slog.Group("response",
			slog.Int("status_code", http.StatusBadRequest),
		),
	)

	builder := NewBuilder(w, r)
	builder.WithStatus(http.StatusBadRequest)
	builder.WithHeader("Content-Security-Policy", ContentSecurityPolicyForUntrustedContent)
	builder.WithHeader("Content-Type", "text/plain; charset=utf-8")
	builder.WithHeader("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")
	builder.WithBodyAsString(html.EscapeString(err.Error()))
	builder.Write()
}

// HTMLForbidden sends a forbidden error to the client.
func HTMLForbidden(w http.ResponseWriter, r *http.Request) {
	slog.Warn(http.StatusText(http.StatusForbidden),
		slog.String("client_ip", request.ClientIP(r)),
		slog.Group("request",
			slog.String("method", r.Method),
			slog.String("uri", r.RequestURI),
			slog.String("user_agent", r.UserAgent()),
		),
		slog.Group("response",
			slog.Int("status_code", http.StatusForbidden),
		),
	)

	builder := NewBuilder(w, r)
	builder.WithStatus(http.StatusForbidden)
	builder.WithHeader("Content-Type", "text/html; charset=utf-8")
	builder.WithHeader("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")
	builder.WithBodyAsString("Access Forbidden")
	builder.Write()
}

// HTMLNotFound sends a page not found error to the client.
func HTMLNotFound(w http.ResponseWriter, r *http.Request) {
	slog.Warn(http.StatusText(http.StatusNotFound),
		slog.String("client_ip", request.ClientIP(r)),
		slog.Group("request",
			slog.String("method", r.Method),
			slog.String("uri", r.RequestURI),
			slog.String("user_agent", r.UserAgent()),
		),
		slog.Group("response",
			slog.Int("status_code", http.StatusNotFound),
		),
	)

	builder := NewBuilder(w, r)
	builder.WithStatus(http.StatusNotFound)
	builder.WithHeader("Content-Type", "text/html; charset=utf-8")
	builder.WithHeader("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")
	builder.WithBodyAsString("Page Not Found")
	builder.Write()
}

// HTMLRedirect redirects the user to a relative path or an absolute http(s) URL.
func HTMLRedirect(w http.ResponseWriter, r *http.Request, uri string) {
	if !urllib.IsRelativePath(uri) && !urllib.IsAbsoluteURL(uri) {
		HTMLBadRequest(w, r, fmt.Errorf("invalid redirect URL: %q", uri))
		return
	}
	http.Redirect(w, r, uri, http.StatusFound)
}

// HTMLRequestedRangeNotSatisfiable sends a range not satisfiable error to the client.
func HTMLRequestedRangeNotSatisfiable(w http.ResponseWriter, r *http.Request, contentRange string) {
	slog.Warn(http.StatusText(http.StatusRequestedRangeNotSatisfiable),
		slog.String("client_ip", request.ClientIP(r)),
		slog.Group("request",
			slog.String("method", r.Method),
			slog.String("uri", r.RequestURI),
			slog.String("user_agent", r.UserAgent()),
		),
		slog.Group("response",
			slog.Int("status_code", http.StatusRequestedRangeNotSatisfiable),
		),
	)

	builder := NewBuilder(w, r)
	builder.WithStatus(http.StatusRequestedRangeNotSatisfiable)
	builder.WithHeader("Content-Type", "text/html; charset=utf-8")
	builder.WithHeader("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")
	builder.WithHeader("Content-Range", contentRange)
	builder.WithBodyAsString("Range Not Satisfiable")
	builder.Write()
}
