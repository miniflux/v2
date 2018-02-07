// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package handler

import (
	"net/http"
	"time"

	"github.com/miniflux/miniflux/config"
	"github.com/miniflux/miniflux/template"
)

// Response handles HTTP responses.
type Response struct {
	cfg      *config.Config
	writer   http.ResponseWriter
	request  *http.Request
	template *template.Engine
}

// SetCookie send a cookie to the client.
func (r *Response) SetCookie(cookie *http.Cookie) {
	http.SetCookie(r.writer, cookie)
}

// JSON returns a JSONResponse.
func (r *Response) JSON() *JSONResponse {
	r.commonHeaders()
	return NewJSONResponse(r.writer, r.request)
}

// HTML returns a HTMLResponse.
func (r *Response) HTML() *HTMLResponse {
	r.commonHeaders()
	return &HTMLResponse{writer: r.writer, request: r.request, template: r.template}
}

// XML returns a XMLResponse.
func (r *Response) XML() *XMLResponse {
	r.commonHeaders()
	return &XMLResponse{writer: r.writer, request: r.request}
}

// Redirect redirects the user to another location.
func (r *Response) Redirect(path string) {
	http.Redirect(r.writer, r.request, path, http.StatusFound)
}

// NotModified sends a response with a 304 status code.
func (r *Response) NotModified() {
	r.commonHeaders()
	r.writer.WriteHeader(http.StatusNotModified)
}

// Cache returns a response with caching headers.
func (r *Response) Cache(mimeType, etag string, content []byte, duration time.Duration) {
	r.writer.Header().Set("Content-Type", mimeType)
	r.writer.Header().Set("ETag", etag)
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

	// Even if the directive "frame-src" has been deprecated in Firefox,
	// we keep it to stay compatible with other browsers.
	r.writer.Header().Set("Content-Security-Policy", "default-src 'self'; img-src *; media-src *; frame-src *; child-src *")

	if r.cfg.IsHTTPS && r.cfg.HasHSTS() {
		r.writer.Header().Set("Strict-Transport-Security", "max-age=31536000")
	}
}

// NewResponse returns a new Response.
func NewResponse(cfg *config.Config, w http.ResponseWriter, r *http.Request, template *template.Engine) *Response {
	return &Response{cfg: cfg, writer: w, request: r, template: template}
}
