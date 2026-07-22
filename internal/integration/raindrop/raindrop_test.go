// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package raindrop

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestNewClientTagParsing(t *testing.T) {
	tests := []struct {
		name string
		tags string
		want []string
	}{
		{name: "empty string produces no tags", tags: "", want: nil},
		{name: "single tag", tags: "news", want: []string{"news"}},
		{name: "multiple tags", tags: "news,tech", want: []string{"news", "tech"}},
		{name: "whitespace is trimmed", tags: " news , tech ", want: []string{"news", "tech"}},
		{name: "empty items are dropped", tags: "news,,tech,", want: []string{"news", "tech"}},
		{name: "only separators produce no tags", tags: ", ,", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient("token", "collection", tt.tags)
			if !slices.Equal(client.tags, tt.want) {
				t.Errorf("NewClient(%q) tags = %#v, want %#v", tt.tags, client.tags, tt.want)
			}
		})
	}
}

func TestPayloadOmitsEmptyTags(t *testing.T) {
	payload, err := json.Marshal(&raindrop{Link: "https://example.com", Title: "Example"})
	if err != nil {
		t.Fatalf("unable to marshal payload: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatalf("unable to unmarshal payload: %v", err)
	}

	if _, found := fields["tags"]; found {
		t.Errorf("payload without tags should omit the tags field, got %s", payload)
	}
}
