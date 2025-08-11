// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package opml // import "miniflux.app/v2/internal/reader/opml"

// subcription represents a feed that will be imported or exported.
type subcription struct {
	Title        string
	SiteURL      string
	FeedURL      string
	CategoryName string
	Description  string
}
