// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package atom // import "miniflux.app/v2/internal/reader/atom"

import (
	"encoding/base64"
	"html"
	"strings"
)

// Specs: http://web.archive.org/web/20060811235523/http://www.mnot.net/drafts/draft-nottingham-atom-format-02.html
type Atom03Feed struct {
	Version string `xml:"version,attr"`

	// The "atom:id" element's content conveys a permanent, globally unique identifier for the feed.
	// It MUST NOT change over time, even if the feed is relocated. atom:feed elements MAY contain an atom:id element,
	// but MUST NOT contain more than one. The content of this element, when present, MUST be a URI.
	ID string `xml:"http://purl.org/atom/ns# id"`

	// The "atom:title" element is a Content construct that conveys a human-readable title for the feed.
	// atom:feed elements MUST contain exactly one atom:title element.
	// If the feed describes a Web resource, its content SHOULD be the same as that resource's title.
	Title Atom03Content `xml:"http://purl.org/atom/ns# title"`

	// The "atom:link" element is a Link construct that conveys a URI associated with the feed.
	// The nature of the relationship as well as the link itself is determined by the element's content.
	// atom:feed elements MUST contain at least one atom:link element with a rel attribute value of "alternate".
	// atom:feed elements MUST NOT contain more than one atom:link element with a rel attribute value of "alternate" that has the same type attribute value.
	// atom:feed elements MAY contain additional atom:link elements beyond those described above.
	Links AtomLinks `xml:"http://purl.org/atom/ns# link"`

	// The "atom:author" element is a Person construct that indicates the default author of the feed.
	// atom:feed elements MUST contain exactly one atom:author element,
	// UNLESS all of the atom:feed element's child atom:entry elements contain an atom:author element.
	// atom:feed elements MUST NOT contain more than one atom:author element.
	Author AtomPerson `xml:"http://purl.org/atom/ns# author"`

	// The "atom:entry" element's represents an individual entry that is contained by the feed.
	// atom:feed elements MAY contain one or more atom:entry elements.
	Entries []Atom03Entry `xml:"http://purl.org/atom/ns# entry"`
}

type Atom03Entry struct {
	// The "atom:id" element's content conveys a permanent, globally unique identifier for the entry.
	// It MUST NOT change over time, even if other representations of the entry (such as a web representation pointed to by the entry's atom:link element) are relocated.
	// If the same entry is syndicated in two atom:feeds published by the same entity, the entry's atom:id MUST be the same in both feeds.
	ID string `xml:"id"`

	// The "atom:title" element is a Content construct that conveys a human-readable title for the entry.
	// atom:entry elements MUST have exactly one "atom:title" element.
	// If an entry describes a Web resource, its content SHOULD be the same as that resource's title.
	Title Atom03Content `xml:"title"`

	// The "atom:modified" element is a Date construct that indicates the time that the entry was last modified.
	// atom:entry elements MUST contain an atom:modified element, but MUST NOT contain more than one.
	// The content of an atom:modified element MUST have a time zone whose value SHOULD be "UTC".
	Modified string `xml:"modified"`

	// The "atom:issued" element is a Date construct that indicates the time that the entry was issued.
	// atom:entry elements MUST contain an atom:issued element, but MUST NOT contain more than one.
	// The content of an atom:issued element MAY omit a time zone.
	Issued string `xml:"issued"`

	// The "atom:created" element is a Date construct that indicates the time that the entry was created.
	// atom:entry elements MAY contain an atom:created element, but MUST NOT contain more than one.
	// The content of an atom:created element MUST have a time zone whose value SHOULD be "UTC".
	// If atom:created is not present, its content MUST considered to be the same as that of atom:modified.
	Created string `xml:"created"`

	// The "atom:link" element is a Link construct that conveys a URI associated with the entry.
	// The nature of the relationship as well as the link itself is determined by the element's content.
	// atom:entry elements MUST contain at least one atom:link element with a rel attribute value of "alternate".
	// atom:entry elements MUST NOT contain more than one atom:link element with a rel attribute value of "alternate" that has the same type attribute value.
	// atom:entry elements MAY contain additional atom:link elements beyond those described above.
	Links AtomLinks `xml:"link"`

	// The "atom:summary" element is a Content construct that conveys a short summary, abstract or excerpt of the entry.
	// atom:entry elements MAY contain an atom:created element, but MUST NOT contain more than one.
	Summary Atom03Content `xml:"summary"`

	// The "atom:content" element is a Content construct that conveys the content of the entry.
	// atom:entry elements MAY contain one or more atom:content elements.
	Content Atom03Content `xml:"content"`

	// The "atom:author" element is a Person construct that indicates the default author of the entry.
	// atom:entry elements MUST contain exactly one atom:author element,
	// UNLESS the atom:feed element containing them contains an atom:author element itself.
	// atom:entry elements MUST NOT contain more than one atom:author element.
	Author AtomPerson `xml:"author"`
}

type Atom03Content struct {
	// Content constructs MAY have a "type" attribute, whose value indicates the media type of the content.
	// When present, this attribute's value MUST be a registered media type [RFC2045].
	// If not present, its value MUST be considered to be "text/plain".
	Type string `xml:"type,attr"`

	// Content constructs MAY have a "mode" attribute, whose value indicates the method used to encode the content.
	// When present, this attribute's value MUST be listed below.
	// If not present, its value MUST be considered to be "xml".
	//
	// "xml": A mode attribute with the value "xml" indicates that the element's content is inline xml (for example, namespace-qualified XHTML).
	//
	// "escaped": A mode attribute with the value "escaped" indicates that the element's content is an escaped string.
	// Processors MUST unescape the element's content before considering it as content of the indicated media type.
	//
	// "base64": A mode attribute with the value "base64" indicates that the element's content is base64-encoded [RFC2045].
	// Processors MUST decode the element's content before considering it as content of the the indicated media type.
	Mode string `xml:"mode,attr"`

	CharData string `xml:",chardata"`
	InnerXML string `xml:",innerxml"`
}

func (a *Atom03Content) Content() string {
	content := ""

	switch {
	case a.Mode == "xml":
		content = a.InnerXML
	case a.Mode == "escaped":
		content = a.CharData
	case a.Mode == "base64":
		b, err := base64.StdEncoding.DecodeString(a.CharData)
		if err == nil {
			content = string(b)
		}
	default:
		content = a.CharData
	}

	if a.Type != "text/html" {
		content = html.EscapeString(content)
	}

	return strings.TrimSpace(content)
}
