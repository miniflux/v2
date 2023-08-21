// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package crypto // import "miniflux.app/v2/internal/crypto"

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashFromBytes returns a SHA-256 checksum of the input.
func HashFromBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return fmt.Sprintf("%x", sum)
}

// Hash returns a SHA-256 checksum of a string.
func Hash(value string) string {
	return HashFromBytes([]byte(value))
}

// GenerateRandomBytes returns random bytes.
func GenerateRandomBytes(size int) []byte {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return b
}

// GenerateRandomString returns a random string.
func GenerateRandomString(size int) string {
	return base64.URLEncoding.EncodeToString(GenerateRandomBytes(size))
}

// GenerateRandomStringHex returns a random hexadecimal string.
func GenerateRandomStringHex(size int) string {
	return hex.EncodeToString(GenerateRandomBytes(size))
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
