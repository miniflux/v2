// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

// EntryStatItem represents an entries statistics of a Feed or Category
type EntryStatItem struct {
	Count    string    `json:"count"`
	Feed     *Feed     `json:"feed,omitempty"`
	Category *Category `json:"category,omitempty"`
}

// EntryStat represents an entries statistics of Feeds or Categories
type EntryStat []*EntryStatItem
