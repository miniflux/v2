// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/storage"
)

func (c *Controller) getEntryPrevNext(user *model.User, builder *storage.EntryQueryBuilder, entryID int64) (prev *model.Entry, next *model.Entry, err error) {
	builder.WithoutStatus(model.EntryStatusRemoved)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(user.EntryDirection)

	entries, err := builder.GetEntries()
	if err != nil {
		return nil, nil, err
	}

	n := len(entries)
	for i := 0; i < n; i++ {
		if entries[i].ID == entryID {
			if i-1 >= 0 {
				prev = entries[i-1]
			}

			if i+1 < n {
				next = entries[i+1]
			}
		}
	}

	return prev, next, nil
}
