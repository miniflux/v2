// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rss // import "miniflux.app/v2/internal/reader/rss"

import (
	"miniflux.app/v2/internal/reader/atom"
)

type AtomAuthor struct {
	Author atom.AtomPerson `xml:"http://www.w3.org/2005/Atom author"`
}

func (a *AtomAuthor) PersonName() string {
	return a.Author.PersonName()
}

type AtomLinks struct {
	Links []*atom.AtomLink `xml:"http://www.w3.org/2005/Atom link"`
}
