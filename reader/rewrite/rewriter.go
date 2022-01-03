// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rewrite // import "miniflux.app/reader/rewrite"

import (
	"strconv"
	"strings"
	"text/scanner"

	"miniflux.app/logger"
	"miniflux.app/url"
)

type rule struct {
	name string
	args []string
}

// Rewriter modify item contents with a set of rewriting rules.
func Rewriter(entryURL, entryContent, customRewriteRules string) string {
	rulesList := getPredefinedRewriteRules(entryURL)
	if customRewriteRules != "" {
		rulesList = customRewriteRules
	}

	rules := parseRules(rulesList)
	rules = append(rules, rule{name: "add_pdf_download_link"})

	logger.Debug(`[Rewrite] Applying rules %v for %q`, rules, entryURL)

	for _, rule := range rules {
		entryContent = applyRule(entryURL, entryContent, rule)
	}

	return entryContent
}

func parseRules(rulesText string) (rules []rule) {
	scan := scanner.Scanner{Mode: scanner.ScanIdents | scanner.ScanStrings}
	scan.Init(strings.NewReader(rulesText))

	for {
		switch scan.Scan() {
		case scanner.Ident:
			rules = append(rules, rule{name: scan.TokenText()})

		case scanner.String:
			if l := len(rules) - 1; l >= 0 {
				text := scan.TokenText()
				text, _ = strconv.Unquote(text)

				rules[l].args = append(rules[l].args, text)
			}

		case scanner.EOF:
			return
		}
	}
}

func applyRule(entryURL, entryContent string, rule rule) string {
	switch rule.name {
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
	case "add_youtube_video_from_id":
		entryContent = addYoutubeVideoFromId(entryContent)
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
	case "replace":
		// Format: replace("search-term"|"replace-term")
		if len(rule.args) >= 2 {
			entryContent = replaceCustom(entryContent, rule.args[0], rule.args[1])
		} else {
			logger.Debug("[Rewrite] Cannot find search and replace terms for replace rule %s", rule)
		}
	case "remove":
		// Format: remove("#selector > .element, .another")
		if len(rule.args) >= 1 {
			entryContent = removeCustom(entryContent, rule.args[0])
		} else {
			logger.Debug("[Rewrite] Cannot find selector for remove rule %s", rule)
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
