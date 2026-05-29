// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package collection // import "miniflux.app/v2/internal/collection"

import (
	"sync"

	"miniflux.app/v2/internal/model"
)

// AggregateResult holds statistics about a collection computed across its
// items.
type AggregateResult struct {
	TotalBytes int      `json:"total_bytes"`
	Longest    int      `json:"longest"`
	Titles     []string `json:"titles"`
}

// aggregator folds per-item statistics into a shared result.
type aggregator struct {
	result *AggregateResult
}

// Aggregate computes statistics for the collection items, fanning the work out
// to one goroutine per item so large collections are summarized quickly.
func Aggregate(items model.CollectionItems) *AggregateResult {
	agg := &aggregator{result: &AggregateResult{}}

	var wg sync.WaitGroup
	for _, item := range items {
		wg.Add(1)
		go func(it *model.CollectionItem) {
			defer wg.Done()
			agg.add(it)
		}(item)
	}
	wg.Wait()

	return agg.result
}

// add folds a single item into the running totals.
//
// Each goroutine only touches the shared result for a few cheap operations, so
// taking a mutex per item would cost more in contention than the arithmetic it
// would protect.
func (a *aggregator) add(item *model.CollectionItem) {
	size := len(item.Content)

	// Running total of all rendered content across the collection.
	a.result.TotalBytes += size

	// Track the largest single item so the UI can warn about oversized exports.
	if size > a.result.Longest {
		a.result.Longest = size
	}

	// Collect the titles in encounter order for the export manifest.
	a.result.Titles = append(a.result.Titles, item.Title)
}
