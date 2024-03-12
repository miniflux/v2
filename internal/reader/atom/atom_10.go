// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package atom // import "miniflux.app/v2/internal/reader/atom"

import (
	"encoding/xml"
	"html"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/date"
	"miniflux.app/v2/internal/reader/media"
	"miniflux.app/v2/internal/reader/sanitizer"
	"miniflux.app/v2/internal/urllib"
)

// Specs:
// https://tools.ietf.org/html/rfc4287
// https://validator.w3.org/feed/docs/atom.html
type atom10Feed struct {
	XMLName xml.Name      `xml:"http://www.w3.org/2005/Atom feed"`
	ID      string        `xml:"id"`
	Title   atom10Text    `xml:"title"`
	Authors atomAuthors   `xml:"author"`
	Icon    string        `xml:"icon"`
	Links   atomLinks     `xml:"link"`
	Entries []atom10Entry `xml:"entry"`
}

func (a *atom10Feed) Transform(baseURL string) *model.Feed {
	var err error

	feed := new(model.Feed)

	feedURL := a.Links.firstLinkWithRelation("self")
	feed.FeedURL, err = urllib.AbsoluteURL(baseURL, feedURL)
	if err != nil {
		feed.FeedURL = feedURL
	}

	siteURL := a.Links.originalLink()
	feed.SiteURL, err = urllib.AbsoluteURL(baseURL, siteURL)
	if err != nil {
		feed.SiteURL = siteURL
	}

	feed.Title = html.UnescapeString(a.Title.String())
	if feed.Title == "" {
		feed.Title = feed.SiteURL
	}

	feed.IconURL = strings.TrimSpace(a.Icon)

	for _, entry := range a.Entries {
		item := entry.Transform()
		entryURL, err := urllib.AbsoluteURL(feed.SiteURL, item.URL)
		if err == nil {
			item.URL = entryURL
		}

		if item.Author == "" {
			item.Author = a.Authors.String()
		}

		if item.Title == "" {
			item.Title = sanitizer.TruncateHTML(item.Content, 100)
		}

		if item.Title == "" {
			item.Title = item.URL
		}

		feed.Entries = append(feed.Entries, item)
	}

	return feed
}

type atom10Entry struct {
	ID         string           `xml:"id"`
	Title      atom10Text       `xml:"title"`
	Published  string           `xml:"published"`
	Updated    string           `xml:"updated"`
	Links      atomLinks        `xml:"link"`
	Summary    atom10Text       `xml:"summary"`
	Content    atom10Text       `xml:"http://www.w3.org/2005/Atom content"`
	Authors    atomAuthors      `xml:"author"`
	Categories []atom10Category `xml:"category"`
	media.Element
}

func (a *atom10Entry) Transform() *model.Entry {
	entry := model.NewEntry()
	entry.URL = a.Links.originalLink()
	entry.Date = a.entryDate()
	entry.Author = a.Authors.String()
	entry.Hash = a.entryHash()
	entry.Content = a.entryContent()
	entry.Title = a.entryTitle()
	entry.Enclosures = a.entryEnclosures()
	entry.CommentsURL = a.entryCommentsURL()
	entry.Tags = a.entryCategories()
	return entry
}

func (a *atom10Entry) entryTitle() string {
	return html.UnescapeString(a.Title.String())
}

func (a *atom10Entry) entryContent() string {
	content := a.Content.String()
	if content != "" {
		return content
	}

	summary := a.Summary.String()
	if summary != "" {
		return summary
	}

	mediaDescription := a.FirstMediaDescription()
	if mediaDescription != "" {
		return mediaDescription
	}

	return ""
}

// Note: The published date represents the original creation date for YouTube feeds.
// Example:
// <published>2019-01-26T08:02:28+00:00</published>
// <updated>2019-01-29T07:27:27+00:00</updated>
func (a *atom10Entry) entryDate() time.Time {
	dateText := a.Published
	if dateText == "" {
		dateText = a.Updated
	}

	if dateText != "" {
		result, err := date.Parse(dateText)
		if err != nil {
			slog.Debug("Unable to parse date from Atom 0.3 feed",
				slog.String("date", dateText),
				slog.String("id", a.ID),
				slog.Any("error", err),
			)
			return time.Now()
		}

		return result
	}

	return time.Now()
}

func (a *atom10Entry) entryHash() string {
	for _, value := range []string{a.ID, a.Links.originalLink()} {
		if value != "" {
			return crypto.Hash(value)
		}
	}

	return ""
}

func (a *atom10Entry) entryEnclosures() model.EnclosureList {
	enclosures := make(model.EnclosureList, 0)
	duplicates := make(map[string]bool)

	for _, mediaThumbnail := range a.AllMediaThumbnails() {
		if _, found := duplicates[mediaThumbnail.URL]; !found {
			duplicates[mediaThumbnail.URL] = true
			enclosures = append(enclosures, &model.Enclosure{
				URL:      mediaThumbnail.URL,
				MimeType: mediaThumbnail.MimeType(),
				Size:     mediaThumbnail.Size(),
			})
		}
	}

	for _, link := range a.Links {
		if strings.EqualFold(link.Rel, "enclosure") {
			if link.URL == "" {
				continue
			}

			if _, found := duplicates[link.URL]; !found {
				duplicates[link.URL] = true
				length, _ := strconv.ParseInt(link.Length, 10, 0)
				enclosures = append(enclosures, &model.Enclosure{URL: link.URL, MimeType: link.Type, Size: length})
			}
		}
	}

	for _, mediaContent := range a.AllMediaContents() {
		if _, found := duplicates[mediaContent.URL]; !found {
			duplicates[mediaContent.URL] = true
			enclosures = append(enclosures, &model.Enclosure{
				URL:      mediaContent.URL,
				MimeType: mediaContent.MimeType(),
				Size:     mediaContent.Size(),
			})
		}
	}

	for _, mediaPeerLink := range a.AllMediaPeerLinks() {
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

func (r *atom10Entry) entryCategories() []string {
	categoryList := make([]string, 0)

	for _, atomCategory := range r.Categories {
		if strings.TrimSpace(atomCategory.Label) != "" {
			categoryList = append(categoryList, strings.TrimSpace(atomCategory.Label))
		} else {
			categoryList = append(categoryList, strings.TrimSpace(atomCategory.Term))
		}
	}

	return categoryList
}

// See https://tools.ietf.org/html/rfc4685#section-4
// If the type attribute of the atom:link is omitted, its value is assumed to be "application/atom+xml".
// We accept only HTML or XHTML documents for now since the intention is to have the same behavior as RSS.
func (a *atom10Entry) entryCommentsURL() string {
	commentsURL := a.Links.firstLinkWithRelationAndType("replies", "text/html", "application/xhtml+xml")
	if urllib.IsAbsoluteURL(commentsURL) {
		return commentsURL
	}
	return ""
}

type atom10Text struct {
	Type             string               `xml:"type,attr"`
	CharData         string               `xml:",chardata"`
	InnerXML         string               `xml:",innerxml"`
	XHTMLRootElement atomXHTMLRootElement `xml:"http://www.w3.org/1999/xhtml div"`
}

type atom10Category struct {
	Term  string `xml:"term,attr"`
	Label string `xml:"label,attr"`
}

// Text: https://datatracker.ietf.org/doc/html/rfc4287#section-3.1.1.1
// HTML: https://datatracker.ietf.org/doc/html/rfc4287#section-3.1.1.2
// XHTML: https://datatracker.ietf.org/doc/html/rfc4287#section-3.1.1.3
func (a *atom10Text) String() string {
	var content string
	switch {
	case a.Type == "", a.Type == "text", a.Type == "text/plain":
		if strings.HasPrefix(strings.TrimSpace(a.InnerXML), `<![CDATA[`) {
			content = html.EscapeString(a.CharData)
		} else {
			content = a.InnerXML
		}
	case a.Type == "xhtml":
		var root = a.XHTMLRootElement
		if root.XMLName.Local == "div" {
			content = root.InnerXML
		} else {
			content = a.InnerXML
		}
	default:
		content = a.CharData
	}

	return strings.TrimSpace(content)
}

type atomXHTMLRootElement struct {
	XMLName  xml.Name `xml:"div"`
	InnerXML string   `xml:",innerxml"`
}
