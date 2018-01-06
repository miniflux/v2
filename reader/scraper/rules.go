// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package scraper

// List of predefined scraper rules (alphabetically sorted)
// domain => CSS selectors
var predefinedRules = map[string]string{
	"cbc.ca":              ".story-content",
	"darkreading.com":     "#article-main:not(header)",
	"developpez.com":      "div[itemprop=articleBody]",
	"francetvinfo.fr":     ".text",
	"github.com":          "article.entry-content",
	"heise.de":            "div.article-content",
	"igen.fr":             "section.corps",
	"ing.dk":              "section.body",
	"lapresse.ca":         ".amorce, .entry",
	"lemonde.fr":          "div#articleBody",
	"lepoint.fr":          ".art-text",
	"lesjoiesducode.fr":   ".blog-post-content img",
	"lesnumeriques.com":   ".text",
	"linux.com":           "div.content, div[property]",
	"medium.com":          ".section-content",
	"mac4ever.com":        "div[itemprop=articleBody]",
	"monwindows.com":      ".blog-post-body",
	"npr.org":             "#storytext",
	"oneindia.com":        ".io-article-body",
	"opensource.com":      "div[property]",
	"osnews.com":          "div.newscontent1",
	"phoronix.com":        "div.content",
	"pseudo-sciences.org": "#art_main",
	"slate.fr":            ".field-items",
	"techcrunch.com":      "div.article-entry",
	"theregister.co.uk":   "#body",
	"universfreebox.com":  "#corps_corps",
	"version2.dk":         "section.body",
	"wired.com":           "main figure, article",
	"zeit.de":             ".summary, .article-body",
	"zdnet.com":           "div.storyBody",
}
