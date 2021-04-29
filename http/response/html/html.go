// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package html // import "miniflux.app/http/response/html"

import (
	"net/http"

	"miniflux.app/http/response"
	"miniflux.app/logger"
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
	logger.Error("[HTTP:Internal Server Error] %s => %v", r.URL, err)

	builder := response.New(w, r)
	builder.WithStatus(http.StatusInternalServerError)
	builder.WithHeader("Content-Type", "text/html; charset=utf-8")
	builder.WithHeader("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")
	builder.WithBody(err)
	builder.Write()
}

// BadRequest sends a bad request error to the client.
func BadRequest(w http.ResponseWriter, r *http.Request, err error) {
	logger.Error("[HTTP:Bad Request] %s => %v", r.URL, err)

	builder := response.New(w, r)
	builder.WithStatus(http.StatusBadRequest)
	builder.WithHeader("Content-Type", "text/html; charset=utf-8")
	builder.WithHeader("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")
	builder.WithBody(err)
	builder.Write()
}

// Forbidden sends a forbidden error to the client.
func Forbidden(w http.ResponseWriter, r *http.Request) {
	logger.Error("[HTTP:Forbidden] %s", r.URL)

	builder := response.New(w, r)
	builder.WithStatus(http.StatusForbidden)
	builder.WithHeader("Content-Type", "text/html; charset=utf-8")
	builder.WithHeader("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store")
	builder.WithBody("Access Forbidden")
	builder.Write()
}

// NotFound sends a page not found error to the client.
func NotFound(w http.ResponseWriter, r *http.Request) {
	logger.Error("[HTTP:Not Found] %s", r.URL)

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

// RedirectWithQueries redirects the user to current location with given query string pairs.
func RedirectWithQueries(w http.ResponseWriter, r *http.Request, qs ...string) {
	u := r.URL
	query := u.Query()
	for i := 0; i < len(qs)-1; i += 2 {
		query.Set(qs[i], qs[i+1])
	}
	u.RawQuery = query.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
}
