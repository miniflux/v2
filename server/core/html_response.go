// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package core

import (
	"log"
	"net/http"

	"github.com/miniflux/miniflux2/server/template"
)

// HTMLResponse handles HTML responses.
type HTMLResponse struct {
	writer   http.ResponseWriter
	request  *http.Request
	template *template.TemplateEngine
}

// Render execute a template and send to the client the generated HTML.
func (h *HTMLResponse) Render(template string, args map[string]interface{}) {
	h.writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.template.Execute(h.writer, template, args)
}

// ServerError sends a 500 error to the browser.
func (h *HTMLResponse) ServerError(err error) {
	h.writer.WriteHeader(http.StatusInternalServerError)
	h.writer.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err != nil {
		log.Println(err)
		h.writer.Write([]byte("Internal Server Error: " + err.Error()))
	} else {
		h.writer.Write([]byte("Internal Server Error"))
	}
}

// BadRequest sends a 400 error to the browser.
func (h *HTMLResponse) BadRequest(err error) {
	h.writer.WriteHeader(http.StatusBadRequest)
	h.writer.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err != nil {
		log.Println(err)
		h.writer.Write([]byte("Bad Request: " + err.Error()))
	} else {
		h.writer.Write([]byte("Bad Request"))
	}
}

// NotFound sends a 404 error to the browser.
func (h *HTMLResponse) NotFound() {
	h.writer.WriteHeader(http.StatusNotFound)
	h.writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.writer.Write([]byte("Page Not Found"))
}

// Forbidden sends a 403 error to the browser.
func (h *HTMLResponse) Forbidden() {
	h.writer.WriteHeader(http.StatusForbidden)
	h.writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.writer.Write([]byte("Access Forbidden"))
}
