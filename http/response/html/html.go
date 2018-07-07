// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package html

import (
	"net/http"

	"github.com/miniflux/miniflux/http/response"
	"github.com/miniflux/miniflux/logger"
)

// OK writes a standard HTML response.
func OK(w http.ResponseWriter, r *http.Request, b []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	response.Compress(w, r, b)
}

// ServerError sends a 500 error to the browser.
func ServerError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)

	if err != nil {
		logger.Error("[Internal Server Error] %v", err)
		w.Write([]byte("Internal Server Error: " + err.Error()))
	} else {
		w.Write([]byte("Internal Server Error"))
	}
}

// BadRequest sends a 400 error to the browser.
func BadRequest(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)

	if err != nil {
		logger.Error("[Bad Request] %v", err)
		w.Write([]byte("Bad Request: " + err.Error()))
	} else {
		w.Write([]byte("Bad Request"))
	}
}

// NotFound sends a 404 error to the browser.
func NotFound(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Page Not Found"))
}

// Forbidden sends a 403 error to the browser.
func Forbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("Access Forbidden"))
}
