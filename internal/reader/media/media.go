// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package media // import "miniflux.app/v2/internal/reader/media"

import (
	"regexp"
	"strconv"
	"strings"
)

var textLinkRegex = regexp.MustCompile(`(?mi)(\bhttps?:\/\/[-A-Z0-9+&@#\/%?=~_|!:,.;]*[-A-Z0-9+&@#\/%=~_|])`)

// Specs: https://www.rssboard.org/media-rss
type MediaItemElement struct {
	MediaCategories   MediaCategoryList `xml:"http://search.yahoo.com/mrss/ category"`
	MediaGroups       []Group           `xml:"http://search.yahoo.com/mrss/ group"`
	MediaContents     []Content         `xml:"http://search.yahoo.com/mrss/ content"`
	MediaThumbnails   []Thumbnail       `xml:"http://search.yahoo.com/mrss/ thumbnail"`
	MediaDescriptions DescriptionList   `xml:"http://search.yahoo.com/mrss/ description"`
	MediaPeerLinks    []PeerLink        `xml:"http://search.yahoo.com/mrss/ peerLink"`
}

// AllMediaThumbnails returns all thumbnail elements merged together.
func (e *MediaItemElement) AllMediaThumbnails() []Thumbnail {
	var items []Thumbnail
	items = append(items, e.MediaThumbnails...)
	for _, mediaGroup := range e.MediaGroups {
		items = append(items, mediaGroup.MediaThumbnails...)
	}
	return items
}

// AllMediaContents returns all content elements merged together.
func (e *MediaItemElement) AllMediaContents() []Content {
	var items []Content
	items = append(items, e.MediaContents...)
	for _, mediaGroup := range e.MediaGroups {
		items = append(items, mediaGroup.MediaContents...)
	}
	return items
}

// AllMediaPeerLinks returns all peer link elements merged together.
func (e *MediaItemElement) AllMediaPeerLinks() []PeerLink {
	var items []PeerLink
	items = append(items, e.MediaPeerLinks...)
	for _, mediaGroup := range e.MediaGroups {
		items = append(items, mediaGroup.MediaPeerLinks...)
	}
	return items
}

// FirstMediaDescription returns the first description element.
func (e *MediaItemElement) FirstMediaDescription() string {
	description := e.MediaDescriptions.First()
	if description != "" {
		return description
	}

	for _, mediaGroup := range e.MediaGroups {
		description = mediaGroup.MediaDescriptions.First()
		if description != "" {
			return description
		}
	}

	return ""
}

// Group represents a XML element "media:group".
type Group struct {
	MediaContents     []Content       `xml:"http://search.yahoo.com/mrss/ content"`
	MediaThumbnails   []Thumbnail     `xml:"http://search.yahoo.com/mrss/ thumbnail"`
	MediaDescriptions DescriptionList `xml:"http://search.yahoo.com/mrss/ description"`
	MediaPeerLinks    []PeerLink      `xml:"http://search.yahoo.com/mrss/ peerLink"`
}

// Content represents a XML element "media:content".
type Content struct {
	URL      string `xml:"url,attr"`
	Type     string `xml:"type,attr"`
	FileSize string `xml:"fileSize,attr"`
	Medium   string `xml:"medium,attr"`
}

// MimeType returns the attachment mime type.
func (mc *Content) MimeType() string {
	if mc.Type != "" {
		return mc.Type
	}

	switch mc.Medium {
	case "image":
		return "image/*"
	case "video":
		return "video/*"
	case "audio":
		return "audio/*"
	default:
		return "application/octet-stream"
	}
}

// Size returns the attachment size.
func (mc *Content) Size() int64 {
	size, _ := strconv.ParseInt(mc.FileSize, 10, 0)
	return size
}

// Thumbnail represents a XML element "media:thumbnail".
type Thumbnail struct {
	URL string `xml:"url,attr"`
}

// MimeType returns the attachment mime type.
func (t *Thumbnail) MimeType() string {
	return "image/*"
}

// Size returns the attachment size.
func (t *Thumbnail) Size() int64 {
	return 0
}

// PeerLink represents a XML element "media:peerLink".
type PeerLink struct {
	URL  string `xml:"href,attr"`
	Type string `xml:"type,attr"`
}

// MimeType returns the attachment mime type.
func (p *PeerLink) MimeType() string {
	if p.Type != "" {
		return p.Type
	}
	return "application/octet-stream"
}

// Size returns the attachment size.
func (p *PeerLink) Size() int64 {
	return 0
}

// Description represents a XML element "media:description".
type Description struct {
	Type        string `xml:"type,attr"`
	Description string `xml:",chardata"`
}

// HTML returns the description as HTML.
func (d *Description) HTML() string {
	if d.Type == "html" {
		return d.Description
	}

	content := strings.ReplaceAll(d.Description, "\n", "<br>")
	return textLinkRegex.ReplaceAllString(content, `<a href="${1}">${1}</a>`)
}

// DescriptionList represents a list of "media:description" XML elements.
type DescriptionList []Description

// First returns the first non-empty description.
func (dl DescriptionList) First() string {
	for _, description := range dl {
		contents := description.HTML()
		if contents != "" {
			return contents
		}
	}
	return ""
}

type MediaCategoryList []MediaCategory

func (mcl MediaCategoryList) Labels() []string {
	var labels []string
	for _, category := range mcl {
		label := strings.TrimSpace(category.Label)
		if label != "" {
			labels = append(labels, label)
		}
	}
	return labels
}

type MediaCategory struct {
	Label string `xml:"label,attr"`
}
