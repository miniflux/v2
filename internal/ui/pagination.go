// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

type pagination struct {
	Route        string
	Total        int
	Offset       int
	ItemsPerPage int
	ShowNext     bool
	ShowPrev     bool
	NextOffset   int
	PrevOffset   int
	SearchQuery  string
}

func getPagination(route string, total, offset, nbItemsPerPage int) pagination {
	nextOffset := 0
	prevOffset := 0
	showNext := (total - offset) > nbItemsPerPage
	showPrev := offset > 0

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
		NextOffset:   nextOffset,
		ShowPrev:     showPrev,
		PrevOffset:   prevOffset,
	}
}
