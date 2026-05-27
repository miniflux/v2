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
		Title:       strings.TrimSpace(j.jsonFeed.Title),
		FeedURL:     strings.TrimSpace(j.jsonFeed.FeedURL),
		SiteURL:     strings.TrimSpace(j.jsonFeed.HomePageURL),
		Description: strings.TrimSpace(j.jsonFeed.Description),
	}

	if feed.FeedURL == "" {
		feed.FeedURL = strings.TrimSpace(baseURL)
	}

	// Fallback to the feed URL if the site URL is empty.
	if feed.SiteURL == "" {
		feed.SiteURL = feed.FeedURL
	}

	if feedURL, err := urllib.ResolveToAbsoluteURL(baseURL, feed.FeedURL); err == nil {
		feed.FeedURL = feedURL
	}

	if siteURL, err := urllib.ResolveToAbsoluteURL(baseURL, feed.SiteURL); err == nil {
		feed.SiteURL = siteURL
	}

	// Fallback to the feed URL if the title is empty.
	if feed.Title == "" {
		feed.Title = feed.SiteURL
	}

	// Populate the icon URL if present.
	for _, iconURL := range []string{j.jsonFeed.FaviconURL, j.jsonFeed.IconURL} {
		if iconURL = strings.TrimSpace(iconURL); iconURL == "" {
			continue
		}

		if absoluteIconURL, err := urllib.ResolveToAbsoluteURL(feed.SiteURL, iconURL); err == nil {
			feed.IconURL = absoluteIconURL
			break
		}
	}

	for _, item := range j.jsonFeed.Items {
		entry := model.NewEntry()

		for _, itemURL := range []string{item.URL, item.ExternalURL} {
			if itemURL = strings.TrimSpace(itemURL); itemURL == "" {
				continue
			}

			// Make sure the entry URL is absolute.
			if entryURL, err := urllib.ResolveToAbsoluteURL(feed.SiteURL, itemURL); err == nil {
				entry.URL = entryURL
				break
			}
		}

		entry.Title = strings.TrimSpace(item.Title)
		if entry.Title == "" {
			// The entry title is optional, so we need to find a fallback.
			for _, value := range []string{item.Summary, item.ContentText, item.ContentHTML} {
				if value = sanitizer.TruncateHTML(value, 100); value == "" {
					continue
				}

				entry.Title = value
				break
			}
		}

		// Fallback to the entry URL if the title is empty.
		if entry.Title == "" {
			entry.Title = entry.URL
		}

		// Populate the entry content.
		for _, value := range []string{item.ContentHTML, item.ContentText, item.Summary} {
			if value = strings.TrimSpace(value); value == "" {
				continue
			}

			entry.Content = value
			break
		}

		// Populate the entry date.
		for _, value := range []string{item.DatePublished, item.DateModified} {
			if value = strings.TrimSpace(value); value == "" {
				continue
			}

			parsedDate, err := date.Parse(value)
			if err != nil {
				slog.Debug("Unable to parse date from JSON feed",
					slog.String("date", value),
					slog.String("url", entry.URL),
					slog.Any("error", err),
				)
				continue
			}

			entry.Date = parsedDate
			break
		}

		if entry.Date.IsZero() {
			entry.Date = time.Now()
		}

		// Populate the entry author.
		itemAuthors := j.jsonFeed.Authors
		itemAuthors = append(itemAuthors, item.Authors...)
		itemAuthors = append(itemAuthors, item.Author, j.jsonFeed.Author)

		var authorNames = make([]string, 0, len(itemAuthors))
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
			if attachmentURL == "" {
				continue
			}

			absoluteAttachmentURL, err := urllib.ResolveToAbsoluteURL(feed.SiteURL, attachmentURL)
			if err != nil {
				slog.Debug("Unable to build absolute URL for attachment",
					slog.String("url", attachmentURL),
					slog.String("site_url", feed.SiteURL),
					slog.Any("error", err),
				)
				continue
			}

			entry.Enclosures = append(entry.Enclosures, &model.Enclosure{
				URL:      absoluteAttachmentURL,
				MimeType: attachment.MimeType,
				Size:     attachment.Size,
			})
		}

		// Populate the entry tags.
		for _, tag := range item.Tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				entry.Tags = append(entry.Tags, tag)
			}
		}

		// Sort and deduplicate tags.
		slices.Sort(entry.Tags)
		entry.Tags = slices.Compact(entry.Tags)

		// Generate a hash for the entry.
		for _, value := range []string{item.ID, item.URL, item.ExternalURL, item.ContentText + item.ContentHTML + item.Summary} {
			value = strings.TrimSpace(value)
			if value != "" {
				entry.Hash = crypto.SHA256(value)
				break
			}
		}

		feed.Entries = append(feed.Entries, entry)
	}

	return feed
}
