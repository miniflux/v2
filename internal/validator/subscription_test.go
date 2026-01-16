// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"testing"

	"miniflux.app/v2/internal/model"
)

func TestValidateSubscriptionDiscovery(t *testing.T) {
	tests := []struct {
		name    string
		req     *model.SubscriptionDiscoveryRequest
		wantErr bool
	}{
		{
			name:    "valid site url",
			req:     &model.SubscriptionDiscoveryRequest{URL: "https://example.org"},
			wantErr: false,
		},
		{
			name:    "invalid site url",
			req:     &model.SubscriptionDiscoveryRequest{URL: "example.org"},
			wantErr: true,
		},
		{
			name:    "invalid proxy url",
			req:     &model.SubscriptionDiscoveryRequest{URL: "https://example.org", ProxyURL: "example.org"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateSubscriptionDiscovery(tc.req); (err != nil) != tc.wantErr {
				t.Fatalf("expected error %v, got %v", tc.wantErr, err)
			}
		})
	}
}
