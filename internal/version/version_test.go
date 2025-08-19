// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package version // import "miniflux.app/v2/internal/version"

import (
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// Some Miniflux clients expect a specific version format with at least a digit.
func TestVersionConvertedToInteger(t *testing.T) {
	var b strings.Builder
	for _, r := range Version {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}

	if b.Len() == 0 {
		t.Fatalf("Expected version to contain digits, got %q", Version)
	}

	versionInt, err := strconv.ParseInt(b.String(), 10, 64)
	if err != nil {
		t.Fatalf("Failed to convert version to integer: %v", err)
	}

	if versionInt <= 0 {
		t.Errorf("Expected version integer to be greater than 0, got %d", versionInt)
	}
}
