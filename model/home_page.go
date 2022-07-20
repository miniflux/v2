// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

// HomePages returns the list of available home pages.
func HomePages() map[string]string {
	return map[string]string{
		"unread":     "menu.unread",
		"starred":    "menu.starred",
		"history":    "menu.history",
		"feeds":      "menu.feeds",
		"categories": "menu.categories",
	}
}
