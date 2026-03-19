// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"miniflux.app/v2/internal/version"
)

func TestNewHandlerHandlesOptionsRequests(t *testing.T) {
	handler := NewHandler(nil, nil)

	r := httptest.NewRequest(http.MethodOptions, "/v1/users", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if got := w.Code; got != http.StatusNoContent {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, got, http.StatusNoContent)
	}

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf(`Unexpected Access-Control-Allow-Origin header, got %q`, got)
	}

	if got := w.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Fatalf(`Unexpected Access-Control-Allow-Methods header, got %q`, got)
	}

	if got := w.Header().Get("Access-Control-Allow-Headers"); got != "X-Auth-Token, Authorization, Content-Type, Accept" {
		t.Fatalf(`Unexpected Access-Control-Allow-Headers header, got %q`, got)
	}

	if got := w.Header().Get("Access-Control-Max-Age"); got != "3600" {
		t.Fatalf(`Unexpected Access-Control-Max-Age header, got %q`, got)
	}
}

func TestVersionHandler(t *testing.T) {
	h := &handler{}
	r := httptest.NewRequest(http.MethodGet, "/v1/version", nil)
	w := httptest.NewRecorder()

	h.versionHandler(w, r)

	if got := w.Code; got != http.StatusOK {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, got, http.StatusOK)
	}

	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf(`Unexpected Content-Type header, got %q`, got)
	}

	var responseBody versionResponse
	if err := json.NewDecoder(w.Body).Decode(&responseBody); err != nil {
		t.Fatalf("Unexpected JSON decoding error: %v", err)
	}

	if responseBody.Version != version.Version {
		t.Fatalf(`Unexpected version, got %q instead of %q`, responseBody.Version, version.Version)
	}

	if responseBody.Commit != version.Commit {
		t.Fatalf(`Unexpected commit, got %q instead of %q`, responseBody.Commit, version.Commit)
	}

	if responseBody.BuildDate != version.BuildDate {
		t.Fatalf(`Unexpected build date, got %q instead of %q`, responseBody.BuildDate, version.BuildDate)
	}

	if responseBody.GoVersion != runtime.Version() {
		t.Fatalf(`Unexpected Go version, got %q instead of %q`, responseBody.GoVersion, runtime.Version())
	}

	if responseBody.Compiler != runtime.Compiler {
		t.Fatalf(`Unexpected compiler, got %q instead of %q`, responseBody.Compiler, runtime.Compiler)
	}

	if responseBody.Arch != runtime.GOARCH {
		t.Fatalf(`Unexpected architecture, got %q instead of %q`, responseBody.Arch, runtime.GOARCH)
	}

	if responseBody.OS != runtime.GOOS {
		t.Fatalf(`Unexpected OS, got %q instead of %q`, responseBody.OS, runtime.GOOS)
	}
}

func TestNewHandlerSupportsBasePathStripping(t *testing.T) {
	scenarios := []struct {
		name   string
		prefix string
		path   string
	}{
		{name: "empty base path", prefix: "", path: "/v1/users"},
		{name: "non empty base path", prefix: "/base", path: "/base/v1/users"},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			handler := http.StripPrefix(scenario.prefix, NewHandler(nil, nil))

			r := httptest.NewRequest(http.MethodOptions, scenario.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, r)

			if got := w.Code; got != http.StatusNoContent {
				t.Fatalf(`Unexpected status code, got %d instead of %d`, got, http.StatusNoContent)
			}
		})
	}
}
