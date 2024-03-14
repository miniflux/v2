// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rss // import "miniflux.app/v2/internal/reader/rss"

import (
	"html"
	"log/slog"
	"path"
	"strconv"
	"strings"
	"time"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/date"
	"miniflux.app/v2/internal/reader/sanitizer"
	"miniflux.app/v2/internal/urllib"
)

type RSSAdapter struct {
	rss *RSS
}

func NewRSSAdapter(rss *RSS) *RSSAdapter {
	return &RSSAdapter{rss}
}

func (r *RSSAdapter) BuildFeed(feedURL string) *model.Feed {
	feed := &model.Feed{
		Title:   html.UnescapeString(strings.TrimSpace(r.rss.Channel.Title)),
		FeedURL: feedURL,
		SiteURL: r.rss.Channel.Link,
	}

	if siteURL, err := urllib.AbsoluteURL(feedURL, r.rss.Channel.Link); err == nil {
		feed.SiteURL = siteURL
	}

	// Try to find the feed URL from the Atom links.
	for _, atomLink := range r.rss.Channel.AtomLinks.Links {
		atomLinkHref := strings.TrimSpace(atomLink.URL)
		if atomLinkHref != "" && atomLink.Rel == "self" {
			if absoluteFeedURL, err := urllib.AbsoluteURL(feedURL, atomLinkHref); err == nil {
				feed.FeedURL = absoluteFeedURL
				break
			}
		}
	}

	// Fallback to the site URL if the title is empty.
	if feed.Title == "" {
		feed.Title = feed.SiteURL
	}

	// Get TTL if defined.
	if r.rss.Channel.TTL != "" {
		if ttl, err := strconv.Atoi(r.rss.Channel.TTL); err == nil {
			feed.TTL = ttl
		}
	}

	// Get the feed icon URL if defined.
	if r.rss.Channel.Image != nil {
		if absoluteIconURL, err := urllib.AbsoluteURL(feed.SiteURL, r.rss.Channel.Image.URL); err == nil {
			feed.IconURL = absoluteIconURL
		}
	}

	for _, item := range r.rss.Channel.Items {
		entry := model.NewEntry()
		entry.Author = findEntryAuthor(&item)
		entry.Date = findEntryDate(&item)
		entry.Content = findEntryContent(&item)
		entry.Enclosures = findEntryEnclosures(&item)

		// Populate the entry URL.
		entryURL := findEntryURL(&item)
		if entryURL == "" {
			entry.URL = feed.SiteURL
		} else {
			if absoluteEntryURL, err := urllib.AbsoluteURL(feed.SiteURL, entryURL); err == nil {
				entry.URL = absoluteEntryURL
			} else {
				entry.URL = entryURL
			}
		}

		// Populate the entry title.
		entry.Title = findEntryTitle(&item)
		if entry.Title == "" {
			entry.Title = sanitizer.TruncateHTML(entry.Content, 100)
		}

		if entry.Title == "" {
			entry.Title = entry.URL
		}

		if entry.Author == "" {
			entry.Author = findFeedAuthor(&r.rss.Channel)
		}

		// Generate the entry hash.
		for _, value := range []string{item.GUID.Data, entryURL} {
			if value != "" {
				entry.Hash = crypto.Hash(value)
				break
			}
		}

		// Find CommentsURL if defined.
		if absoluteCommentsURL := strings.TrimSpace(item.CommentsURL); absoluteCommentsURL != "" && urllib.IsAbsoluteURL(absoluteCommentsURL) {
			entry.CommentsURL = absoluteCommentsURL
		}

		// Set podcast listening time.
		if item.ItunesDuration != "" {
			if duration, err := getDurationInMinutes(item.ItunesDuration); err == nil {
				entry.ReadingTime = duration
			}
		}

		// Populate entry categories.
		entry.Tags = append(entry.Tags, item.Categories...)
		entry.Tags = append(entry.Tags, item.MediaCategories.Labels()...)
		entry.Tags = append(entry.Tags, r.rss.Channel.Categories...)
		entry.Tags = append(entry.Tags, r.rss.Channel.GetItunesCategories()...)

		if r.rss.Channel.GooglePlayCategory.Text != "" {
			entry.Tags = append(entry.Tags, r.rss.Channel.GooglePlayCategory.Text)
		}

		feed.Entries = append(feed.Entries, entry)
	}

	return feed
}

func findFeedAuthor(rssChannel *RSSChannel) string {
	var author string
	switch {
	case rssChannel.ItunesAuthor != "":
		author = rssChannel.ItunesAuthor
	case rssChannel.GooglePlayAuthor != "":
		author = rssChannel.GooglePlayAuthor
	case rssChannel.ItunesOwner.String() != "":
		author = rssChannel.ItunesOwner.String()
	case rssChannel.ManagingEditor != "":
		author = rssChannel.ManagingEditor
	case rssChannel.Webmaster != "":
		author = rssChannel.Webmaster
	}
	return sanitizer.StripTags(strings.TrimSpace(author))
}

func findEntryTitle(rssItem *RSSItem) string {
	title := rssItem.Title

	if rssItem.DublinCoreTitle != "" {
		title = rssItem.DublinCoreTitle
	}

	return html.UnescapeString(strings.TrimSpace(title))
}

func findEntryURL(rssItem *RSSItem) string {
	for _, link := range []string{rssItem.FeedBurnerLink, rssItem.Link} {
		if link != "" {
			return strings.TrimSpace(link)
		}
	}

	for _, atomLink := range rssItem.AtomLinks.Links {
		if atomLink.URL != "" && (strings.EqualFold(atomLink.Rel, "alternate") || atomLink.Rel == "") {
			return strings.TrimSpace(atomLink.URL)
		}
	}

	// Specs: https://cyber.harvard.edu/rss/rss.html#ltguidgtSubelementOfLtitemgt
	// isPermaLink is optional, its default value is true.
	// If its value is false, the guid may not be assumed to be a url, or a url to anything in particular.
	if rssItem.GUID.IsPermaLink == "true" || rssItem.GUID.IsPermaLink == "" {
		return strings.TrimSpace(rssItem.GUID.Data)
	}

	return ""
}

func findEntryContent(rssItem *RSSItem) string {
	for _, value := range []string{
		rssItem.DublinCoreContent,
		rssItem.Description,
		rssItem.GooglePlayDescription,
		rssItem.ItunesSummary,
		rssItem.ItunesSubtitle,
	} {
		if value != "" {
			return value
		}
	}
	return ""
}

func findEntryDate(rssItem *RSSItem) time.Time {
	value := rssItem.PubDate
	if rssItem.DublinCoreDate != "" {
		value = rssItem.DublinCoreDate
	}

	if value != "" {
		result, err := date.Parse(value)
		if err != nil {
			slog.Debug("Unable to parse date from RSS feed",
				slog.String("date", value),
				slog.String("guid", rssItem.GUID.Data),
				slog.Any("error", err),
			)
			return time.Now()
		}

		return result
	}

	return time.Now()
}

func findEntryAuthor(rssItem *RSSItem) string {
	var author string

	switch {
	case rssItem.GooglePlayAuthor != "":
		author = rssItem.GooglePlayAuthor
	case rssItem.ItunesAuthor != "":
		author = rssItem.ItunesAuthor
	case rssItem.DublinCoreCreator != "":
		author = rssItem.DublinCoreCreator
	case rssItem.AtomAuthor.String() != "":
		author = rssItem.AtomAuthor.String()
	case strings.Contains(rssItem.Author.Inner, "<![CDATA["):
		author = rssItem.Author.Data
	default:
		author = rssItem.Author.Inner
	}

	return strings.TrimSpace(sanitizer.StripTags(author))
}

func findEntryEnclosures(rssItem *RSSItem) model.EnclosureList {
	enclosures := make(model.EnclosureList, 0)
	duplicates := make(map[string]bool)

	for _, mediaThumbnail := range rssItem.AllMediaThumbnails() {
		if _, found := duplicates[mediaThumbnail.URL]; !found {
			duplicates[mediaThumbnail.URL] = true
			enclosures = append(enclosures, &model.Enclosure{
				URL:      mediaThumbnail.URL,
				MimeType: mediaThumbnail.MimeType(),
				Size:     mediaThumbnail.Size(),
			})
		}
	}

	for _, enclosure := range rssItem.Enclosures {
		enclosureURL := enclosure.URL

		if rssItem.FeedBurnerEnclosureLink != "" {
			filename := path.Base(rssItem.FeedBurnerEnclosureLink)
			if strings.Contains(enclosureURL, filename) {
				enclosureURL = rssItem.FeedBurnerEnclosureLink
			}
		}

		if enclosureURL == "" {
			continue
		}

		if _, found := duplicates[enclosureURL]; !found {
			duplicates[enclosureURL] = true

			enclosures = append(enclosures, &model.Enclosure{
				URL:      enclosureURL,
				MimeType: enclosure.Type,
				Size:     enclosure.Size(),
			})
		}
	}

	for _, mediaContent := range rssItem.AllMediaContents() {
		if _, found := duplicates[mediaContent.URL]; !found {
			duplicates[mediaContent.URL] = true
			enclosures = append(enclosures, &model.Enclosure{
				URL:      mediaContent.URL,
				MimeType: mediaContent.MimeType(),
				Size:     mediaContent.Size(),
			})
		}
	}

	for _, mediaPeerLink := range rssItem.AllMediaPeerLinks() {
		if _, found := duplicates[mediaPeerLink.URL]; !found {
			duplicates[mediaPeerLink.URL] = true
			enclosures = append(enclosures, &model.Enclosure{
				URL:      mediaPeerLink.URL,
				MimeType: mediaPeerLink.MimeType(),
				Size:     mediaPeerLink.Size(),
			})
		}
	}

	return enclosures
}
