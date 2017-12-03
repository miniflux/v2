// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package http

import "testing"

func TestHasServerFailureWith200Status(t *testing.T) {
	r := &Response{StatusCode: 200}
	if r.HasServerFailure() {
		t.Error("200 is not a failure")
	}
}

func TestHasServerFailureWith404Status(t *testing.T) {
	r := &Response{StatusCode: 404}
	if !r.HasServerFailure() {
		t.Error("404 is a failure")
	}
}

func TestHasServerFailureWith500Status(t *testing.T) {
	r := &Response{StatusCode: 500}
	if !r.HasServerFailure() {
		t.Error("500 is a failure")
	}
}

func TestIsModifiedWith304Status(t *testing.T) {
	r := &Response{StatusCode: 304}
	if r.IsModified("etag", "lastModified") {
		t.Error("The resource should not be considered modified")
	}
}

func TestIsModifiedWithIdenticalEtag(t *testing.T) {
	r := &Response{StatusCode: 200, ETag: "etag"}
	if r.IsModified("etag", "lastModified") {
		t.Error("The resource should not be considered modified")
	}
}

func TestIsModifiedWithIdenticalLastModified(t *testing.T) {
	r := &Response{StatusCode: 200, LastModified: "lastModified"}
	if r.IsModified("etag", "lastModified") {
		t.Error("The resource should not be considered modified")
	}
}

func TestIsModifiedWithDifferentHeaders(t *testing.T) {
	r := &Response{StatusCode: 200, ETag: "some etag", LastModified: "some date"}
	if !r.IsModified("etag", "lastModified") {
		t.Error("The resource should be considered modified")
	}
}
