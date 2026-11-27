// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package response // import "miniflux.app/v2/internal/http/response"

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf(`Unexpected status code, got %d instead of %d`, resp.StatusCode, http.StatusNoContent)
	}

	if actualBody := w.Body.String(); actualBody != `` {
		t.Fatalf(`Unexpected body, got %s instead of %s`, actualBody, ``)
	}

	if actualContentType := resp.Header.Get("Content-Type"); actualContentType != "" {
		t.Fatalf(`Unexpected content type, got %q instead of empty string`, actualContentType)
	}
}

func TestErrorResponsesDoNotLogQueryString(t *testing.T) {
	const secret = "Abcdef123456"

	helpers := map[string]func(http.ResponseWriter, *http.Request){
		"HTMLServerError": func(writer http.ResponseWriter, request *http.Request) {
			HTMLServerError(writer, request, errors.New("some error"))
		},
		"HTMLBadRequest": func(writer http.ResponseWriter, request *http.Request) {
			HTMLBadRequest(writer, request, errors.New("some error"))
		},
		"HTMLForbidden":     HTMLForbidden,
		"HTMLNotFound":      HTMLNotFound,
		"HTMLNotAcceptable": HTMLNotAcceptable,
		"HTMLRequestedRangeNotSatisfiable": func(writer http.ResponseWriter, request *http.Request) {
			HTMLRequestedRangeNotSatisfiable(writer, request, "bytes */0")
		},
		"JSONServerError": func(writer http.ResponseWriter, request *http.Request) {
			JSONServerError(writer, request, errors.New("some error"))
		},
		"JSONBadRequest": func(writer http.ResponseWriter, request *http.Request) {
			JSONBadRequest(writer, request, errors.New("some error"))
		},
		"JSONUnauthorized": JSONUnauthorized,
		"JSONForbidden":    JSONForbidden,
		"JSONNotFound":     JSONNotFound,
	}

	for name, helper := range helpers {
		t.Run(name, func(t *testing.T) {
			var logs bytes.Buffer
			previousLogger := slog.Default()
			slog.SetDefault(slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
			t.Cleanup(func() { slog.SetDefault(previousLogger) })

			request := httptest.NewRequest(http.MethodPost, "/accounts/ClientLogin?Email=alice&Passwd="+secret, nil)
			helper(httptest.NewRecorder(), request)

			if !strings.Contains(logs.String(), "/accounts/ClientLogin") {
				t.Fatalf("Expected the request path to be logged, got:\n%s", logs.String())
			}
			if strings.Contains(logs.String(), secret) {
				t.Errorf("Logs contain the query string:\n%s", logs.String())
			}
		})
	}
}
