// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginResponseString(t *testing.T) {
	result := loginResponse{SID: "user/sid", LSID: "user/lsid", Auth: "user/auth"}.String()
	expected := "SID=user/sid\nLSID=user/lsid\nAuth=user/auth\n"

	if result != expected {
		t.Errorf(`Unexpected login response, got %q instead of %q`, result, expected)
	}
}

func TestSendUnauthorizedResponse(t *testing.T) {
	recorder := httptest.NewRecorder()
	sendUnauthorizedResponse(recorder, httptest.NewRequest(http.MethodGet, "/reader/api/0/user-info", nil))

	checkUnauthorizedResponse(t, recorder)
}

func checkUnauthorizedResponse(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf(`Unexpected status code, got %d instead of %d`, recorder.Code, http.StatusUnauthorized)
	}
	if value := recorder.Header().Get("X-Reader-Google-Bad-Token"); value != "true" {
		t.Errorf(`Unexpected X-Reader-Google-Bad-Token header, got %q`, value)
	}
	if value := recorder.Header().Get("Content-Type"); value != "text/plain; charset=utf-8" {
		t.Errorf(`Unexpected Content-Type header, got %q`, value)
	}
	if body := recorder.Body.String(); body != "Unauthorized" {
		t.Errorf(`Unexpected body, got %q`, body)
	}
}
