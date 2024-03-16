// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package atom // import "miniflux.app/v2/internal/reader/atom"

import (
	"encoding/xml"
	"html"
	"strings"

	"miniflux.app/v2/internal/reader/media"
	"miniflux.app/v2/internal/reader/sanitizer"
)

// The "atom:feed" element is the document (i.e., top-level) element of
// an Atom Feed Document, acting as a container for metadata and data
// associated with the feed. Its element children consist of metadata
// elements followed by zero or more atom:entry child elements.
//
// Specs:
// https://tools.ietf.org/html/rfc4287
// https://validator.w3.org/feed/docs/atom.html
type Atom10Feed struct {
	XMLName xml.Name `xml:"http://www.w3.org/2005/Atom feed"`

	// The "atom:id" element conveys a permanent, universally unique
	// identifier for an entry or feed.
	//
	// Its content MUST be an IRI, as defined by [RFC3987].  Note that the
	// definition of "IRI" excludes relative references.  Though the IRI
	// might use a dereferencable scheme, Atom Processors MUST NOT assume it
	// can be dereferenced.
	//
	// atom:feed elements MUST contain exactly one atom:id element.
	ID string `xml:"http://www.w3.org/2005/Atom id"`

	// The "atom:title" element is a Text construct that conveys a human-
	// readable title for an entry or feed.
	//
	// atom:feed elements MUST contain exactly one atom:title element.
	Title Atom10Text `xml:"http://www.w3.org/2005/Atom title"`

	// The "atom:author" element is a Person construct that indicates the
	// author of the entry or feed.
	//
	// atom:feed elements MUST contain one or more atom:author elements,
	// unless all of the atom:feed element's child atom:entry elements
	// contain at least one atom:author element.
	Authors AtomPersons `xml:"http://www.w3.org/2005/Atom author"`

	// The "atom:icon" element's content is an IRI reference [RFC3987] that
	// identifies an image that provides iconic visual identification for a
	// feed.
	//
	// atom:feed elements MUST NOT contain more than one atom:icon element.
	Icon string `xml:"http://www.w3.org/2005/Atom icon"`

	// The "atom:logo" element's content is an IRI reference [RFC3987] that
	// identifies an image that provides visual identification for a feed.
	//
	// atom:feed elements MUST NOT contain more than one atom:logo element.
	Logo string `xml:"http://www.w3.org/2005/Atom logo"`

	// atom:feed elements SHOULD contain one atom:link element with a rel
	// attribute value of "self". This is the preferred URI for
	// retrieving Atom Feed Documents representing this Atom feed.
	//
	// atom:feed elements MUST NOT contain more than one atom:link
	// element with a rel attribute value of "alternate" that has the
	// same combination of type and hreflang attribute values.
	Links AtomLinks `xml:"http://www.w3.org/2005/Atom link"`

	// The "atom:category" element conveys information about a category
	// associated with an entry or feed.  This specification assigns no
	// meaning to the content (if any) of this element.
	//
	// atom:feed elements MAY contain any number of atom:category
	// elements.
	Categories AtomCategories `xml:"http://www.w3.org/2005/Atom category"`

	Entries []Atom10Entry `xml:"http://www.w3.org/2005/Atom entry"`
}

type Atom10Entry struct {
	// The "atom:id" element conveys a permanent, universally unique
	// identifier for an entry or feed.
	//
	// Its content MUST be an IRI, as defined by [RFC3987].  Note that the
	// definition of "IRI" excludes relative references.  Though the IRI
	// might use a dereferencable scheme, Atom Processors MUST NOT assume it
	// can be dereferenced.
	//
	// atom:entry elements MUST contain exactly one atom:id element.
	ID string `xml:"http://www.w3.org/2005/Atom id"`

	// The "atom:title" element is a Text construct that conveys a human-
	// readable title for an entry or feed.
	//
	// atom:entry elements MUST contain exactly one atom:title element.
	Title Atom10Text `xml:"http://www.w3.org/2005/Atom title"`

	// The "atom:published" element is a Date construct indicating an
	// instant in time associated with an event early in the life cycle of
	// the entry.
	Published string `xml:"http://www.w3.org/2005/Atom published"`

	// The "atom:updated" element is a Date construct indicating the most
	// recent instant in time when an entry or feed was modified in a way
	// the publisher considers significant. Therefore, not all
	// modifications necessarily result in a changed atom:updated value.
	//
	// atom:entry elements MUST contain exactly one atom:updated element.
	Updated string `xml:"http://www.w3.org/2005/Atom updated"`

	// atom:entry elements MUST NOT contain more than one atom:link
	// element with a rel attribute value of "alternate" that has the
	// same combination of type and hreflang attribute values.
	Links AtomLinks `xml:"http://www.w3.org/2005/Atom link"`

	// atom:entry elements MUST contain an atom:summary element in either
	// of the following cases:
	// *  the atom:entry contains an atom:content that has a "src"
	//    attribute (and is thus empty).
	// *  the atom:entry contains content that is encoded in Base64;
	//    i.e., the "type" attribute of atom:content is a MIME media type
	//    [MIMEREG], but is not an XML media type [RFC3023], does not
	//    begin with "text/", and does not end with "/xml" or "+xml".
	//
	// atom:entry elements MUST NOT contain more than one atom:summary
	// element.
	Summary Atom10Text `xml:"http://www.w3.org/2005/Atom summary"`

	// atom:entry elements MUST NOT contain more than one atom:content
	// element.
	Content Atom10Text `xml:"http://www.w3.org/2005/Atom content"`

	// The "atom:author" element is a Person construct that indicates the
	// author of the entry or feed.
	//
	// atom:entry elements MUST contain one or more atom:author elements
	Authors AtomPersons `xml:"http://www.w3.org/2005/Atom author"`

	// The "atom:category" element conveys information about a category
	// associated with an entry or feed.  This specification assigns no
	// meaning to the content (if any) of this element.
	//
	// atom:entry elements MAY contain any number of atom:category
	// elements.
	Categories AtomCategories `xml:"http://www.w3.org/2005/Atom category"`

	media.MediaItemElement
}

// A Text construct contains human-readable text, usually in small
// quantities. The content of Text constructs is Language-Sensitive.
// Specs: https://datatracker.ietf.org/doc/html/rfc4287#section-3.1
// Text: https://datatracker.ietf.org/doc/html/rfc4287#section-3.1.1.1
// HTML: https://datatracker.ietf.org/doc/html/rfc4287#section-3.1.1.2
// XHTML: https://datatracker.ietf.org/doc/html/rfc4287#section-3.1.1.3
type Atom10Text struct {
	Type             string               `xml:"type,attr"`
	CharData         string               `xml:",chardata"`
	InnerXML         string               `xml:",innerxml"`
	XHTMLRootElement AtomXHTMLRootElement `xml:"http://www.w3.org/1999/xhtml div"`
}

func (a *Atom10Text) Body() string {
	var content string

	if strings.EqualFold(a.Type, "xhtml") {
		content = a.xhtmlContent()
	} else {
		content = a.CharData
	}

	return strings.TrimSpace(content)
}

func (a *Atom10Text) Title() string {
	var content string

	if strings.EqualFold(a.Type, "xhtml") {
		content = a.xhtmlContent()
	} else if strings.Contains(a.InnerXML, "<![CDATA[") {
		content = html.UnescapeString(a.CharData)
	} else {
		content = a.CharData
	}

	content = sanitizer.StripTags(content)
	return strings.TrimSpace(content)
}

func (a *Atom10Text) xhtmlContent() string {
	if a.XHTMLRootElement.XMLName.Local == "div" {
		return a.XHTMLRootElement.InnerXML
	}
	return a.InnerXML
}

type AtomXHTMLRootElement struct {
	XMLName  xml.Name `xml:"div"`
	InnerXML string   `xml:",innerxml"`
}
