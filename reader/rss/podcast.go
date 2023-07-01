// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rss // import "miniflux.app/reader/rss"

import "strings"

// PodcastFeedElement represents iTunes and GooglePlay feed XML elements.
// Specs:
// - https://github.com/simplepie/simplepie-ng/wiki/Spec:-iTunes-Podcast-RSS
// - https://developers.google.com/search/reference/podcast/rss-feed
type PodcastFeedElement struct {
	ItunesAuthor     string       `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd channel>author"`
	Subtitle         string       `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd channel>subtitle"`
	Summary          string       `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd channel>summary"`
	PodcastOwner     PodcastOwner `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd channel>owner"`
	GooglePlayAuthor string       `xml:"http://www.google.com/schemas/play-podcasts/1.0 channel>author"`
}

// PodcastEntryElement represents iTunes and GooglePlay entry XML elements.
type PodcastEntryElement struct {
	Subtitle              string `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd subtitle"`
	Summary               string `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd summary"`
	GooglePlayDescription string `xml:"http://www.google.com/schemas/play-podcasts/1.0 description"`
}

// PodcastOwner represents contact information for the podcast owner.
type PodcastOwner struct {
	Name  string `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd name"`
	Email string `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd email"`
}

// Image represents podcast artwork.
type Image struct {
	URL string `xml:"href,attr"`
}

// PodcastAuthor returns the author of the podcast.
func (e *PodcastFeedElement) PodcastAuthor() string {
	author := ""

	switch {
	case e.ItunesAuthor != "":
		author = e.ItunesAuthor
	case e.GooglePlayAuthor != "":
		author = e.GooglePlayAuthor
	case e.PodcastOwner.Name != "":
		author = e.PodcastOwner.Name
	case e.PodcastOwner.Email != "":
		author = e.PodcastOwner.Email
	}

	return strings.TrimSpace(author)
}

// PodcastDescription returns the description of the podcast.
func (e *PodcastEntryElement) PodcastDescription() string {
	description := ""

	switch {
	case e.GooglePlayDescription != "":
		description = e.GooglePlayDescription
	case e.Summary != "":
		description = e.Summary
	case e.Subtitle != "":
		description = e.Subtitle
	}
	return strings.TrimSpace(description)
}
