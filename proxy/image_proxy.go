// Copyright 2020 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package proxy // import "miniflux.app/proxy"

import (
	"strings"

	"miniflux.app/config"
	"miniflux.app/reader/sanitizer"
	"miniflux.app/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

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
		if srcAttrValue, ok := img.Attr("src"); ok {
			if !isDataURL(srcAttrValue) && (proxyImages == "all" || !url.IsHTTPS(srcAttrValue)) {
				img.SetAttr("src", ProxifyURL(router, srcAttrValue))
			}
		}

		if srcsetAttrValue, ok := img.Attr("srcset"); ok {
			proxifySourceSet(img, router, proxyImages, srcsetAttrValue)
		}
	})

	doc.Find("picture source").Each(func(i int, sourceElement *goquery.Selection) {
		if srcsetAttrValue, ok := sourceElement.Attr("srcset"); ok {
			proxifySourceSet(sourceElement, router, proxyImages, srcsetAttrValue)
		}
	})

	output, err := doc.Find("body").First().Html()
	if err != nil {
		return data
	}

	return output
}

func proxifySourceSet(element *goquery.Selection, router *mux.Router, proxyImages, srcsetAttrValue string) {
	imageCandidates := sanitizer.ParseSrcSetAttribute(srcsetAttrValue)

	for _, imageCandidate := range imageCandidates {
		if !isDataURL(imageCandidate.ImageURL) && (proxyImages == "all" || !url.IsHTTPS(imageCandidate.ImageURL)) {
			imageCandidate.ImageURL = ProxifyURL(router, imageCandidate.ImageURL)
		}
	}

	element.SetAttr("srcset", imageCandidates.String())
}

func isDataURL(s string) bool {
	return strings.HasPrefix(s, "data:")
}
