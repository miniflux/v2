// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package json // import "miniflux.app/http/response/json"

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOKResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		OK(w, r, map[string]string{"key": "value"})
	})

	handler.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	expectedStatusCode := http.StatusOK
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}

	expectedBody := `{"key":"value"}`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %q instead of %q`, actualBody, expectedBody)
	}

	expectedContentType := contentTypeHeader
	actualContentType := resp.Header.Get("Content-Type")
	if actualContentType != expectedContentType {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, expectedContentType)
	}
}

func TestCreatedResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Created(w, r, map[string]string{"key": "value"})
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expectedStatusCode := http.StatusCreated
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}

	expectedBody := `{"key":"value"}`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, expectedBody)
	}

	expectedContentType := contentTypeHeader
	actualContentType := resp.Header.Get("Content-Type")
	if actualContentType != expectedContentType {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, expectedContentType)
	}
}

func TestNoContentResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NoContent(w, r)
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expectedStatusCode := http.StatusNoContent
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}

	expectedBody := ``
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, expectedBody)
	}

	expectedContentType := contentTypeHeader
	actualContentType := resp.Header.Get("Content-Type")
	if actualContentType != expectedContentType {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, expectedContentType)
	}
}

func TestServerErrorResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServerError(w, r, errors.New("some error"))
	})

	handler.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	expectedStatusCode := http.StatusInternalServerError
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}

	expectedBody := `{"error_message":"some error"}`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %q instead of %q`, actualBody, expectedBody)
	}

	expectedContentType := contentTypeHeader
	actualContentType := resp.Header.Get("Content-Type")
	if actualContentType != expectedContentType {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, expectedContentType)
	}
}

func TestBadRequestResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		BadRequest(w, r, errors.New("Some Error"))
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expectedStatusCode := http.StatusBadRequest
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}

	expectedBody := `{"error_message":"Some Error"}`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, expectedBody)
	}

	expectedContentType := contentTypeHeader
	actualContentType := resp.Header.Get("Content-Type")
	if actualContentType != expectedContentType {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, expectedContentType)
	}
}

func TestUnauthorizedResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Unauthorized(w, r)
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expectedStatusCode := http.StatusUnauthorized
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}

	expectedBody := `{"error_message":"Access Unauthorized"}`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, expectedBody)
	}

	expectedContentType := contentTypeHeader
	actualContentType := resp.Header.Get("Content-Type")
	if actualContentType != expectedContentType {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, expectedContentType)
	}
}

func TestForbiddenResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Forbidden(w, r)
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expectedStatusCode := http.StatusForbidden
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}

	expectedBody := `{"error_message":"Access Forbidden"}`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, expectedBody)
	}

	expectedContentType := contentTypeHeader
	actualContentType := resp.Header.Get("Content-Type")
	if actualContentType != expectedContentType {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, expectedContentType)
	}
}

func TestNotFoundResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NotFound(w, r)
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expectedStatusCode := http.StatusNotFound
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}

	expectedBody := `{"error_message":"Resource Not Found"}`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, expectedBody)
	}

	expectedContentType := contentTypeHeader
	actualContentType := resp.Header.Get("Content-Type")
	if actualContentType != expectedContentType {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, expectedContentType)
	}
}

func TestBuildInvalidJSONResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		OK(w, r, make(chan int))
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expectedStatusCode := http.StatusOK
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}

	expectedBody := ``
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, expectedBody)
	}

	expectedContentType := contentTypeHeader
	actualContentType := resp.Header.Get("Content-Type")
	if actualContentType != expectedContentType {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, expectedContentType)
	}
}
