// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package client // import "miniflux.app/http/client"

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html/charset"
)

var xmlEncodingRegex = regexp.MustCompile(`<\?xml(.*)encoding=["'](.+)["'](.*)\?>`)

// Response wraps a server response.
type Response struct {
	Body          io.Reader
	StatusCode    int
	EffectiveURL  string
	LastModified  string
	ETag          string
	Expires       string
	ContentType   string
	ContentLength int64
}

func (r *Response) String() string {
	return fmt.Sprintf(
		`StatusCode=%d EffectiveURL=%q LastModified=%q ETag=%s Expires=%s ContentType=%q ContentLength=%d`,
		r.StatusCode,
		r.EffectiveURL,
		r.LastModified,
		r.ETag,
		r.Expires,
		r.ContentType,
		r.ContentLength,
	)
}

// IsNotFound returns true if the resource doesn't exist anymore.
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

// EnsureUnicodeBody makes sure the body is encoded in UTF-8.
//
// If a charset other than UTF-8 is detected, we convert the document to UTF-8.
// This is used by the scraper and feed readers.
//
// Do not forget edge cases:
//
// - Feeds with encoding specified only in Content-Type header and not in XML document
// - Feeds with encoding specified in both places
// - Feeds with encoding specified only in XML document and not in HTTP header
// - Feeds with wrong encoding defined and already in UTF-8
func (r *Response) EnsureUnicodeBody() (err error) {
	buffer, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	r.Body = bytes.NewReader(buffer)
	if utf8.Valid(buffer) {
		return nil
	}

	if strings.Contains(r.ContentType, "xml") {
		// We ignore documents with encoding specified in XML prolog.
		// This is going to be handled by the XML parser.
		length := 1024
		if len(buffer) < 1024 {
			length = len(buffer)
		}

		if xmlEncodingRegex.Match(buffer[0:length]) {
			return nil
		}
	}

	r.Body, err = charset.NewReader(r.Body, r.ContentType)
	return err
}

// BodyAsString returns the response body as string.
func (r *Response) BodyAsString() string {
	bytes, _ := io.ReadAll(r.Body)
	return string(bytes)
}
