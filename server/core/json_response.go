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

type JsonResponse struct {
	writer  http.ResponseWriter
	request *http.Request
}

func (j *JsonResponse) Standard(v interface{}) {
	j.writer.WriteHeader(http.StatusOK)
	j.commonHeaders()
	j.writer.Write(j.toJSON(v))
}

func (j *JsonResponse) Created(v interface{}) {
	j.writer.WriteHeader(http.StatusCreated)
	j.commonHeaders()
	j.writer.Write(j.toJSON(v))
}

func (j *JsonResponse) NoContent() {
	j.writer.WriteHeader(http.StatusNoContent)
	j.commonHeaders()
}

func (j *JsonResponse) BadRequest(err error) {
	log.Println("[API:BadRequest]", err)
	j.writer.WriteHeader(http.StatusBadRequest)
	j.commonHeaders()

	if err != nil {
		j.writer.Write(j.encodeError(err))
	}
}

func (j *JsonResponse) NotFound(err error) {
	log.Println("[API:NotFound]", err)
	j.writer.WriteHeader(http.StatusNotFound)
	j.commonHeaders()
	j.writer.Write(j.encodeError(err))
}

func (j *JsonResponse) ServerError(err error) {
	log.Println("[API:ServerError]", err)
	j.writer.WriteHeader(http.StatusInternalServerError)
	j.commonHeaders()

	if err != nil {
		j.writer.Write(j.encodeError(err))
	}
}

func (j *JsonResponse) Forbidden() {
	log.Println("[API:Forbidden]")
	j.writer.WriteHeader(http.StatusForbidden)
	j.commonHeaders()
	j.writer.Write(j.encodeError(errors.New("Access Forbidden")))
}

func (j *JsonResponse) commonHeaders() {
	j.writer.Header().Set("Accept", "application/json")
	j.writer.Header().Set("Content-Type", "application/json")
}

func (j *JsonResponse) encodeError(err error) []byte {
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

func (j *JsonResponse) toJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		log.Println("Unable to convert interface to JSON:", err)
		return []byte("")
	}

	return b
}
