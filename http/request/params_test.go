// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package request // import "miniflux.app/http/request"

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

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

func TestQueryBooleanParam(t *testing.T) {
	u, _ := url.Parse("http://example.org/?t=t")
	r := &http.Request{URL: u}

	result := QueryBooleanParam(r, "t")
	expected := true

	if result != expected {
		t.Errorf(`Unexpected result, got %t instead of %t`, result, expected)
	}

	result = QueryBooleanParam(r, "f")
	expected = false

	if result != expected {
		t.Errorf(`Unexpected result, got %t instead of %t`, result, expected)
	}
}

func TestQueryTimestampParam(t *testing.T) {
	anyTime := time.Now()
	u, _ := url.Parse(fmt.Sprintf("http://example.org/?t=%d&invalid=invalidformat", anyTime.Unix()))
	r := &http.Request{URL: u}

	result := QueryTimestampParam(r, "t")

	if result.Unix() != anyTime.Unix() {
		t.Errorf(`Unexpected result, got %v instead of %v`, result, anyTime)
	}

	result = QueryTimestampParam(r, "invalid")

	if result != nil {
		t.Errorf(`Unexpected result, got %v instead of %v`, result, nil)
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
