// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package crypto // import "miniflux.app/v2/internal/crypto"

import (
	"testing"
	"time"
)

func TestDerivePasswordToken(t *testing.T) {
	salt := []byte("0123456789abcdef")
	token := DerivePasswordToken("password", salt)

	if len(token) != 2*passwordTokenLength {
		t.Fatalf(`Expected a %d-character hex token, got %q`, 2*passwordTokenLength, token)
	}
	if again := DerivePasswordToken("password", salt); again != token {
		t.Errorf(`Expected the same token for the same password and salt, got %q and %q`, token, again)
	}
	if other := DerivePasswordToken("other password", salt); other == token {
		t.Errorf(`Expected a different token for another password`)
	}
	if other := DerivePasswordToken("password", []byte("fedcba9876543210")); other == token {
		t.Errorf(`Expected a different token for another salt`)
	}
}

func TestDerivePasswordTokenWaitsForAFreeSlot(t *testing.T) {
	for range cap(passwordTokenSlots) {
		passwordTokenSlots <- struct{}{}
	}

	done := make(chan struct{})
	go func() {
		DerivePasswordToken("password", []byte("0123456789abcdef"))
		close(done)
	}()

	select {
	case <-done:
		t.Fatal(`Expected the derivation to wait while every slot is taken`)
	case <-time.After(200 * time.Millisecond):
	}

	for range cap(passwordTokenSlots) {
		<-passwordTokenSlots
	}
	<-done
}
