// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googleplay // import "miniflux.app/v2/internal/reader/googleplay"

// Specs:
// https://support.google.com/googleplay/podcasts/answer/6260341
// https://www.google.com/schemas/play-podcasts/1.0/play-podcasts.xsd
type GooglePlayChannelElement struct {
	GooglePlayAuthor      string                    `xml:"http://www.google.com/schemas/play-podcasts/1.0 author"`
	GooglePlayEmail       string                    `xml:"http://www.google.com/schemas/play-podcasts/1.0 email"`
	GooglePlayImage       GooglePlayImageElement    `xml:"http://www.google.com/schemas/play-podcasts/1.0 image"`
	GooglePlayDescription string                    `xml:"http://www.google.com/schemas/play-podcasts/1.0 description"`
	GooglePlayCategory    GooglePlayCategoryElement `xml:"http://www.google.com/schemas/play-podcasts/1.0 category"`
}

type GooglePlayItemElement struct {
	GooglePlayAuthor      string `xml:"http://www.google.com/schemas/play-podcasts/1.0 author"`
	GooglePlayDescription string `xml:"http://www.google.com/schemas/play-podcasts/1.0 description"`
	GooglePlayExplicit    string `xml:"http://www.google.com/schemas/play-podcasts/1.0 explicit"`
	GooglePlayBlock       string `xml:"http://www.google.com/schemas/play-podcasts/1.0 block"`
	GooglePlayNewFeedURL  string `xml:"http://www.google.com/schemas/play-podcasts/1.0 new-feed-url"`
}

type GooglePlayImageElement struct {
	Href string `xml:"href,attr"`
}

type GooglePlayCategoryElement struct {
	Text string `xml:"text,attr"`
}
