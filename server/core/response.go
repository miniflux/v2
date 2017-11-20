// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package core

import (
	"github.com/miniflux/miniflux2/server/template"
	"net/http"
	"time"
)

type Response struct {
	writer   http.ResponseWriter
	request  *http.Request
	template *template.TemplateEngine
}

func (r *Response) SetCookie(cookie *http.Cookie) {
	http.SetCookie(r.writer, cookie)
}

func (r *Response) Json() *JsonResponse {
	r.commonHeaders()
	return &JsonResponse{writer: r.writer, request: r.request}
}

func (r *Response) Html() *HtmlResponse {
	r.commonHeaders()
	return &HtmlResponse{writer: r.writer, request: r.request, template: r.template}
}

func (r *Response) Xml() *XmlResponse {
	r.commonHeaders()
	return &XmlResponse{writer: r.writer, request: r.request}
}

func (r *Response) Redirect(path string) {
	http.Redirect(r.writer, r.request, path, http.StatusFound)
}

func (r *Response) Cache(mime_type, etag string, content []byte, duration time.Duration) {
	r.writer.Header().Set("Content-Type", mime_type)
	r.writer.Header().Set("Etag", etag)
	r.writer.Header().Set("Cache-Control", "public")
	r.writer.Header().Set("Expires", time.Now().Add(duration).Format(time.RFC1123))

	if etag == r.request.Header.Get("If-None-Match") {
		r.writer.WriteHeader(http.StatusNotModified)
	} else {
		r.writer.Write(content)
	}
}

func (r *Response) commonHeaders() {
	r.writer.Header().Set("X-XSS-Protection", "1; mode=block")
	r.writer.Header().Set("X-Content-Type-Options", "nosniff")
	r.writer.Header().Set("X-Frame-Options", "DENY")
}

func NewResponse(w http.ResponseWriter, r *http.Request, template *template.TemplateEngine) *Response {
	return &Response{writer: w, request: r, template: template}
}
