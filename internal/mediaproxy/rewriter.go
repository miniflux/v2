// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package mediaproxy // import "miniflux.app/v2/internal/mediaproxy"

import (
	"slices"
	"strings"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/reader/sanitizer"
	"miniflux.app/v2/internal/urllib"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

type urlProxyRewriter func(router *mux.Router, url string) string

func RewriteDocumentWithRelativeProxyURL(router *mux.Router, htmlDocument string) string {
	return genericProxyRewriter(router, ProxifyRelativeURL, htmlDocument)
}

func RewriteDocumentWithAbsoluteProxyURL(router *mux.Router, htmlDocument string) string {
	return genericProxyRewriter(router, ProxifyAbsoluteURL, htmlDocument)
}

func genericProxyRewriter(router *mux.Router, proxifyFunction urlProxyRewriter, htmlDocument string) string {
	proxyOption := config.Opts.MediaProxyMode()
	if proxyOption == "none" {
		return htmlDocument
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlDocument))
	if err != nil {
		return htmlDocument
	}

	for _, mediaType := range config.Opts.MediaProxyResourceTypes() {
		switch mediaType {
		case "image":
			doc.Find("img, picture source").Each(func(i int, img *goquery.Selection) {
				if srcAttrValue, ok := img.Attr("src"); ok {
					if shouldProxifyURL(srcAttrValue, proxyOption) {
						img.SetAttr("src", proxifyFunction(router, srcAttrValue))
					}
				}

				if srcsetAttrValue, ok := img.Attr("srcset"); ok {
					proxifySourceSet(img, router, proxifyFunction, proxyOption, srcsetAttrValue)
				}
			})

			if !slices.Contains(config.Opts.MediaProxyResourceTypes(), "video") {
				doc.Find("video").Each(func(i int, video *goquery.Selection) {
					if posterAttrValue, ok := video.Attr("poster"); ok {
						if shouldProxifyURL(posterAttrValue, proxyOption) {
							video.SetAttr("poster", proxifyFunction(router, posterAttrValue))
						}
					}
				})
			}

		case "audio":
			doc.Find("audio, audio source").Each(func(i int, audio *goquery.Selection) {
				if srcAttrValue, ok := audio.Attr("src"); ok {
					if shouldProxifyURL(srcAttrValue, proxyOption) {
						audio.SetAttr("src", proxifyFunction(router, srcAttrValue))
					}
				}
			})

		case "video":
			doc.Find("video, video source").Each(func(i int, video *goquery.Selection) {
				if srcAttrValue, ok := video.Attr("src"); ok {
					if shouldProxifyURL(srcAttrValue, proxyOption) {
						video.SetAttr("src", proxifyFunction(router, srcAttrValue))
					}
				}

				if posterAttrValue, ok := video.Attr("poster"); ok {
					if shouldProxifyURL(posterAttrValue, proxyOption) {
						video.SetAttr("poster", proxifyFunction(router, posterAttrValue))
					}
				}
			})
		}
	}

	output, err := doc.FindMatcher(goquery.Single("body")).Html()
	if err != nil {
		return htmlDocument
	}

	return output
}

func proxifySourceSet(element *goquery.Selection, router *mux.Router, proxifyFunction urlProxyRewriter, proxyOption, srcsetAttrValue string) {
	imageCandidates := sanitizer.ParseSrcSetAttribute(srcsetAttrValue)

	for _, imageCandidate := range imageCandidates {
		if shouldProxifyURL(imageCandidate.ImageURL, proxyOption) {
			imageCandidate.ImageURL = proxifyFunction(router, imageCandidate.ImageURL)
		}
	}

	element.SetAttr("srcset", imageCandidates.String())
}

// shouldProxifyURL checks if the media URL should be proxified based on the media proxy option and URL scheme.
func shouldProxifyURL(mediaURL, mediaProxyOption string) bool {
	switch {
	case mediaURL == "":
		return false
	case strings.HasPrefix(mediaURL, "data:"):
		return false
	case mediaProxyOption == "all":
		return true
	case mediaProxyOption != "none" && !urllib.IsHTTPS(mediaURL):
		return true
	default:
		return false
	}
}

// ShouldProxifyURLWithMimeType checks if the media URL should be proxified based on the media proxy option, URL scheme, and MIME type.
func ShouldProxifyURLWithMimeType(mediaURL, mediaMimeType, mediaProxyOption string, mediaProxyResourceTypes []string) bool {
	if !shouldProxifyURL(mediaURL, mediaProxyOption) {
		return false
	}

	for _, mediaType := range mediaProxyResourceTypes {
		if strings.HasPrefix(mediaMimeType, mediaType+"/") {
			return true
		}
	}

	return false
}
