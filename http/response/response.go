// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package response

import (
	"compress/flate"
	"compress/gzip"
	"net/http"
	"strings"
	"time"
)

// Redirect redirects the user to another location.
func Redirect(w http.ResponseWriter, r *http.Request, path string) {
	http.Redirect(w, r, path, http.StatusFound)
}

// NotModified sends a response with a 304 status code.
func NotModified(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotModified)
}

// Cache returns a response with caching headers.
func Cache(w http.ResponseWriter, r *http.Request, mimeType, etag string, data []byte, duration time.Duration) {
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "public")
	w.Header().Set("Expires", time.Now().Add(duration).Format(time.RFC1123))

	if etag == r.Header.Get("If-None-Match") {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	switch mimeType {
	case "text/javascript; charset=utf-8", "text/css; charset=utf-8":
		Compress(w, r, data)
	default:
		w.Write(data)
	}
}

// Compress the response sent to the browser.
func Compress(w http.ResponseWriter, r *http.Request, data []byte) {
	acceptEncoding := r.Header.Get("Accept-Encoding")

	switch {
	case strings.Contains(acceptEncoding, "gzip"):
		w.Header().Set("Content-Encoding", "gzip")
		gzipWriter := gzip.NewWriter(w)
		defer gzipWriter.Close()
		gzipWriter.Write(data)
	case strings.Contains(acceptEncoding, "deflate"):
		w.Header().Set("Content-Encoding", "deflate")
		flateWriter, _ := flate.NewWriter(w, -1)
		defer flateWriter.Close()
		flateWriter.Write(data)
	default:
		w.Write(data)
	}
}
