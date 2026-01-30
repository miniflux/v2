// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sanitizer // import "miniflux.app/v2/internal/reader/sanitizer"

import (
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
	allowedHTMLTagsAndAttributes = map[string][]string{
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
		"i":          {},
		"iframe":     {"width", "height", "frameborder", "src", "allowfullscreen"},
		"img":        {"alt", "title", "src", "srcset", "sizes", "width", "height", "fetchpriority", "decoding"},
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
		"small":      {},
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

	// See https://www.iana.org/assignments/uri-schemes/uri-schemes.xhtml
	validURISchemes = []string{
		// Most commong schemes on top.
		"https:",
		"http:",

		// Then the rest.
		"apt:",
		"bitcoin:",
		"callto:",
		"dav:",
		"davs:",
		"ed2k:",
		"facetime:",
		"feed:",
		"ftp:",
		"geo:",
		"git:",
		"gopher:",
		"irc:",
		"irc6:",
		"ircs:",
		"itms-apps:",
		"itms:",
		"magnet:",
		"mailto:",
		"news:",
		"nntp:",
		"rtmp:",
		"sftp:",
		"sip:",
		"sips:",
		"skype:",
		"spotify:",
		"ssh:",
		"steam:",
		"svn:",
		"svn+ssh:",
		"tel:",
		"webcal:",
		"xmpp:",

		// iOS Apps
		"opener:", // https://www.opener.link
		"hack:",   // https://apps.apple.com/it/app/hack-for-hacker-news-reader/id1464477788?l=en-GB
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

// SanitizerOptions holds options for the HTML sanitizer.
type SanitizerOptions struct {
	OpenLinksInNewTab bool
}

// SanitizeHTML takes raw HTML input and removes any disallowed tags and attributes.
func SanitizeHTML(baseURL, rawHTML string, sanitizerOptions *SanitizerOptions) string {
	var buffer strings.Builder

	// Educated guess about how big the sanitized HTML will be,
	// to reduce the amount of buffer re-allocations in this function.
	estimatedRatio := len(rawHTML) * 3 / 4
	buffer.Grow(estimatedRatio)

	// We need to surround `rawHTML` with body tags so that html.Parse
	// will consider it a valid html document.
	doc, err := html.Parse(strings.NewReader("<body>" + rawHTML + "</body>"))
	if err != nil {
		return ""
	}

	/* The structure of `doc` is always:
	<html>
	<head>...</head>
	<body>..</body>
	</html>
	*/
	body := doc.FirstChild.FirstChild.NextSibling

	// Errors are a non-issue, so they're handled in filterAndRenderHTML
	parsedBaseUrl, _ := url.Parse(baseURL)
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		filterAndRenderHTML(&buffer, c, parsedBaseUrl, sanitizerOptions)
	}

	return buffer.String()
}

func isHidden(n *html.Node) bool {
	for _, attr := range n.Attr {
		if strings.ToLower(attr.Key) == "hidden" {
			return true
		}
	}
	return false
}

func shouldIgnoreTag(n *html.Node, tag string) bool {
	if isPixelTracker(tag, n.Attr) {
		return true
	}
	if isBlockedTag(tag) {
		return true
	}
	if isHidden(n) {
		return true
	}

	return false
}

func isSelfContainedTag(tag string) bool {
	switch tag {
	case "area", "base", "br", "col", "embed", "hr", "img", "input",
		"link", "meta", "param", "source", "track", "wbr":
		return true
	}
	return false
}

func filterAndRenderHTMLChildren(buf *strings.Builder, n *html.Node, parsedBaseUrl *url.URL, sanitizerOptions *SanitizerOptions) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		filterAndRenderHTML(buf, c, parsedBaseUrl, sanitizerOptions)
	}
}

func filterAndRenderHTML(buf *strings.Builder, n *html.Node, parsedBaseUrl *url.URL, sanitizerOptions *SanitizerOptions) {
	if n == nil {
		return
	}

	switch n.Type {
	case html.TextNode:
		buf.WriteString(html.EscapeString(n.Data))
	case html.ElementNode:
		tag := strings.ToLower(n.Data)
		if shouldIgnoreTag(n, tag) {
			return
		}

		_, ok := allowedHTMLTagsAndAttributes[tag]
		if !ok {
			// The tag isn't allowed, but we're still interested in its content
			filterAndRenderHTMLChildren(buf, n, parsedBaseUrl, sanitizerOptions)
			return
		}

		attrNames, htmlAttributes := sanitizeAttributes(parsedBaseUrl, tag, n.Attr, sanitizerOptions)
		if !hasRequiredAttributes(tag, attrNames) {
			// The tag doesn't have every required attributes but we're still interested in its content
			filterAndRenderHTMLChildren(buf, n, parsedBaseUrl, sanitizerOptions)
			return
		}
		buf.WriteString("<")
		buf.WriteString(n.Data)
		if len(attrNames) > 0 {
			buf.WriteString(" " + htmlAttributes)
		}
		buf.WriteString(">")

		if isSelfContainedTag(tag) {
			return
		}

		if tag != "iframe" {
			// iframes aren't allowed to have child nodes.
			filterAndRenderHTMLChildren(buf, n, parsedBaseUrl, sanitizerOptions)
		}

		buf.WriteString("</")
		buf.WriteString(n.Data)
		buf.WriteString(">")
	case html.DocumentNode:
		filterAndRenderHTMLChildren(buf, n, parsedBaseUrl, sanitizerOptions)
	default:
	}
}

func sanitizeAttributes(parsedBaseUrl *url.URL, tagName string, attributes []html.Attribute, sanitizerOptions *SanitizerOptions) ([]string, string) {
	htmlAttrs := make([]string, 0, len(attributes))
	attrNames := make([]string, 0, len(attributes))

	var err error
	var isAnchorLink bool
	var isYouTubeEmbed bool

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
				iframeSourceDomain, trustedIframeDomain := findAllowedIframeSourceDomain(attribute.Val)
				if !trustedIframeDomain {
					continue
				}

				value = rewriteIframeURL(attribute.Val)

				if iframeSourceDomain == "youtube.com" || iframeSourceDomain == "youtube-nocookie.com" {
					isYouTubeEmbed = true
				}
			case tagName == "img" && attribute.Key == "src" && isValidDataAttribute(attribute.Val):
				value = attribute.Val
			case tagName == "a" && attribute.Key == "href" && strings.HasPrefix(attribute.Val, "#"):
				value = attribute.Val
				isAnchorLink = true
			default:
				value, err = urllib.ResolveToAbsoluteURLWithParsedBaseURL(parsedBaseUrl, value)
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
		extraAttrNames, extraHTMLAttributes := getExtraAttributes(tagName, isYouTubeEmbed, sanitizerOptions)
		if len(extraAttrNames) > 0 {
			attrNames = append(attrNames, extraAttrNames...)
			htmlAttrs = append(htmlAttrs, extraHTMLAttributes...)
		}
	}

	return attrNames, strings.Join(htmlAttrs, " ")
}

func getExtraAttributes(tagName string, isYouTubeEmbed bool, sanitizerOptions *SanitizerOptions) ([]string, []string) {
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
		extraAttrNames := []string{}
		extraHTMLAttributes := []string{}

		// Note: the referrerpolicy seems to be required to avoid YouTube error 153 video player configuration error
		// See https://developers.google.com/youtube/terms/required-minimum-functionality#embedded-player-api-client-identity
		if isYouTubeEmbed {
			extraAttrNames = append(extraAttrNames, "referrerpolicy")
			extraHTMLAttributes = append(extraHTMLAttributes, `referrerpolicy="strict-origin-when-cross-origin"`)
		}

		extraAttrNames = append(extraAttrNames, "sandbox", "loading")
		extraHTMLAttributes = append(extraHTMLAttributes, `sandbox="allow-scripts allow-same-origin allow-popups allow-popups-to-escape-sandbox"`, `loading="lazy"`)

		return extraAttrNames, extraHTMLAttributes
	case "img":
		return []string{"loading"}, []string{`loading="lazy"`}
	default:
		return nil, nil
	}
}

func isValidAttribute(tagName, attributeName string) bool {
	if attributes, ok := allowedHTMLTagsAndAttributes[tagName]; ok {
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
		for _, attribute := range attributes {
			if attribute == "src" || attribute == "srcset" {
				return true
			}
		}
		return false
	default:
		return true
	}
}

func hasValidURIScheme(absoluteURL string) bool {
	for _, scheme := range validURISchemes {
		if strings.HasPrefix(absoluteURL, scheme) {
			return true
		}
	}
	return false
}

func isBlockedResource(absoluteURL string) bool {
	for _, blockedURL := range blockedResourceURLSubstrings {
		if strings.Contains(absoluteURL, blockedURL) {
			return true
		}
	}
	return false
}

func findAllowedIframeSourceDomain(iframeSourceURL string) (string, bool) {
	iframeSourceDomain := urllib.DomainWithoutWWW(iframeSourceURL)

	if _, ok := iframeAllowList[iframeSourceDomain]; ok {
		return iframeSourceDomain, true
	}

	if ytDomain := config.Opts.YouTubeEmbedDomain(); ytDomain != "" && iframeSourceDomain == strings.TrimPrefix(ytDomain, "www.") {
		return iframeSourceDomain, true
	}

	if invidiousInstance := config.Opts.InvidiousInstance(); invidiousInstance != "" && iframeSourceDomain == strings.TrimPrefix(invidiousInstance, "www.") {
		return iframeSourceDomain, true
	}

	return "", false
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
		if absoluteURL, err := urllib.ResolveToAbsoluteURLWithParsedBaseURL(parsedBaseURL, imageCandidate.ImageURL); err == nil {
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
