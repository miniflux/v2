// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package json // import "miniflux.app/v2/internal/http/response/json"

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
)

const contentTypeHeader = `application/json`

// OK creates a new JSON response with a 200 status code.
func OK(w http.ResponseWriter, r *http.Request, body interface{}) {
	builder := response.New(w, r)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSON(body))
	builder.Write()
}

// Created sends a created response to the client.
func Created(w http.ResponseWriter, r *http.Request, body interface{}) {
	builder := response.New(w, r)
	builder.WithStatus(http.StatusCreated)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSON(body))
	builder.Write()
}

// NoContent sends a no content response to the client.
func NoContent(w http.ResponseWriter, r *http.Request) {
	builder := response.New(w, r)
	builder.WithStatus(http.StatusNoContent)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.Write()
}

func Accepted(w http.ResponseWriter, r *http.Request) {
	builder := response.New(w, r)
	builder.WithStatus(http.StatusAccepted)
	builder.WithHeader("Content-Type", contentTypeHeader)
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
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSONError(err))
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
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSONError(err))
	builder.Write()
}

// Unauthorized sends a not authorized error to the client.
func Unauthorized(w http.ResponseWriter, r *http.Request) {
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

	builder := response.New(w, r)
	builder.WithStatus(http.StatusUnauthorized)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSONError(errors.New("access unauthorized")))
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
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSONError(errors.New("access forbidden")))
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
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSONError(errors.New("resource not found")))
	builder.Write()
}

func toJSONError(err error) []byte {
	type errorMsg struct {
		ErrorMessage string `json:"error_message"`
	}

	return toJSON(errorMsg{ErrorMessage: err.Error()})
}

func toJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		slog.Error("Unable to marshal JSON response", slog.Any("error", err))
		return []byte("")
	}

	return b
}
