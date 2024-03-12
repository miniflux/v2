// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rss // import "miniflux.app/v2/internal/reader/rss"

import (
	"encoding/xml"
	"html"
	"log/slog"
	"path"
	"strconv"
	"strings"
	"time"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/date"
	"miniflux.app/v2/internal/reader/dublincore"
	"miniflux.app/v2/internal/reader/googleplay"
	"miniflux.app/v2/internal/reader/itunes"
	"miniflux.app/v2/internal/reader/media"
	"miniflux.app/v2/internal/reader/sanitizer"
	"miniflux.app/v2/internal/urllib"
)

// Specs: https://www.rssboard.org/rss-specification
type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"rss version,attr"`
	Channel rssChannel `xml:"rss channel"`
}

type rssChannel struct {
	Categories     []string  `xml:"rss category"`
	Title          string    `xml:"rss title"`
	Link           string    `xml:"rss link"`
	ImageURL       string    `xml:"rss image>url"`
	Language       string    `xml:"rss language"`
	Description    string    `xml:"rss description"`
	PubDate        string    `xml:"rss pubDate"`
	ManagingEditor string    `xml:"rss managingEditor"`
	Webmaster      string    `xml:"rss webMaster"`
	TimeToLive     rssTTL    `xml:"rss ttl"`
	Items          []rssItem `xml:"rss item"`
	AtomLinks
	itunes.ItunesFeedElement
	googleplay.GooglePlayFeedElement
}

type rssTTL struct {
	Data string `xml:",chardata"`
}

func (r *rssTTL) Value() int {
	if r.Data == "" {
		return 0
	}

	value, err := strconv.Atoi(r.Data)
	if err != nil {
		return 0
	}

	return value
}

func (r *rssFeed) Transform(baseURL string) *model.Feed {
	var err error

	feed := new(model.Feed)

	siteURL := r.siteURL()
	feed.SiteURL, err = urllib.AbsoluteURL(baseURL, siteURL)
	if err != nil {
		feed.SiteURL = siteURL
	}

	feedURL := r.feedURL()
	feed.FeedURL, err = urllib.AbsoluteURL(baseURL, feedURL)
	if err != nil {
		feed.FeedURL = feedURL
	}

	feed.Title = html.UnescapeString(strings.TrimSpace(r.Channel.Title))
	if feed.Title == "" {
		feed.Title = feed.SiteURL
	}

	feed.IconURL = strings.TrimSpace(r.Channel.ImageURL)
	feed.TTL = r.Channel.TimeToLive.Value()

	for _, item := range r.Channel.Items {
		entry := item.Transform()
		if entry.Author == "" {
			entry.Author = r.feedAuthor()
		}

		if entry.URL == "" {
			entry.URL = feed.SiteURL
		} else {
			entryURL, err := urllib.AbsoluteURL(feed.SiteURL, entry.URL)
			if err == nil {
				entry.URL = entryURL
			}
		}

		if entry.Title == "" {
			entry.Title = sanitizer.TruncateHTML(entry.Content, 100)
		}

		if entry.Title == "" {
			entry.Title = entry.URL
		}

		entry.Tags = append(entry.Tags, r.Channel.Categories...)
		entry.Tags = append(entry.Tags, r.Channel.GetItunesCategories()...)

		if r.Channel.GooglePlayCategory.Text != "" {
			entry.Tags = append(entry.Tags, r.Channel.GooglePlayCategory.Text)
		}

		feed.Entries = append(feed.Entries, entry)
	}

	return feed
}

func (r *rssFeed) siteURL() string {
	return strings.TrimSpace(r.Channel.Link)
}

func (r *rssFeed) feedURL() string {
	for _, atomLink := range r.Channel.AtomLinks.Links {
		if atomLink.Rel == "self" {
			return strings.TrimSpace(atomLink.URL)
		}
	}
	return ""
}

func (r rssFeed) feedAuthor() string {
	var author string
	switch {
	case r.Channel.ItunesAuthor != "":
		author = r.Channel.ItunesAuthor
	case r.Channel.GooglePlayAuthor != "":
		author = r.Channel.GooglePlayAuthor
	case r.Channel.ItunesOwner.String() != "":
		author = r.Channel.ItunesOwner.String()
	case r.Channel.ManagingEditor != "":
		author = r.Channel.ManagingEditor
	case r.Channel.Webmaster != "":
		author = r.Channel.Webmaster
	}
	return sanitizer.StripTags(strings.TrimSpace(author))
}

type rssGUID struct {
	XMLName     xml.Name
	Data        string `xml:",chardata"`
	IsPermaLink string `xml:"isPermaLink,attr"`
}

type rssAuthor struct {
	XMLName xml.Name
	Data    string `xml:",chardata"`
	Inner   string `xml:",innerxml"`
}

type rssEnclosure struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Length string `xml:"length,attr"`
}

func (enclosure *rssEnclosure) Size() int64 {
	if enclosure.Length == "" {
		return 0
	}
	size, _ := strconv.ParseInt(enclosure.Length, 10, 0)
	return size
}

type rssItem struct {
	GUID           rssGUID        `xml:"rss guid"`
	Title          string         `xml:"rss title"`
	Link           string         `xml:"rss link"`
	Description    string         `xml:"rss description"`
	PubDate        string         `xml:"rss pubDate"`
	Author         rssAuthor      `xml:"rss author"`
	Comments       string         `xml:"rss comments"`
	EnclosureLinks []rssEnclosure `xml:"rss enclosure"`
	Categories     []string       `xml:"rss category"`
	dublincore.DublinCoreItemElement
	FeedBurnerElement
	media.Element
	AtomAuthor
	AtomLinks
	itunes.ItunesItemElement
	googleplay.GooglePlayItemElement
}

func (r *rssItem) Transform() *model.Entry {
	entry := model.NewEntry()
	entry.URL = r.entryURL()
	entry.CommentsURL = r.entryCommentsURL()
	entry.Date = r.entryDate()
	entry.Author = r.entryAuthor()
	entry.Hash = r.entryHash()
	entry.Content = r.entryContent()
	entry.Title = r.entryTitle()
	entry.Enclosures = r.entryEnclosures()
	entry.Tags = r.Categories
	if duration, err := normalizeDuration(r.ItunesDuration); err == nil {
		entry.ReadingTime = duration
	}

	return entry
}

func (r *rssItem) entryDate() time.Time {
	value := r.PubDate
	if r.DublinCoreDate != "" {
		value = r.DublinCoreDate
	}

	if value != "" {
		result, err := date.Parse(value)
		if err != nil {
			slog.Debug("Unable to parse date from RSS feed",
				slog.String("date", value),
				slog.String("guid", r.GUID.Data),
				slog.Any("error", err),
			)
			return time.Now()
		}

		return result
	}

	return time.Now()
}

func (r *rssItem) entryAuthor() string {
	var author string

	switch {
	case r.GooglePlayAuthor != "":
		author = r.GooglePlayAuthor
	case r.ItunesAuthor != "":
		author = r.ItunesAuthor
	case r.DublinCoreCreator != "":
		author = r.DublinCoreCreator
	case r.AtomAuthor.String() != "":
		author = r.AtomAuthor.String()
	case strings.Contains(r.Author.Inner, "<![CDATA["):
		author = r.Author.Data
	default:
		author = r.Author.Inner
	}

	return strings.TrimSpace(sanitizer.StripTags(author))
}

func (r *rssItem) entryHash() string {
	for _, value := range []string{r.GUID.Data, r.entryURL()} {
		if value != "" {
			return crypto.Hash(value)
		}
	}

	return ""
}

func (r *rssItem) entryTitle() string {
	title := r.Title

	if r.DublinCoreTitle != "" {
		title = r.DublinCoreTitle
	}

	return html.UnescapeString(strings.TrimSpace(title))
}

func (r *rssItem) entryContent() string {
	for _, value := range []string{
		r.DublinCoreContent,
		r.Description,
		r.GooglePlayDescription,
		r.ItunesSummary,
		r.ItunesSubtitle,
	} {
		if value != "" {
			return value
		}
	}
	return ""
}

func (r *rssItem) entryURL() string {
	for _, link := range []string{r.FeedBurnerLink, r.Link} {
		if link != "" {
			return strings.TrimSpace(link)
		}
	}

	for _, atomLink := range r.AtomLinks.Links {
		if atomLink.URL != "" && (strings.EqualFold(atomLink.Rel, "alternate") || atomLink.Rel == "") {
			return strings.TrimSpace(atomLink.URL)
		}
	}

	// Specs: https://cyber.harvard.edu/rss/rss.html#ltguidgtSubelementOfLtitemgt
	// isPermaLink is optional, its default value is true.
	// If its value is false, the guid may not be assumed to be a url, or a url to anything in particular.
	if r.GUID.IsPermaLink == "true" || r.GUID.IsPermaLink == "" {
		return strings.TrimSpace(r.GUID.Data)
	}

	return ""
}

func (r *rssItem) entryEnclosures() model.EnclosureList {
	enclosures := make(model.EnclosureList, 0)
	duplicates := make(map[string]bool)

	for _, mediaThumbnail := range r.AllMediaThumbnails() {
		if _, found := duplicates[mediaThumbnail.URL]; !found {
			duplicates[mediaThumbnail.URL] = true
			enclosures = append(enclosures, &model.Enclosure{
				URL:      mediaThumbnail.URL,
				MimeType: mediaThumbnail.MimeType(),
				Size:     mediaThumbnail.Size(),
			})
		}
	}

	for _, enclosure := range r.EnclosureLinks {
		enclosureURL := enclosure.URL

		if r.FeedBurnerEnclosureLink != "" {
			filename := path.Base(r.FeedBurnerEnclosureLink)
			if strings.Contains(enclosureURL, filename) {
				enclosureURL = r.FeedBurnerEnclosureLink
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

	for _, mediaContent := range r.AllMediaContents() {
		if _, found := duplicates[mediaContent.URL]; !found {
			duplicates[mediaContent.URL] = true
			enclosures = append(enclosures, &model.Enclosure{
				URL:      mediaContent.URL,
				MimeType: mediaContent.MimeType(),
				Size:     mediaContent.Size(),
			})
		}
	}

	for _, mediaPeerLink := range r.AllMediaPeerLinks() {
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

func (r *rssItem) entryCommentsURL() string {
	commentsURL := strings.TrimSpace(r.Comments)
	if commentsURL != "" && urllib.IsAbsoluteURL(commentsURL) {
		return commentsURL
	}

	return ""
}
