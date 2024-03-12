// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rss // import "miniflux.app/v2/internal/reader/rss"

import "strings"

type AtomAuthor struct {
	Author AtomPerson `xml:"http://www.w3.org/2005/Atom author"`
}

func (a *AtomAuthor) String() string {
	return a.Author.String()
}

type AtomPerson struct {
	Name  string `xml:"name"`
	Email string `xml:"email"`
}

func (a *AtomPerson) String() string {
	var name string

	switch {
	case a.Name != "":
		name = a.Name
	case a.Email != "":
		name = a.Email
	}

	return strings.TrimSpace(name)
}

type AtomLink struct {
	URL    string `xml:"href,attr"`
	Type   string `xml:"type,attr"`
	Rel    string `xml:"rel,attr"`
	Length string `xml:"length,attr"`
}

type AtomLinks struct {
	Links []*AtomLink `xml:"http://www.w3.org/2005/Atom link"`
}
