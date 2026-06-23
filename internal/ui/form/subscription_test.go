// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import "testing"

func TestSubscriptionFormValidateInvalidBlockFilterRules(t *testing.T) {
	s := &SubscriptionForm{URL: "https://example.com/feed", CategoryID: 1, BlockFilterEntryRules: "BadField=foo"}
	if err := s.Validate(); err == nil {
		t.Error("Validate should return an error for an invalid block filter rule")
	}
}

func TestSubscriptionFormValidateInvalidKeepFilterRules(t *testing.T) {
	s := &SubscriptionForm{URL: "https://example.com/feed", CategoryID: 1, KeepFilterEntryRules: "BadField=foo"}
	if err := s.Validate(); err == nil {
		t.Error("Validate should return an error for an invalid keep filter rule")
	}
}

func TestSubscriptionFormValidateValidFilterRules(t *testing.T) {
	s := &SubscriptionForm{URL: "https://example.com/feed", CategoryID: 1, BlockFilterEntryRules: "EntryTitle=add"}
	if err := s.Validate(); err != nil {
		t.Errorf("Validate should not return an error for a valid filter rule, got: %v", err)
	}
}
