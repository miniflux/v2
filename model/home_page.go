// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

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
