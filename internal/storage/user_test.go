// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
	"miniflux.app/v2/internal/crypto"
)

// The dummy comparison made for unknown users must cost as much as a real one, or the response time reveals which users exist.
func TestDummyBcryptHashHasSameCostAsPasswordHashes(t *testing.T) {
	passwordHash, err := crypto.HashPassword("password")
	if err != nil {
		t.Fatal(err)
	}

	expectedCost, err := bcrypt.Cost([]byte(passwordHash))
	if err != nil {
		t.Fatal(err)
	}

	cost, err := bcrypt.Cost(dummyBcryptHash)
	if err != nil {
		t.Fatalf(`The dummy hash is not a valid bcrypt hash, so comparing with it returns immediately: %v`, err)
	}

	if cost != expectedCost {
		t.Errorf(`The dummy hash has a cost of %d instead of %d, the cost of password hashes`, cost, expectedCost)
	}
}
