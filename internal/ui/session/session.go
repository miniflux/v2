// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package session // import "miniflux.app/v2/internal/ui/session"

import (
	"time"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
)

// Session handles session data.
type Session struct {
	store     *storage.Storage
	sessionID string
}

// New returns a new session handler.
func New(store *storage.Storage, sessionID string) *Session {
	return &Session{store, sessionID}
}

func (s *Session) SetLastForceRefresh() {
	s.store.UpdateAppSessionField(s.sessionID, "last_force_refresh", time.Now().UTC().Unix())
}

func (s *Session) SetOAuth2State(state string) {
	s.store.UpdateAppSessionField(s.sessionID, "oauth2_state", state)
}

func (s *Session) SetOAuth2CodeVerifier(codeVerfier string) {
	s.store.UpdateAppSessionField(s.sessionID, "oauth2_code_verifier", codeVerfier)
}

// NewFlashMessage creates a new flash message.
func (s *Session) NewFlashMessage(message string) {
	s.store.UpdateAppSessionField(s.sessionID, "flash_message", message)
}

// FlashMessage returns the current flash message if any.
func (s *Session) FlashMessage(message string) string {
	if message != "" {
		s.store.UpdateAppSessionField(s.sessionID, "flash_message", "")
	}
	return message
}

// NewFlashErrorMessage creates a new flash error message.
func (s *Session) NewFlashErrorMessage(message string) {
	s.store.UpdateAppSessionField(s.sessionID, "flash_error_message", message)
}

// FlashErrorMessage returns the last flash error message if any.
func (s *Session) FlashErrorMessage(message string) string {
	if message != "" {
		s.store.UpdateAppSessionField(s.sessionID, "flash_error_message", "")
	}
	return message
}

// SetLanguage updates the language field in session.
func (s *Session) SetLanguage(language string) {
	s.store.UpdateAppSessionField(s.sessionID, "language", language)
}

// SetTheme updates the theme field in session.
func (s *Session) SetTheme(theme string) {
	s.store.UpdateAppSessionField(s.sessionID, "theme", theme)
}

// SetPocketRequestToken updates Pocket Request Token.
func (s *Session) SetPocketRequestToken(requestToken string) {
	s.store.UpdateAppSessionField(s.sessionID, "pocket_request_token", requestToken)
}

func (s *Session) SetWebAuthnSessionData(sessionData *model.WebAuthnSession) {
	s.store.UpdateAppSessionObjectField(s.sessionID, "webauthn_session_data", sessionData)
}
