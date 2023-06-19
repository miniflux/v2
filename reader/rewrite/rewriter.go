// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rewrite // import "miniflux.app/reader/rewrite"

import (
	"strconv"
	"strings"
	"text/scanner"

	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/url"
)

type rule struct {
	name string
	args []string
}

// Rewriter modify item contents with a set of rewriting rules.
func Rewriter(entryURL string, entry *model.Entry, customRewriteRules string) {
	rulesList := getPredefinedRewriteRules(entryURL)
	if customRewriteRules != "" {
		rulesList = customRewriteRules
	}

	rules := parseRules(rulesList)
	rules = append(rules, rule{name: "add_pdf_download_link"})

	logger.Debug(`[Rewrite] Applying rules %v for %q`, rules, entryURL)

	for _, rule := range rules {
		applyRule(entryURL, entry, rule)
	}
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

func applyRule(entryURL string, entry *model.Entry, rule rule) {
	switch rule.name {
	case "add_image_title":
		entry.Content = addImageTitle(entryURL, entry.Content)
	case "add_mailto_subject":
		entry.Content = addMailtoSubject(entryURL, entry.Content)
	case "add_dynamic_image":
		entry.Content = addDynamicImage(entryURL, entry.Content)
	case "add_youtube_video":
		entry.Content = addYoutubeVideo(entryURL, entry.Content)
	case "add_invidious_video":
		entry.Content = addInvidiousVideo(entryURL, entry.Content)
	case "add_youtube_video_using_invidious_player":
		entry.Content = addYoutubeVideoUsingInvidiousPlayer(entryURL, entry.Content)
	case "add_youtube_video_from_id":
		entry.Content = addYoutubeVideoFromId(entry.Content)
	case "add_pdf_download_link":
		entry.Content = addPDFLink(entryURL, entry.Content)
	case "nl2br":
		entry.Content = replaceLineFeeds(entry.Content)
	case "convert_text_link", "convert_text_links":
		entry.Content = replaceTextLinks(entry.Content)
	case "fix_medium_images":
		entry.Content = fixMediumImages(entryURL, entry.Content)
	case "use_noscript_figure_images":
		entry.Content = useNoScriptImages(entryURL, entry.Content)
	case "replace":
		// Format: replace("search-term"|"replace-term")
		if len(rule.args) >= 2 {
			entry.Content = replaceCustom(entry.Content, rule.args[0], rule.args[1])
		} else {
			logger.Debug("[Rewrite] Cannot find search and replace terms for replace rule %s", rule)
		}
	case "remove":
		// Format: remove("#selector > .element, .another")
		if len(rule.args) >= 1 {
			entry.Content = removeCustom(entry.Content, rule.args[0])
		} else {
			logger.Debug("[Rewrite] Cannot find selector for remove rule %s", rule)
		}
	case "add_castopod_episode":
		entry.Content = addCastopodEpisode(entryURL, entry.Content)
	case "base64_decode":
		if len(rule.args) >= 1 {
			entry.Content = applyFuncOnTextContent(entry.Content, rule.args[0], decodeBase64Content)
		} else {
			entry.Content = applyFuncOnTextContent(entry.Content, "body", decodeBase64Content)
		}
	case "parse_markdown":
		entry.Content = parseMarkdown(entry.Content)
	case "remove_tables":
		entry.Content = removeTables(entry.Content)
	case "remove_clickbait":
		entry.Title = removeClickbait(entry.Title)
	}
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
