// Copyright 2019 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package atom // import "miniflux.app/reader/atom"

import "strings"

type atomPerson struct {
	Name  string `xml:"name"`
	Email string `xml:"email"`
}

func (a *atomPerson) String() string {
	name := ""

	switch {
	case a.Name != "":
		name = a.Name
	case a.Email != "":
		name = a.Email
	}

	return strings.TrimSpace(name)
}

type atomAuthors []*atomPerson

func (a atomAuthors) String() string {
	var authors []string

	for _, person := range a {
		authors = append(authors, person.String())
	}

	return strings.Join(authors, ", ")
}

type atomLink struct {
	URL    string `xml:"href,attr"`
	Type   string `xml:"type,attr"`
	Rel    string `xml:"rel,attr"`
	Length string `xml:"length,attr"`
}

type atomLinks []*atomLink

func (a atomLinks) originalLink() string {
	for _, link := range a {
		if strings.ToLower(link.Rel) == "alternate" {
			return strings.TrimSpace(link.URL)
		}

		if link.Rel == "" && (link.Type == "" || link.Type == "text/html") {
			return strings.TrimSpace(link.URL)
		}
	}

	return ""
}

func (a atomLinks) firstLinkWithRelation(relation string) string {
	for _, link := range a {
		if strings.ToLower(link.Rel) == relation {
			return strings.TrimSpace(link.URL)
		}
	}

	return ""
}

func (a atomLinks) firstLinkWithRelationAndType(relation string, contentTypes ...string) string {
	for _, link := range a {
		if strings.ToLower(link.Rel) == relation {
			for _, contentType := range contentTypes {
				if strings.ToLower(link.Type) == contentType {
					return strings.TrimSpace(link.URL)
				}
			}
		}
	}

	return ""
}
