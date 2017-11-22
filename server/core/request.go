// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package core

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Request is a thin wrapper around "http.Request".
type Request struct {
	writer  http.ResponseWriter
	request *http.Request
}

// Request returns the raw Request struct.
func (r *Request) Request() *http.Request {
	return r.request
}

// Body returns the request body.
func (r *Request) Body() io.ReadCloser {
	return r.request.Body
}

// File returns uploaded file properties.
func (r *Request) File(name string) (multipart.File, *multipart.FileHeader, error) {
	return r.request.FormFile(name)
}

// IsHTTPS returns if the request is made over HTTPS.
func (r *Request) IsHTTPS() bool {
	return r.request.URL.Scheme == "https"
}

// Cookie returns the cookie value.
func (r *Request) Cookie(name string) string {
	cookie, err := r.request.Cookie(name)
	if err == http.ErrNoCookie {
		return ""
	}

	return cookie.Value
}

// IntegerParam returns an URL parameter as integer.
func (r *Request) IntegerParam(param string) (int64, error) {
	vars := mux.Vars(r.request)
	value, err := strconv.Atoi(vars[param])
	if err != nil {
		log.Println(err)
		return 0, fmt.Errorf("%s parameter is not an integer", param)
	}

	if value < 0 {
		return 0, nil
	}

	return int64(value), nil
}

// StringParam returns an URL parameter as string.
func (r *Request) StringParam(param, defaultValue string) string {
	vars := mux.Vars(r.request)
	value := vars[param]
	if value == "" {
		value = defaultValue
	}
	return value
}

// QueryStringParam returns a querystring parameter as string.
func (r *Request) QueryStringParam(param, defaultValue string) string {
	value := r.request.URL.Query().Get(param)
	if value == "" {
		value = defaultValue
	}
	return value
}

// QueryIntegerParam returns a querystring parameter as string.
func (r *Request) QueryIntegerParam(param string, defaultValue int) int {
	value := r.request.URL.Query().Get(param)
	if value == "" {
		return defaultValue
	}

	val, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	if val < 0 {
		return defaultValue
	}

	return val
}

// NewRequest returns a new Request struct.
func NewRequest(w http.ResponseWriter, r *http.Request) *Request {
	return &Request{writer: w, request: r}
}
