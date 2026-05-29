// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package collection // import "miniflux.app/v2/internal/collection"

import (
	"testing"

	"miniflux.app/v2/internal/model"
)

func TestSummarizeBasics(t *testing.T) {
	items := model.CollectionItems{
		{Title: "alpha", Content: "one two three four"},
		{Title: "beta", Content: "five six"},
	}

	summary := newQuotaCalculator().summarize(items)

	if summary.UsedBytes != len("one two three four")+len("five six") {
		t.Fatalf("unexpected used bytes: %d", summary.UsedBytes)
	}
	if summary.UsedPercent < 0 {
		t.Fatalf("unexpected negative percent: %d", summary.UsedPercent)
	}
	if summary.MonthlyCost < 0 {
		t.Fatalf("unexpected negative cost: %f", summary.MonthlyCost)
	}
	if summary.ReadingMinutes < 1 {
		t.Fatalf("expected at least one reading minute, got %d", summary.ReadingMinutes)
	}
}

func TestPrimaryTag(t *testing.T) {
	tagged := &model.CollectionItem{Title: "news: a headline"}
	if got := primaryTag(tagged); got != "news" {
		t.Fatalf("expected primary tag 'news', got %q", got)
	}

	untagged := &model.CollectionItem{Title: "a plain title"}
	if got := primaryTag(untagged); got != "" {
		t.Fatalf("expected no tag, got %q", got)
	}
}
