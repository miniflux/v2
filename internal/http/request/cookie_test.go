// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package request // import "miniflux.app/v2/internal/http/request"

import (
	"net/http"
	"testing"
)

func TestGetCookieValue(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)
	r.AddCookie(&http.Cookie{Value: "cookie_value", Name: "my_cookie"})

	result := CookieValue(r, "my_cookie")
	expected := "cookie_value"

	if result != expected {
		t.Errorf(`Unexpected cookie value, got %q instead of %q`, result, expected)
	}
}

func TestGetCookieValueWhenUnset(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := CookieValue(r, "my_cookie")
	expected := ""

	if result != expected {
		t.Errorf(`Unexpected cookie value, got %q instead of %q`, result, expected)
	}
}
