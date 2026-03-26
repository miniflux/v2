// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package mediaproxy // import "miniflux.app/v2/internal/mediaproxy"

import (
	"net/url"
	"slices"
	"strings"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/reader/sanitizer"

	"github.com/PuerkitoBio/goquery"
)

type urlProxyRewriter func(url string) string

func RewriteDocumentWithRelativeProxyURL(htmlDocument string) string {
	return genericProxyRewriter(ProxifyRelativeURL, htmlDocument)
}

func RewriteDocumentWithAbsoluteProxyURL(htmlDocument string) string {
	return genericProxyRewriter(ProxifyAbsoluteURL, htmlDocument)
}

func genericProxyRewriter(proxifyFunction urlProxyRewriter, htmlDocument string) string {
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
						img.SetAttr("src", proxifyFunction(srcAttrValue))
					}
				}

				if srcsetAttrValue, ok := img.Attr("srcset"); ok {
					proxifySourceSet(img, proxifyFunction, proxyOption, srcsetAttrValue)
				}
			})

			if !slices.Contains(config.Opts.MediaProxyResourceTypes(), "video") {
				doc.Find("video").Each(func(i int, video *goquery.Selection) {
					if posterAttrValue, ok := video.Attr("poster"); ok {
						if shouldProxifyURL(posterAttrValue, proxyOption) {
							video.SetAttr("poster", proxifyFunction(posterAttrValue))
						}
					}
				})
			}

		case "audio":
			doc.Find("audio, audio source").Each(func(i int, audio *goquery.Selection) {
				if srcAttrValue, ok := audio.Attr("src"); ok {
					if shouldProxifyURL(srcAttrValue, proxyOption) {
						audio.SetAttr("src", proxifyFunction(srcAttrValue))
					}
				}
			})

		case "video":
			doc.Find("video, video source").Each(func(i int, video *goquery.Selection) {
				if srcAttrValue, ok := video.Attr("src"); ok {
					if shouldProxifyURL(srcAttrValue, proxyOption) {
						video.SetAttr("src", proxifyFunction(srcAttrValue))
					}
				}

				if posterAttrValue, ok := video.Attr("poster"); ok {
					if shouldProxifyURL(posterAttrValue, proxyOption) {
						video.SetAttr("poster", proxifyFunction(posterAttrValue))
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

func proxifySourceSet(element *goquery.Selection, proxifyFunction urlProxyRewriter, proxyOption, srcsetAttrValue string) {
	imageCandidates := sanitizer.ParseSrcSetAttribute(srcsetAttrValue)

	for _, imageCandidate := range imageCandidates {
		if shouldProxifyURL(imageCandidate.ImageURL, proxyOption) {
			imageCandidate.ImageURL = proxifyFunction(imageCandidate.ImageURL)
		}
	}

	element.SetAttr("srcset", imageCandidates.String())
}

// shouldProxifyURL checks if the media URL should be proxified based on the media proxy option and URL scheme.
func shouldProxifyURL(mediaURL, mediaProxyOption string) bool {
	parsedURL, err := url.Parse(mediaURL)
	if err != nil || !parsedURL.IsAbs() || parsedURL.Host == "" {
		return false
	}

	switch {
	case mediaProxyOption == "all" && (strings.EqualFold(parsedURL.Scheme, "http") || strings.EqualFold(parsedURL.Scheme, "https")):
		return true
	case mediaProxyOption != "none" && strings.EqualFold(parsedURL.Scheme, "http"):
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

	mediaMimeType = strings.ToLower(mediaMimeType)

	for _, mediaType := range mediaProxyResourceTypes {
		if strings.HasPrefix(mediaMimeType, mediaType+"/") {
			return true
		}
	}

	return false
}
