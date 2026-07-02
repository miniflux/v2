// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rss // import "miniflux.app/v2/internal/reader/rss"

import (
	"cmp"
	"html"
	"iter"
	"log/slog"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/date"
	"miniflux.app/v2/internal/reader/language"
	"miniflux.app/v2/internal/reader/sanitizer"
	"miniflux.app/v2/internal/urllib"
)

type rssAdapter struct {
	rss *rss
}

func (r *rssAdapter) buildFeed(baseURL string) *model.Feed {
	feed := &model.Feed{
		Title:       html.UnescapeString(strings.TrimSpace(r.rss.Channel.Title)),
		FeedURL:     strings.TrimSpace(baseURL),
		SiteURL:     strings.TrimSpace(r.rss.Channel.Link),
		Description: strings.TrimSpace(r.rss.Channel.Description),
		Language:    language.Normalize(r.rss.Channel.Language),
	}

	// Hybrid feeds declare the channel language with <dc:language>
	// instead of <language>.
	if feed.Language == "" {
		feed.Language = language.Normalize(r.rss.Channel.DublinCoreLanguage)
	}

	// Ensure the Site URL is absolute.
	if absoluteSiteURL, err := urllib.ResolveToAbsoluteURL(baseURL, feed.SiteURL); err == nil {
		feed.SiteURL = absoluteSiteURL
	}

	// Try to find the feed URL from the Channel links.
	for _, link := range r.rss.Channel.Links {
		href := strings.TrimSpace(link.Href)
		if href == "" || link.Rel != "self" {
			continue
		}

		if absoluteFeedURL, err := urllib.ResolveToAbsoluteURL(feed.FeedURL, href); err == nil {
			feed.FeedURL = absoluteFeedURL
			break
		}
	}

	// Fallback to the site URL if the title is empty.
	if feed.Title == "" {
		feed.Title = feed.SiteURL
	}

	// Get TTL if defined.
	if r.rss.Channel.TTL != "" {
		if ttl, err := strconv.Atoi(r.rss.Channel.TTL); err == nil {
			feed.TTL = time.Duration(ttl) * time.Minute
		}
	}

	// Get the feed icon URL if defined.
	if r.rss.Channel.Image != nil {
		if absoluteIconURL, err := urllib.ResolveToAbsoluteURL(feed.SiteURL, r.rss.Channel.Image.URL); err == nil {
			feed.IconURL = absoluteIconURL
		}
	}

	// Track GUIDs already seen in this feed to disambiguate items from
	// non-conformant feeds that reuse the same <guid> for every entry.
	seenGUIDs := make(map[string]int)

	for _, item := range r.rss.Channel.Items {
		entry := model.NewEntry()
		entry.Date = findEntryDate(&item)
		entry.Content = findEntryContent(&item)
		entry.Enclosures = findEntryEnclosures(&item, feed.SiteURL)

		// Populate the entry URL.
		entryURL := findEntryURL(&item)
		if entryURL != "" {
			entry.URL = entryURL
			if absoluteEntryURL, err := urllib.ResolveToAbsoluteURL(feed.SiteURL, entryURL); err == nil {
				entry.URL = absoluteEntryURL
			}
		}

		if entry.URL == "" {
			// Fallback to the feed URL if no entry URL is found.
			entry.URL = feed.SiteURL

			// Fallback to the first enclosure URL if it exists.
			if len(entry.Enclosures) > 0 && entry.Enclosures[0].URL != "" {
				entry.URL = entry.Enclosures[0].URL
			}
		}

		// Populate the entry title.
		entry.Title = findEntryTitle(&item)
		if entry.Title == "" {
			entry.Title = sanitizer.TruncateHTML(entry.Content, 100)
			if entry.Title == "" {
				entry.Title = entry.URL
			}
		}

		entry.Author = findEntryAuthor(&item)
		if entry.Author == "" {
			entry.Author = findFeedAuthor(&r.rss.Channel)
		}

		// Populate the entry language, falling back to the channel
		// language: items are part of the channel's content.
		entry.Language = language.Normalize(item.DublinCoreLanguage)
		if entry.Language == "" {
			entry.Language = feed.Language
		}

		// Generate the entry hash.
		//
		// The RSS 2.0 spec requires <guid> to uniquely identify the item, but
		// some feeds ship the same GUID for every entry. Keep the first
		// occurrence stable (so existing stored entries still match) and
		// disambiguate later collisions using the entry URL or, as a last
		// resort, the item position.
		switch {
		case item.GUID.Data != "":
			n := seenGUIDs[item.GUID.Data]
			seenGUIDs[item.GUID.Data] = n + 1
			switch {
			case n == 0:
				entry.Hash = crypto.SHA256(item.GUID.Data)
			case entry.URL != "":
				entry.Hash = crypto.SHA256(item.GUID.Data + "|" + entry.URL)
			default:
				entry.Hash = crypto.SHA256(item.GUID.Data + "|" + strconv.Itoa(n))
			}
		case entryURL != "":
			entry.Hash = crypto.SHA256(entryURL)
		default:
			entry.Hash = crypto.SHA256(entry.Title + entry.Content)
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
		entry.Tags = findEntryTags(&item)
		if len(entry.Tags) == 0 {
			entry.Tags = findFeedTags(&r.rss.Channel)
		}

		feed.Entries = append(feed.Entries, entry)
	}

	return feed
}

func findFeedAuthor(rssChannel *rssChannel) string {
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
	default:
		return ""
	}

	return sanitizer.StripTags(author)
}

func findFeedTags(rssChannel *rssChannel) []string {
	tags := make([]string, 0, len(rssChannel.Categories)+2*len(rssChannel.ItunesCategories)+1)

	tags = appendSorted(tags, strings.TrimSpace, rssChannel.Categories...)
	tags = appendSortedSeq(tags, strings.TrimSpace, rssChannel.ItunesCategoriesSeq())

	tags = appendSorted(tags, strings.TrimSpace, rssChannel.GooglePlayCategory.Text)

	return tags
}

func findEntryTitle(rssItem *rssItem) string {
	title := rssItem.Title.Content

	if rssItem.DublinCoreTitle != "" {
		title = rssItem.DublinCoreTitle
	}

	return html.UnescapeString(html.UnescapeString(strings.TrimSpace(title)))
}

func findEntryURL(rssItem *rssItem) string {
	for _, link := range []string{rssItem.FeedBurnerLink, rssItem.Link} {
		if link != "" {
			return strings.TrimSpace(link)
		}
	}

	for _, atomLink := range rssItem.Links {
		if atomLink.Href != "" && (strings.EqualFold(atomLink.Rel, "alternate") || atomLink.Rel == "") {
			return strings.TrimSpace(atomLink.Href)
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

func findEntryContent(rssItem *rssItem) string {
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

func findEntryDate(rssItem *rssItem) time.Time {
	value := rssItem.PubDate
	if rssItem.DublinCoreDate != "" {
		value = rssItem.DublinCoreDate
	}

	if value = strings.TrimSpace(value); value == "" {
		return time.Now()
	}

	parsedDate, err := date.Parse(value)
	if err != nil {
		slog.Debug("Unable to parse date from RSS feed",
			slog.String("date", value),
			slog.String("guid", rssItem.GUID.Data),
			slog.Any("error", err),
		)
		return time.Now()
	}

	return parsedDate
}

func findEntryAuthor(rssItem *rssItem) string {
	var author string

	switch {
	case rssItem.GooglePlayAuthor != "":
		author = rssItem.GooglePlayAuthor
	case rssItem.ItunesAuthor != "":
		author = rssItem.ItunesAuthor
	case rssItem.DublinCoreCreator != "":
		author = rssItem.DublinCoreCreator
	case rssItem.PersonName() != "":
		author = rssItem.PersonName()
	case strings.Contains(rssItem.Author.Inner, "<![CDATA["):
		author = rssItem.Author.Data
	case rssItem.Author.Inner != "":
		author = rssItem.Author.Inner
	default:
		return ""
	}

	return sanitizer.StripTags(author)
}

func findEntryTags(rssItem *rssItem) []string {
	tags := make([]string, 0, len(rssItem.Categories)+len(rssItem.MediaCategories))

	tags = appendSorted(tags, strings.TrimSpace, rssItem.Categories...)
	tags = appendSortedSeq(tags, strings.TrimSpace, rssItem.MediaCategories.LabelsSeq())

	return tags
}

func findEntryEnclosures(rssItem *rssItem, siteURL string) model.EnclosureList {
	mediaThumbnails := rssItem.AllMediaThumbnails()
	mediaContents := rssItem.AllMediaContents()
	mediaPeerLinks := rssItem.AllMediaPeerLinks()
	capacity := len(mediaThumbnails) + len(rssItem.Enclosures) + len(mediaContents) + len(mediaPeerLinks)
	enclosures := make(model.EnclosureList, 0, capacity)
	duplicates := make(map[string]bool, capacity)

	for _, mediaThumbnail := range mediaThumbnails {
		mediaURL := strings.TrimSpace(mediaThumbnail.URL)
		if mediaURL == "" {
			continue
		}

		mediaURL, err := urllib.ResolveToAbsoluteURL(siteURL, mediaURL)
		if err != nil {
			slog.Debug("Unable to build absolute URL for media thumbnail",
				slog.String("url", mediaThumbnail.URL),
				slog.String("site_url", siteURL),
				slog.Any("error", err),
			)
			continue
		}

		if _, found := duplicates[mediaURL]; found {
			continue
		}

		duplicates[mediaURL] = true

		enclosures = append(enclosures, &model.Enclosure{
			URL:      mediaURL,
			MimeType: mediaThumbnail.MimeType(),
			Size:     mediaThumbnail.Size(),
		})
	}

	for _, enclosure := range rssItem.Enclosures {
		enclosureURL := enclosure.URL

		if rssItem.FeedBurnerEnclosureLink != "" {
			filename := path.Base(rssItem.FeedBurnerEnclosureLink)
			if strings.HasSuffix(enclosureURL, filename) {
				enclosureURL = rssItem.FeedBurnerEnclosureLink
			}
		}

		enclosureURL = strings.TrimSpace(enclosureURL)
		if enclosureURL == "" {
			continue
		}

		if absoluteEnclosureURL, err := urllib.ResolveToAbsoluteURL(siteURL, enclosureURL); err == nil {
			enclosureURL = absoluteEnclosureURL
		}

		if _, found := duplicates[enclosureURL]; found {
			continue
		}

		duplicates[enclosureURL] = true

		enclosures = append(enclosures, &model.Enclosure{
			URL:      enclosureURL,
			MimeType: enclosure.Type,
			Size:     enclosure.Size(),
		})
	}

	for _, mediaContent := range mediaContents {
		mediaURL := strings.TrimSpace(mediaContent.URL)
		if mediaURL == "" {
			continue
		}

		mediaURL, err := urllib.ResolveToAbsoluteURL(siteURL, mediaURL)
		if err != nil {
			slog.Debug("Unable to build absolute URL for media content",
				slog.String("url", mediaContent.URL),
				slog.String("site_url", siteURL),
				slog.Any("error", err),
			)
			continue
		}

		if _, found := duplicates[mediaURL]; found {
			continue
		}

		duplicates[mediaURL] = true

		enclosures = append(enclosures, &model.Enclosure{
			URL:      mediaURL,
			MimeType: mediaContent.MimeType(),
			Size:     mediaContent.Size(),
		})
	}

	for _, mediaPeerLink := range mediaPeerLinks {
		mediaURL := strings.TrimSpace(mediaPeerLink.URL)
		if mediaURL == "" {
			continue
		}

		mediaURL, err := urllib.ResolveToAbsoluteURL(siteURL, mediaURL)
		if err != nil {
			slog.Debug("Unable to build absolute URL for media peer link",
				slog.String("url", mediaPeerLink.URL),
				slog.String("site_url", siteURL),
				slog.Any("error", err),
			)
			continue
		}

		if _, found := duplicates[mediaURL]; found {
			continue
		}

		duplicates[mediaURL] = true

		enclosures = append(enclosures, &model.Enclosure{
			URL:      mediaURL,
			MimeType: mediaPeerLink.MimeType(),
			Size:     mediaPeerLink.Size(),
		})
	}

	return enclosures
}

// appendSorted is identical to [appendSortedSeq] except receives variadic values rather than [iter.Seq].
func appendSorted[I any, O cmp.Ordered](sorted []O, fn func(I) O, values ...I) []O {
	sorted = slices.Grow(sorted, len(values))
	return appendSortedSeq(sorted, fn, slices.Values(values))
}

// appendSortedSeq appends elements from "values" iterator into "sorted" slice.
//   - "fn" applied to every element of "values"
//   - elements inserted into "sorted" slice so it stays sorted
//   - duplicate elements are not inserted
func appendSortedSeq[I any, O cmp.Ordered](sorted []O, fn func(I) O, values iter.Seq[I]) []O {
	var zero O

	for in := range values {
		out := fn(in)
		if out == zero {
			continue
		}

		where, found := slices.BinarySearch(sorted, out)
		if found {
			continue
		}

		// Insert sorted to avoid duplicates.
		sorted = slices.Insert(sorted, where, out)
	}

	return sorted
}
