// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package session // import "miniflux.app/ui/session"

import (
	"miniflux.app/crypto"
	"miniflux.app/storage"
)

// Session handles session data.
type Session struct {
	store     *storage.Storage
	sessionID string
}

// NewOAuth2State generates a new OAuth2 state and stores the value into the database.
func (s *Session) NewOAuth2State() string {
	state := crypto.GenerateRandomString(32)
	s.store.UpdateAppSessionField(s.sessionID, "oauth2_state", state)
	return state
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

// New returns a new session handler.
func New(store *storage.Storage, sessionID string) *Session {
	return &Session{store, sessionID}
}
