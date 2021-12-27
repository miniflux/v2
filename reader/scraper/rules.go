// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package scraper // import "miniflux.app/reader/scraper"

// List of predefined scraper rules (alphabetically sorted)
// domain => CSS selectors
var predefinedRules = map[string]string{
	"bbc.co.uk":            "div.vxp-column--single, div.story-body__inner, ul.gallery-images__list",
	"cbc.ca":               ".story-content",
	"darkreading.com":      "#article-main:not(header)",
	"developpez.com":       "div[itemprop=articleBody]",
	"dilbert.com":          "span.comic-title-name, img.img-comic",
	"financialsamurai.com": "article",
	"francetvinfo.fr":      ".text",
	"github.com":           "article.entry-content",
	"heise.de":             "header .article-content__lead, header .article-image, div.article-layout__content.article-content",
	"igen.fr":              "section.corps",
	"ikiwiki.iki.fi":       ".page.group",
	"ing.dk":               "section.body",
	"lapresse.ca":          ".amorce, .entry",
	"lemonde.fr":           "article",
	"lepoint.fr":           ".art-text",
	"lesjoiesducode.fr":    ".blog-post-content img",
	"lesnumeriques.com":    ".text",
	"linux.com":            "div.content, div[property]",
	"mac4ever.com":         "div[itemprop=articleBody]",
	"monwindows.com":       ".blog-post-body",
	"npr.org":              "#storytext",
	"oneindia.com":         ".io-article-body",
	"opensource.com":       "div[property]",
	"osnews.com":           "div.newscontent1",
	"phoronix.com":         "div.content",
	"pseudo-sciences.org":  "#art_main",
	"raywenderlich.com":    "article",
	"slate.fr":             ".field-items",
	"techcrunch.com":       "div.article-entry",
	"theoatmeal.com":       "div#comic",
	"theregister.com":      "#top-col-story h2, #body",
	"turnoff.us":           "article.post-content",
	"universfreebox.com":   "#corps_corps",
	"version2.dk":          "section.body",
	"wdwnt.com":            "div.entry-content",
	"wired.com":            "main figure, article",
	"zeit.de":              ".summary, .article-body",
	"zdnet.com":            "div.storyBody",
	"openingsource.org":    "article.suxing-popup-gallery",
}
