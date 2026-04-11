// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

func TestNewWebSession(t *testing.T) {
	const userAgent = "test-agent"
	const ip = "127.0.0.1"

	session, secret := NewWebSession(userAgent, ip)

	if session == nil {
		t.Fatal("NewWebSession returned a nil session")
	}
	if secret == "" {
		t.Error("NewWebSession returned an empty secret")
	}
	if session.ID == "" {
		t.Error("NewWebSession produced an empty ID")
	}
	if session.ID == secret {
		t.Error("session ID and secret must not be equal")
	}
	if len(session.SecretHash) == 0 {
		t.Error("NewWebSession produced an empty SecretHash")
	}
	if session.CSRF() == "" {
		t.Error("NewWebSession produced an empty CSRF token")
	}
	if session.UserAgent != userAgent {
		t.Errorf("UserAgent = %q, want %q", session.UserAgent, userAgent)
	}
	if session.IP != ip {
		t.Errorf("IP = %q, want %q", session.IP, ip)
	}
	if session.IsAuthenticated() {
		t.Error("a fresh session must not be authenticated")
	}
	if session.IsDirty() {
		t.Error("a fresh session must not be dirty")
	}
	if !session.VerifySecret(secret) {
		t.Error("VerifySecret rejected the secret returned by NewWebSession")
	}
}

func TestNewWebSession_ProducesUniqueIdentities(t *testing.T) {
	s1, secret1 := NewWebSession("", "")
	s2, secret2 := NewWebSession("", "")

	if s1.ID == s2.ID {
		t.Error("successive NewWebSession calls produced the same ID")
	}
	if secret1 == secret2 {
		t.Error("successive NewWebSession calls produced the same secret")
	}
	if bytes.Equal(s1.SecretHash, s2.SecretHash) {
		t.Error("successive NewWebSession calls produced the same SecretHash")
	}
	if s1.CSRF() == s2.CSRF() {
		t.Error("successive NewWebSession calls produced the same CSRF token")
	}
}

func TestWebSession_Rotate(t *testing.T) {
	session, originalSecret := NewWebSession("agent", "ip")
	originalID := session.ID
	originalHash := bytes.Clone(session.SecretHash)
	originalCSRF := session.CSRF()

	// Bind a user so we can verify Rotate preserves the user binding.
	session.SetUser(&User{ID: 42})

	oldID, newSecret := session.Rotate()

	if oldID != originalID {
		t.Errorf("Rotate returned oldID = %q, want %q", oldID, originalID)
	}
	if newSecret == "" {
		t.Error("Rotate returned an empty new secret")
	}
	if newSecret == originalSecret {
		t.Error("Rotate returned the same secret as before")
	}
	if session.ID == originalID {
		t.Error("Rotate did not change the session ID")
	}
	if bytes.Equal(session.SecretHash, originalHash) {
		t.Error("Rotate did not change the SecretHash")
	}
	if session.VerifySecret(originalSecret) {
		t.Error("VerifySecret must reject the pre-rotation secret")
	}
	if !session.VerifySecret(newSecret) {
		t.Error("VerifySecret must accept the post-rotation secret")
	}
	if session.CSRF() != originalCSRF {
		t.Error("Rotate must preserve the CSRF token so in-flight forms remain valid")
	}
	if !session.IsAuthenticated() {
		t.Error("Rotate must preserve the user binding")
	}
	if id, _ := session.UserID(); id != 42 {
		t.Errorf("Rotate corrupted user ID: got %d, want 42", id)
	}
}

func TestWebSession_VerifySecret(t *testing.T) {
	good, goodSecret := NewWebSession("", "")

	testCases := []struct {
		name   string
		hash   []byte
		secret string
		want   bool
	}{
		{"correct secret", good.SecretHash, goodSecret, true},
		{"wrong secret", good.SecretHash, "not-the-right-secret", false},
		{"empty secret", good.SecretHash, "", false},
		{"nil hash", nil, goodSecret, false},
		{"empty hash and secret", nil, "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := &WebSession{SecretHash: tc.hash}
			if got := s.VerifySecret(tc.secret); got != tc.want {
				t.Errorf("VerifySecret(%q) = %v, want %v", tc.secret, got, tc.want)
			}
		})
	}
}

func TestWebSession_UserBindingLifecycle(t *testing.T) {
	session, _ := NewWebSession("", "")

	if session.IsAuthenticated() {
		t.Error("a fresh session must not be authenticated")
	}
	if id, ok := session.UserID(); ok || id != 0 {
		t.Errorf("UserID() = (%d, %v), want (0, false)", id, ok)
	}

	user := &User{ID: 99, Language: "fr_FR", Theme: "dark_serif"}
	session.SetUser(user)

	if !session.IsAuthenticated() {
		t.Error("session must be authenticated after SetUser")
	}
	if id, ok := session.UserID(); !ok || id != 99 {
		t.Errorf("UserID() = (%d, %v), want (99, true)", id, ok)
	}
	if session.Language() != "fr_FR" {
		t.Errorf("SetUser did not copy Language: got %q, want %q", session.Language(), "fr_FR")
	}
	if session.Theme() != "dark_serif" {
		t.Errorf("SetUser did not copy Theme: got %q, want %q", session.Theme(), "dark_serif")
	}
	if !session.IsDirty() {
		t.Error("SetUser must mark the session dirty")
	}

	session.ClearUser()
	if session.IsAuthenticated() {
		t.Error("session must not be authenticated after ClearUser")
	}
	if id, ok := session.UserID(); ok || id != 0 {
		t.Errorf("UserID() after ClearUser = (%d, %v), want (0, false)", id, ok)
	}
}

func TestWebSession_SetUser_NilIsNoop(t *testing.T) {
	session, _ := NewWebSession("", "")
	session.SetUser(nil)

	if session.IsAuthenticated() {
		t.Error("SetUser(nil) must not authenticate the session")
	}
	if session.IsDirty() {
		t.Error("SetUser(nil) must not mark the session dirty")
	}
}

func TestWebSession_UserIDStorageRoundTrip(t *testing.T) {
	testCases := []struct {
		name string
		in   sql.NullInt64
	}{
		{"null", sql.NullInt64{}},
		{"zero valid", sql.NullInt64{Int64: 0, Valid: true}},
		{"positive valid", sql.NullInt64{Int64: 42, Valid: true}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			session := &WebSession{}
			session.ScanUserID(tc.in)

			if got := session.NullUserID(); got != tc.in {
				t.Errorf("round-trip = %+v, want %+v", got, tc.in)
			}
			if got := session.IsAuthenticated(); got != tc.in.Valid {
				t.Errorf("IsAuthenticated() = %v, want %v", got, tc.in.Valid)
			}
		})
	}
}

func TestWebSession_ScanUserID_ClearsPreviousValue(t *testing.T) {
	session := &WebSession{}
	session.ScanUserID(sql.NullInt64{Int64: 1, Valid: true})
	session.ScanUserID(sql.NullInt64{})

	if session.IsAuthenticated() {
		t.Error("ScanUserID with an invalid value must clear the user binding")
	}
}

func TestWebSession_LanguageAndThemeDefaults(t *testing.T) {
	session := &WebSession{}

	if got := session.Language(); got != defaultSessionLanguage {
		t.Errorf("default Language() = %q, want %q", got, defaultSessionLanguage)
	}
	if got := session.Theme(); got != defaultSessionTheme {
		t.Errorf("default Theme() = %q, want %q", got, defaultSessionTheme)
	}

	session.SetLanguage("de_DE")
	session.SetTheme("light_sans_serif")

	if got := session.Language(); got != "de_DE" {
		t.Errorf("Language() = %q, want %q", got, "de_DE")
	}
	if got := session.Theme(); got != "light_sans_serif" {
		t.Errorf("Theme() = %q, want %q", got, "light_sans_serif")
	}
	if !session.IsDirty() {
		t.Error("SetLanguage/SetTheme must mark the session dirty")
	}
}

func TestWebSession_OAuth2FlowLifecycle(t *testing.T) {
	session := &WebSession{}

	if session.OAuth2State() != "" {
		t.Error("OAuth2State() must be empty by default")
	}
	if session.OAuth2CodeVerifier() != "" {
		t.Error("OAuth2CodeVerifier() must be empty by default")
	}

	session.StartOAuth2Flow("state-token", "code-verifier")

	if got := session.OAuth2State(); got != "state-token" {
		t.Errorf("OAuth2State() = %q, want %q", got, "state-token")
	}
	if got := session.OAuth2CodeVerifier(); got != "code-verifier" {
		t.Errorf("OAuth2CodeVerifier() = %q, want %q", got, "code-verifier")
	}
	if !session.IsDirty() {
		t.Error("StartOAuth2Flow must mark the session dirty")
	}

	session.ClearOAuth2Flow()

	if session.OAuth2State() != "" {
		t.Errorf("OAuth2State() after Clear = %q, want empty", session.OAuth2State())
	}
	if session.OAuth2CodeVerifier() != "" {
		t.Errorf("OAuth2CodeVerifier() after Clear = %q, want empty", session.OAuth2CodeVerifier())
	}
}

func TestWebSession_ConsumeMessages(t *testing.T) {
	t.Run("no messages", func(t *testing.T) {
		session := &WebSession{}

		success, errMsg := session.ConsumeMessages()
		if success != "" || errMsg != "" {
			t.Errorf("ConsumeMessages() = (%q, %q), want empty", success, errMsg)
		}
		if session.IsDirty() {
			t.Error("ConsumeMessages with no messages must not mark the session dirty")
		}
	})

	t.Run("returns and clears", func(t *testing.T) {
		session := &WebSession{}
		session.SetSuccessMessage("saved")
		session.SetErrorMessage("nope")
		session.dirty = false // isolate the dirty contribution of ConsumeMessages

		success, errMsg := session.ConsumeMessages()
		if success != "saved" || errMsg != "nope" {
			t.Errorf("ConsumeMessages() = (%q, %q), want (%q, %q)", success, errMsg, "saved", "nope")
		}
		if !session.IsDirty() {
			t.Error("ConsumeMessages with messages must mark the session dirty")
		}

		success, errMsg = session.ConsumeMessages()
		if success != "" || errMsg != "" {
			t.Errorf("second ConsumeMessages() = (%q, %q), want empty", success, errMsg)
		}
	})
}

func TestWebSession_ConsumeWebAuthnSession(t *testing.T) {
	t.Run("no data", func(t *testing.T) {
		session := &WebSession{}

		if got := session.ConsumeWebAuthnSession(); got != nil {
			t.Errorf("ConsumeWebAuthnSession() = %v, want nil", got)
		}
		if session.IsDirty() {
			t.Error("ConsumeWebAuthnSession with no data must not mark the session dirty")
		}
	})

	t.Run("returns and clears", func(t *testing.T) {
		data := &webauthn.SessionData{}
		session := &WebSession{}
		session.SetWebAuthn(data)
		session.dirty = false // isolate the dirty contribution of ConsumeWebAuthnSession

		if got := session.ConsumeWebAuthnSession(); got != data {
			t.Errorf("ConsumeWebAuthnSession() = %p, want %p", got, data)
		}
		if !session.IsDirty() {
			t.Error("ConsumeWebAuthnSession with data must mark the session dirty")
		}
		if got := session.ConsumeWebAuthnSession(); got != nil {
			t.Errorf("second ConsumeWebAuthnSession() = %v, want nil", got)
		}
	})
}

func TestWebSession_MarkForceRefreshed(t *testing.T) {
	session := &WebSession{}

	if got := session.LastForceRefresh(); !got.IsZero() {
		t.Errorf("default LastForceRefresh() = %v, want zero time", got)
	}

	before := time.Now().UTC()
	session.MarkForceRefreshed()
	after := time.Now().UTC()

	got := session.LastForceRefresh()
	if got.Before(before) || got.After(after) {
		t.Errorf("LastForceRefresh() = %v, want between %v and %v", got, before, after)
	}
	if !session.IsDirty() {
		t.Error("MarkForceRefreshed must mark the session dirty")
	}
}

func TestWebSession_StateRoundTrip(t *testing.T) {
	original := &WebSession{}
	original.SetLanguage("de_DE")
	original.SetTheme("light_sans_serif")
	original.SetSuccessMessage("saved")
	original.SetErrorMessage("oops")
	original.StartOAuth2Flow("state-token", "code-verifier")
	original.MarkForceRefreshed()
	originalRefreshAt := original.LastForceRefresh()

	data, err := original.MarshalState()
	if err != nil {
		t.Fatalf("MarshalState() error: %v", err)
	}
	if !json.Valid(data) {
		t.Errorf("MarshalState() produced invalid JSON: %s", data)
	}

	restored := &WebSession{}
	if err := restored.UnmarshalState(data); err != nil {
		t.Fatalf("UnmarshalState() error: %v", err)
	}

	if got := restored.Language(); got != "de_DE" {
		t.Errorf("Language() = %q, want %q", got, "de_DE")
	}
	if got := restored.Theme(); got != "light_sans_serif" {
		t.Errorf("Theme() = %q, want %q", got, "light_sans_serif")
	}
	if got := restored.OAuth2State(); got != "state-token" {
		t.Errorf("OAuth2State() = %q, want %q", got, "state-token")
	}
	if got := restored.OAuth2CodeVerifier(); got != "code-verifier" {
		t.Errorf("OAuth2CodeVerifier() = %q, want %q", got, "code-verifier")
	}
	if got := restored.LastForceRefresh(); !got.Equal(originalRefreshAt) {
		t.Errorf("LastForceRefresh() = %v, want %v", got, originalRefreshAt)
	}

	success, errMsg := restored.ConsumeMessages()
	if success != "saved" || errMsg != "oops" {
		t.Errorf("ConsumeMessages() = (%q, %q), want (%q, %q)", success, errMsg, "saved", "oops")
	}
}

func TestWebSession_UnmarshalState_EmptyDataResetsState(t *testing.T) {
	session := &WebSession{}
	session.SetLanguage("fr_FR")
	session.StartOAuth2Flow("s", "v")

	if err := session.UnmarshalState(nil); err != nil {
		t.Fatalf("UnmarshalState(nil) error: %v", err)
	}

	if got := session.Language(); got != defaultSessionLanguage {
		t.Errorf("UnmarshalState(nil) did not reset Language: got %q", got)
	}
	if session.OAuth2State() != "" {
		t.Error("UnmarshalState(nil) did not reset OAuth2 state")
	}
}
