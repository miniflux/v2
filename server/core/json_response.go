// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package core

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// JSONResponse handles JSON responses.
type JSONResponse struct {
	writer  http.ResponseWriter
	request *http.Request
}

// Standard sends a JSON response with the status code 200.
func (j *JSONResponse) Standard(v interface{}) {
	j.writer.WriteHeader(http.StatusOK)
	j.commonHeaders()
	j.writer.Write(j.toJSON(v))
}

// Created sends a JSON response with the status code 201.
func (j *JSONResponse) Created(v interface{}) {
	j.writer.WriteHeader(http.StatusCreated)
	j.commonHeaders()
	j.writer.Write(j.toJSON(v))
}

// NoContent sends a JSON response with the status code 204.
func (j *JSONResponse) NoContent() {
	j.writer.WriteHeader(http.StatusNoContent)
	j.commonHeaders()
}

// BadRequest sends a JSON response with the status code 400.
func (j *JSONResponse) BadRequest(err error) {
	log.Println("[API:BadRequest]", err)
	j.writer.WriteHeader(http.StatusBadRequest)
	j.commonHeaders()

	if err != nil {
		j.writer.Write(j.encodeError(err))
	}
}

// NotFound sends a JSON response with the status code 404.
func (j *JSONResponse) NotFound(err error) {
	log.Println("[API:NotFound]", err)
	j.writer.WriteHeader(http.StatusNotFound)
	j.commonHeaders()
	j.writer.Write(j.encodeError(err))
}

// ServerError sends a JSON response with the status code 500.
func (j *JSONResponse) ServerError(err error) {
	log.Println("[API:ServerError]", err)
	j.writer.WriteHeader(http.StatusInternalServerError)
	j.commonHeaders()

	if err != nil {
		j.writer.Write(j.encodeError(err))
	}
}

// Forbidden sends a JSON response with the status code 403.
func (j *JSONResponse) Forbidden() {
	log.Println("[API:Forbidden]")
	j.writer.WriteHeader(http.StatusForbidden)
	j.commonHeaders()
	j.writer.Write(j.encodeError(errors.New("Access Forbidden")))
}

func (j *JSONResponse) commonHeaders() {
	j.writer.Header().Set("Accept", "application/json")
	j.writer.Header().Set("Content-Type", "application/json")
}

func (j *JSONResponse) encodeError(err error) []byte {
	type errorMsg struct {
		ErrorMessage string `json:"error_message"`
	}

	tmp := errorMsg{ErrorMessage: err.Error()}
	data, err := json.Marshal(tmp)
	if err != nil {
		log.Println("encodeError:", err)
	}

	return data
}

func (j *JSONResponse) toJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		log.Println("Unable to convert interface to JSON:", err)
		return []byte("")
	}

	return b
}

// NewJSONResponse returns a new JSONResponse.
func NewJSONResponse(w http.ResponseWriter, r *http.Request) *JSONResponse {
	return &JSONResponse{request: r, writer: w}
}
