// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"

	"miniflux.app/v2/internal/timezone"
)

const (
	defaultSessionLanguage = "en_US"
	defaultSessionTheme    = "system_serif"
)

// WebSession represents a browser session persisted in the web_sessions table.
type WebSession struct {
	ID         string
	SecretHash []byte
	CreatedAt  time.Time
	UserAgent  string
	IP         string
	userID     *int64
	state      webSessionState
	dirty      bool
}

// webSessionState stores transient browser session state as a JSON blob.
type webSessionState struct {
	CSRF               string                `json:"csrf,omitempty"`
	SuccessMessage     string                `json:"success_message,omitempty"`
	ErrorMessage       string                `json:"error_message,omitempty"`
	OAuth2             *WebSessionOAuth2     `json:"oauth2,omitempty"`
	WebAuthn           *webauthn.SessionData `json:"webauthn,omitempty"`
	LastForceRefreshAt *time.Time            `json:"last_force_refresh_at,omitempty"`
	Language           string                `json:"language,omitempty"`
	Theme              string                `json:"theme,omitempty"`
}

// WebSessionOAuth2 stores transient OAuth2 flow state.
type WebSessionOAuth2 struct {
	State        string `json:"state,omitempty"`
	CodeVerifier string `json:"code_verifier,omitempty"`
}

// NewWebSession builds an unauthenticated browser session with a fresh
// identity and returns it along with the raw session secret.
func NewWebSession(userAgent, ip string) (*WebSession, string) {
	secret := rand.Text()
	session := &WebSession{
		ID:         rand.Text(),
		SecretHash: hashWebSessionSecret(secret),
		UserAgent:  userAgent,
		IP:         ip,
	}
	session.state.CSRF = rand.Text()
	return session, secret
}

// Rotate assigns a new ID and secret in place, returning the previous ID
// and the new raw secret. Rotating on authentication prevents session fixation.
func (s *WebSession) Rotate() (oldID, newSecret string) {
	oldID = s.ID
	newSecret = rand.Text()
	s.ID = rand.Text()
	s.SecretHash = hashWebSessionSecret(newSecret)
	return oldID, newSecret
}

// VerifySecret reports whether the given raw secret matches the stored hash.
func (s *WebSession) VerifySecret(secret string) bool {
	if secret == "" || len(s.SecretHash) == 0 {
		return false
	}
	actual := hashWebSessionSecret(secret)
	return subtle.ConstantTimeCompare(actual, s.SecretHash) == 1
}

func hashWebSessionSecret(secret string) []byte {
	sum := sha256.Sum256([]byte(secret))
	return sum[:]
}

// IsDirty reports whether the session has been modified since it was loaded.
func (s *WebSession) IsDirty() bool {
	return s.dirty
}

// IsAuthenticated reports whether the session is bound to a user.
func (s *WebSession) IsAuthenticated() bool {
	return s.userID != nil
}

// UserID returns the authenticated user ID and whether the session is bound to a user.
func (s *WebSession) UserID() (int64, bool) {
	if s.userID == nil {
		return 0, false
	}
	return *s.userID, true
}

// NullUserID returns the session user ID as a sql.NullInt64 for storage writes.
func (s *WebSession) NullUserID() sql.NullInt64 {
	if s.userID == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *s.userID, Valid: true}
}

// ScanUserID sets the session user ID from a sql.NullInt64 loaded from storage.
func (s *WebSession) ScanUserID(v sql.NullInt64) {
	if !v.Valid {
		s.userID = nil
		return
	}
	id := v.Int64
	s.userID = &id
}

// UseTimezone converts creation date to the given timezone.
func (s *WebSession) UseTimezone(tz string) {
	s.CreatedAt = timezone.Convert(tz, s.CreatedAt)
}

// CSRF returns the CSRF token for this session.
func (s *WebSession) CSRF() string {
	return s.state.CSRF
}

// Language returns the session language, or a default when unset.
func (s *WebSession) Language() string {
	if s.state.Language != "" {
		return s.state.Language
	}
	return defaultSessionLanguage
}

// Theme returns the session theme, or a default when unset.
func (s *WebSession) Theme() string {
	if s.state.Theme != "" {
		return s.state.Theme
	}
	return defaultSessionTheme
}

// OAuth2State returns the OAuth2 state parameter, or empty if not in an OAuth2 flow.
func (s *WebSession) OAuth2State() string {
	if s.state.OAuth2 != nil {
		return s.state.OAuth2.State
	}
	return ""
}

// OAuth2CodeVerifier returns the PKCE code verifier, or empty if not in an OAuth2 flow.
func (s *WebSession) OAuth2CodeVerifier() string {
	if s.state.OAuth2 != nil {
		return s.state.OAuth2.CodeVerifier
	}
	return ""
}

// ConsumeWebAuthnSession returns and clears the pending WebAuthn session data.
func (s *WebSession) ConsumeWebAuthnSession() *webauthn.SessionData {
	data := s.state.WebAuthn
	if data == nil {
		return nil
	}

	s.dirty = true
	s.state.WebAuthn = nil
	return data
}

// LastForceRefresh returns the last force refresh timestamp, or zero time if unset.
func (s *WebSession) LastForceRefresh() time.Time {
	if s.state.LastForceRefreshAt != nil {
		return *s.state.LastForceRefreshAt
	}
	return time.Time{}
}

// ConsumeMessages returns and clears the success and error messages.
func (s *WebSession) ConsumeMessages() (string, string) {
	successMessage := s.state.SuccessMessage
	errorMessage := s.state.ErrorMessage

	if successMessage != "" || errorMessage != "" {
		s.dirty = true
		s.state.SuccessMessage = ""
		s.state.ErrorMessage = ""
	}

	return successMessage, errorMessage
}

// SetLanguage updates the language.
func (s *WebSession) SetLanguage(language string) {
	s.dirty = true
	s.state.Language = language
}

// SetTheme updates the theme.
func (s *WebSession) SetTheme(theme string) {
	s.dirty = true
	s.state.Theme = theme
}

// SetSuccessMessage stores a success message shown on the next page load.
func (s *WebSession) SetSuccessMessage(message string) {
	s.dirty = true
	s.state.SuccessMessage = message
}

// SetErrorMessage stores an error message shown on the next page load.
func (s *WebSession) SetErrorMessage(message string) {
	s.dirty = true
	s.state.ErrorMessage = message
}

// StartOAuth2Flow stores the OAuth2 state parameter and PKCE code verifier.
func (s *WebSession) StartOAuth2Flow(state, codeVerifier string) {
	s.dirty = true
	s.state.OAuth2 = &WebSessionOAuth2{
		State:        state,
		CodeVerifier: codeVerifier,
	}
}

// ClearOAuth2Flow discards any pending OAuth2 flow state.
func (s *WebSession) ClearOAuth2Flow() {
	s.dirty = true
	s.state.OAuth2 = nil
}

// SetUser binds the session to an authenticated user and copies their preferences.
func (s *WebSession) SetUser(user *User) {
	if user == nil {
		return
	}

	s.dirty = true
	userID := user.ID
	s.userID = &userID
	s.state.Language = user.Language
	s.state.Theme = user.Theme
}

// ClearUser removes the user binding from the session.
func (s *WebSession) ClearUser() {
	s.dirty = true
	s.userID = nil
}

// MarkForceRefreshed records the current time as the last force refresh.
func (s *WebSession) MarkForceRefreshed() {
	s.dirty = true
	now := time.Now().UTC()
	s.state.LastForceRefreshAt = &now
}

// SetWebAuthn stores or clears WebAuthn session data.
func (s *WebSession) SetWebAuthn(data *webauthn.SessionData) {
	s.dirty = true
	s.state.WebAuthn = data
}

// MarshalState serializes the session state to JSON for storage.
func (s *WebSession) MarshalState() ([]byte, error) {
	return json.Marshal(s.state)
}

// UnmarshalState populates the session state from raw JSON bytes.
func (s *WebSession) UnmarshalState(data []byte) error {
	s.state = webSessionState{}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, &s.state)
}
