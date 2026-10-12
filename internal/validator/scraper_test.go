// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"strings"
	"testing"
)

func TestIsValidScraperRules(t *testing.T) {
	tests := []struct {
		rules   string
		wantErr bool
	}{
		{"", false},
		{"article", false},
		{"div.content, .post-body > p", false},
		{"div:has(> p)", true},
		{"div[", true},
		{"p:first", true},
	}

	for _, tc := range tests {
		err := IsValidScraperRules(tc.rules)
		if (err != nil) != tc.wantErr {
			t.Errorf("IsValidScraperRules(%q): expected error %v, got %v", tc.rules, tc.wantErr, err)
		}
		if err != nil && !strings.Contains(err.String(), "not a valid CSS selector") {
			t.Errorf("IsValidScraperRules(%q): unexpected error message: %q", tc.rules, err.String())
		}
	}
}

func TestIsValidRewriteRulesInvalidSelector(t *testing.T) {
	err := IsValidRewriteRules(`remove("div:has(> p)")`)
	if err == nil {
		t.Fatal("expected an error for an invalid selector")
	}
	if msg := err.String(); !strings.Contains(msg, "invalid CSS selector") {
		t.Errorf("unexpected error message: %q", msg)
	}
}
