// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sanitizer // import "miniflux.app/v2/internal/reader/sanitizer"

import (
	"io"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/reader/urlcleaner"
	"miniflux.app/v2/internal/urllib"

	"golang.org/x/net/html"
)

var (
	tagAllowList = map[string][]string{
		"a":          {"href", "title", "id"},
		"abbr":       {"title"},
		"acronym":    {"title"},
		"aside":      {},
		"audio":      {"src"},
		"blockquote": {},
		"b":          {},
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
		"hr":         {},
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
		"tfoot":      {},
		"th":         {"rowspan", "colspan"},
		"thead":      {},
		"time":       {"datetime"},
		"tr":         {},
		"u":          {},
		"ul":         {"id"},
		"var":        {},
		"video":      {"poster", "height", "width", "src"},
		"wbr":        {},

		// MathML: https://w3c.github.io/mathml-core/ and https://developer.mozilla.org/en-US/docs/Web/MathML/Reference/Element
		"annotation":     {},
		"annotation-xml": {},
		"maction":        {},
		"math":           {"xmlns"},
		"merror":         {},
		"mfrac":          {},
		"mi":             {},
		"mmultiscripts":  {},
		"mn":             {},
		"mo":             {},
		"mover":          {},
		"mpadded":        {},
		"mphantom":       {},
		"mprescripts":    {},
		"mroot":          {},
		"mrow":           {},
		"ms":             {},
		"mspace":         {},
		"msqrt":          {},
		"mstyle":         {},
		"msub":           {},
		"msubsup":        {},
		"msup":           {},
		"mtable":         {},
		"mtd":            {},
		"mtext":          {},
		"mtr":            {},
		"munder":         {},
		"munderover":     {},
		"semantics":      {},
	}
)

type SanitizerOptions struct {
	OpenLinksInNewTab bool
}

func SanitizeHTMLWithDefaultOptions(baseURL, rawHTML string) string {
	return SanitizeHTML(baseURL, rawHTML, &SanitizerOptions{
		OpenLinksInNewTab: true,
	})
}

func SanitizeHTML(baseURL, rawHTML string, sanitizerOptions *SanitizerOptions) string {
	var buffer strings.Builder
	var tagStack []string
	var parentTag string
	var blockedStack []string

	tokenizer := html.NewTokenizer(strings.NewReader(rawHTML))
	for {
		if tokenizer.Next() == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				return buffer.String()
			}

			return ""
		}

		token := tokenizer.Token()

		// Note: MathML elements are not fully supported by golang.org/x/net/html.
		// See https://github.com/golang/net/blob/master/html/atom/gen.go
		// and https://github.com/golang/net/blob/master/html/atom/table.go
		tagName := token.Data
		if tagName == "" {
			continue
		}

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

			buffer.WriteString(token.String())
		case html.StartTagToken:
			parentTag = tagName

			if isPixelTracker(tagName, token.Attr) {
				continue
			}

			if isBlockedTag(tagName) || slices.ContainsFunc(token.Attr, func(attr html.Attribute) bool { return attr.Key == "hidden" }) {
				blockedStack = append(blockedStack, tagName)
				continue
			}

			if len(blockedStack) == 0 && isValidTag(tagName) {
				attrNames, htmlAttributes := sanitizeAttributes(baseURL, tagName, token.Attr, sanitizerOptions)
				if hasRequiredAttributes(tagName, attrNames) {
					if len(attrNames) > 0 {
						// Rewrite the start tag with allowed attributes.
						buffer.WriteString("<" + tagName + " " + htmlAttributes + ">")
					} else {
						// Rewrite the start tag without any attributes.
						buffer.WriteString("<" + tagName + ">")
					}

					tagStack = append(tagStack, tagName)
				}
			}
		case html.EndTagToken:
			if len(blockedStack) == 0 {
				if isValidTag(tagName) && slices.Contains(tagStack, tagName) {
					buffer.WriteString("</" + tagName + ">")
				}
			} else {
				if blockedStack[len(blockedStack)-1] == tagName {
					blockedStack = blockedStack[:len(blockedStack)-1]
				}
			}
		case html.SelfClosingTagToken:
			if isPixelTracker(tagName, token.Attr) {
				continue
			}
			if len(blockedStack) == 0 && isValidTag(tagName) {
				attrNames, htmlAttributes := sanitizeAttributes(baseURL, tagName, token.Attr, sanitizerOptions)
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

func sanitizeAttributes(baseURL, tagName string, attributes []html.Attribute, sanitizerOptions *SanitizerOptions) ([]string, string) {
	var htmlAttrs, attrNames []string
	var err error
	var isImageLargerThanLayout bool
	var isAnchorLink bool

	if tagName == "img" {
		imgWidth := getIntegerAttributeValue("width", attributes)
		isImageLargerThanLayout = imgWidth > 750
	}

	parsedBaseUrl, _ := url.Parse(baseURL)

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

				// TODO use feedURL instead of baseURL twice.
				parsedValueUrl, _ := url.Parse(value)
				if cleanedURL, err := urlcleaner.RemoveTrackingParameters(parsedBaseUrl, parsedBaseUrl, parsedValueUrl); err == nil {
					value = cleanedURL
				}
			}
		}

		attrNames = append(attrNames, attribute.Key)
		htmlAttrs = append(htmlAttrs, attribute.Key+`="`+html.EscapeString(value)+`"`)
	}

	if !isAnchorLink {
		extraAttrNames, extraHTMLAttributes := getExtraAttributes(tagName, sanitizerOptions)
		if len(extraAttrNames) > 0 {
			attrNames = append(attrNames, extraAttrNames...)
			htmlAttrs = append(htmlAttrs, extraHTMLAttributes...)
		}
	}

	return attrNames, strings.Join(htmlAttrs, " ")
}

func getExtraAttributes(tagName string, sanitizerOptions *SanitizerOptions) ([]string, []string) {
	switch tagName {
	case "a":
		attributeNames := []string{"rel", "referrerpolicy"}
		htmlAttributes := []string{`rel="noopener noreferrer"`, `referrerpolicy="no-referrer"`}
		if sanitizerOptions.OpenLinksInNewTab {
			attributeNames = append(attributeNames, "target")
			htmlAttributes = append(htmlAttributes, `target="_blank"`)
		}
		return attributeNames, htmlAttributes
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
	_, ok := tagAllowList[tagName]
	return ok
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
			switch attribute.Key {
			case "height":
				hasHeight = true
			case "width":
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
	case "iframe":
		return slices.Contains(attributes, "src")
	case "source", "img":
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
		"open.spotify.com",
	}
	domain := urllib.Domain(src)

	// allow iframe from same origin
	if urllib.Domain(baseURL) == domain {
		return true
	}

	// allow iframe from custom invidious instance
	if config.Opts.InvidiousInstance() == domain {
		return true
	}

	return slices.Contains(whitelist, strings.TrimPrefix(domain, "www."))
}

func rewriteIframeURL(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		return link
	}

	switch strings.TrimPrefix(u.Hostname(), "www.") {
	case "youtube.com":
		if strings.HasPrefix(u.Path, "/embed/") {
			if len(u.RawQuery) > 0 {
				return config.Opts.YouTubeEmbedUrlOverride() + strings.TrimPrefix(u.Path, "/embed/") + "?" + u.RawQuery
			}
			return config.Opts.YouTubeEmbedUrlOverride() + strings.TrimPrefix(u.Path, "/embed/")
		}
	case "player.vimeo.com":
		// See https://help.vimeo.com/hc/en-us/articles/12426260232977-About-Player-parameters
		if strings.HasPrefix(u.Path, "/video/") {
			if len(u.RawQuery) > 0 {
				return link + "&dnt=1"
			}
			return link + "?dnt=1"
		}
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
		if absoluteURL, err := urllib.AbsoluteURL(baseURL, imageCandidate.ImageURL); err == nil {
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

func getIntegerAttributeValue(name string, attributes []html.Attribute) int {
	for _, attribute := range attributes {
		if attribute.Key == name {
			number, _ := strconv.Atoi(attribute.Val)
			return number
		}
	}
	return 0
}
