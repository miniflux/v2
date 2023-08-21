// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package html // import "miniflux.app/v2/internal/http/response/html"

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
		OK(w, r, "Some HTML")
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expectedStatusCode := http.StatusOK
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}

	expectedBody := `Some HTML`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, expectedBody)
	}

	headers := map[string]string{
		"Content-Type":  "text/html; charset=utf-8",
		"Cache-Control": "no-cache, max-age=0, must-revalidate, no-store",
	}

	for header, expected := range headers {
		actual := resp.Header.Get(header)
		if actual != expected {
			t.Fatalf(`Unexpected header value, got %q instead of %q`, actual, expected)
		}
	}
}

func TestServerErrorResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServerError(w, r, errors.New("Some error"))
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expectedStatusCode := http.StatusInternalServerError
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}

	expectedBody := `Some error`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, expectedBody)
	}

	expectedContentType := "text/html; charset=utf-8"
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
		BadRequest(w, r, errors.New("Some error"))
	})

	handler.ServeHTTP(w, r)
	resp := w.Result()

	expectedStatusCode := http.StatusBadRequest
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}

	expectedBody := `Some error`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, expectedBody)
	}

	expectedContentType := "text/html; charset=utf-8"
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

	expectedBody := `Access Forbidden`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, expectedBody)
	}

	expectedContentType := "text/html; charset=utf-8"
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

	expectedBody := `Page Not Found`
	actualBody := w.Body.String()
	if actualBody != expectedBody {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, expectedBody)
	}

	expectedContentType := "text/html; charset=utf-8"
	actualContentType := resp.Header.Get("Content-Type")
	if actualContentType != expectedContentType {
		t.Fatalf(`Unexpected content type, got %q instead of %q`, actualContentType, expectedContentType)
	}
}

func TestRedirectResponse(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Redirect(w, r, "/path")
	})

	handler.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	expectedStatusCode := http.StatusFound
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}

	expectedResult := "/path"
	actualResult := resp.Header.Get("Location")
	if actualResult != expectedResult {
		t.Fatalf(`Unexpected redirect location, got %q instead of %q`, actualResult, expectedResult)
	}
}

func TestRequestedRangeNotSatisfiable(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RequestedRangeNotSatisfiable(w, r, "bytes */12777")
	})

	handler.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	expectedStatusCode := http.StatusRequestedRangeNotSatisfiable
	if resp.StatusCode != expectedStatusCode {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, expectedStatusCode)
	}

	expectedContentRangeHeader := "bytes */12777"
	actualContentRangeHeader := resp.Header.Get("Content-Range")
	if actualContentRangeHeader != expectedContentRangeHeader {
		t.Fatalf(`Unexpected content range header, got %q instead of %q`, actualContentRangeHeader, expectedContentRangeHeader)
	}
}
