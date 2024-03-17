// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package atom // import "miniflux.app/v2/internal/reader/atom"

import (
	"log/slog"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/date"
	"miniflux.app/v2/internal/reader/sanitizer"
	"miniflux.app/v2/internal/urllib"
)

type Atom10Adapter struct {
	atomFeed *Atom10Feed
}

func NewAtom10Adapter(atomFeed *Atom10Feed) *Atom10Adapter {
	return &Atom10Adapter{atomFeed}
}

func (a *Atom10Adapter) BuildFeed(baseURL string) *model.Feed {
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
	feed.Title = a.atomFeed.Title.Body()
	if feed.Title == "" {
		feed.Title = feed.SiteURL
	}

	// Populate the feed icon.
	if a.atomFeed.Icon != "" {
		if absoluteIconURL, err := urllib.AbsoluteURL(feed.SiteURL, a.atomFeed.Icon); err == nil {
			feed.IconURL = absoluteIconURL
		}
	} else if a.atomFeed.Logo != "" {
		if absoluteLogoURL, err := urllib.AbsoluteURL(feed.SiteURL, a.atomFeed.Logo); err == nil {
			feed.IconURL = absoluteLogoURL
		}
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
		entry.Content = atomEntry.Content.Body()
		if entry.Content == "" {
			entry.Content = atomEntry.Summary.Body()
		}
		if entry.Content == "" {
			entry.Content = atomEntry.FirstMediaDescription()
		}

		// Populate the entry title.
		entry.Title = atomEntry.Title.Title()
		if entry.Title == "" {
			entry.Title = sanitizer.TruncateHTML(entry.Content, 100)
		}
		if entry.Title == "" {
			entry.Title = entry.URL
		}

		// Populate the entry author.
		authors := atomEntry.Authors.PersonNames()
		if len(authors) == 0 {
			authors = append(authors, a.atomFeed.Authors.PersonNames()...)
		}
		authors = slices.Compact(authors)
		sort.Strings(authors)
		entry.Author = strings.Join(authors, ", ")

		// Populate the entry date.
		for _, value := range []string{atomEntry.Published, atomEntry.Updated} {
			if value != "" {
				if parsedDate, err := date.Parse(value); err != nil {
					slog.Debug("Unable to parse date from Atom 1.0 feed",
						slog.String("date", value),
						slog.String("url", entry.URL),
						slog.Any("error", err),
					)
				} else {
					entry.Date = parsedDate
					break
				}
			}
		}
		if entry.Date.IsZero() {
			entry.Date = time.Now()
		}

		// Populate categories.
		categories := atomEntry.Categories.CategoryNames()
		if len(categories) == 0 {
			categories = append(categories, a.atomFeed.Categories.CategoryNames()...)
		}
		if len(categories) > 0 {
			categories = slices.Compact(categories)
			sort.Strings(categories)
			entry.Tags = categories
		}

		// Populate the commentsURL if defined.
		// See https://tools.ietf.org/html/rfc4685#section-4
		// If the type attribute of the atom:link is omitted, its value is assumed to be "application/atom+xml".
		// We accept only HTML or XHTML documents for now since the intention is to have the same behavior as RSS.
		commentsURL := atomEntry.Links.firstLinkWithRelationAndType("replies", "text/html", "application/xhtml+xml")
		if urllib.IsAbsoluteURL(commentsURL) {
			entry.CommentsURL = commentsURL
		}

		// Generate the entry hash.
		for _, value := range []string{atomEntry.ID, atomEntry.Links.OriginalLink()} {
			if value != "" {
				entry.Hash = crypto.Hash(value)
				break
			}
		}

		// Populate the entry enclosures.
		uniqueEnclosuresMap := make(map[string]bool)

		for _, mediaThumbnail := range atomEntry.AllMediaThumbnails() {
			if _, found := uniqueEnclosuresMap[mediaThumbnail.URL]; !found {
				uniqueEnclosuresMap[mediaThumbnail.URL] = true
				entry.Enclosures = append(entry.Enclosures, &model.Enclosure{
					URL:      mediaThumbnail.URL,
					MimeType: mediaThumbnail.MimeType(),
					Size:     mediaThumbnail.Size(),
				})
			}
		}

		for _, link := range atomEntry.Links {
			if strings.EqualFold(link.Rel, "enclosure") {
				if link.Href == "" {
					continue
				}

				if _, found := uniqueEnclosuresMap[link.Href]; !found {
					uniqueEnclosuresMap[link.Href] = true
					length, _ := strconv.ParseInt(link.Length, 10, 0)
					entry.Enclosures = append(entry.Enclosures, &model.Enclosure{
						URL:      link.Href,
						MimeType: link.Type,
						Size:     length,
					})
				}
			}
		}

		for _, mediaContent := range atomEntry.AllMediaContents() {
			if _, found := uniqueEnclosuresMap[mediaContent.URL]; !found {
				uniqueEnclosuresMap[mediaContent.URL] = true
				entry.Enclosures = append(entry.Enclosures, &model.Enclosure{
					URL:      mediaContent.URL,
					MimeType: mediaContent.MimeType(),
					Size:     mediaContent.Size(),
				})
			}
		}

		for _, mediaPeerLink := range atomEntry.AllMediaPeerLinks() {
			if _, found := uniqueEnclosuresMap[mediaPeerLink.URL]; !found {
				uniqueEnclosuresMap[mediaPeerLink.URL] = true
				entry.Enclosures = append(entry.Enclosures, &model.Enclosure{
					URL:      mediaPeerLink.URL,
					MimeType: mediaPeerLink.MimeType(),
					Size:     mediaPeerLink.Size(),
				})
			}
		}

		feed.Entries = append(feed.Entries, entry)
	}

	return feed
}
