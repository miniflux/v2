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
	allowedHTMLTagsAndAttributes = map[string]map[string]struct{}{
		"a":          {"href": {}, "title": {}, "id": {}},
		"abbr":       {"title": {}},
		"acronym":    {"title": {}},
		"aside":      {},
		"audio":      {"src": {}},
		"blockquote": {},
		"b":          {},
		"br":         {},
		"caption":    {},
		"cite":       {},
		"code":       {},
		"dd":         {"id": {}},
		"del":        {},
		"dfn":        {},
		"dl":         {"id": {}},
		"dt":         {"id": {}},
		"em":         {},
		"figcaption": {},
		"figure":     {},
		"h1":         {"id": {}},
		"h2":         {"id": {}},
		"h3":         {"id": {}},
		"h4":         {"id": {}},
		"h5":         {"id": {}},
		"h6":         {"id": {}},
		"hr":         {},
		"iframe":     {"width": {}, "height": {}, "frameborder": {}, "src": {}, "allowfullscreen": {}},
		"img":        {"alt": {}, "title": {}, "src": {}, "srcset": {}, "sizes": {}, "width": {}, "height": {}, "fetchpriority": {}, "decoding": {}},
		"ins":        {},
		"kbd":        {},
		"li":         {"id": {}},
		"ol":         {"id": {}},
		"p":          {},
		"picture":    {},
		"pre":        {},
		"q":          {"cite": {}},
		"rp":         {},
		"rt":         {},
		"rtc":        {},
		"ruby":       {},
		"s":          {},
		"samp":       {},
		"source":     {"src": {}, "type": {}, "srcset": {}, "sizes": {}, "media": {}},
		"strong":     {},
		"sub":        {},
		"sup":        {"id": {}},
		"table":      {},
		"td":         {"rowspan": {}, "colspan": {}},
		"tfoot":      {},
		"th":         {"rowspan": {}, "colspan": {}},
		"thead":      {},
		"time":       {"datetime": {}},
		"tr":         {},
		"u":          {},
		"ul":         {"id": {}},
		"var":        {},
		"video":      {"poster": {}, "height": {}, "width": {}, "src": {}},
		"wbr":        {},

		// MathML: https://w3c.github.io/mathml-core/ and https://developer.mozilla.org/en-US/docs/Web/MathML/Reference/Element
		"annotation":     {},
		"annotation-xml": {},
		"maction":        {},
		"math":           {"xmlns": {}},
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

	iframeAllowList = map[string]struct{}{
		"bandcamp.com":         {},
		"cdn.embedly.com":      {},
		"dailymotion.com":      {},
		"open.spotify.com":     {},
		"player.bilibili.com":  {},
		"player.twitch.tv":     {},
		"player.vimeo.com":     {},
		"soundcloud.com":       {},
		"vk.com":               {},
		"w.soundcloud.com":     {},
		"youtube-nocookie.com": {},
		"youtube.com":          {},
	}

	blockedResourceURLSubstrings = []string{
		"api.flattr.com",
		"feeds.feedburner.com",
		"feedsportal.com",
		"pinterest.com/pin/create/button/",
		"stats.wordpress.com",
		"twitter.com/intent/tweet",
		"twitter.com/share",
		"facebook.com/sharer.php",
		"linkedin.com/shareArticle",
	}

	validURISchemes = map[string]struct{}{
		"apt":       {},
		"bitcoin":   {},
		"callto":    {},
		"dav":       {},
		"davs":      {},
		"ed2k":      {},
		"facetime":  {},
		"feed":      {},
		"ftp":       {},
		"geo":       {},
		"git":       {},
		"gopher":    {},
		"http":      {},
		"https":     {},
		"irc":       {},
		"irc6":      {},
		"ircs":      {},
		"itms-apps": {},
		"itms":      {},
		"magnet":    {},
		"mailto":    {},
		"news":      {},
		"nntp":      {},
		"rtmp":      {},
		"sftp":      {},
		"sip":       {},
		"sips":      {},
		"skype":     {},
		"spotify":   {},
		"ssh":       {},
		"steam":     {},
		"svn":       {},
		"svn+ssh":   {},
		"tel":       {},
		"webcal":    {},
		"xmpp":      {},
		// iOS Apps
		"opener": {}, // https://www.opener.link
		"hack":   {}, // https://apps.apple.com/it/app/hack-for-hacker-news-reader/id1464477788?l=en-GB
	}

	dataAttributeAllowedPrefixes = []string{
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
	var tagStack []string
	var parentTag string
	var blockedStack []string
	var buffer strings.Builder

	// Educated guess about how big the sanitized HTML will be,
	// to reduce the amount of buffer re-allocations in this function.
	estimatedRatio := len(rawHTML) * 3 / 4
	buffer.Grow(estimatedRatio)

	// Errors are a non-issue, so they're handled later in the function.
	parsedBaseUrl, _ := url.Parse(baseURL)

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
				attrNames, htmlAttributes := sanitizeAttributes(parsedBaseUrl, tagName, token.Attr, sanitizerOptions)
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
				attrNames, htmlAttributes := sanitizeAttributes(parsedBaseUrl, tagName, token.Attr, sanitizerOptions)
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

func sanitizeAttributes(parsedBaseUrl *url.URL, tagName string, attributes []html.Attribute, sanitizerOptions *SanitizerOptions) ([]string, string) {
	var htmlAttrs, attrNames []string
	var err error
	var isAnchorLink bool

	for _, attribute := range attributes {
		if !isValidAttribute(tagName, attribute.Key) {
			continue
		}

		value := attribute.Val

		switch tagName {
		case "math":
			if attribute.Key == "xmlns" {
				if value != "http://www.w3.org/1998/Math/MathML" {
					value = "http://www.w3.org/1998/Math/MathML"
				}
			}
		case "img":
			switch attribute.Key {
			case "fetchpriority":
				if !isValidFetchPriorityValue(value) {
					continue
				}
			case "decoding":
				if !isValidDecodingValue(value) {
					continue
				}
			case "width", "height":
				if !isPositiveInteger(value) {
					continue
				}

				// Discard width and height attributes when width is larger than Miniflux layout (750px)
				if imgWidth := getIntegerAttributeValue("width", attributes); imgWidth > 750 {
					continue
				}
			case "srcset":
				value = sanitizeSrcsetAttr(parsedBaseUrl, value)
			}
		case "source":
			if attribute.Key == "srcset" {
				value = sanitizeSrcsetAttr(parsedBaseUrl, value)
			}
		}

		if isExternalResourceAttribute(attribute.Key) {
			switch {
			case tagName == "iframe":
				if !isValidIframeSource(attribute.Val) {
					continue
				}
				value = rewriteIframeURL(attribute.Val)
			case tagName == "img" && attribute.Key == "src" && isValidDataAttribute(attribute.Val):
				value = attribute.Val
			case tagName == "a" && attribute.Key == "href" && strings.HasPrefix(attribute.Val, "#"):
				value = attribute.Val
				isAnchorLink = true
			default:
				value, err = absoluteURLParsedBase(parsedBaseUrl, value)
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
	_, ok := allowedHTMLTagsAndAttributes[tagName]
	return ok
}

func isValidAttribute(tagName, attributeName string) bool {
	if attributes, ok := allowedHTMLTagsAndAttributes[tagName]; ok {
		_, allowed := attributes[attributeName]
		return allowed
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
		if attribute.Val == "1" || attribute.Val == "0" {
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
func hasValidURIScheme(absoluteURL string) bool {
	colonIndex := strings.IndexByte(absoluteURL, ':')
	// Scheme must exist (colonIndex > 0). An empty scheme (e.g. ":foo") is not allowed.
	if colonIndex <= 0 {
		return false
	}

	scheme := absoluteURL[:colonIndex]
	_, ok := validURISchemes[strings.ToLower(scheme)]
	return ok
}

func isBlockedResource(absoluteURL string) bool {
	return slices.ContainsFunc(blockedResourceURLSubstrings, func(element string) bool {
		return strings.Contains(absoluteURL, element)
	})
}

func isValidIframeSource(iframeSourceURL string) bool {
	iframeSourceDomain := urllib.DomainWithoutWWW(iframeSourceURL)

	if _, ok := iframeAllowList[iframeSourceDomain]; ok {
		return true
	}

	if ytDomain := config.Opts.YouTubeEmbedDomain(); ytDomain != "" && iframeSourceDomain == strings.TrimPrefix(ytDomain, "www.") {
		return true
	}

	if invidiousInstance := config.Opts.InvidiousInstance(); invidiousInstance != "" && iframeSourceDomain == strings.TrimPrefix(invidiousInstance, "www.") {
		return true
	}

	return false
}

func rewriteIframeURL(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		return link
	}

	switch strings.TrimPrefix(u.Hostname(), "www.") {
	case "youtube.com":
		if pathWithoutEmbed, ok := strings.CutPrefix(u.Path, "/embed/"); ok {
			if len(u.RawQuery) > 0 {
				return config.Opts.YouTubeEmbedUrlOverride() + pathWithoutEmbed + "?" + u.RawQuery
			}
			return config.Opts.YouTubeEmbedUrlOverride() + pathWithoutEmbed
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
	switch tagName {
	case "noscript", "script", "style":
		return true
	}
	return false
}

func sanitizeSrcsetAttr(parsedBaseURL *url.URL, value string) string {
	imageCandidates := ParseSrcSetAttribute(value)

	for _, imageCandidate := range imageCandidates {
		if absoluteURL, err := absoluteURLParsedBase(parsedBaseURL, imageCandidate.ImageURL); err == nil {
			imageCandidate.ImageURL = absoluteURL
		}
	}

	return imageCandidates.String()
}

func isValidDataAttribute(value string) bool {
	for _, prefix := range dataAttributeAllowedPrefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func isPositiveInteger(value string) bool {
	if value == "" {
		return false
	}
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

func isValidFetchPriorityValue(value string) bool {
	switch value {
	case "high", "low", "auto":
		return true
	}
	return false
}

func isValidDecodingValue(value string) bool {
	switch value {
	case "sync", "async", "auto":
		return true
	}
	return false
}

// absoluteURLParsedBase is used instead of urllib.AbsoluteURL to avoid parsing baseURL over and over.
func absoluteURLParsedBase(parsedBaseURL *url.URL, input string) (string, error) {
	absURL, u, err := urllib.GetAbsoluteURL(input)
	if err != nil {
		return "", err
	}
	if absURL != "" {
		return absURL, nil
	}
	if parsedBaseURL == nil {
		return "", nil
	}

	return parsedBaseURL.ResolveReference(u).String(), nil
}
