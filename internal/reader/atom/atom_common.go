// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package atom // import "miniflux.app/v2/internal/reader/atom"

import (
	"cmp"
	"slices"
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

type atomPersons []*AtomPerson

// personNames returns sorted and deduplicated author names.
func (a atomPersons) personNames() []string {
	return makeSorted((*AtomPerson).PersonName, a)
}

// Specs: https://datatracker.ietf.org/doc/html/rfc4287#section-4.2.7
type AtomLink struct {
	Href   string `xml:"href,attr"`
	Type   string `xml:"type,attr"`
	Rel    string `xml:"rel,attr"`
	Length string `xml:"length,attr"`
	Title  string `xml:"title,attr"`
}

type atomLinks []*AtomLink

func (a atomLinks) originalLink() string {
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

func (a atomLinks) firstLinkWithRelation(relation string) string {
	for _, link := range a {
		if strings.EqualFold(link.Rel, relation) {
			return strings.TrimSpace(link.Href)
		}
	}

	return ""
}

func (a atomLinks) firstLinkWithRelationAndType(relation string, contentTypes ...string) string {
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

func (a atomLinks) findAllLinksWithRelation(relation string) []*AtomLink {
	links := make([]*AtomLink, 0, len(a))

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
type atomCategory struct {
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

func (ac atomCategory) name() string {
	name := strings.TrimSpace(ac.Label)
	if name != "" {
		return name
	}

	name = strings.TrimSpace(ac.Term)
	if name != "" {
		return name
	}

	return ""
}

type atomCategories []atomCategory

// CategoryNames returns sorted and deduplicated category names.
func (ac atomCategories) CategoryNames() []string {
	return makeSorted(atomCategory.name, ac)
}

func makeSorted[I any, O cmp.Ordered](fn func(I) O, values []I) []O {
	var zero O

	sorted := make([]O, 0, len(values))
	for _, in := range values {
		out := fn(in)
		if out == zero {
			continue
		}

		where, found := slices.BinarySearch(sorted, out)
		if found {
			continue
		}

		// Insert sorted to avoid duplicates.
		sorted = slices.Insert(sorted, where, out)
	}

	return sorted
}
