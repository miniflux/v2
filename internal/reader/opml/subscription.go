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

	// Miniflux-specific feed settings
	ScraperRules                string
	RewriteRules                string
	UrlRewriteRules             string
	BlocklistRules              string
	KeeplistRules               string
	BlockFilterEntryRules       string
	KeepFilterEntryRules        string
	UserAgent                   string
	Cookie                      string
	ProxyURL                    string
	Crawler                     bool
	IgnoreHTTPCache             bool
	FetchViaProxy               bool
	Disabled                    bool
	NoMediaPlayer               bool
	HideGlobally                bool
	AllowSelfSignedCertificates bool
	DisableHTTP2                bool
	IgnoreEntryUpdates          bool
}
