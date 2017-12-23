// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package core

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/miniflux/miniflux/logger"
)

// JSONResponse handles JSON responses.
type JSONResponse struct {
	writer  http.ResponseWriter
	request *http.Request
}

// Standard sends a JSON response with the status code 200.
func (j *JSONResponse) Standard(v interface{}) {
	j.commonHeaders()
	j.writer.WriteHeader(http.StatusOK)
	j.writer.Write(j.toJSON(v))
}

// Created sends a JSON response with the status code 201.
func (j *JSONResponse) Created(v interface{}) {
	j.commonHeaders()
	j.writer.WriteHeader(http.StatusCreated)
	j.writer.Write(j.toJSON(v))
}

// NoContent sends a JSON response with the status code 204.
func (j *JSONResponse) NoContent() {
	j.commonHeaders()
	j.writer.WriteHeader(http.StatusNoContent)
}

// BadRequest sends a JSON response with the status code 400.
func (j *JSONResponse) BadRequest(err error) {
	logger.Error("[Bad Request] %v", err)
	j.commonHeaders()
	j.writer.WriteHeader(http.StatusBadRequest)

	if err != nil {
		j.writer.Write(j.encodeError(err))
	}
}

// NotFound sends a JSON response with the status code 404.
func (j *JSONResponse) NotFound(err error) {
	logger.Error("[Not Found] %v", err)
	j.commonHeaders()
	j.writer.WriteHeader(http.StatusNotFound)
	j.writer.Write(j.encodeError(err))
}

// ServerError sends a JSON response with the status code 500.
func (j *JSONResponse) ServerError(err error) {
	logger.Error("[Internal Server Error] %v", err)
	j.commonHeaders()
	j.writer.WriteHeader(http.StatusInternalServerError)

	if err != nil {
		j.writer.Write(j.encodeError(err))
	}
}

// Forbidden sends a JSON response with the status code 403.
func (j *JSONResponse) Forbidden() {
	logger.Info("[API:Forbidden]")
	j.commonHeaders()
	j.writer.WriteHeader(http.StatusForbidden)
	j.writer.Write(j.encodeError(errors.New("Access Forbidden")))
}

func (j *JSONResponse) commonHeaders() {
	j.writer.Header().Set("Accept", "application/json")
	j.writer.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func (j *JSONResponse) encodeError(err error) []byte {
	type errorMsg struct {
		ErrorMessage string `json:"error_message"`
	}

	tmp := errorMsg{ErrorMessage: err.Error()}
	data, err := json.Marshal(tmp)
	if err != nil {
		logger.Error("encoding error: %v", err)
	}

	return data
}

func (j *JSONResponse) toJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		logger.Error("encoding error: %v", err)
		return []byte("")
	}

	return b
}

// NewJSONResponse returns a new JSONResponse.
func NewJSONResponse(w http.ResponseWriter, r *http.Request) *JSONResponse {
	return &JSONResponse{request: r, writer: w}
}
