// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package http

import "io"

type ServerResponse struct {
	Body         io.Reader
	StatusCode   int
	EffectiveURL string
	LastModified string
	ETag         string
	ContentType  string
}

func (s *ServerResponse) HasServerFailure() bool {
	return s.StatusCode >= 400
}

func (s *ServerResponse) IsModified(etag, lastModified string) bool {
	if s.StatusCode == 304 {
		return false
	}

	if s.ETag != "" && s.LastModified != "" && (s.ETag == etag || s.LastModified == lastModified) {
		return false
	}

	return true
}
