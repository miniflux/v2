// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

const (
	nbItemsPerPage = 100
)

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

func (c *Controller) getPagination(route string, total, offset int) pagination {
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
