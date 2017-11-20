// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package controller

const (
	NbItemsPerPage = 100
)

type Pagination struct {
	Route        string
	Total        int
	Offset       int
	ItemsPerPage int
	ShowNext     bool
	ShowPrev     bool
	NextOffset   int
	PrevOffset   int
}

func (c *Controller) getPagination(route string, total, offset int) Pagination {
	nextOffset := 0
	prevOffset := 0
	showNext := (total - offset) > NbItemsPerPage
	showPrev := offset > 0

	if showNext {
		nextOffset = offset + NbItemsPerPage
	}

	if showPrev {
		prevOffset = offset - NbItemsPerPage
	}

	return Pagination{
		Route:        route,
		Total:        total,
		Offset:       offset,
		ItemsPerPage: NbItemsPerPage,
		ShowNext:     showNext,
		NextOffset:   nextOffset,
		ShowPrev:     showPrev,
		PrevOffset:   prevOffset,
	}
}
