// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

type pagination struct {
	Route        string
	Total        int
	Offset       int
	ItemsPerPage int
	ShowNext     bool
	ShowLast     bool
	ShowFirst    bool
	ShowPrev     bool
	NextOffset   int
	LastOffset   int
	PrevOffset   int
	FirstOffset  int
	SearchQuery  string
}

func getPagination(route string, total, offset, nbItemsPerPage int) pagination {
	nextOffset := 0
	prevOffset := 0

	firstOffset := 0
	lastOffset := (total / nbItemsPerPage) * nbItemsPerPage
	if lastOffset == total {
		lastOffset -= nbItemsPerPage
	}

	showNext := (total - offset) > nbItemsPerPage
	showPrev := offset > 0
	showLast := showNext
	showFirst := showPrev

	if showNext {
		nextOffset = offset + nbItemsPerPage
	}

	if showPrev {
		prevOffset = offset - nbItemsPerPage
	}

	return pagination{
		Route:        route,
		Total:        total,
		Offset:       offset,
		ItemsPerPage: nbItemsPerPage,
		ShowNext:     showNext,
		ShowLast:     showLast,
		NextOffset:   nextOffset,
		LastOffset:   lastOffset,
		ShowPrev:     showPrev,
		ShowFirst:    showFirst,
		PrevOffset:   prevOffset,
		FirstOffset:  firstOffset,
	}
}
