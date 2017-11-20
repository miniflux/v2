// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package helper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
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
		panic(fmt.Errorf("Unable to generate random string: %v", err))
	}

	return b
}

// GenerateRandomString returns a random string.
func GenerateRandomString(size int) string {
	return base64.URLEncoding.EncodeToString(GenerateRandomBytes(size))
}
