// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package client // import "miniflux.app/http/client"

import (
	"io"
	"mime"
	"strings"

	"golang.org/x/net/html/charset"
	"miniflux.app/logger"
)

// Response wraps a server response.
type Response struct {
	Body          io.Reader
	StatusCode    int
	EffectiveURL  string
	LastModified  string
	ETag          string
	ContentType   string
	ContentLength int64
}

// IsNotFound returns true if the resource doesn't exists anymore.
func (r *Response) IsNotFound() bool {
	return r.StatusCode == 404 || r.StatusCode == 410
}

// IsNotAuthorized returns true if the resource require authentication.
func (r *Response) IsNotAuthorized() bool {
	return r.StatusCode == 401
}

// HasServerFailure returns true if the status code represents a failure.
func (r *Response) HasServerFailure() bool {
	return r.StatusCode >= 400
}

// IsModified returns true if the resource has been modified.
func (r *Response) IsModified(etag, lastModified string) bool {
	if r.StatusCode == 304 {
		return false
	}

	if r.ETag != "" && r.ETag == etag {
		return false
	}

	if r.LastModified != "" && r.LastModified == lastModified {
		return false
	}

	return true
}

// NormalizeBodyEncoding make sure the body is encoded in UTF-8.
//
// If a charset other than UTF-8 is detected, we convert the document to UTF-8.
// This is used by the scraper and feed readers.
//
// Do not forget edge cases:
// - Some non-utf8 feeds specify encoding only in Content-Type, not in XML document.
func (r *Response) NormalizeBodyEncoding() (io.Reader, error) {
	_, params, err := mime.ParseMediaType(r.ContentType)
	if err == nil {
		if enc, found := params["charset"]; found {
			enc = strings.ToLower(enc)
			if enc != "utf-8" && enc != "utf8" && enc != "" {
				logger.Debug("[NormalizeBodyEncoding] Convert body to UTF-8 from %s", enc)
				return charset.NewReader(r.Body, r.ContentType)
			}
		}
	}
	return r.Body, nil
}
