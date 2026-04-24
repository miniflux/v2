// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package opml // import "miniflux.app/v2/internal/reader/opml"

import (
	"encoding/xml"
	"strings"
)

// Specs: http://opml.org/spec2.opml
type opmlDocument struct {
	XMLName  xml.Name              `xml:"opml"`
	Version  string                `xml:"version,attr"`
	Header   opmlHeader            `xml:"head"`
	Outlines opmlOutlineCollection `xml:"body>outline"`
}

type opmlHeader struct {
	Title       string `xml:"title,omitempty"`
	DateCreated string `xml:"dateCreated,omitempty"`
	OwnerName   string `xml:"ownerName,omitempty"`
}

type opmlOutline struct {
	Title       string                `xml:"title,attr,omitempty"`
	Text        string                `xml:"text,attr"`
	FeedURL     string                `xml:"xmlUrl,attr,omitempty"`
	SiteURL     string                `xml:"htmlUrl,attr,omitempty"`
	Description string                `xml:"description,attr,omitempty"`
	Outlines    opmlOutlineCollection `xml:"outline,omitempty"`

	// Miniflux-specific feed settings
	ScraperRules                string `xml:"scraperRules,attr,omitempty"`
	RewriteRules                string `xml:"rewriteRules,attr,omitempty"`
	UrlRewriteRules             string `xml:"urlRewriteRules,attr,omitempty"`
	BlocklistRules              string `xml:"blocklistRules,attr,omitempty"`
	KeeplistRules               string `xml:"keeplistRules,attr,omitempty"`
	BlockFilterEntryRules       string `xml:"blockFilterEntryRules,attr,omitempty"`
	KeepFilterEntryRules        string `xml:"keepFilterEntryRules,attr,omitempty"`
	UserAgent                   string `xml:"userAgent,attr,omitempty"`
	Cookie                      string `xml:"cookie,attr,omitempty"`
	ProxyURL                    string `xml:"proxyUrl,attr,omitempty"`
	Crawler                     bool   `xml:"crawler,attr,omitempty"`
	IgnoreHTTPCache             bool   `xml:"ignoreHTTPCache,attr,omitempty"`
	FetchViaProxy               bool   `xml:"fetchViaProxy,attr,omitempty"`
	Disabled                    bool   `xml:"disabled,attr,omitempty"`
	NoMediaPlayer               bool   `xml:"noMediaPlayer,attr,omitempty"`
	HideGlobally                bool   `xml:"hideGlobally,attr,omitempty"`
	AllowSelfSignedCertificates bool   `xml:"allowSelfSignedCertificates,attr,omitempty"`
	DisableHTTP2                bool   `xml:"disableHTTP2,attr,omitempty"`
	IgnoreEntryUpdates          bool   `xml:"ignoreEntryUpdates,attr,omitempty"`
}

func (o opmlOutline) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type opmlOutlineXml opmlOutline

	outlineType := ""
	if o.IsSubscription() {
		outlineType = "rss"
	}

	return e.EncodeElement(struct {
		opmlOutlineXml
		Type string `xml:"type,attr,omitempty"`
	}{
		opmlOutlineXml: opmlOutlineXml(o),
		Type:           outlineType,
	}, start)
}

func (o opmlOutline) IsSubscription() bool {
	return strings.TrimSpace(o.FeedURL) != ""
}

func (o opmlOutline) GetTitle() string {
	if o.Title != "" {
		return o.Title
	}

	if o.Text != "" {
		return o.Text
	}

	if o.SiteURL != "" {
		return o.SiteURL
	}

	if o.FeedURL != "" {
		return o.FeedURL
	}

	return ""
}

func (o opmlOutline) GetSiteURL() string {
	if o.SiteURL != "" {
		return o.SiteURL
	}

	return o.FeedURL
}

type opmlOutlineCollection []opmlOutline

func (o opmlOutlineCollection) HasChildren() bool {
	return len(o) > 0
}
