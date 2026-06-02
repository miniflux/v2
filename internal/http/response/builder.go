// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package response // import "miniflux.app/v2/internal/http/response"

import (
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"mime"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/andybalholm/brotli/matchfinder"
)

const compressionThreshold = 1024

// Compression writers are pooled so each request reuses their internal
// state (brotli sliding window + hash tables, flate dictionary, etc.)
// instead of allocating it from scratch. Reset(dst) rebinds the
// destination without re-allocating the buffers.
var (
	brotliWriterPool = sync.Pool{
		New: func() any {
			return brotli.NewWriterV2(io.Discard, brotli.DefaultCompression)
		},
	}
	gzipWriterPool = sync.Pool{
		New: func() any {
			return gzip.NewWriter(io.Discard)
		},
	}
	flateWriterPool = sync.Pool{
		New: func() any {
			w, _ := flate.NewWriter(io.Discard, flate.DefaultCompression)
			return w
		},
	}
)

// Builder generates HTTP responses.
type Builder struct {
	w                 http.ResponseWriter
	r                 *http.Request
	statusCode        int
	headers           http.Header
	enableCompression bool
	body              any
}

// NewBuilder creates a new response builder.
func NewBuilder(w http.ResponseWriter, r *http.Request) *Builder {
	return &Builder{w: w, r: r, statusCode: http.StatusOK, headers: make(http.Header), enableCompression: true}
}

// WithStatus uses the given status code to build the response.
func (b *Builder) WithStatus(statusCode int) *Builder {
	b.statusCode = statusCode
	return b
}

// WithHeader adds the given HTTP header to the response.
func (b *Builder) WithHeader(key, value string) *Builder {
	b.headers.Set(key, value)
	return b
}

// WithBodyAsBytes uses the given bytes to build the response.
func (b *Builder) WithBodyAsBytes(body []byte) *Builder {
	b.body = body
	return b
}

// WithBodyAsString uses the given string to build the response.
func (b *Builder) WithBodyAsString(body string) *Builder {
	b.body = body
	return b
}

// WithBodyAsReader uses the given reader to build the response.
func (b *Builder) WithBodyAsReader(body io.Reader) *Builder {
	b.body = body
	return b
}

// WithAttachment forces the document to be downloaded by the web browser.
func (b *Builder) WithAttachment(filename string) *Builder {
	b.headers.Set("Content-Disposition", formatContentDisposition("attachment", filename))
	return b
}

// WithInline suggests an inline filename for the current response.
func (b *Builder) WithInline(filename string) *Builder {
	b.headers.Set("Content-Disposition", formatContentDisposition("inline", filename))
	return b
}

// WithoutCompression disables HTTP compression.
func (b *Builder) WithoutCompression() *Builder {
	b.enableCompression = false
	return b
}

// WithCaching adds caching headers to the response.
func (b *Builder) WithCaching(etag string, duration time.Duration, callback func(*Builder)) {
	etag = normalizeETag(etag)
	b.headers.Set("ETag", etag)
	// max-age is required for the "immutable" directive to take effect: without
	// it, browsers still revalidate content-hashed assets on every reload.
	b.headers.Set("Cache-Control", fmt.Sprintf("public, max-age=%d, immutable", int64(duration.Seconds())))
	b.headers.Set("Expires", time.Now().Add(duration).UTC().Format(http.TimeFormat))

	if ifNoneMatch(b.r.Header.Get("If-None-Match"), etag) {
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
	case io.Reader:
		// Compression not implemented in this case
		b.writeHeaders()
		_, err := io.Copy(b.w, v)
		if err != nil {
			slog.Error("Unable to write response body", slog.Any("error", err))
		}
	}
}

func (b *Builder) writeHeaders() {
	b.headers.Set("X-Content-Type-Options", "nosniff")
	b.headers.Set("X-Frame-Options", "DENY")
	b.headers.Set("Referrer-Policy", "no-referrer")

	maps.Copy(b.w.Header(), b.headers)

	b.w.WriteHeader(b.statusCode)
}

// values should be in sync with [Builder.compress] switch/case.
var acceptEncoding = AcceptEncoding("br", "gzip", "deflate")

func (b *Builder) compress(data []byte) {
	if b.enableCompression && len(data) > compressionThreshold {
		b.headers.Set("Vary", "Accept-Encoding")

		encoding := acceptEncoding.Parse(b.r.Header.Get("Accept-Encoding"))
		switch encoding {
		case "br":
			b.headers.Set("Content-Encoding", "br")
			b.writeHeaders()

			brotliWriter := brotliWriterPool.Get().(*matchfinder.Writer)
			brotliWriter.Reset(b.w)
			brotliWriter.Write(data)
			brotliWriter.Close()
			brotliWriter.Reset(io.Discard)
			brotliWriterPool.Put(brotliWriter)
			return
		case "gzip":
			b.headers.Set("Content-Encoding", "gzip")
			b.writeHeaders()

			gzipWriter := gzipWriterPool.Get().(*gzip.Writer)
			gzipWriter.Reset(b.w)
			gzipWriter.Write(data)
			gzipWriter.Close()
			gzipWriter.Reset(io.Discard)
			gzipWriterPool.Put(gzipWriter)
			return
		case "deflate":
			b.headers.Set("Content-Encoding", "deflate")
			b.writeHeaders()

			flateWriter := flateWriterPool.Get().(*flate.Writer)
			flateWriter.Reset(b.w)
			flateWriter.Write(data)
			flateWriter.Close()
			flateWriter.Reset(io.Discard)
			flateWriterPool.Put(flateWriter)
			return
		}
	}

	b.writeHeaders()
	b.w.Write(data)
}

func normalizeETag(etag string) string {
	etag = strings.TrimSpace(etag)
	if etag == "" {
		return ""
	}
	if strings.HasPrefix(etag, `"`) || strings.HasPrefix(etag, `W/"`) {
		return etag
	}
	return `"` + etag + `"`
}

func ifNoneMatch(headerValue, etag string) bool {
	if headerValue == "" || etag == "" {
		return false
	}
	if strings.TrimSpace(headerValue) == "*" {
		return true
	}
	// Weak ETag comparison: the opaque-tag (quoted string without W/ prefix) must match.
	return strings.Contains(headerValue, strings.TrimPrefix(etag, `W/`))
}

func formatContentDisposition(dispositionType, filename string) string {
	if filename == "" {
		return dispositionType
	}

	if value := mime.FormatMediaType(dispositionType, map[string]string{"filename": filename}); value != "" {
		return value
	}

	return dispositionType
}
