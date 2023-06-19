// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package request // import "miniflux.app/http/request"

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
)

func TestFormInt64Value(t *testing.T) {
	f := url.Values{}
	f.Set("integer value", "42")
	f.Set("invalid value", "invalid integer")

	r := &http.Request{Form: f}

	result := FormInt64Value(r, "integer value")
	expected := int64(42)

	if result != expected {
		t.Errorf(`Unexpected result, got %d instead of %d`, result, expected)
	}

	result = FormInt64Value(r, "invalid value")
	expected = int64(0)

	if result != expected {
		t.Errorf(`Unexpected result, got %d instead of %d`, result, expected)
	}

	result = FormInt64Value(r, "missing value")
	expected = int64(0)

	if result != expected {
		t.Errorf(`Unexpected result, got %d instead of %d`, result, expected)
	}
}

func TestRouteStringParam(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/route/{variable}/index", func(w http.ResponseWriter, r *http.Request) {
		result := RouteStringParam(r, "variable")
		expected := "value"

		if result != expected {
			t.Errorf(`Unexpected result, got %q instead of %q`, result, expected)
		}

		result = RouteStringParam(r, "missing variable")
		expected = ""

		if result != expected {
			t.Errorf(`Unexpected result, got %q instead of %q`, result, expected)
		}
	})

	r, err := http.NewRequest("GET", "/route/value/index", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
}

func TestRouteInt64Param(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/a/{variable1}/b/{variable2}/c/{variable3}", func(w http.ResponseWriter, r *http.Request) {
		result := RouteInt64Param(r, "variable1")
		expected := int64(42)

		if result != expected {
			t.Errorf(`Unexpected result, got %d instead of %d`, result, expected)
		}

		result = RouteInt64Param(r, "missing variable")
		expected = 0

		if result != expected {
			t.Errorf(`Unexpected result, got %d instead of %d`, result, expected)
		}

		result = RouteInt64Param(r, "variable2")
		expected = 0

		if result != expected {
			t.Errorf(`Unexpected result, got %d instead of %d`, result, expected)
		}

		result = RouteInt64Param(r, "variable3")
		expected = 0

		if result != expected {
			t.Errorf(`Unexpected result, got %d instead of %d`, result, expected)
		}
	})

	r, err := http.NewRequest("GET", "/a/42/b/not-int/c/-10", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
}

func TestQueryStringParam(t *testing.T) {
	u, _ := url.Parse("http://example.org/?key=value")
	r := &http.Request{URL: u}

	result := QueryStringParam(r, "key", "fallback")
	expected := "value"

	if result != expected {
		t.Errorf(`Unexpected result, got %q instead of %q`, result, expected)
	}

	result = QueryStringParam(r, "missing key", "fallback")
	expected = "fallback"

	if result != expected {
		t.Errorf(`Unexpected result, got %q instead of %q`, result, expected)
	}
}

func TestQueryIntParam(t *testing.T) {
	u, _ := url.Parse("http://example.org/?key=42&invalid=value&negative=-5")
	r := &http.Request{URL: u}

	result := QueryIntParam(r, "key", 84)
	expected := 42

	if result != expected {
		t.Errorf(`Unexpected result, got %d instead of %d`, result, expected)
	}

	result = QueryIntParam(r, "missing key", 84)
	expected = 84

	if result != expected {
		t.Errorf(`Unexpected result, got %d instead of %d`, result, expected)
	}

	result = QueryIntParam(r, "negative", 69)
	expected = 69

	if result != expected {
		t.Errorf(`Unexpected result, got %d instead of %d`, result, expected)
	}

	result = QueryIntParam(r, "invalid", 99)
	expected = 99

	if result != expected {
		t.Errorf(`Unexpected result, got %d instead of %d`, result, expected)
	}
}

func TestQueryInt64Param(t *testing.T) {
	u, _ := url.Parse("http://example.org/?key=42&invalid=value&negative=-5")
	r := &http.Request{URL: u}

	result := QueryInt64Param(r, "key", int64(84))
	expected := int64(42)

	if result != expected {
		t.Errorf(`Unexpected result, got %d instead of %d`, result, expected)
	}

	result = QueryInt64Param(r, "missing key", int64(84))
	expected = int64(84)

	if result != expected {
		t.Errorf(`Unexpected result, got %d instead of %d`, result, expected)
	}

	result = QueryInt64Param(r, "invalid", int64(69))
	expected = int64(69)

	if result != expected {
		t.Errorf(`Unexpected result, got %d instead of %d`, result, expected)
	}

	result = QueryInt64Param(r, "invalid", int64(99))
	expected = int64(99)

	if result != expected {
		t.Errorf(`Unexpected result, got %d instead of %d`, result, expected)
	}
}

func TestHasQueryParam(t *testing.T) {
	u, _ := url.Parse("http://example.org/?key=42")
	r := &http.Request{URL: u}

	result := HasQueryParam(r, "key")
	expected := true

	if result != expected {
		t.Errorf(`Unexpected result, got %v instead of %v`, result, expected)
	}

	result = HasQueryParam(r, "missing key")
	expected = false

	if result != expected {
		t.Errorf(`Unexpected result, got %v instead of %v`, result, expected)
	}
}
