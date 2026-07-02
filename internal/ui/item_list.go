// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

type itemListAction struct {
	Class    string
	URL      string
	Icon     string
	LabelKey string
	Confirm  bool
}

type itemListItem struct {
	ID            string
	Title         string
	TitleURL      string
	UnreadCount   int
	MetaCount     int
	MetaPluralKey string
	MetaEmptyKey  string
	Actions       []itemListAction
}

func intPtrValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
