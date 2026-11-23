// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"testing"

	"miniflux.app/v2/internal/config"
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

func TestIsValidRefreshInterval(t *testing.T) {
	// Ensure config is initialized with default POLLING_FREQUENCY=60.
	if config.Opts == nil {
		parser := config.NewConfigParser()
		opts, err := parser.ParseEnvironmentVariables()
		if err != nil {
			t.Fatalf("failed to parse config: %v", err)
		}
		config.Opts = opts
	}

	minInterval := config.Opts.PollingFrequency().Minutes()
	cases := []struct {
		name  string
		value int
		want  bool
	}{
		{"below minimum", int(minInterval) - 1, false},
		{"at minimum", int(minInterval), true},
		{"normal", 60, true},
		{"at maximum", MaxFeedRefreshIntervalMinutes, true},
		{"above maximum", MaxFeedRefreshIntervalMinutes + 1, false},
		{"zero", 0, false},
		{"negative", -1, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isValidRefreshInterval(c.value); got != c.want {
				t.Errorf(`isValidRefreshInterval(%d) = %v, want %v`, c.value, got, c.want)
			}
		})
	}
}
