// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package atom // import "miniflux.app/v2/internal/reader/atom"

import (
	"log/slog"
	"time"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/date"
	"miniflux.app/v2/internal/reader/sanitizer"
	"miniflux.app/v2/internal/urllib"
)

type Atom03Adapter struct {
	atomFeed *Atom03Feed
}

func NewAtom03Adapter(atomFeed *Atom03Feed) *Atom03Adapter {
	return &Atom03Adapter{atomFeed}
}

func (a *Atom03Adapter) BuildFeed(baseURL string) *model.Feed {
	feed := new(model.Feed)

	// Populate the feed URL.
	feedURL := a.atomFeed.Links.firstLinkWithRelation("self")
	if feedURL != "" {
		if absoluteFeedURL, err := urllib.AbsoluteURL(baseURL, feedURL); err == nil {
			feed.FeedURL = absoluteFeedURL
		}
	} else {
		feed.FeedURL = baseURL
	}

	// Populate the site URL.
	siteURL := a.atomFeed.Links.OriginalLink()
	if siteURL != "" {
		if absoluteSiteURL, err := urllib.AbsoluteURL(baseURL, siteURL); err == nil {
			feed.SiteURL = absoluteSiteURL
		}
	} else {
		feed.SiteURL = baseURL
	}

	// Populate the feed title.
	feed.Title = a.atomFeed.Title.Content()
	if feed.Title == "" {
		feed.Title = feed.SiteURL
	}

	for _, atomEntry := range a.atomFeed.Entries {
		entry := model.NewEntry()

		// Populate the entry URL.
		entry.URL = atomEntry.Links.OriginalLink()
		if entry.URL != "" {
			if absoluteEntryURL, err := urllib.AbsoluteURL(feed.SiteURL, entry.URL); err == nil {
				entry.URL = absoluteEntryURL
			}
		}

		// Populate the entry content.
		entry.Content = atomEntry.Content.Content()
		if entry.Content == "" {
			entry.Content = atomEntry.Summary.Content()
		}

		// Populate the entry title.
		entry.Title = atomEntry.Title.Content()
		if entry.Title == "" {
			entry.Title = sanitizer.TruncateHTML(entry.Content, 100)
		}
		if entry.Title == "" {
			entry.Title = entry.URL
		}

		// Populate the entry author.
		entry.Author = atomEntry.Author.PersonName()
		if entry.Author == "" {
			entry.Author = a.atomFeed.Author.PersonName()
		}

		// Populate the entry date.
		for _, value := range []string{atomEntry.Issued, atomEntry.Modified, atomEntry.Created} {
			if parsedDate, err := date.Parse(value); err == nil {
				entry.Date = parsedDate
				break
			} else {
				slog.Debug("Unable to parse date from Atom 0.3 feed",
					slog.String("date", value),
					slog.String("id", atomEntry.ID),
					slog.Any("error", err),
				)
			}
		}
		if entry.Date.IsZero() {
			entry.Date = time.Now()
		}

		// Generate the entry hash.
		for _, value := range []string{atomEntry.ID, atomEntry.Links.OriginalLink()} {
			if value != "" {
				entry.Hash = crypto.Hash(value)
				break
			}
		}

		feed.Entries = append(feed.Entries, entry)
	}

	return feed
}
