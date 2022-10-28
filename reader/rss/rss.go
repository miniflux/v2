// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rss // import "miniflux.app/reader/rss"

import (
	"encoding/xml"
	"html"
	"path"
	"strconv"
	"strings"
	"time"

	"miniflux.app/crypto"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/reader/date"
	"miniflux.app/reader/media"
	"miniflux.app/reader/sanitizer"
	"miniflux.app/url"
)

// Specs: https://cyber.harvard.edu/rss/rss.html
type rssFeed struct {
	XMLName        xml.Name  `xml:"rss"`
	Version        string    `xml:"version,attr"`
	Title          string    `xml:"channel>title"`
	Links          []rssLink `xml:"channel>link"`
	Language       string    `xml:"channel>language"`
	Description    string    `xml:"channel>description"`
	PubDate        string    `xml:"channel>pubDate"`
	ManagingEditor string    `xml:"channel>managingEditor"`
	Webmaster      string    `xml:"channel>webMaster"`
	Items          []rssItem `xml:"channel>item"`
	PodcastFeedElement
}

func (r *rssFeed) Transform(baseURL string) *model.Feed {
	var err error

	feed := new(model.Feed)

	siteURL := r.siteURL()
	feed.SiteURL, err = url.AbsoluteURL(baseURL, siteURL)
	if err != nil {
		feed.SiteURL = siteURL
	}

	feedURL := r.feedURL()
	feed.FeedURL, err = url.AbsoluteURL(baseURL, feedURL)
	if err != nil {
		feed.FeedURL = feedURL
	}

	feed.Title = html.UnescapeString(strings.TrimSpace(r.Title))
	if feed.Title == "" {
		feed.Title = feed.SiteURL
	}

	for _, item := range r.Items {
		entry := item.Transform()
		if entry.Author == "" {
			entry.Author = r.feedAuthor()
		}

		if entry.URL == "" {
			entry.URL = feed.SiteURL
		} else {
			entryURL, err := url.AbsoluteURL(feed.SiteURL, entry.URL)
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

		feed.Entries = append(feed.Entries, entry)
	}

	return feed
}

func (r *rssFeed) siteURL() string {
	for _, element := range r.Links {
		if element.XMLName.Space == "" {
			return strings.TrimSpace(element.Data)
		}
	}

	return ""
}

func (r *rssFeed) feedURL() string {
	for _, element := range r.Links {
		if element.XMLName.Space == "http://www.w3.org/2005/Atom" {
			return strings.TrimSpace(element.Href)
		}
	}

	return ""
}

func (r rssFeed) feedAuthor() string {
	author := r.PodcastAuthor()
	switch {
	case r.ManagingEditor != "":
		author = r.ManagingEditor
	case r.Webmaster != "":
		author = r.Webmaster
	}
	return sanitizer.StripTags(strings.TrimSpace(author))
}

type rssGUID struct {
	XMLName     xml.Name
	Data        string `xml:",chardata"`
	IsPermaLink string `xml:"isPermaLink,attr"`
}

type rssLink struct {
	XMLName xml.Name
	Data    string `xml:",chardata"`
	Href    string `xml:"href,attr"`
	Rel     string `xml:"rel,attr"`
}

type rssCommentLink struct {
	XMLName xml.Name
	Data    string `xml:",chardata"`
}

type rssAuthor struct {
	XMLName xml.Name
	Data    string `xml:",chardata"`
	Name    string `xml:"name"`
	Email   string `xml:"email"`
	Inner   string `xml:",innerxml"`
}

type rssTitle struct {
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
	GUID           rssGUID          `xml:"guid"`
	Title          []rssTitle       `xml:"title"`
	Links          []rssLink        `xml:"link"`
	Description    string           `xml:"description"`
	PubDate        string           `xml:"pubDate"`
	Authors        []rssAuthor      `xml:"author"`
	CommentLinks   []rssCommentLink `xml:"comments"`
	EnclosureLinks []rssEnclosure   `xml:"enclosure"`
	DublinCoreElement
	FeedBurnerElement
	PodcastEntryElement
	media.Element
}

func (r *rssItem) Transform() *model.Entry {
	entry := new(model.Entry)
	entry.URL = r.entryURL()
	entry.CommentsURL = r.entryCommentsURL()
	entry.Date = r.entryDate()
	entry.Author = r.entryAuthor()
	entry.Hash = r.entryHash()
	entry.Content = r.entryContent()
	entry.Title = r.entryTitle()
	entry.Enclosures = r.entryEnclosures()
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
			logger.Error("rss: %v (entry GUID = %s)", err, r.GUID)
			return time.Now()
		}

		return result
	}

	return time.Now()
}

func (r *rssItem) entryAuthor() string {
	author := ""

	for _, rssAuthor := range r.Authors {
		switch rssAuthor.XMLName.Space {
		case "http://www.itunes.com/dtds/podcast-1.0.dtd", "http://www.google.com/schemas/play-podcasts/1.0":
			author = rssAuthor.Data
		case "http://www.w3.org/2005/Atom":
			if rssAuthor.Name != "" {
				author = rssAuthor.Name
			} else if rssAuthor.Email != "" {
				author = rssAuthor.Email
			}
		default:
			if rssAuthor.Name != "" {
				author = rssAuthor.Name
			} else if strings.Contains(rssAuthor.Inner, "<![CDATA[") {
				author = rssAuthor.Data
			} else {
				author = rssAuthor.Inner
			}
		}
	}

	if author == "" {
		author = r.DublinCoreCreator
	}

	return sanitizer.StripTags(strings.TrimSpace(author))
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
	var title string

	for _, rssTitle := range r.Title {
		switch rssTitle.XMLName.Space {
		case "http://search.yahoo.com/mrss/":
			// Ignore title in media namespace
		case "http://purl.org/dc/elements/1.1/":
			title = rssTitle.Data
		default:
			title = rssTitle.Data
		}

		if title != "" {
			break
		}
	}

	return html.UnescapeString(strings.TrimSpace(title))
}

func (r *rssItem) entryContent() string {
	for _, value := range []string{r.DublinCoreContent, r.Description, r.PodcastDescription()} {
		if value != "" {
			return value
		}
	}
	return ""
}

func (r *rssItem) entryURL() string {
	if r.FeedBurnerLink != "" {
		return r.FeedBurnerLink
	}

	for _, link := range r.Links {
		if link.XMLName.Space == "http://www.w3.org/2005/Atom" && link.Href != "" && isValidLinkRelation(link.Rel) {
			return strings.TrimSpace(link.Href)
		}

		if link.Data != "" {
			return strings.TrimSpace(link.Data)
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
	for _, commentLink := range r.CommentLinks {
		if commentLink.XMLName.Space == "" {
			commentsURL := strings.TrimSpace(commentLink.Data)
			// The comments URL is supposed to be absolute (some feeds publishes incorrect comments URL)
			// See https://cyber.harvard.edu/rss/rss.html#ltcommentsgtSubelementOfLtitemgt
			if url.IsAbsoluteURL(commentsURL) {
				return commentsURL
			}
		}
	}

	return ""
}

func isValidLinkRelation(rel string) bool {
	switch rel {
	case "", "alternate", "enclosure", "related", "self", "via":
		return true
	default:
		if strings.HasPrefix(rel, "http") {
			return true
		}
		return false
	}
}
