// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package sanitizer // import "miniflux.app/reader/sanitizer"

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"

	"miniflux.app/url"

	"golang.org/x/net/html"
)

var (
	youtubeEmbedRegex = regexp.MustCompile(`//www\.youtube\.com/embed/(.*)`)
)

// Sanitize returns safe HTML.
func Sanitize(baseURL, input string) string {
	tokenizer := html.NewTokenizer(bytes.NewBufferString(input))
	var buffer bytes.Buffer
	var tagStack []string
	blacklistedTagDepth := 0

	for {
		if tokenizer.Next() == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				return buffer.String()
			}

			return ""
		}

		token := tokenizer.Token()
		switch token.Type {
		case html.TextToken:
			if blacklistedTagDepth > 0 {
				continue
			}

			buffer.WriteString(html.EscapeString(token.Data))
		case html.StartTagToken:
			tagName := token.DataAtom.String()

			if !isPixelTracker(tagName, token.Attr) && isValidTag(tagName) {
				attrNames, htmlAttributes := sanitizeAttributes(baseURL, tagName, token.Attr)

				if hasRequiredAttributes(tagName, attrNames) {
					if len(attrNames) > 0 {
						buffer.WriteString("<" + tagName + " " + htmlAttributes + ">")
					} else {
						buffer.WriteString("<" + tagName + ">")
					}

					tagStack = append(tagStack, tagName)
				}
			} else if isBlacklistedTag(tagName) {
				blacklistedTagDepth++
			}
		case html.EndTagToken:
			tagName := token.DataAtom.String()
			if isValidTag(tagName) && inList(tagName, tagStack) {
				buffer.WriteString(fmt.Sprintf("</%s>", tagName))
			} else if isBlacklistedTag(tagName) {
				blacklistedTagDepth--
			}
		case html.SelfClosingTagToken:
			tagName := token.DataAtom.String()
			if !isPixelTracker(tagName, token.Attr) && isValidTag(tagName) {
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

	for _, attribute := range attributes {
		value := attribute.Val

		if !isValidAttribute(tagName, attribute.Key) {
			continue
		}

		if isExternalResourceAttribute(attribute.Key) {
			if tagName == "iframe" {
				if isValidIframeSource(attribute.Val) {
					value = rewriteIframeURL(attribute.Val)
				} else {
					continue
				}
			} else {
				value, err = url.AbsoluteURL(baseURL, value)
				if err != nil {
					continue
				}

				if !hasValidScheme(value) || isBlacklistedResource(value) {
					continue
				}
			}
		}

		attrNames = append(attrNames, attribute.Key)
		htmlAttrs = append(htmlAttrs, fmt.Sprintf(`%s="%s"`, attribute.Key, html.EscapeString(value)))
	}

	extraAttrNames, extraHTMLAttributes := getExtraAttributes(tagName)
	if len(extraAttrNames) > 0 {
		attrNames = append(attrNames, extraAttrNames...)
		htmlAttrs = append(htmlAttrs, extraHTMLAttributes...)
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
		return []string{"sandbox"}, []string{`sandbox="allow-scripts allow-same-origin"`}
	default:
		return nil, nil
	}
}

func isValidTag(tagName string) bool {
	for element := range getTagWhitelist() {
		if tagName == element {
			return true
		}
	}

	return false
}

func isValidAttribute(tagName, attributeName string) bool {
	for element, attributes := range getTagWhitelist() {
		if tagName == element {
			if inList(attributeName, attributes) {
				return true
			}
		}
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
	if tagName == "img" {
		hasHeight := false
		hasWidth := false

		for _, attribute := range attributes {
			if attribute.Key == "height" && attribute.Val == "1" {
				hasHeight = true
			}

			if attribute.Key == "width" && attribute.Val == "1" {
				hasWidth = true
			}
		}

		return hasHeight && hasWidth
	}

	return false
}

func hasRequiredAttributes(tagName string, attributes []string) bool {
	elements := make(map[string][]string)
	elements["a"] = []string{"href"}
	elements["iframe"] = []string{"src"}
	elements["img"] = []string{"src"}
	elements["source"] = []string{"src"}

	for element, attrs := range elements {
		if tagName == element {
			for _, attribute := range attributes {
				for _, attr := range attrs {
					if attr == attribute {
						return true
					}
				}
			}

			return false
		}
	}

	return true
}

func hasValidScheme(src string) bool {
	// See https://www.iana.org/assignments/uri-schemes/uri-schemes.xhtml
	whitelist := []string{
		"apt://",
		"bitcoin://",
		"callto://",
		"ed2k://",
		"facetime://",
		"feed://",
		"ftp://",
		"geo://",
		"gopher://",
		"git://",
		"http://",
		"https://",
		"irc://",
		"irc6://",
		"ircs://",
		"itms://",
		"jabber://",
		"magnet://",
		"mailto://",
		"maps://",
		"news://",
		"nfs://",
		"nntp://",
		"rtmp://",
		"sip://",
		"sips://",
		"skype://",
		"smb://",
		"sms://",
		"spotify://",
		"ssh://",
		"sftp://",
		"steam://",
		"svn://",
		"tel://",
		"webcal://",
		"xmpp://",
	}

	for _, prefix := range whitelist {
		if strings.HasPrefix(src, prefix) {
			return true
		}
	}

	return false
}

func isBlacklistedResource(src string) bool {
	blacklist := []string{
		"feedsportal.com",
		"api.flattr.com",
		"stats.wordpress.com",
		"plus.google.com/share",
		"twitter.com/share",
		"feeds.feedburner.com",
	}

	for _, element := range blacklist {
		if strings.Contains(src, element) {
			return true
		}
	}

	return false
}

func isValidIframeSource(src string) bool {
	whitelist := []string{
		"//www.youtube.com",
		"http://www.youtube.com",
		"https://www.youtube.com",
		"https://www.youtube-nocookie.com",
		"http://player.vimeo.com",
		"https://player.vimeo.com",
		"http://www.dailymotion.com",
		"https://www.dailymotion.com",
		"http://vk.com",
		"https://vk.com",
		"http://soundcloud.com",
		"https://soundcloud.com",
		"http://w.soundcloud.com",
		"https://w.soundcloud.com",
		"http://bandcamp.com",
		"https://bandcamp.com",
		"https://cdn.embedly.com",
	}

	for _, prefix := range whitelist {
		if strings.HasPrefix(src, prefix) {
			return true
		}
	}

	return false
}

func getTagWhitelist() map[string][]string {
	whitelist := make(map[string][]string)
	whitelist["img"] = []string{"alt", "title", "src"}
	whitelist["audio"] = []string{"src"}
	whitelist["video"] = []string{"poster", "height", "width", "src"}
	whitelist["source"] = []string{"src", "type"}
	whitelist["dt"] = []string{}
	whitelist["dd"] = []string{}
	whitelist["dl"] = []string{}
	whitelist["table"] = []string{}
	whitelist["caption"] = []string{}
	whitelist["thead"] = []string{}
	whitelist["tfooter"] = []string{}
	whitelist["tr"] = []string{}
	whitelist["td"] = []string{"rowspan", "colspan"}
	whitelist["th"] = []string{"rowspan", "colspan"}
	whitelist["h1"] = []string{}
	whitelist["h2"] = []string{}
	whitelist["h3"] = []string{}
	whitelist["h4"] = []string{}
	whitelist["h5"] = []string{}
	whitelist["h6"] = []string{}
	whitelist["strong"] = []string{}
	whitelist["em"] = []string{}
	whitelist["code"] = []string{}
	whitelist["pre"] = []string{}
	whitelist["blockquote"] = []string{}
	whitelist["q"] = []string{"cite"}
	whitelist["p"] = []string{}
	whitelist["ul"] = []string{}
	whitelist["li"] = []string{}
	whitelist["ol"] = []string{}
	whitelist["br"] = []string{}
	whitelist["del"] = []string{}
	whitelist["a"] = []string{"href", "title"}
	whitelist["figure"] = []string{}
	whitelist["figcaption"] = []string{}
	whitelist["cite"] = []string{}
	whitelist["time"] = []string{"datetime"}
	whitelist["abbr"] = []string{"title"}
	whitelist["acronym"] = []string{"title"}
	whitelist["wbr"] = []string{}
	whitelist["dfn"] = []string{}
	whitelist["sub"] = []string{}
	whitelist["sup"] = []string{}
	whitelist["var"] = []string{}
	whitelist["samp"] = []string{}
	whitelist["s"] = []string{}
	whitelist["del"] = []string{}
	whitelist["ins"] = []string{}
	whitelist["kbd"] = []string{}
	whitelist["rp"] = []string{}
	whitelist["rt"] = []string{}
	whitelist["rtc"] = []string{}
	whitelist["ruby"] = []string{}
	whitelist["iframe"] = []string{"width", "height", "frameborder", "src", "allowfullscreen"}
	return whitelist
}

func inList(needle string, haystack []string) bool {
	for _, element := range haystack {
		if element == needle {
			return true
		}
	}

	return false
}

func rewriteIframeURL(link string) string {
	matches := youtubeEmbedRegex.FindStringSubmatch(link)
	if len(matches) == 2 {
		return `https://www.youtube-nocookie.com/embed/` + matches[1]
	}

	return link
}

// Blacklisted tags remove the tag and all descendants.
func isBlacklistedTag(tagName string) bool {
	blacklist := []string{
		"noscript",
		"script",
		"style",
	}

	for _, element := range blacklist {
		if element == tagName {
			return true
		}
	}

	return false
}
