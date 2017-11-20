// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package core

import (
	"github.com/miniflux/miniflux2/server/template"
	"log"
	"net/http"
)

type HtmlResponse struct {
	writer   http.ResponseWriter
	request  *http.Request
	template *template.TemplateEngine
}

func (h *HtmlResponse) Render(template string, args map[string]interface{}) {
	h.writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.template.Execute(h.writer, template, args)
}

func (h *HtmlResponse) ServerError(err error) {
	h.writer.WriteHeader(http.StatusInternalServerError)
	h.writer.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err != nil {
		log.Println(err)
		h.writer.Write([]byte("Internal Server Error: " + err.Error()))
	} else {
		h.writer.Write([]byte("Internal Server Error"))
	}
}

func (h *HtmlResponse) BadRequest(err error) {
	h.writer.WriteHeader(http.StatusBadRequest)
	h.writer.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err != nil {
		log.Println(err)
		h.writer.Write([]byte("Bad Request: " + err.Error()))
	} else {
		h.writer.Write([]byte("Bad Request"))
	}
}

func (h *HtmlResponse) NotFound() {
	h.writer.WriteHeader(http.StatusNotFound)
	h.writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.writer.Write([]byte("Page Not Found"))
}

func (h *HtmlResponse) Forbidden() {
	h.writer.WriteHeader(http.StatusForbidden)
	h.writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.writer.Write([]byte("Access Forbidden"))
}
