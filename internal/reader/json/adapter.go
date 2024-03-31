// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package json // import "miniflux.app/v2/internal/reader/json"

import (
	"log/slog"
	"slices"
	"strings"
	"time"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/date"
	"miniflux.app/v2/internal/reader/sanitizer"
	"miniflux.app/v2/internal/urllib"
)

type JSONAdapter struct {
	jsonFeed *JSONFeed
}

func NewJSONAdapter(jsonFeed *JSONFeed) *JSONAdapter {
	return &JSONAdapter{jsonFeed}
}

func (j *JSONAdapter) BuildFeed(baseURL string) *model.Feed {
	feed := &model.Feed{
		Title:   strings.TrimSpace(j.jsonFeed.Title),
		FeedURL: strings.TrimSpace(j.jsonFeed.FeedURL),
		SiteURL: strings.TrimSpace(j.jsonFeed.HomePageURL),
	}

	if feed.FeedURL == "" {
		feed.FeedURL = strings.TrimSpace(baseURL)
	}

	// Fallback to the feed URL if the site URL is empty.
	if feed.SiteURL == "" {
		feed.SiteURL = feed.FeedURL
	}

	if feedURL, err := urllib.AbsoluteURL(baseURL, feed.FeedURL); err == nil {
		feed.FeedURL = feedURL
	}

	if siteURL, err := urllib.AbsoluteURL(baseURL, feed.SiteURL); err == nil {
		feed.SiteURL = siteURL
	}

	// Fallback to the feed URL if the title is empty.
	if feed.Title == "" {
		feed.Title = feed.SiteURL
	}

	// Populate the icon URL if present.
	for _, iconURL := range []string{j.jsonFeed.FaviconURL, j.jsonFeed.IconURL} {
		iconURL = strings.TrimSpace(iconURL)
		if iconURL != "" {
			if absoluteIconURL, err := urllib.AbsoluteURL(feed.SiteURL, iconURL); err == nil {
				feed.IconURL = absoluteIconURL
				break
			}
		}
	}

	for _, item := range j.jsonFeed.Items {
		entry := model.NewEntry()
		entry.Title = strings.TrimSpace(item.Title)
		entry.URL = strings.TrimSpace(item.URL)

		// Make sure the entry URL is absolute.
		if entryURL, err := urllib.AbsoluteURL(feed.SiteURL, entry.URL); err == nil {
			entry.URL = entryURL
		}

		// The entry title is optional, so we need to find a fallback.
		if entry.Title == "" {
			for _, value := range []string{item.Summary, item.ContentText, item.ContentHTML} {
				if value != "" {
					entry.Title = sanitizer.TruncateHTML(value, 100)
				}
			}
		}

		// Fallback to the entry URL if the title is empty.
		if entry.Title == "" {
			entry.Title = entry.URL
		}

		// Populate the entry content.
		for _, value := range []string{item.ContentHTML, item.ContentText, item.Summary} {
			value = strings.TrimSpace(value)
			if value != "" {
				entry.Content = value
				break
			}
		}

		// Populate the entry date.
		for _, value := range []string{item.DatePublished, item.DateModified} {
			value = strings.TrimSpace(value)
			if value != "" {
				if date, err := date.Parse(value); err != nil {
					slog.Debug("Unable to parse date from JSON feed",
						slog.String("date", value),
						slog.String("url", entry.URL),
						slog.Any("error", err),
					)
				} else {
					entry.Date = date
					break
				}
			}
		}
		if entry.Date.IsZero() {
			entry.Date = time.Now()
		}

		// Populate the entry author.
		itemAuthors := j.jsonFeed.Authors
		itemAuthors = append(itemAuthors, item.Authors...)
		itemAuthors = append(itemAuthors, item.Author, j.jsonFeed.Author)

		var authorNames []string
		for _, author := range itemAuthors {
			authorName := strings.TrimSpace(author.Name)
			if authorName != "" {
				authorNames = append(authorNames, authorName)
			}
		}

		slices.Sort(authorNames)
		authorNames = slices.Compact(authorNames)
		entry.Author = strings.Join(authorNames, ", ")

		// Populate the entry enclosures.
		for _, attachment := range item.Attachments {
			attachmentURL := strings.TrimSpace(attachment.URL)
			if attachmentURL != "" {
				if absoluteAttachmentURL, err := urllib.AbsoluteURL(feed.SiteURL, attachmentURL); err == nil {
					entry.Enclosures = append(entry.Enclosures, &model.Enclosure{
						URL:      absoluteAttachmentURL,
						MimeType: attachment.MimeType,
						Size:     attachment.Size,
					})
				}
			}
		}

		// Populate the entry tags.
		for _, tag := range item.Tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				entry.Tags = append(entry.Tags, tag)
			}
		}

		// Generate a hash for the entry.
		for _, value := range []string{item.ID, item.URL, item.ContentText + item.ContentHTML + item.Summary} {
			value = strings.TrimSpace(value)
			if value != "" {
				entry.Hash = crypto.Hash(value)
				break
			}
		}

		feed.Entries = append(feed.Entries, entry)
	}

	return feed
}
