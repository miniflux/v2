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

type Request struct {
	writer  http.ResponseWriter
	request *http.Request
}

func (r *Request) GetRequest() *http.Request {
	return r.request
}

func (r *Request) GetBody() io.ReadCloser {
	return r.request.Body
}

func (r *Request) GetHeaders() http.Header {
	return r.request.Header
}

func (r *Request) GetScheme() string {
	return r.request.URL.Scheme
}

func (r *Request) GetFile(name string) (multipart.File, *multipart.FileHeader, error) {
	return r.request.FormFile(name)
}

func (r *Request) IsHTTPS() bool {
	return r.request.URL.Scheme == "https"
}

func (r *Request) GetCookie(name string) string {
	cookie, err := r.request.Cookie(name)
	if err == http.ErrNoCookie {
		return ""
	}

	return cookie.Value
}

func (r *Request) GetIntegerParam(param string) (int64, error) {
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

func (r *Request) GetStringParam(param, defaultValue string) string {
	vars := mux.Vars(r.request)
	value := vars[param]
	if value == "" {
		value = defaultValue
	}
	return value
}

func (r *Request) GetQueryStringParam(param, defaultValue string) string {
	value := r.request.URL.Query().Get(param)
	if value == "" {
		value = defaultValue
	}
	return value
}

func (r *Request) GetQueryIntegerParam(param string, defaultValue int) int {
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

func NewRequest(w http.ResponseWriter, r *http.Request) *Request {
	return &Request{writer: w, request: r}
}
