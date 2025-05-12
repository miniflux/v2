// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package html // import "miniflux.app/v2/internal/http/response/html"

import (
	"html"
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
)

// OK creates a new HTML response with a 200 status code.
func OK(w http.ResponseWriter, r *http.Request, body interface{}) {
	builder := response.New(w, r)
	builder.WithHeader("Content-Type", "text/html; charset=utf-8")
	builder.WithHeader("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")
	builder.WithBody(body)
	builder.Write()
}

// ServerError sends an internal error to the client.
func ServerError(w http.ResponseWriter, r *http.Request, err error) {
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

	builder := response.New(w, r)
	builder.WithStatus(http.StatusInternalServerError)
	builder.WithHeader("Content-Security-Policy", response.ContentSecurityPolicyForUntrustedContent)
	builder.WithHeader("Content-Type", "text/plain; charset=utf-8")
	builder.WithHeader("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")
	builder.WithBody(html.EscapeString(err.Error()))
	builder.Write()
}

// BadRequest sends a bad request error to the client.
func BadRequest(w http.ResponseWriter, r *http.Request, err error) {
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

	builder := response.New(w, r)
	builder.WithStatus(http.StatusBadRequest)
	builder.WithHeader("Content-Security-Policy", response.ContentSecurityPolicyForUntrustedContent)
	builder.WithHeader("Content-Type", "text/plain; charset=utf-8")
	builder.WithHeader("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")
	builder.WithBody(html.EscapeString(err.Error()))
	builder.Write()
}

// Forbidden sends a forbidden error to the client.
func Forbidden(w http.ResponseWriter, r *http.Request) {
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

	builder := response.New(w, r)
	builder.WithStatus(http.StatusForbidden)
	builder.WithHeader("Content-Type", "text/html; charset=utf-8")
	builder.WithHeader("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")
	builder.WithBody("Access Forbidden")
	builder.Write()
}

// NotFound sends a page not found error to the client.
func NotFound(w http.ResponseWriter, r *http.Request) {
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

	builder := response.New(w, r)
	builder.WithStatus(http.StatusNotFound)
	builder.WithHeader("Content-Type", "text/html; charset=utf-8")
	builder.WithHeader("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")
	builder.WithBody("Page Not Found")
	builder.Write()
}

// Redirect redirects the user to another location.
func Redirect(w http.ResponseWriter, r *http.Request, uri string) {
	http.Redirect(w, r, uri, http.StatusFound)
}

// RequestedRangeNotSatisfiable sends a range not satisfiable error to the client.
func RequestedRangeNotSatisfiable(w http.ResponseWriter, r *http.Request, contentRange string) {
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

	builder := response.New(w, r)
	builder.WithStatus(http.StatusRequestedRangeNotSatisfiable)
	builder.WithHeader("Content-Type", "text/html; charset=utf-8")
	builder.WithHeader("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")
	builder.WithHeader("Content-Range", contentRange)
	builder.WithBody("Range Not Satisfiable")
	builder.Write()
}
