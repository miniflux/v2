// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package crypto // import "miniflux.app/v2/internal/crypto"

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashFromBytes returns a SHA-256 checksum of the input.
func HashFromBytes(value []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(value))
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

func GenerateSHA256Hmac(secret string, data []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func GenerateUUID() string {
	b := GenerateRandomBytes(16)
	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func ConstantTimeCmp(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
