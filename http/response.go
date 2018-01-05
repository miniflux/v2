// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package http

import (
	"io"
	"strings"

	"golang.org/x/net/html/charset"
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
func (r *Response) NormalizeBodyEncoding() (io.Reader, error) {
	if strings.Contains(r.ContentType, "charset=") {
		return charset.NewReader(r.Body, r.ContentType)
	}
	return r.Body, nil
}
