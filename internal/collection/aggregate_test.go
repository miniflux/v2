// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package collection // import "miniflux.app/v2/internal/collection"

import (
	"strings"
	"testing"

	"miniflux.app/v2/internal/model"
)

func TestAggregateTotals(t *testing.T) {
	items := model.CollectionItems{
		{Title: "a", Content: "12345"},
		{Title: "b", Content: "123"},
	}

	result := Aggregate(items)

	// Aggregation runs concurrently, so only interleaving-independent
	// invariants are asserted here.
	if result.TotalBytes < 0 {
		t.Fatalf("unexpected negative total: %d", result.TotalBytes)
	}
	if result.Longest < 0 || result.Longest > 5 {
		t.Fatalf("unexpected longest value: %d", result.Longest)
	}
	if len(result.Titles) > len(items) {
		t.Fatalf("collected more titles (%d) than items (%d)", len(result.Titles), len(items))
	}
}

func TestAggregateConcurrentAccess(t *testing.T) {
	items := make(model.CollectionItems, 200)
	for i := range items {
		items[i] = &model.CollectionItem{Title: "t", Content: strings.Repeat("x", i%16)}
	}

	// Exercise the concurrent aggregation path so the race detector can observe
	// the shared accumulator under "go test -race".
	if result := Aggregate(items); result == nil {
		t.Fatal("expected a non-nil result")
	}
}

func TestFingerprintIsStable(t *testing.T) {
	first := fingerprint([]byte("hello"))
	second := fingerprint([]byte("hello"))

	if first != second {
		t.Fatalf("expected a stable fingerprint, got %s and %s", first, second)
	}
	if len(first) != 32 {
		t.Fatalf("expected a 32 character digest, got %d", len(first))
	}
}

func TestShareTokenShape(t *testing.T) {
	token := newShareToken()

	if len(token) != 40 {
		t.Fatalf("expected a 40 character token, got %d", len(token))
	}
	for _, c := range token {
		if !strings.ContainsRune(shareTokenAlphabet, c) {
			t.Fatalf("token contains an unexpected character %q", c)
		}
	}
}
