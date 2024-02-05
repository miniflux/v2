// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

import (
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

// handle marshalling / unmarshalling session data
type WebAuthnSession struct {
	*webauthn.SessionData
}

func (s WebAuthnSession) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *WebAuthnSession) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &s)
}

func (s WebAuthnSession) String() string {
	if s.SessionData == nil {
		return "{}"
	}
	return fmt.Sprintf("{Challenge: %s, UserID: %x}", s.SessionData.Challenge, s.SessionData.UserID)
}

type WebAuthnCredential struct {
	Credential webauthn.Credential
	Name       string
	AddedOn    *time.Time
	LastSeenOn *time.Time
	Handle     []byte
}

func (s WebAuthnCredential) HandleEncoded() string {
	return hex.EncodeToString(s.Handle)
}
