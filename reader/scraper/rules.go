// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package scraper

// List of predefined scraper rules (alphabetically sorted)
// domain => CSS selectors
var predefinedRules = map[string]string{
	"lemonde.fr":        "div#articleBody",
	"lesjoiesducode.fr": ".blog-post-content img",
	"linux.com":         "div.content, div[property]",
	"opensource.com":    "div[property]",
	"phoronix.com":      "div.content",
	"techcrunch.com":    "div.article-entry",
}
