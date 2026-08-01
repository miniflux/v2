// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"testing"

	"miniflux.app/v2/internal/model"
)

func TestValidateFeedModificationProxyURL(t *testing.T) {
	tests := []struct {
		name     string
		proxyURL string
		wantErr  bool
	}{
		{
			name:     "empty proxy URL",
			proxyURL: "",
			wantErr:  false,
		},
		{
			name:     "valid proxy URL",
			proxyURL: "http://127.0.0.1:3128",
			wantErr:  false,
		},
		{
			name:     "invalid proxy URL",
			proxyURL: "example.org",
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := &model.FeedModificationRequest{ProxyURL: &tc.proxyURL}
			if err := ValidateFeedModification(nil, 0, 0, request); (err != nil) != tc.wantErr {
				t.Fatalf("expected error %v, got %v", tc.wantErr, err)
			}
		})
	}
}
