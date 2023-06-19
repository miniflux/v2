// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package response // import "miniflux.app/http/response"

import (
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"miniflux.app/logger"
)

const compressionThreshold = 1024

// Builder generates HTTP responses.
type Builder struct {
	w                 http.ResponseWriter
	r                 *http.Request
	statusCode        int
	headers           map[string]string
	enableCompression bool
	body              interface{}
}

// WithStatus uses the given status code to build the response.
func (b *Builder) WithStatus(statusCode int) *Builder {
	b.statusCode = statusCode
	return b
}

// WithHeader adds the given HTTP header to the response.
func (b *Builder) WithHeader(key, value string) *Builder {
	b.headers[key] = value
	return b
}

// WithBody uses the given body to build the response.
func (b *Builder) WithBody(body interface{}) *Builder {
	b.body = body
	return b
}

// WithAttachment forces the document to be downloaded by the web browser.
func (b *Builder) WithAttachment(filename string) *Builder {
	b.headers["Content-Disposition"] = fmt.Sprintf("attachment; filename=%s", filename)
	return b
}

// WithoutCompression disables HTTP compression.
func (b *Builder) WithoutCompression() *Builder {
	b.enableCompression = false
	return b
}

// WithCaching adds caching headers to the response.
func (b *Builder) WithCaching(etag string, duration time.Duration, callback func(*Builder)) {
	b.headers["ETag"] = etag
	b.headers["Cache-Control"] = "public"
	b.headers["Expires"] = time.Now().Add(duration).Format(time.RFC1123)

	if etag == b.r.Header.Get("If-None-Match") {
		b.statusCode = http.StatusNotModified
		b.body = nil
		b.Write()
	} else {
		callback(b)
	}
}

// Write generates the HTTP response.
func (b *Builder) Write() {
	if b.body == nil {
		b.writeHeaders()
		return
	}

	switch v := b.body.(type) {
	case []byte:
		b.compress(v)
	case string:
		b.compress([]byte(v))
	case error:
		b.compress([]byte(v.Error()))
	case io.Reader:
		// Compression not implemented in this case
		b.writeHeaders()
		_, err := io.Copy(b.w, v)
		if err != nil {
			logger.Error("%v", err)
		}
	}
}

func (b *Builder) writeHeaders() {
	b.headers["X-XSS-Protection"] = "1; mode=block"
	b.headers["X-Content-Type-Options"] = "nosniff"
	b.headers["X-Frame-Options"] = "DENY"
	b.headers["Referrer-Policy"] = "no-referrer"

	for key, value := range b.headers {
		b.w.Header().Set(key, value)
	}

	b.w.WriteHeader(b.statusCode)
}

func (b *Builder) compress(data []byte) {
	if b.enableCompression && len(data) > compressionThreshold {
		acceptEncoding := b.r.Header.Get("Accept-Encoding")

		switch {
		case strings.Contains(acceptEncoding, "gzip"):
			b.headers["Content-Encoding"] = "gzip"
			b.writeHeaders()

			gzipWriter := gzip.NewWriter(b.w)
			defer gzipWriter.Close()
			gzipWriter.Write(data)
			return
		case strings.Contains(acceptEncoding, "deflate"):
			b.headers["Content-Encoding"] = "deflate"
			b.writeHeaders()

			flateWriter, _ := flate.NewWriter(b.w, -1)
			defer flateWriter.Close()
			flateWriter.Write(data)
			return
		}
	}

	b.writeHeaders()
	b.w.Write(data)
}

// New creates a new response builder.
func New(w http.ResponseWriter, r *http.Request) *Builder {
	return &Builder{w: w, r: r, statusCode: http.StatusOK, headers: make(map[string]string), enableCompression: true}
}
