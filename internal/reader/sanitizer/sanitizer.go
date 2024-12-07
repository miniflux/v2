// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sanitizer // import "miniflux.app/v2/internal/reader/sanitizer"

import (
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/reader/urlcleaner"
	"miniflux.app/v2/internal/urllib"

	"golang.org/x/net/html"
)

var (
	youtubeEmbedRegex = regexp.MustCompile(`//(?:www\.)?youtube\.com/embed/(.+)$`)
	tagAllowList      = map[string][]string{
		"a":          {"href", "title", "id"},
		"abbr":       {"title"},
		"acronym":    {"title"},
		"aside":      {},
		"audio":      {"src"},
		"blockquote": {},
		"br":         {},
		"caption":    {},
		"cite":       {},
		"code":       {},
		"dd":         {"id"},
		"del":        {},
		"dfn":        {},
		"dl":         {"id"},
		"dt":         {"id"},
		"em":         {},
		"figcaption": {},
		"figure":     {},
		"h1":         {"id"},
		"h2":         {"id"},
		"h3":         {"id"},
		"h4":         {"id"},
		"h5":         {"id"},
		"h6":         {"id"},
		"iframe":     {"width", "height", "frameborder", "src", "allowfullscreen"},
		"img":        {"alt", "title", "src", "srcset", "sizes", "width", "height"},
		"ins":        {},
		"kbd":        {},
		"li":         {"id"},
		"ol":         {"id"},
		"p":          {},
		"picture":    {},
		"pre":        {},
		"q":          {"cite"},
		"rp":         {},
		"rt":         {},
		"rtc":        {},
		"ruby":       {},
		"s":          {},
		"samp":       {},
		"source":     {"src", "type", "srcset", "sizes", "media"},
		"strong":     {},
		"sub":        {},
		"sup":        {"id"},
		"table":      {},
		"td":         {"rowspan", "colspan"},
		"tfooter":    {},
		"th":         {"rowspan", "colspan"},
		"thead":      {},
		"time":       {"datetime"},
		"tr":         {},
		"ul":         {"id"},
		"var":        {},
		"video":      {"poster", "height", "width", "src"},
		"wbr":        {},
	}
)

// Sanitize returns safe HTML.
func Sanitize(baseURL, input string) string {
	var buffer strings.Builder
	var tagStack []string
	var parentTag string
	var blockedStack []string

	tokenizer := html.NewTokenizer(strings.NewReader(input))
	for {
		if tokenizer.Next() == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				return buffer.String()
			}

			return ""
		}

		token := tokenizer.Token()
		tagName := token.DataAtom.String()
		switch token.Type {
		case html.TextToken:
			if len(blockedStack) > 0 {
				continue
			}

			// An iframe element never has fallback content.
			// See https://www.w3.org/TR/2010/WD-html5-20101019/the-iframe-element.html#the-iframe-element
			if parentTag == "iframe" {
				continue
			}

			buffer.WriteString(html.EscapeString(token.Data))
		case html.StartTagToken:
			parentTag = tagName

			if isPixelTracker(tagName, token.Attr) {
				continue
			}

			if isBlockedTag(tagName) || slices.ContainsFunc(token.Attr, func(attr html.Attribute) bool { return attr.Key == "hidden" }) {
				blockedStack = append(blockedStack, tagName)
			} else if len(blockedStack) == 0 && isValidTag(tagName) {
				attrNames, htmlAttributes := sanitizeAttributes(baseURL, tagName, token.Attr)

				if hasRequiredAttributes(tagName, attrNames) {
					if len(attrNames) > 0 {
						buffer.WriteString("<" + tagName + " " + htmlAttributes + ">")
					} else {
						buffer.WriteString("<" + tagName + ">")
					}

					tagStack = append(tagStack, tagName)
				}
			}
		case html.EndTagToken:
			if len(blockedStack) > 0 && blockedStack[len(blockedStack)-1] == tagName {
				blockedStack = blockedStack[:len(blockedStack)-1]
			} else if len(blockedStack) == 0 && isValidTag(tagName) && slices.Contains(tagStack, tagName) {
				buffer.WriteString("</" + tagName + ">")
			}
		case html.SelfClosingTagToken:
			if isPixelTracker(tagName, token.Attr) {
				continue
			}
			if isValidTag(tagName) && len(blockedStack) == 0 {
				attrNames, htmlAttributes := sanitizeAttributes(baseURL, tagName, token.Attr)
				if hasRequiredAttributes(tagName, attrNames) {
					if len(attrNames) > 0 {
						buffer.WriteString("<" + tagName + " " + htmlAttributes + "/>")
					} else {
						buffer.WriteString("<" + tagName + "/>")
					}
				}
			}
		}
	}
}

func sanitizeAttributes(baseURL, tagName string, attributes []html.Attribute) ([]string, string) {
	var htmlAttrs, attrNames []string
	var err error
	var isImageLargerThanLayout bool
	var isAnchorLink bool

	if tagName == "img" {
		imgWidth := getIntegerAttributeValue("width", attributes)
		isImageLargerThanLayout = imgWidth > 750
	}

	for _, attribute := range attributes {
		value := attribute.Val

		if !isValidAttribute(tagName, attribute.Key) {
			continue
		}

		if (tagName == "img" || tagName == "source") && attribute.Key == "srcset" {
			value = sanitizeSrcsetAttr(baseURL, value)
		}

		if tagName == "img" && (attribute.Key == "width" || attribute.Key == "height") {
			if isImageLargerThanLayout || !isPositiveInteger(value) {
				continue
			}
		}

		if isExternalResourceAttribute(attribute.Key) {
			switch {
			case tagName == "iframe":
				if !isValidIframeSource(baseURL, attribute.Val) {
					continue
				}
				value = rewriteIframeURL(attribute.Val)
			case tagName == "img" && attribute.Key == "src" && isValidDataAttribute(attribute.Val):
				value = attribute.Val
			case tagName == "a" && attribute.Key == "href" && strings.HasPrefix(attribute.Val, "#"):
				value = attribute.Val
				isAnchorLink = true
			default:
				value, err = urllib.AbsoluteURL(baseURL, value)
				if err != nil {
					continue
				}

				if !hasValidURIScheme(value) || isBlockedResource(value) {
					continue
				}

				if cleanedURL, err := urlcleaner.RemoveTrackingParameters(value); err == nil {
					value = cleanedURL
				}
			}
		}

		attrNames = append(attrNames, attribute.Key)
		htmlAttrs = append(htmlAttrs, fmt.Sprintf(`%s=%q`, attribute.Key, html.EscapeString(value)))
	}

	if !isAnchorLink {
		extraAttrNames, extraHTMLAttributes := getExtraAttributes(tagName)
		if len(extraAttrNames) > 0 {
			attrNames = append(attrNames, extraAttrNames...)
			htmlAttrs = append(htmlAttrs, extraHTMLAttributes...)
		}
	}

	return attrNames, strings.Join(htmlAttrs, " ")
}

func getExtraAttributes(tagName string) ([]string, []string) {
	switch tagName {
	case "a":
		return []string{"rel", "target", "referrerpolicy"}, []string{`rel="noopener noreferrer"`, `target="_blank"`, `referrerpolicy="no-referrer"`}
	case "video", "audio":
		return []string{"controls"}, []string{"controls"}
	case "iframe":
		return []string{"sandbox", "loading"}, []string{`sandbox="allow-scripts allow-same-origin allow-popups allow-popups-to-escape-sandbox"`, `loading="lazy"`}
	case "img":
		return []string{"loading"}, []string{`loading="lazy"`}
	default:
		return nil, nil
	}
}

func isValidTag(tagName string) bool {
	if _, ok := tagAllowList[tagName]; ok {
		return true
	}
	return false
}

func isValidAttribute(tagName, attributeName string) bool {
	if attributes, ok := tagAllowList[tagName]; ok {
		return slices.Contains(attributes, attributeName)
	}
	return false
}

func isExternalResourceAttribute(attribute string) bool {
	switch attribute {
	case "src", "href", "poster", "cite":
		return true
	default:
		return false
	}
}

func isPixelTracker(tagName string, attributes []html.Attribute) bool {
	if tagName != "img" {
		return false
	}
	hasHeight := false
	hasWidth := false

	for _, attribute := range attributes {
		if attribute.Val == "1" {
			if attribute.Key == "height" {
				hasHeight = true
			} else if attribute.Key == "width" {
				hasWidth = true
			}
		}
	}

	return hasHeight && hasWidth
}

func hasRequiredAttributes(tagName string, attributes []string) bool {
	switch tagName {
	case "a":
		return slices.Contains(attributes, "href")
	case "iframe", "img":
		return slices.Contains(attributes, "src")
	case "source":
		return slices.Contains(attributes, "src") || slices.Contains(attributes, "srcset")
	default:
		return true
	}
}

// See https://www.iana.org/assignments/uri-schemes/uri-schemes.xhtml
func hasValidURIScheme(src string) bool {
	whitelist := []string{
		"apt:",
		"bitcoin:",
		"callto:",
		"dav:",
		"davs:",
		"ed2k://",
		"facetime://",
		"feed:",
		"ftp://",
		"geo:",
		"gopher://",
		"git://",
		"http://",
		"https://",
		"irc://",
		"irc6://",
		"ircs://",
		"itms://",
		"itms-apps://",
		"magnet:",
		"mailto:",
		"news:",
		"nntp:",
		"rtmp://",
		"sip:",
		"sips:",
		"skype:",
		"spotify:",
		"ssh://",
		"sftp://",
		"steam://",
		"svn://",
		"svn+ssh://",
		"tel:",
		"webcal://",
		"xmpp:",

		// iOS Apps
		"opener://", // https://www.opener.link
		"hack://",   // https://apps.apple.com/it/app/hack-for-hacker-news-reader/id1464477788?l=en-GB
	}

	return slices.ContainsFunc(whitelist, func(prefix string) bool {
		return strings.HasPrefix(src, prefix)
	})
}

func isBlockedResource(src string) bool {
	blacklist := []string{
		"feedsportal.com",
		"api.flattr.com",
		"stats.wordpress.com",
		"plus.google.com/share",
		"twitter.com/share",
		"feeds.feedburner.com",
	}

	return slices.ContainsFunc(blacklist, func(element string) bool {
		return strings.Contains(src, element)
	})
}

func isValidIframeSource(baseURL, src string) bool {
	whitelist := []string{
		"bandcamp.com",
		"cdn.embedly.com",
		"player.bilibili.com",
		"player.twitch.tv",
		"player.vimeo.com",
		"soundcloud.com",
		"vk.com",
		"w.soundcloud.com",
		"dailymotion.com",
		"youtube-nocookie.com",
		"youtube.com",
	}
	domain := urllib.Domain(src)

	// allow iframe from same origin
	if urllib.Domain(baseURL) == domain {
		return true
	}

	// allow iframe from custom invidious instance
	if config.Opts != nil && config.Opts.InvidiousInstance() == domain {
		return true
	}

	return slices.Contains(whitelist, strings.TrimPrefix(domain, "www."))
}

func rewriteIframeURL(link string) string {
	matches := youtubeEmbedRegex.FindStringSubmatch(link)
	if len(matches) == 2 {
		return config.Opts.YouTubeEmbedUrlOverride() + matches[1]
	}

	return link
}

func isBlockedTag(tagName string) bool {
	blacklist := []string{
		"noscript",
		"script",
		"style",
	}

	return slices.Contains(blacklist, tagName)
}

func sanitizeSrcsetAttr(baseURL, value string) string {
	imageCandidates := ParseSrcSetAttribute(value)

	for _, imageCandidate := range imageCandidates {
		absoluteURL, err := urllib.AbsoluteURL(baseURL, imageCandidate.ImageURL)
		if err == nil {
			imageCandidate.ImageURL = absoluteURL
		}
	}

	return imageCandidates.String()
}

func isValidDataAttribute(value string) bool {
	var dataAttributeAllowList = []string{
		"data:image/avif",
		"data:image/apng",
		"data:image/png",
		"data:image/svg",
		"data:image/svg+xml",
		"data:image/jpg",
		"data:image/jpeg",
		"data:image/gif",
		"data:image/webp",
	}
	return slices.ContainsFunc(dataAttributeAllowList, func(prefix string) bool {
		return strings.HasPrefix(value, prefix)
	})
}

func isPositiveInteger(value string) bool {
	if number, err := strconv.Atoi(value); err == nil {
		return number > 0
	}
	return false
}

func getAttributeValue(name string, attributes []html.Attribute) string {
	for _, attribute := range attributes {
		if attribute.Key == name {
			return attribute.Val
		}
	}
	return ""
}

func getIntegerAttributeValue(name string, attributes []html.Attribute) int {
	number, _ := strconv.Atoi(getAttributeValue(name, attributes))
	return number
}
