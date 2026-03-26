// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package response // import "miniflux.app/v2/internal/http/response"

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		JSON(w, r, map[string]string{"key": "value"})
	})

	handler.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusOK)
	}

	if actualBody := w.Body.String(); actualBody != `{"key":"value"}` {
		t.Fatalf(`Unexpected body, got %q instead of %q`, actualBody, `{"key":"value"}`)
	}

	if actualContentType := resp.Header.Get("Content-Type"); actualContentType != jsonContentTypeHeader {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, jsonContentTypeHeader)
	}
}

func TestJSONCreatedResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		JSONCreated(w, r, map[string]string{"key": "value"})
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusCreated)
	}

	if actualBody := w.Body.String(); actualBody != `{"key":"value"}` {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, `{"key":"value"}`)
	}

	if actualContentType := resp.Header.Get("Content-Type"); actualContentType != jsonContentTypeHeader {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, jsonContentTypeHeader)
	}
}

func TestJSONAcceptedResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		JSONAccepted(w, r)
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusAccepted)
	}

	if actualBody := w.Body.String(); actualBody != `` {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, ``)
	}

	if actualContentType := resp.Header.Get("Content-Type"); actualContentType != jsonContentTypeHeader {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, jsonContentTypeHeader)
	}
}

func TestJSONServerErrorResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		JSONServerError(w, r, errors.New("some error"))
	})

	handler.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusInternalServerError)
	}

	if actualBody := w.Body.String(); actualBody != `{"error_message":"some error"}` {
		t.Fatalf(`Unexpected body, got %q instead of %q`, actualBody, `{"error_message":"some error"}`)
	}

	if actualContentType := resp.Header.Get("Content-Type"); actualContentType != jsonContentTypeHeader {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, jsonContentTypeHeader)
	}
}

func TestJSONBadRequestResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		JSONBadRequest(w, r, errors.New("Some Error"))
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusBadRequest)
	}

	if actualBody := w.Body.String(); actualBody != `{"error_message":"Some Error"}` {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, `{"error_message":"Some Error"}`)
	}

	if actualContentType := resp.Header.Get("Content-Type"); actualContentType != jsonContentTypeHeader {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, jsonContentTypeHeader)
	}
}

func TestJSONUnauthorizedResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		JSONUnauthorized(w, r)
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusUnauthorized)
	}

	if actualBody := w.Body.String(); actualBody != `{"error_message":"access unauthorized"}` {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, `{"error_message":"access unauthorized"}`)
	}

	if actualContentType := resp.Header.Get("Content-Type"); actualContentType != jsonContentTypeHeader {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, jsonContentTypeHeader)
	}
}

func TestJSONForbiddenResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		JSONForbidden(w, r)
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusForbidden)
	}

	if actualBody := w.Body.String(); actualBody != `{"error_message":"access forbidden"}` {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, `{"error_message":"access forbidden"}`)
	}

	if actualContentType := resp.Header.Get("Content-Type"); actualContentType != jsonContentTypeHeader {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, jsonContentTypeHeader)
	}
}

func TestJSONNotFoundResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		JSONNotFound(w, r)
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusNotFound)
	}

	if actualBody := w.Body.String(); actualBody != `{"error_message":"resource not found"}` {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, `{"error_message":"resource not found"}`)
	}

	if actualContentType := resp.Header.Get("Content-Type"); actualContentType != jsonContentTypeHeader {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, jsonContentTypeHeader)
	}
}

func TestBuildInvalidJSONResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		JSON(w, r, make(chan int))
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusInternalServerError)
	}

	if actualBody := w.Body.String(); actualBody != `{"error_message":"json: unsupported type: chan int"}` {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, `{"error_message":"json: unsupported type: chan int"}`)
	}

	if actualContentType := resp.Header.Get("Content-Type"); actualContentType != jsonContentTypeHeader {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, jsonContentTypeHeader)
	}
}

func TestBuildInvalidJSONCreatedResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		JSONCreated(w, r, make(chan int))
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusInternalServerError)
	}

	if actualBody := w.Body.String(); actualBody != `{"error_message":"json: unsupported type: chan int"}` {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, `{"error_message":"json: unsupported type: chan int"}`)
	}

	if actualContentType := resp.Header.Get("Content-Type"); actualContentType != jsonContentTypeHeader {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, jsonContentTypeHeader)
	}
}

func TestGenerateJSONError(t *testing.T) {
	actualBody := string(generateJSONError(errors.New("some error")))
	if actualBody != `{"error_message":"some error"}` {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, `{"error_message":"some error"}`)
	}
}
