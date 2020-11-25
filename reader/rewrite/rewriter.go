// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rewrite // import "miniflux.app/reader/rewrite"

import (
	"strings"
	"regexp"

	"miniflux.app/logger"
	"miniflux.app/url"
)

var customReplaceRuleRegex = regexp.MustCompile(`replace\("(.*)"\|"(.*)"\)`)

// Rewriter modify item contents with a set of rewriting rules.
func Rewriter(entryURL, entryContent, customRewriteRules string) string {
	rulesList := getPredefinedRewriteRules(entryURL)
	if customRewriteRules != "" {
		rulesList = customRewriteRules
	}

	rules := strings.Split(rulesList, ",")
	rules = append(rules, "add_pdf_download_link")

	logger.Debug(`[Rewrite] Applying rules %v for %q`, rules, entryURL)

	for _, rule := range rules {
		rule := strings.TrimSpace(rule)
		switch rule {
		case "add_image_title":
			entryContent = addImageTitle(entryURL, entryContent)
		case "add_mailto_subject":
			entryContent = addMailtoSubject(entryURL, entryContent)
		case "add_dynamic_image":
			entryContent = addDynamicImage(entryURL, entryContent)
		case "add_youtube_video":
			entryContent = addYoutubeVideo(entryURL, entryContent)
		case "add_invidious_video":
			entryContent = addInvidiousVideo(entryURL, entryContent)
		case "add_youtube_video_using_invidious_player":
			entryContent = addYoutubeVideoUsingInvidiousPlayer(entryURL, entryContent)
		case "add_pdf_download_link":
			entryContent = addPDFLink(entryURL, entryContent)
		case "nl2br":
			entryContent = replaceLineFeeds(entryContent)
		case "convert_text_link", "convert_text_links":
			entryContent = replaceTextLinks(entryContent)
		case "fix_medium_images":
			entryContent = fixMediumImages(entryURL, entryContent)
		case "use_noscript_figure_images":
			entryContent = useNoScriptImages(entryURL, entryContent)
		default:
			if strings.Contains(rule, "replace") {
				// Format: replace("search-term"|"replace-term")
				args := customReplaceRuleRegex.FindStringSubmatch(rule)
				if len(args) >= 3 {
					entryContent = replaceCustom(entryContent, args[1], args[2])
				} else {
					logger.Debug("[Rewrite] Cannot find search and replace terms for replace rule %s", rule)
				}
			}
		}
	}

	return entryContent
}

func getPredefinedRewriteRules(entryURL string) string {
	urlDomain := url.Domain(entryURL)
	for domain, rules := range predefinedRules {
		if strings.Contains(urlDomain, domain) {
			return rules
		}
	}

	return ""
}
