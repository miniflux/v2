// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rewrite

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	youtubeRegex = regexp.MustCompile(`youtube\.com/watch\?v=(.*)`)
)

func addImageTitle(entryURL, entryContent string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(entryContent))
	if err != nil {
		return entryContent
	}

	imgTag := doc.Find("img").First()
	if titleAttr, found := imgTag.Attr("title"); found {
		return entryContent + `<blockquote cite="` + entryURL + `">` + titleAttr + "</blockquote>"
	}

	return entryContent
}

func addYoutubeVideo(entryURL, entryContent string) string {
	matches := youtubeRegex.FindStringSubmatch(entryURL)

	if len(matches) == 2 {
		video := `<iframe width="650" height="350" frameborder="0" src="https://www.youtube-nocookie.com/embed/` + matches[1] + `" allowfullscreen></iframe>`
		return video + "<p>" + entryContent + "</p>"
	}
	return entryContent
}
