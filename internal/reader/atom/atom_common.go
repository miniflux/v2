// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package atom // import "miniflux.app/v2/internal/reader/atom"

import (
	"strings"
)

// Specs: https://datatracker.ietf.org/doc/html/rfc4287#section-3.2
type AtomPerson struct {
	// The "atom:name" element's content conveys a human-readable name for the author.
	// It MAY be the name of a corporation or other entity no individual authors can be named.
	// Person constructs MUST contain exactly one "atom:name" element, whose content MUST be a string.
	Name string `xml:"name"`

	// The "atom:email" element's content conveys an e-mail address associated with the Person construct.
	// Person constructs MAY contain an atom:email element, but MUST NOT contain more than one.
	// Its content MUST be an e-mail address [RFC2822].
	// Ordering of the element children of Person constructs MUST NOT be considered significant.
	Email string `xml:"email"`
}

func (a *AtomPerson) PersonName() string {
	name := strings.TrimSpace(a.Name)
	if name != "" {
		return name
	}

	return strings.TrimSpace(a.Email)
}

type AtomPersons []*AtomPerson

func (a AtomPersons) PersonNames() []string {
	var names []string
	authorNamesMap := make(map[string]bool)

	for _, person := range a {
		personName := person.PersonName()
		if _, ok := authorNamesMap[personName]; !ok {
			names = append(names, personName)
			authorNamesMap[personName] = true
		}
	}

	return names
}

// Specs: https://datatracker.ietf.org/doc/html/rfc4287#section-4.2.7
type AtomLink struct {
	Href   string `xml:"href,attr"`
	Type   string `xml:"type,attr"`
	Rel    string `xml:"rel,attr"`
	Length string `xml:"length,attr"`
	Title  string `xml:"title,attr"`
}

type AtomLinks []*AtomLink

func (a AtomLinks) OriginalLink() string {
	for _, link := range a {
		if strings.EqualFold(link.Rel, "alternate") {
			return strings.TrimSpace(link.Href)
		}

		if link.Rel == "" && (link.Type == "" || link.Type == "text/html") {
			return strings.TrimSpace(link.Href)
		}
	}

	return ""
}

func (a AtomLinks) firstLinkWithRelation(relation string) string {
	for _, link := range a {
		if strings.EqualFold(link.Rel, relation) {
			return strings.TrimSpace(link.Href)
		}
	}

	return ""
}

func (a AtomLinks) firstLinkWithRelationAndType(relation string, contentTypes ...string) string {
	for _, link := range a {
		if strings.EqualFold(link.Rel, relation) {
			for _, contentType := range contentTypes {
				if strings.EqualFold(link.Type, contentType) {
					return strings.TrimSpace(link.Href)
				}
			}
		}
	}

	return ""
}

func (a AtomLinks) findAllLinksWithRelation(relation string) []*AtomLink {
	var links []*AtomLink

	for _, link := range a {
		if strings.EqualFold(link.Rel, relation) {
			link.Href = strings.TrimSpace(link.Href)
			if link.Href != "" {
				links = append(links, link)
			}
		}
	}

	return links
}

// The "atom:category" element conveys information about a category
// associated with an entry or feed.  This specification assigns no
// meaning to the content (if any) of this element.
//
// Specs: https://datatracker.ietf.org/doc/html/rfc4287#section-4.2.2
type AtomCategory struct {
	// The "term" attribute is a string that identifies the category to
	// which the entry or feed belongs. Category elements MUST have a
	// "term" attribute.
	Term string `xml:"term,attr"`

	// The "scheme" attribute is an IRI that identifies a categorization
	// scheme. Category elements MAY have a "scheme" attribute.
	Scheme string `xml:"scheme,attr"`

	// The "label" attribute provides a human-readable label for display in
	// end-user applications. The content of the "label" attribute is
	// Language-Sensitive. Entities such as "&amp;" and "&lt;" represent
	// their corresponding characters ("&" and "<", respectively), not
	// markup. Category elements MAY have a "label" attribute.
	Label string `xml:"label,attr"`
}

type AtomCategories []AtomCategory

func (ac AtomCategories) CategoryNames() []string {
	var categories []string

	for _, category := range ac {
		label := strings.TrimSpace(category.Label)
		if label != "" {
			categories = append(categories, label)
		} else {
			term := strings.TrimSpace(category.Term)
			if term != "" {
				categories = append(categories, term)
			}
		}
	}

	return categories
}
