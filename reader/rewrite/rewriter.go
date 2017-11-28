// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rewrite

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var rewriteRules = []func(string, string) string{
	func(url, content string) string {
		re := regexp.MustCompile(`youtube\.com/watch\?v=(.*)`)
		matches := re.FindStringSubmatch(url)

		if len(matches) == 2 {
			video := `<iframe width="650" height="350" frameborder="0" src="https://www.youtube-nocookie.com/embed/` + matches[1] + `" allowfullscreen></iframe>`
			return video + "<p>" + content + "</p>"
		}
		return content
	},
	func(url, content string) string {
		if strings.HasPrefix(url, "https://xkcd.com") {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
			if err != nil {
				return content
			}

			imgTag := doc.Find("img").First()
			if titleAttr, found := imgTag.Attr("title"); found {
				return content + `<blockquote cite="` + url + `">` + titleAttr + "</blockquote>"
			}
		}
		return content
	},
}

// Rewriter modify item contents with a set of rewriting rules.
func Rewriter(url, content string) string {
	for _, rewriteRule := range rewriteRules {
		content = rewriteRule(url, content)
	}

	return content
}
