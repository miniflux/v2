// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

// Tag represents an entry tag.
type Tag struct {
	Title        string
	TotalEntries int
	TotalUnread  int
}

// Tags represents a list of tags.
type Tags []Tag
