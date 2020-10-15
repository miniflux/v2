// Copyright 2020 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package proxy // import "miniflux.app/proxy"

import (
	"regexp"
	"strings"

	"miniflux.app/config"
	"miniflux.app/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

var regexSplitSrcset = regexp.MustCompile(`,\s+`)

// ImageProxyRewriter replaces image URLs with internal proxy URLs.
func ImageProxyRewriter(router *mux.Router, data string) string {
	proxyImages := config.Opts.ProxyImages()
	if proxyImages == "none" {
		return data
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data))
	if err != nil {
		return data
	}

	doc.Find("img").Each(func(i int, img *goquery.Selection) {
		if srcAttr, ok := img.Attr("src"); ok {
			if !isDataURL(srcAttr) && (proxyImages == "all" || !url.IsHTTPS(srcAttr)) {
				img.SetAttr("src", ProxifyURL(router, srcAttr))
			}
		}

		if srcsetAttr, ok := img.Attr("srcset"); ok {
			if proxyImages == "all" || !url.IsHTTPS(srcsetAttr) {
				proxifySourceSet(img, router, srcsetAttr)
			}
		}
	})

	doc.Find("picture source").Each(func(i int, sourceElement *goquery.Selection) {
		if srcsetAttr, ok := sourceElement.Attr("srcset"); ok {
			if proxyImages == "all" || !url.IsHTTPS(srcsetAttr) {
				proxifySourceSet(sourceElement, router, srcsetAttr)
			}
		}
	})

	output, err := doc.Find("body").First().Html()
	if err != nil {
		return data
	}

	return output
}

func proxifySourceSet(element *goquery.Selection, router *mux.Router, attributeValue string) {
	var proxifiedSources []string

	for _, source := range regexSplitSrcset.Split(attributeValue, -1) {
		parts := strings.Split(strings.TrimSpace(source), " ")
		nbParts := len(parts)

		if nbParts > 0 {
			rewrittenSource := parts[0]
			if !isDataURL(rewrittenSource) {
				rewrittenSource = ProxifyURL(router, rewrittenSource)
			}

			if nbParts > 1 {
				rewrittenSource += " " + parts[1]
			}

			proxifiedSources = append(proxifiedSources, rewrittenSource)
		}
	}

	if len(proxifiedSources) > 0 {
		element.SetAttr("srcset", strings.Join(proxifiedSources, ", "))
	}
}

func isDataURL(s string) bool {
	return strings.HasPrefix(s, "data:")
}
