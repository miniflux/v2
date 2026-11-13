// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rssfeed // import "miniflux.app/v2/internal/rssfeed"

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/mediaproxy"
	"miniflux.app/v2/internal/model"
)

const (
	generatorName = "Miniflux"

	// documentOverhead approximates the size of the feed-level elements, and
	// entryOverhead the size of the per-entry elements (links, dates,
	// identifiers, categories) that do not depend on the entry content.
	documentOverhead = 1024
	entryOverhead    = 512

	rssContentNamespace = "http://purl.org/rss/1.0/modules/content/"
	rssAtomNamespace    = "http://www.w3.org/2005/Atom"
	rssDCNamespace      = "http://purl.org/dc/elements/1.1/"
	atomNamespace       = "http://www.w3.org/2005/Atom"
)

// feedMetadata holds everything needed to render a feed document.
type feedMetadata struct {
	Title       string
	Description string
	Link        string
	ID          string
	SelfURL     string
	IconURL     string
	Language    string
	Updated     time.Time
	Entries     model.Entries
}

type rssDocument struct {
	XMLName          xml.Name   `xml:"rss"`
	Version          string     `xml:"version,attr"`
	ContentNamespace string     `xml:"xmlns:content,attr"`
	AtomNamespace    string     `xml:"xmlns:atom,attr"`
	DCNamespace      string     `xml:"xmlns:dc,attr"`
	Channel          rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title         string      `xml:"title"`
	Link          string      `xml:"link"`
	Description   string      `xml:"description"`
	Language      string      `xml:"language,omitempty"`
	LastBuildDate string      `xml:"lastBuildDate"`
	Generator     string      `xml:"generator"`
	Image         *rssImage   `xml:"image,omitempty"`
	AtomLink      rssAtomLink `xml:"atom:link"`
	Items         []rssItem   `xml:"item"`
}

type rssImage struct {
	URL   string `xml:"url"`
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

type rssAtomLink struct {
	HREF string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type rssItem struct {
	Title       string         `xml:"title"`
	Link        string         `xml:"link"`
	GUID        rssGUID        `xml:"guid"`
	PubDate     string         `xml:"pubDate"`
	DCreator    string         `xml:"dc:creator,omitempty"`
	Categories  []string       `xml:"category,omitempty"`
	Comments    string         `xml:"comments,omitempty"`
	Description string         `xml:"description"`
	Content     string         `xml:"content:encoded,omitempty"`
	Enclosures  []rssEnclosure `xml:"enclosure,omitempty"`
}

type rssGUID struct {
	IsPermaLink string `xml:"isPermaLink,attr"`
	Value       string `xml:",chardata"`
}

type rssEnclosure struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr,omitempty"`
	Length int64  `xml:"length,attr"`
}

type atomFeed struct {
	XMLName   xml.Name    `xml:"feed"`
	XMLNS     string      `xml:"xmlns,attr"`
	Title     string      `xml:"title"`
	Subtitle  string      `xml:"subtitle,omitempty"`
	ID        string      `xml:"id"`
	Updated   string      `xml:"updated"`
	Icon      string      `xml:"icon,omitempty"`
	Logo      string      `xml:"logo,omitempty"`
	Generator string      `xml:"generator"`
	Author    *atomAuthor `xml:"author,omitempty"`
	Links     []atomLink  `xml:"link"`
	Entries   []atomEntry `xml:"entry"`
}

type atomEntry struct {
	ID         string         `xml:"id"`
	Title      string         `xml:"title"`
	Published  string         `xml:"published"`
	Updated    string         `xml:"updated"`
	Author     *atomAuthor    `xml:"author,omitempty"`
	Links      []atomLink     `xml:"link"`
	Categories []atomCategory `xml:"category,omitempty"`
	Summary    atomText       `xml:"summary"`
	Content    atomText       `xml:"content"`
}

type atomAuthor struct {
	Name  string `xml:"name"`
	Email string `xml:"email,omitempty"`
	URI   string `xml:"uri,omitempty"`
}

type atomLink struct {
	HREF   string `xml:"href,attr"`
	Rel    string `xml:"rel,attr,omitempty"`
	Type   string `xml:"type,attr,omitempty"`
	Length int64  `xml:"length,attr,omitempty"`
}

type atomCategory struct {
	Term  string `xml:"term,attr"`
	Label string `xml:"label,attr,omitempty"`
}

type atomText struct {
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

// renderRSS returns the RSS 2.0 representation of the given feed.
func renderRSS(meta feedMetadata) ([]byte, error) {
	now := meta.Updated
	if now.IsZero() {
		now = time.Now()
	}

	channel := rssChannel{
		Title:         meta.Title,
		Link:          meta.Link,
		Description:   meta.Description,
		Language:      meta.Language,
		LastBuildDate: now.UTC().Format(time.RFC1123Z),
		Generator:     generatorName,
		AtomLink: rssAtomLink{
			HREF: meta.SelfURL,
			Rel:  "self",
			Type: "application/rss+xml",
		},
		Items: make([]rssItem, 0, len(meta.Entries)),
	}

	if meta.IconURL != "" {
		channel.Image = &rssImage{URL: meta.IconURL, Title: meta.Title, Link: meta.Link}
	}

	// Every item carries its content twice (description and content:encoded),
	// so the estimate accounts for both copies.
	estimatedSize := documentOverhead

	for _, entry := range meta.Entries {
		content := entryContent(entry)
		estimatedSize += 2*len(content) + entryOverhead

		item := rssItem{
			Title:       entryTitle(entry),
			Link:        entry.URL,
			GUID:        rssGUID{IsPermaLink: "false", Value: entryGUID(entry)},
			PubDate:     entryPublishedAt(entry).UTC().Format(time.RFC1123Z),
			DCreator:    entry.Author,
			Comments:    entry.CommentsURL,
			Description: content,
			Content:     content,
			Categories:  entryCategories(entry),
		}

		for _, enclosure := range entry.Enclosures {
			item.Enclosures = append(item.Enclosures, rssEnclosure{
				URL:    enclosure.URL,
				Type:   enclosure.MimeType,
				Length: enclosure.Size,
			})
		}

		channel.Items = append(channel.Items, item)
	}

	document := rssDocument{
		Version:          "2.0",
		ContentNamespace: rssContentNamespace,
		AtomNamespace:    rssAtomNamespace,
		DCNamespace:      rssDCNamespace,
		Channel:          channel,
	}

	payload, err := encodeXMLDocument(document, estimatedSize)
	if err != nil {
		return nil, fmt.Errorf("rssfeed: unable to marshal RSS document: %w", err)
	}

	return payload, nil
}

// renderAtom returns the Atom 1.0 representation of the given feed.
func renderAtom(meta feedMetadata) ([]byte, error) {
	now := meta.Updated
	if now.IsZero() {
		now = time.Now()
	}

	document := atomFeed{
		XMLNS:     atomNamespace,
		Title:     meta.Title,
		Subtitle:  meta.Description,
		ID:        meta.ID,
		Updated:   now.UTC().Format(time.RFC3339),
		Generator: generatorName,
		Links: []atomLink{
			{HREF: meta.Link, Rel: "alternate", Type: "text/html"},
			{HREF: meta.SelfURL, Rel: "self", Type: "application/atom+xml"},
		},
		Entries: make([]atomEntry, 0, len(meta.Entries)),
	}

	if meta.IconURL != "" {
		document.Icon = meta.IconURL
		document.Logo = meta.IconURL
	}

	estimatedSize := documentOverhead

	for _, entry := range meta.Entries {
		publishedAt := entryPublishedAt(entry)
		updatedAt := entry.ChangedAt
		if updatedAt.IsZero() {
			updatedAt = publishedAt
		}

		content := entryContent(entry)
		estimatedSize += 2*len(content) + entryOverhead

		item := atomEntry{
			ID:        entryGUID(entry),
			Title:     entryTitle(entry),
			Published: publishedAt.UTC().Format(time.RFC3339),
			Updated:   updatedAt.UTC().Format(time.RFC3339),
			Summary:   atomText{Type: "html", Content: content},
			Content:   atomText{Type: "html", Content: content},
			Links: []atomLink{
				{HREF: entry.URL, Rel: "alternate", Type: "text/html"},
			},
		}

		if entry.Author != "" {
			item.Author = &atomAuthor{Name: entry.Author}
		}

		for _, enclosure := range entry.Enclosures {
			item.Links = append(item.Links, atomLink{
				HREF:   enclosure.URL,
				Rel:    "enclosure",
				Type:   enclosure.MimeType,
				Length: enclosure.Size,
			})
		}

		if entry.Feed != nil && entry.Feed.Title != "" {
			item.Categories = append(item.Categories, atomCategory{Term: entry.Feed.Title})
		}

		if entry.Feed != nil && entry.Feed.Category != nil && entry.Feed.Category.Title != "" {
			item.Categories = append(item.Categories, atomCategory{Term: entry.Feed.Category.Title})
		}

		document.Entries = append(document.Entries, item)
	}

	// Atom requires a feed-level author when entries do not all provide one.
	document.Author = &atomAuthor{Name: generatorName}
	if len(meta.Entries) > 0 && meta.Entries[0].Feed != nil && meta.Entries[0].Feed.Title != "" {
		document.Author = &atomAuthor{Name: meta.Entries[0].Feed.Title}
	}

	payload, err := encodeXMLDocument(document, estimatedSize)
	if err != nil {
		return nil, fmt.Errorf("rssfeed: unable to marshal Atom document: %w", err)
	}

	return payload, nil
}

// encodeXMLDocument serializes the document into a buffer of the estimated
// size, so that rendering a large feed does not repeatedly grow and copy the
// payload. The output is not indented: feed clients do not need it, and
// indentation adds work and bytes to every response.
func encodeXMLDocument(document any, estimatedSize int) ([]byte, error) {
	// Escaping HTML content into XML character data expands it noticeably (a
	// single "<" becomes "&lt;"), so leave enough headroom for the buffer to
	// hold the whole document without reallocating it.
	buffer := bytes.NewBuffer(make([]byte, 0, estimatedSize+estimatedSize/3+4096))
	buffer.WriteString(xml.Header)

	if err := xml.NewEncoder(buffer).Encode(document); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func entryGUID(entry *model.Entry) string {
	return "urn:miniflux:entry:" + strconv.FormatInt(entry.ID, 10)
}

func entryTitle(entry *model.Entry) string {
	if strings.TrimSpace(entry.Title) != "" {
		return entry.Title
	}
	if entry.URL != "" {
		return entry.URL
	}
	return "(untitled)"
}

func entryPublishedAt(entry *model.Entry) time.Time {
	if !entry.Date.IsZero() {
		return entry.Date
	}
	if !entry.CreatedAt.IsZero() {
		return entry.CreatedAt
	}
	return time.Now()
}

func entryContent(entry *model.Entry) string {
	return mediaproxy.RewriteDocumentWithAbsoluteProxyURL(entry.Content)
}

func entryCategories(entry *model.Entry) []string {
	var categories []string
	if entry.Feed != nil {
		if entry.Feed.Title != "" {
			categories = append(categories, entry.Feed.Title)
		}
		if entry.Feed.Category != nil && entry.Feed.Category.Title != "" {
			categories = append(categories, entry.Feed.Category.Title)
		}
	}
	return categories
}

func feedIconURL(feed *model.Feed) string {
	if feed == nil || feed.Icon == nil || feed.Icon.ExternalIconID == "" {
		return ""
	}
	return strings.TrimRight(config.Opts.BaseURL(), "/") + "/feed-icon/" + feed.Icon.ExternalIconID
}
