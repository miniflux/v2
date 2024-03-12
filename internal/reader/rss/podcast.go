// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rss // import "miniflux.app/v2/internal/reader/rss"

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

var ErrInvalidDurationFormat = errors.New("rss: invalid duration format")

// PodcastFeedElement represents iTunes and GooglePlay feed XML elements.
// Specs:
// - https://github.com/simplepie/simplepie-ng/wiki/Spec:-iTunes-Podcast-RSS
// - https://support.google.com/podcast-publishers/answer/9889544
type PodcastFeedElement struct {
	ItunesAuthor     string       `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd author"`
	Subtitle         string       `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd subtitle"`
	Summary          string       `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd summary"`
	PodcastOwner     PodcastOwner `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd owner"`
	GooglePlayAuthor string       `xml:"http://www.google.com/schemas/play-podcasts/1.0 author"`
}

// PodcastEntryElement represents iTunes and GooglePlay entry XML elements.
type PodcastEntryElement struct {
	ItunesAuthor          string       `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd author"`
	Subtitle              string       `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd subtitle"`
	Summary               string       `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd summary"`
	Duration              string       `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd duration"`
	PodcastOwner          PodcastOwner `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd owner"`
	GooglePlayAuthor      string       `xml:"http://www.google.com/schemas/play-podcasts/1.0 author"`
	GooglePlayDescription string       `xml:"http://www.google.com/schemas/play-podcasts/1.0 description"`
}

// PodcastOwner represents contact information for the podcast owner.
type PodcastOwner struct {
	Name  string `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd name"`
	Email string `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd email"`
}

func (p *PodcastOwner) String() string {
	var name string

	switch {
	case p.Name != "":
		name = p.Name
	case p.Email != "":
		name = p.Email
	}

	return strings.TrimSpace(name)
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
	case e.PodcastOwner.String() != "":
		author = e.PodcastOwner.String()
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

// normalizeDuration returns the duration tag value as a number of minutes
func normalizeDuration(rawDuration string) (int, error) {
	var sumSeconds int

	durationParts := strings.Split(rawDuration, ":")
	if len(durationParts) > 3 {
		return 0, ErrInvalidDurationFormat
	}

	for i, durationPart := range durationParts {
		durationPartValue, err := strconv.Atoi(durationPart)
		if err != nil {
			return 0, ErrInvalidDurationFormat
		}

		sumSeconds += int(math.Pow(60, float64(len(durationParts)-i-1))) * durationPartValue
	}

	return sumSeconds / 60, nil
}
