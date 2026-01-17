// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package request // import "miniflux.app/v2/internal/http/request"

import (
	"context"
	"net/http"
	"testing"
	"time"

	"miniflux.app/v2/internal/model"
)

func TestContextStringValue(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)
	ctx := r.Context()
	ctx = context.WithValue(ctx, ClientIPContextKey, "IP")
	r = r.WithContext(ctx)

	result := getContextStringValue(r, ClientIPContextKey)
	expected := "IP"

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}
}

func TestContextStringValueWithInvalidType(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)
	ctx := r.Context()
	ctx = context.WithValue(ctx, ClientIPContextKey, 0)
	r = r.WithContext(ctx)

	result := getContextStringValue(r, ClientIPContextKey)
	expected := ""

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}
}

func TestContextStringValueWhenUnset(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := getContextStringValue(r, ClientIPContextKey)
	expected := ""

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}
}

func TestContextBoolValue(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)
	ctx := r.Context()
	ctx = context.WithValue(ctx, IsAdminUserContextKey, true)
	r = r.WithContext(ctx)

	result := getContextBoolValue(r, IsAdminUserContextKey)
	expected := true

	if result != expected {
		t.Errorf(`Unexpected context value, got %v instead of %v`, result, expected)
	}
}

func TestContextBoolValueWithInvalidType(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)
	ctx := r.Context()
	ctx = context.WithValue(ctx, IsAdminUserContextKey, "invalid")
	r = r.WithContext(ctx)

	result := getContextBoolValue(r, IsAdminUserContextKey)
	expected := false

	if result != expected {
		t.Errorf(`Unexpected context value, got %v instead of %v`, result, expected)
	}
}

func TestContextBoolValueWhenUnset(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := getContextBoolValue(r, IsAdminUserContextKey)
	expected := false

	if result != expected {
		t.Errorf(`Unexpected context value, got %v instead of %v`, result, expected)
	}
}

func TestContextInt64Value(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)
	ctx := r.Context()
	ctx = context.WithValue(ctx, UserIDContextKey, int64(1234))
	r = r.WithContext(ctx)

	result := getContextInt64Value(r, UserIDContextKey)
	expected := int64(1234)

	if result != expected {
		t.Errorf(`Unexpected context value, got %d instead of %d`, result, expected)
	}
}

func TestContextInt64ValueWithInvalidType(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)
	ctx := r.Context()
	ctx = context.WithValue(ctx, UserIDContextKey, "invalid")
	r = r.WithContext(ctx)

	result := getContextInt64Value(r, UserIDContextKey)
	expected := int64(0)

	if result != expected {
		t.Errorf(`Unexpected context value, got %d instead of %d`, result, expected)
	}
}

func TestContextInt64ValueWhenUnset(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := getContextInt64Value(r, UserIDContextKey)
	expected := int64(0)

	if result != expected {
		t.Errorf(`Unexpected context value, got %d instead of %d`, result, expected)
	}
}

func TestIsAdmin(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := IsAdminUser(r)
	expected := false

	if result != expected {
		t.Errorf(`Unexpected context value, got %v instead of %v`, result, expected)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, IsAdminUserContextKey, true)
	r = r.WithContext(ctx)

	result = IsAdminUser(r)
	expected = true

	if result != expected {
		t.Errorf(`Unexpected context value, got %v instead of %v`, result, expected)
	}
}

func TestIsAuthenticated(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := IsAuthenticated(r)
	expected := false

	if result != expected {
		t.Errorf(`Unexpected context value, got %v instead of %v`, result, expected)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, IsAuthenticatedContextKey, true)
	r = r.WithContext(ctx)

	result = IsAuthenticated(r)
	expected = true

	if result != expected {
		t.Errorf(`Unexpected context value, got %v instead of %v`, result, expected)
	}
}

func TestUserID(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := UserID(r)
	expected := int64(0)

	if result != expected {
		t.Errorf(`Unexpected context value, got %v instead of %v`, result, expected)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, UserIDContextKey, int64(123))
	r = r.WithContext(ctx)

	result = UserID(r)
	expected = int64(123)

	if result != expected {
		t.Errorf(`Unexpected context value, got %v instead of %v`, result, expected)
	}
}

func TestUserName(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := UserName(r)
	expected := "unknown"

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, UserNameContextKey, "jane")
	r = r.WithContext(ctx)

	result = UserName(r)
	expected = "jane"

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}
}

func TestUserTimezone(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := UserTimezone(r)
	expected := "UTC"

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, UserTimezoneContextKey, "Europe/Paris")
	r = r.WithContext(ctx)

	result = UserTimezone(r)
	expected = "Europe/Paris"

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}
}

func TestUserLanguage(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := UserLanguage(r)
	expected := "en_US"

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, UserLanguageContextKey, "fr_FR")
	r = r.WithContext(ctx)

	result = UserLanguage(r)
	expected = "fr_FR"

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}
}

func TestUserTheme(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := UserTheme(r)
	expected := "system_serif"

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, UserThemeContextKey, "dark_serif")
	r = r.WithContext(ctx)

	result = UserTheme(r)
	expected = "dark_serif"

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}
}

func TestCSRF(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := CSRF(r)
	expected := ""

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, CSRFContextKey, "secret")
	r = r.WithContext(ctx)

	result = CSRF(r)
	expected = "secret"

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}
}

func TestSessionID(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := SessionID(r)
	expected := ""

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, SessionIDContextKey, "id")
	r = r.WithContext(ctx)

	result = SessionID(r)
	expected = "id"

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}
}

func TestUserSessionToken(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := UserSessionToken(r)
	expected := ""

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, UserSessionTokenContextKey, "token")
	r = r.WithContext(ctx)

	result = UserSessionToken(r)
	expected = "token"

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}
}

func TestOAuth2State(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := OAuth2State(r)
	expected := ""

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, OAuth2StateContextKey, "state")
	r = r.WithContext(ctx)

	result = OAuth2State(r)
	expected = "state"

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}
}

func TestOAuth2CodeVerifier(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := OAuth2CodeVerifier(r)
	expected := ""

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, OAuth2CodeVerifierContextKey, "verifier")
	r = r.WithContext(ctx)

	result = OAuth2CodeVerifier(r)
	expected = "verifier"

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}
}

func TestFlashMessage(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := FlashMessage(r)
	expected := ""

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, FlashMessageContextKey, "message")
	r = r.WithContext(ctx)

	result = FlashMessage(r)
	expected = "message"

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}
}

func TestFlashErrorMessage(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := FlashErrorMessage(r)
	expected := ""

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, FlashErrorMessageContextKey, "error message")
	r = r.WithContext(ctx)

	result = FlashErrorMessage(r)
	expected = "error message"

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}
}

func TestLastForceRefresh(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := LastForceRefresh(r)
	expected := time.Time{}

	if !result.Equal(expected) {
		t.Errorf(`Unexpected context value, got %v instead of %v`, result, expected)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, LastForceRefreshContextKey, "not-a-timestamp")
	r = r.WithContext(ctx)

	result = LastForceRefresh(r)
	expected = time.Time{}

	if !result.Equal(expected) {
		t.Errorf(`Unexpected context value, got %v instead of %v`, result, expected)
	}

	ctx = r.Context()
	ctx = context.WithValue(ctx, LastForceRefreshContextKey, "1700000000")
	r = r.WithContext(ctx)

	result = LastForceRefresh(r)
	expected = time.Unix(1700000000, 0)

	if !result.Equal(expected) {
		t.Errorf(`Unexpected context value, got %v instead of %v`, result, expected)
	}
}

func TestWebAuthnSessionData(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := WebAuthnSessionData(r)
	if result != nil {
		t.Errorf("Unexpected context value, got %v instead of nil", result)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, WebAuthnDataContextKey, "invalid")
	r = r.WithContext(ctx)

	result = WebAuthnSessionData(r)
	if result != nil {
		t.Errorf("Unexpected context value, got %v instead of nil", result)
	}

	session := model.WebAuthnSession{}
	ctx = r.Context()
	ctx = context.WithValue(ctx, WebAuthnDataContextKey, session)
	r = r.WithContext(ctx)

	result = WebAuthnSessionData(r)
	if result == nil {
		t.Errorf("Unexpected context value, got nil instead of session")
	}
}

func TestClientIP(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := ClientIP(r)
	expected := ""

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, ClientIPContextKey, "127.0.0.1")
	r = r.WithContext(ctx)

	result = ClientIP(r)
	expected = "127.0.0.1"

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}
}

func TestGoogleReaderToken(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.org", nil)

	result := GoogleReaderToken(r)
	expected := ""

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, GoogleReaderTokenKey, "token")
	r = r.WithContext(ctx)

	result = GoogleReaderToken(r)
	expected = "token"

	if result != expected {
		t.Errorf(`Unexpected context value, got %q instead of %q`, result, expected)
	}
}
