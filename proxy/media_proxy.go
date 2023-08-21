// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package proxy // import "miniflux.app/proxy"

import (
	"strings"

	"miniflux.app/config"
	"miniflux.app/reader/sanitizer"
	"miniflux.app/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

type urlProxyRewriter func(router *mux.Router, url string) string

// ProxyRewriter replaces media URLs with internal proxy URLs.
func ProxyRewriter(router *mux.Router, data string) string {
	return genericProxyRewriter(router, ProxifyURL, data)
}

// AbsoluteProxyRewriter do the same as ProxyRewriter except it uses absolute URLs.
func AbsoluteProxyRewriter(router *mux.Router, host, data string) string {
	proxifyFunction := func(router *mux.Router, url string) string {
		return AbsoluteProxifyURL(router, host, url)
	}
	return genericProxyRewriter(router, proxifyFunction, data)
}

func genericProxyRewriter(router *mux.Router, proxifyFunction urlProxyRewriter, data string) string {
	proxyOption := config.Opts.ProxyOption()
	if proxyOption == "none" {
		return data
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data))
	if err != nil {
		return data
	}

	for _, mediaType := range config.Opts.ProxyMediaTypes() {
		switch mediaType {
		case "image":
			doc.Find("img").Each(func(i int, img *goquery.Selection) {
				if srcAttrValue, ok := img.Attr("src"); ok {
					if !isDataURL(srcAttrValue) && (proxyOption == "all" || !url.IsHTTPS(srcAttrValue)) {
						img.SetAttr("src", proxifyFunction(router, srcAttrValue))
					}
				}

				if srcsetAttrValue, ok := img.Attr("srcset"); ok {
					proxifySourceSet(img, router, proxifyFunction, proxyOption, srcsetAttrValue)
				}
			})

			doc.Find("picture source").Each(func(i int, sourceElement *goquery.Selection) {
				if srcsetAttrValue, ok := sourceElement.Attr("srcset"); ok {
					proxifySourceSet(sourceElement, router, proxifyFunction, proxyOption, srcsetAttrValue)
				}
			})

		case "audio":
			doc.Find("audio").Each(func(i int, audio *goquery.Selection) {
				if srcAttrValue, ok := audio.Attr("src"); ok {
					if !isDataURL(srcAttrValue) && (proxyOption == "all" || !url.IsHTTPS(srcAttrValue)) {
						audio.SetAttr("src", proxifyFunction(router, srcAttrValue))
					}
				}
			})

			doc.Find("audio source").Each(func(i int, sourceElement *goquery.Selection) {
				if srcAttrValue, ok := sourceElement.Attr("src"); ok {
					if !isDataURL(srcAttrValue) && (proxyOption == "all" || !url.IsHTTPS(srcAttrValue)) {
						sourceElement.SetAttr("src", proxifyFunction(router, srcAttrValue))
					}
				}
			})

		case "video":
			doc.Find("video").Each(func(i int, video *goquery.Selection) {
				if srcAttrValue, ok := video.Attr("src"); ok {
					if !isDataURL(srcAttrValue) && (proxyOption == "all" || !url.IsHTTPS(srcAttrValue)) {
						video.SetAttr("src", proxifyFunction(router, srcAttrValue))
					}
				}
			})

			doc.Find("video source").Each(func(i int, sourceElement *goquery.Selection) {
				if srcAttrValue, ok := sourceElement.Attr("src"); ok {
					if !isDataURL(srcAttrValue) && (proxyOption == "all" || !url.IsHTTPS(srcAttrValue)) {
						sourceElement.SetAttr("src", proxifyFunction(router, srcAttrValue))
					}
				}
			})
		}
	}

	output, err := doc.Find("body").First().Html()
	if err != nil {
		return data
	}

	return output
}

func proxifySourceSet(element *goquery.Selection, router *mux.Router, proxifyFunction urlProxyRewriter, proxyOption, srcsetAttrValue string) {
	imageCandidates := sanitizer.ParseSrcSetAttribute(srcsetAttrValue)

	for _, imageCandidate := range imageCandidates {
		if !isDataURL(imageCandidate.ImageURL) && (proxyOption == "all" || !url.IsHTTPS(imageCandidate.ImageURL)) {
			imageCandidate.ImageURL = proxifyFunction(router, imageCandidate.ImageURL)
		}
	}

	element.SetAttr("srcset", imageCandidates.String())
}

func isDataURL(s string) bool {
	return strings.HasPrefix(s, "data:")
}
