// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package proxy // import "miniflux.app/v2/internal/proxy"

import (
	"strings"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/reader/sanitizer"
	"miniflux.app/v2/internal/urllib"

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
			doc.Find("img, picture source").Each(func(i int, img *goquery.Selection) {
				if srcAttrValue, ok := img.Attr("src"); ok {
					if shouldProxy(srcAttrValue, proxyOption) {
						img.SetAttr("src", proxifyFunction(router, srcAttrValue))
					}
				}

				if srcsetAttrValue, ok := img.Attr("srcset"); ok {
					proxifySourceSet(img, router, proxifyFunction, proxyOption, srcsetAttrValue)
				}
			})

			doc.Find("video").Each(func(i int, video *goquery.Selection) {
				if posterAttrValue, ok := video.Attr("poster"); ok {
					if shouldProxy(posterAttrValue, proxyOption) {
						video.SetAttr("poster", proxifyFunction(router, posterAttrValue))
					}
				}
			})

		case "audio":
			doc.Find("audio, audio source").Each(func(i int, audio *goquery.Selection) {
				if srcAttrValue, ok := audio.Attr("src"); ok {
					if shouldProxy(srcAttrValue, proxyOption) {
						audio.SetAttr("src", proxifyFunction(router, srcAttrValue))
					}
				}
			})

		case "video":
			doc.Find("video, video source").Each(func(i int, video *goquery.Selection) {
				if srcAttrValue, ok := video.Attr("src"); ok {
					if shouldProxy(srcAttrValue, proxyOption) {
						video.SetAttr("src", proxifyFunction(router, srcAttrValue))
					}
				}

				if posterAttrValue, ok := video.Attr("poster"); ok {
					if shouldProxy(posterAttrValue, proxyOption) {
						video.SetAttr("poster", proxifyFunction(router, posterAttrValue))
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
		if shouldProxy(imageCandidate.ImageURL, proxyOption) {
			imageCandidate.ImageURL = proxifyFunction(router, imageCandidate.ImageURL)
		}
	}

	element.SetAttr("srcset", imageCandidates.String())
}

func shouldProxy(attrValue, proxyOption string) bool {
	return !strings.HasPrefix(attrValue, "data:") &&
		(proxyOption == "all" || !urllib.IsHTTPS(attrValue))
}
