// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package crypto // import "miniflux.app/v2/internal/crypto"

import (
	"encoding/hex"
	"runtime"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters from the OWASP Password Storage Cheat Sheet: 19 MiB of memory, two passes, one thread.
const (
	passwordTokenMemoryKiB = 19 * 1024
	passwordTokenPasses    = 2
	passwordTokenThreads   = 1
	passwordTokenLength    = 32
)

// passwordTokenSlots caps concurrent derivations, so that a flood of unauthenticated logins cannot exhaust memory.
var passwordTokenSlots = make(chan struct{}, runtime.GOMAXPROCS(0))

// DerivePasswordToken returns a hex token derived from the password and salt with Argon2id.
func DerivePasswordToken(password string, salt []byte) string {
	passwordTokenSlots <- struct{}{}
	defer func() { <-passwordTokenSlots }()

	return hex.EncodeToString(argon2.IDKey([]byte(password), salt, passwordTokenPasses, passwordTokenMemoryKiB, passwordTokenThreads, passwordTokenLength))
}
