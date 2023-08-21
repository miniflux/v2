// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

func CategoriesSortingOptions() map[string]string {
	return map[string]string{
		"unread_count": "form.prefs.select.unread_count",
		"alphabetical": "form.prefs.select.alphabetical",
	}
}
