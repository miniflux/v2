// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package response // import "miniflux.app/v2/internal/http/response"

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/http/request"
)

const jsonContentTypeHeader = `application/json`

// JSON creates a new JSON response with a 200 status code.
func JSON(w http.ResponseWriter, r *http.Request, body any) {
	responseBody, err := json.Marshal(body)
	if err != nil {
		JSONServerError(w, r, err)
		return
	}

	builder := NewBuilder(w, r)
	builder.WithHeader("Content-Type", jsonContentTypeHeader)
	builder.WithBodyAsBytes(responseBody)
	builder.Write()
}

// JSONCreated sends a created response to the client.
func JSONCreated(w http.ResponseWriter, r *http.Request, body any) {
	responseBody, err := json.Marshal(body)
	if err != nil {
		JSONServerError(w, r, err)
		return
	}

	builder := NewBuilder(w, r)
	builder.WithStatus(http.StatusCreated)
	builder.WithHeader("Content-Type", jsonContentTypeHeader)
	builder.WithBodyAsBytes(responseBody)
	builder.Write()
}

// JSONAccepted sends an accepted response to the client.
func JSONAccepted(w http.ResponseWriter, r *http.Request) {
	builder := NewBuilder(w, r)
	builder.WithStatus(http.StatusAccepted)
	builder.WithHeader("Content-Type", jsonContentTypeHeader)
	builder.Write()
}

// JSONServerError sends an internal error to the client.
func JSONServerError(w http.ResponseWriter, r *http.Request, err error) {
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
	builder.WithHeader("Content-Type", jsonContentTypeHeader)
	builder.WithBodyAsBytes(generateJSONError(err))
	builder.Write()
}

// JSONBadRequest sends a bad request error to the client.
func JSONBadRequest(w http.ResponseWriter, r *http.Request, err error) {
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
	builder.WithHeader("Content-Type", jsonContentTypeHeader)
	builder.WithBodyAsBytes(generateJSONError(err))
	builder.Write()
}

// JSONUnauthorized sends a not authorized error to the client.
func JSONUnauthorized(w http.ResponseWriter, r *http.Request) {
	slog.Warn(http.StatusText(http.StatusUnauthorized),
		slog.String("client_ip", request.ClientIP(r)),
		slog.Group("request",
			slog.String("method", r.Method),
			slog.String("uri", r.RequestURI),
			slog.String("user_agent", r.UserAgent()),
		),
		slog.Group("response",
			slog.Int("status_code", http.StatusUnauthorized),
		),
	)

	builder := NewBuilder(w, r)
	builder.WithStatus(http.StatusUnauthorized)
	builder.WithHeader("Content-Type", jsonContentTypeHeader)
	builder.WithBodyAsBytes(generateJSONError(errors.New("access unauthorized")))
	builder.Write()
}

// JSONForbidden sends a forbidden error to the client.
func JSONForbidden(w http.ResponseWriter, r *http.Request) {
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
	builder.WithHeader("Content-Type", jsonContentTypeHeader)
	builder.WithBodyAsBytes(generateJSONError(errors.New("access forbidden")))
	builder.Write()
}

// JSONNotFound sends a page not found error to the client.
func JSONNotFound(w http.ResponseWriter, r *http.Request) {
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
	builder.WithHeader("Content-Type", jsonContentTypeHeader)
	builder.WithBodyAsBytes(generateJSONError(errors.New("resource not found")))
	builder.Write()
}

func generateJSONError(err error) []byte {
	type errorMsg struct {
		ErrorMessage string `json:"error_message"`
	}

	encodedBody, _ := json.Marshal(errorMsg{ErrorMessage: err.Error()})
	return encodedBody
}
