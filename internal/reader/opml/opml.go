// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package opml // import "miniflux.app/v2/internal/reader/opml"

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

const minifluxOPMLNamespace = "https://miniflux.app/opml"

// Specs: http://opml.org/spec2.opml
type opmlDocument struct {
	XMLName           xml.Name              `xml:"opml"`
	Version           string                `xml:"version,attr"`
	MinifluxNamespace string                `xml:"xmlns:miniflux,attr,omitempty"`
	Header            opmlHeader            `xml:"head"`
	Outlines          opmlOutlineCollection `xml:"body>outline"`
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

type opmlOutlineXML struct {
	Title       string                `xml:"title,attr,omitempty"`
	Text        string                `xml:"text,attr"`
	Type        string                `xml:"type,attr,omitempty"`
	FeedURL     string                `xml:"xmlUrl,attr,omitempty"`
	SiteURL     string                `xml:"htmlUrl,attr,omitempty"`
	Description string                `xml:"description,attr,omitempty"`
	Outlines    opmlOutlineCollection `xml:"outline,omitempty"`

	ScraperRules                string `xml:"miniflux:scraperRules,attr,omitempty"`
	RewriteRules                string `xml:"miniflux:rewriteRules,attr,omitempty"`
	UrlRewriteRules             string `xml:"miniflux:urlRewriteRules,attr,omitempty"`
	BlocklistRules              string `xml:"miniflux:blocklistRules,attr,omitempty"`
	KeeplistRules               string `xml:"miniflux:keeplistRules,attr,omitempty"`
	BlockFilterEntryRules       string `xml:"miniflux:blockFilterEntryRules,attr,omitempty"`
	KeepFilterEntryRules        string `xml:"miniflux:keepFilterEntryRules,attr,omitempty"`
	UserAgent                   string `xml:"miniflux:userAgent,attr,omitempty"`
	ProxyURL                    string `xml:"miniflux:proxyUrl,attr,omitempty"`
	Crawler                     bool   `xml:"miniflux:crawler,attr,omitempty"`
	IgnoreHTTPCache             bool   `xml:"miniflux:ignoreHTTPCache,attr,omitempty"`
	FetchViaProxy               bool   `xml:"miniflux:fetchViaProxy,attr,omitempty"`
	Disabled                    bool   `xml:"miniflux:disabled,attr,omitempty"`
	NoMediaPlayer               bool   `xml:"miniflux:noMediaPlayer,attr,omitempty"`
	HideGlobally                bool   `xml:"miniflux:hideGlobally,attr,omitempty"`
	AllowSelfSignedCertificates bool   `xml:"miniflux:allowSelfSignedCertificates,attr,omitempty"`
	DisableHTTP2                bool   `xml:"miniflux:disableHTTP2,attr,omitempty"`
	IgnoreEntryUpdates          bool   `xml:"miniflux:ignoreEntryUpdates,attr,omitempty"`
}

func (o opmlOutline) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	outlineType := ""
	if o.IsSubscription() {
		outlineType = "rss"
	}

	return e.EncodeElement(opmlOutlineXML{
		Title:                       o.Title,
		Text:                        o.Text,
		Type:                        outlineType,
		FeedURL:                     o.FeedURL,
		SiteURL:                     o.SiteURL,
		Description:                 o.Description,
		Outlines:                    o.Outlines,
		ScraperRules:                o.ScraperRules,
		RewriteRules:                o.RewriteRules,
		UrlRewriteRules:             o.UrlRewriteRules,
		BlocklistRules:              o.BlocklistRules,
		KeeplistRules:               o.KeeplistRules,
		BlockFilterEntryRules:       o.BlockFilterEntryRules,
		KeepFilterEntryRules:        o.KeepFilterEntryRules,
		UserAgent:                   o.UserAgent,
		ProxyURL:                    o.ProxyURL,
		Crawler:                     o.Crawler,
		IgnoreHTTPCache:             o.IgnoreHTTPCache,
		FetchViaProxy:               o.FetchViaProxy,
		Disabled:                    o.Disabled,
		NoMediaPlayer:               o.NoMediaPlayer,
		HideGlobally:                o.HideGlobally,
		AllowSelfSignedCertificates: o.AllowSelfSignedCertificates,
		DisableHTTP2:                o.DisableHTTP2,
		IgnoreEntryUpdates:          o.IgnoreEntryUpdates,
	}, start)
}

func (o *opmlOutline) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*o = opmlOutline{}

	for _, attr := range start.Attr {
		if attr.Name.Space == minifluxOPMLNamespace {
			if err := o.setMinifluxAttribute(attr.Name.Local, attr.Value); err != nil {
				return err
			}
			continue
		}

		if attr.Name.Space != "" {
			continue
		}

		switch attr.Name.Local {
		case "title":
			o.Title = attr.Value
		case "text":
			o.Text = attr.Value
		case "xmlUrl":
			o.FeedURL = attr.Value
		case "htmlUrl":
			o.SiteURL = attr.Value
		case "description":
			o.Description = attr.Value
		}
	}

	for {
		token, err := d.Token()
		if err != nil {
			return err
		}

		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local != "outline" {
				if err := d.Skip(); err != nil {
					return err
				}
				continue
			}

			var child opmlOutline
			if err := d.DecodeElement(&child, &element); err != nil {
				return err
			}
			o.Outlines = append(o.Outlines, child)
		case xml.EndElement:
			if element.Name == start.Name {
				return nil
			}
		}
	}
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

func (o *opmlOutline) setMinifluxAttribute(name, value string) error {
	switch name {
	case "scraperRules":
		o.ScraperRules = value
	case "rewriteRules":
		o.RewriteRules = value
	case "urlRewriteRules":
		o.UrlRewriteRules = value
	case "blocklistRules":
		o.BlocklistRules = value
	case "keeplistRules":
		o.KeeplistRules = value
	case "blockFilterEntryRules":
		o.BlockFilterEntryRules = value
	case "keepFilterEntryRules":
		o.KeepFilterEntryRules = value
	case "userAgent":
		o.UserAgent = value
	case "proxyUrl":
		o.ProxyURL = value
	case "crawler":
		return setMinifluxBoolAttribute(name, value, &o.Crawler)
	case "ignoreHTTPCache":
		return setMinifluxBoolAttribute(name, value, &o.IgnoreHTTPCache)
	case "fetchViaProxy":
		return setMinifluxBoolAttribute(name, value, &o.FetchViaProxy)
	case "disabled":
		return setMinifluxBoolAttribute(name, value, &o.Disabled)
	case "noMediaPlayer":
		return setMinifluxBoolAttribute(name, value, &o.NoMediaPlayer)
	case "hideGlobally":
		return setMinifluxBoolAttribute(name, value, &o.HideGlobally)
	case "allowSelfSignedCertificates":
		return setMinifluxBoolAttribute(name, value, &o.AllowSelfSignedCertificates)
	case "disableHTTP2":
		return setMinifluxBoolAttribute(name, value, &o.DisableHTTP2)
	case "ignoreEntryUpdates":
		return setMinifluxBoolAttribute(name, value, &o.IgnoreEntryUpdates)
	}

	return nil
}

func setMinifluxBoolAttribute(name, value string, target *bool) error {
	parsedValue, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("opml: invalid miniflux attribute %q: %w", name, err)
	}

	*target = parsedValue
	return nil
}
