// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package json // import "miniflux.app/http/response/json"

import (
	"encoding/json"
	"errors"
	"net/http"

	"miniflux.app/http/response"
	"miniflux.app/logger"
)

// OK sends a JSON response with the status code 200.
func OK(w http.ResponseWriter, r *http.Request, v interface{}) {
	commonHeaders(w)
	response.Compress(w, r, toJSON(v))
}

// Created sends a JSON response with the status code 201.
func Created(w http.ResponseWriter, v interface{}) {
	commonHeaders(w)
	w.WriteHeader(http.StatusCreated)
	w.Write(toJSON(v))
}

// NoContent sends a JSON response with the status code 204.
func NoContent(w http.ResponseWriter) {
	commonHeaders(w)
	w.WriteHeader(http.StatusNoContent)
}

// NotFound sends a JSON response with the status code 404.
func NotFound(w http.ResponseWriter, err error) {
	logger.Error("[Not Found] %v", err)
	commonHeaders(w)
	w.WriteHeader(http.StatusNotFound)
	w.Write(encodeError(err))
}

// ServerError sends a JSON response with the status code 500.
func ServerError(w http.ResponseWriter, err error) {
	logger.Error("[Internal Server Error] %v", err)
	commonHeaders(w)
	w.WriteHeader(http.StatusInternalServerError)

	if err != nil {
		w.Write(encodeError(err))
	}
}

// Forbidden sends a JSON response with the status code 403.
func Forbidden(w http.ResponseWriter) {
	logger.Info("[Forbidden]")
	commonHeaders(w)
	w.WriteHeader(http.StatusForbidden)
	w.Write(encodeError(errors.New("Access Forbidden")))
}

// Unauthorized sends a JSON response with the status code 401.
func Unauthorized(w http.ResponseWriter) {
	commonHeaders(w)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write(encodeError(errors.New("Access Unauthorized")))
}

// BadRequest sends a JSON response with the status code 400.
func BadRequest(w http.ResponseWriter, err error) {
	logger.Error("[Bad Request] %v", err)
	commonHeaders(w)
	w.WriteHeader(http.StatusBadRequest)

	if err != nil {
		w.Write(encodeError(err))
	}
}

func commonHeaders(w http.ResponseWriter) {
	w.Header().Set("Accept", "application/json")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func encodeError(err error) []byte {
	type errorMsg struct {
		ErrorMessage string `json:"error_message"`
	}

	tmp := errorMsg{ErrorMessage: err.Error()}
	data, err := json.Marshal(tmp)
	if err != nil {
		logger.Error("json encoding error: %v", err)
	}

	return data
}

func toJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		logger.Error("json encoding error: %v", err)
		return []byte("")
	}

	return b
}
