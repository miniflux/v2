// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import "testing"

func TestIsValidFilterRules(t *testing.T) {
	tests := []struct {
		name    string
		rules   string
		wantErr bool
	}{
		{
			name:    "valid single rule",
			rules:   "EntryTitle=foo",
			wantErr: false,
		},
		{
			name:    "valid multiple rules",
			rules:   "EntryTitle=foo\nEntryContent=bar",
			wantErr: false,
		},
		{
			name:    "invalid field name",
			rules:   "Title=foo",
			wantErr: true,
		},
		{
			name:    "missing separator",
			rules:   "EntryTitle:foo",
			wantErr: true,
		},
		{
			name:    "empty regex",
			rules:   "EntryTitle=",
			wantErr: true,
		},
		{
			name:    "invalid regex",
			rules:   "EntryTitle=[",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			err := isValidFilterRules(tc.rules, "block")
			if (err != nil) != tc.wantErr {
				t.Fatalf("expected error=%v, got %v", tc.wantErr, err)
			}
		})
	}
}
