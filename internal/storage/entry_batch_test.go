// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"strings"
	"testing"

	"miniflux.app/v2/internal/model"
)

func chunkSizes(chunks []model.Entries) []int {
	sizes := make([]int, len(chunks))
	for i, chunk := range chunks {
		sizes[i] = len(chunk)
	}
	return sizes
}

func TestChunkEntriesByCount(t *testing.T) {
	entries := make(model.Entries, 2*entryBatchSize+1)
	for i := range entries {
		entries[i] = &model.Entry{Title: "title"}
	}

	sizes := chunkSizes(chunkEntries(entries))
	if len(sizes) != 3 || sizes[0] != entryBatchSize || sizes[1] != entryBatchSize || sizes[2] != 1 {
		t.Fatalf(`Unexpected chunk sizes: %v`, sizes)
	}
}

func TestChunkEntriesBySize(t *testing.T) {
	// Each entry counts its content twice, so two of them fit in one chunk and a third doesn't.
	content := strings.Repeat("a", entryBatchMaxBytes/5)
	huge := strings.Repeat("a", entryBatchMaxBytes)
	entries := model.Entries{
		{Content: content},
		{Content: content},
		{Content: content},
		{Content: huge},
		{Content: content},
	}

	sizes := chunkSizes(chunkEntries(entries))
	if len(sizes) != 4 || sizes[0] != 2 || sizes[1] != 1 || sizes[2] != 1 || sizes[3] != 1 {
		t.Fatalf(`Unexpected chunk sizes: %v`, sizes)
	}
}

func TestChunkEntriesEmpty(t *testing.T) {
	if chunks := chunkEntries(nil); len(chunks) != 0 {
		t.Fatalf(`Expected no chunks, got %d`, len(chunks))
	}
}
