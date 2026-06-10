// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

import (
	"encoding/hex"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

type WebAuthnCredential struct {
	Credential webauthn.Credential
	Name       string
	AddedOn    *time.Time
	LastSeenOn *time.Time
	Handle     []byte

	// False for rows predating the backup_eligible column; the login handler backfills from the assertion on first use.
	BackupEligibleKnown bool
}

func (s WebAuthnCredential) HandleEncoded() string {
	return hex.EncodeToString(s.Handle)
}
