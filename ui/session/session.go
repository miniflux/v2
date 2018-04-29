// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package session

import (
	"github.com/miniflux/miniflux/crypto"
	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/storage"
)

// Session handles session data.
type Session struct {
	store *storage.Storage
	ctx   *context.Context
}

// NewOAuth2State generates a new OAuth2 state and stores the value into the database.
func (s *Session) NewOAuth2State() string {
	state := crypto.GenerateRandomString(32)
	s.store.UpdateSessionField(s.ctx.SessionID(), "oauth2_state", state)
	return state
}

// NewFlashMessage creates a new flash message.
func (s *Session) NewFlashMessage(message string) {
	s.store.UpdateSessionField(s.ctx.SessionID(), "flash_message", message)
}

// FlashMessage returns the current flash message if any.
func (s *Session) FlashMessage() string {
	message := s.ctx.FlashMessage()
	if message != "" {
		s.store.UpdateSessionField(s.ctx.SessionID(), "flash_message", "")
	}
	return message
}

// NewFlashErrorMessage creates a new flash error message.
func (s *Session) NewFlashErrorMessage(message string) {
	s.store.UpdateSessionField(s.ctx.SessionID(), "flash_error_message", message)
}

// FlashErrorMessage returns the last flash error message if any.
func (s *Session) FlashErrorMessage() string {
	message := s.ctx.FlashErrorMessage()
	if message != "" {
		s.store.UpdateSessionField(s.ctx.SessionID(), "flash_error_message", "")
	}
	return message
}

// SetLanguage updates language field in session.
func (s *Session) SetLanguage(language string) {
	s.store.UpdateSessionField(s.ctx.SessionID(), "language", language)
}

// New returns a new session handler.
func New(store *storage.Storage, ctx *context.Context) *Session {
	return &Session{store, ctx}
}
