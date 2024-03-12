// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package itunes // import "miniflux.app/v2/internal/reader/itunes"

import "strings"

// Specs: https://help.apple.com/itc/podcasts_connect/#/itcb54353390
type ItunesFeedElement struct {
	ItunesAuthor     string                  `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd author"`
	ItunesBlock      string                  `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd block"`
	ItunesCategories []ItunesCategoryElement `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd category"`
	ItunesComplete   string                  `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd complete"`
	ItunesCopyright  string                  `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd copyright"`
	ItunesExplicit   string                  `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd explicit"`
	ItunesImage      ItunesImageElement      `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd image"`
	Keywords         string                  `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd keywords"`
	ItunesNewFeedURL string                  `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd new-feed-url"`
	ItunesOwner      ItunesOwnerElement      `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd owner"`
	ItunesSummary    string                  `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd summary"`
	ItunesTitle      string                  `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd title"`
	ItunesType       string                  `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd type"`
}

func (i *ItunesFeedElement) GetItunesCategories() []string {
	var categories []string
	for _, category := range i.ItunesCategories {
		categories = append(categories, category.Text)
		if category.SubCategory != nil {
			categories = append(categories, category.SubCategory.Text)
		}
	}
	return categories
}

type ItunesItemElement struct {
	ItunesAuthor      string             `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd author"`
	ItunesEpisode     string             `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd episode"`
	ItunesEpisodeType string             `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd episodeType"`
	ItunesExplicit    string             `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd explicit"`
	ItunesDuration    string             `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd duration"`
	ItunesImage       ItunesImageElement `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd image"`
	ItunesSeason      string             `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd season"`
	ItunesSubtitle    string             `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd subtitle"`
	ItunesSummary     string             `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd summary"`
	ItunesTitle       string             `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd title"`
	ItunesTranscript  string             `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd transcript"`
}

type ItunesImageElement struct {
	Href string `xml:"href,attr"`
}

type ItunesCategoryElement struct {
	Text        string                 `xml:"text,attr"`
	SubCategory *ItunesCategoryElement `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd category"`
}

type ItunesOwnerElement struct {
	Name  string `xml:"name"`
	Email string `xml:"email"`
}

func (i *ItunesOwnerElement) String() string {
	var name string

	switch {
	case i.Name != "":
		name = i.Name
	case i.Email != "":
		name = i.Email
	}

	return strings.TrimSpace(name)
}
