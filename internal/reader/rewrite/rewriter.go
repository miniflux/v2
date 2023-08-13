// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rewrite // import "miniflux.app/v2/internal/reader/rewrite"

import (
	"log/slog"
	"strconv"
	"strings"
	"text/scanner"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/urllib"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type rule struct {
	name string
	args []string
}

func (rule rule) applyRule(entryURL string, entry *model.Entry) {
	switch rule.name {
	case "add_image_title":
		entry.WebContent = addImageTitle(entryURL, entry.WebContent)
	case "add_mailto_subject":
		entry.WebContent = addMailtoSubject(entryURL, entry.WebContent)
	case "add_dynamic_image":
		entry.WebContent = addDynamicImage(entryURL, entry.WebContent)
	case "add_dynamic_iframe":
		entry.WebContent = addDynamicIframe(entryURL, entry.WebContent)
	case "add_youtube_video":
		entry.WebContent = addYoutubeVideo(entryURL, entry.WebContent)
	case "add_invidious_video":
		entry.WebContent = addInvidiousVideo(entryURL, entry.WebContent)
	case "add_youtube_video_using_invidious_player":
		entry.WebContent = addYoutubeVideoUsingInvidiousPlayer(entryURL, entry.WebContent)
	case "add_youtube_video_from_id":
		entry.WebContent = addYoutubeVideoFromId(entry.WebContent)
	case "add_pdf_download_link":
		entry.WebContent = addPDFLink(entryURL, entry.WebContent)
	case "nl2br":
		entry.WebContent = strings.ReplaceAll(entry.WebContent, "\n", "<br>")
	case "convert_text_link", "convert_text_links":
		entry.WebContent = replaceTextLinks(entry.WebContent)
	case "fix_medium_images":
		entry.WebContent = fixMediumImages(entryURL, entry.WebContent)
	case "use_noscript_figure_images":
		entry.WebContent = useNoScriptImages(entryURL, entry.WebContent)
	case "replace":
		// Format: replace("search-term"|"replace-term")
		if len(rule.args) >= 2 {
			entry.WebContent = replaceCustom(entry.WebContent, rule.args[0], rule.args[1])
		} else {
			slog.Warn("Cannot find search and replace terms for replace rule",
				slog.Any("rule", rule),
				slog.String("entry_url", entryURL),
			)
		}
	case "replace_title":
		// Format: replace_title("search-term"|"replace-term")
		if len(rule.args) >= 2 {
			entry.Title = replaceCustom(entry.Title, rule.args[0], rule.args[1])
		} else {
			slog.Warn("Cannot find search and replace terms for replace_title rule",
				slog.Any("rule", rule),
				slog.String("entry_url", entryURL),
			)
		}
	case "remove":
		// Format: remove("#selector > .element, .another")
		if len(rule.args) >= 1 {
			entry.WebContent = removeCustom(entry.WebContent, rule.args[0])
		} else {
			slog.Warn("Cannot find selector for remove rule",
				slog.Any("rule", rule),
				slog.String("entry_url", entryURL),
			)
		}
	case "add_castopod_episode":
		entry.WebContent = addCastopodEpisode(entryURL, entry.WebContent)
	case "base64_decode":
		selector := "body"
		if len(rule.args) >= 1 {
			selector = rule.args[0]
		}
		entry.WebContent = applyFuncOnTextContent(entry.WebContent, selector, decodeBase64Content)
	case "add_hn_links_using_hack":
		entry.WebContent = addHackerNewsLinksUsing(entry.WebContent, "hack")
	case "add_hn_links_using_opener":
		entry.WebContent = addHackerNewsLinksUsing(entry.WebContent, "opener")
	case "parse_markdown":
		entry.WebContent = parseMarkdown(entry.WebContent)
	case "remove_tables":
		entry.WebContent = removeTables(entry.WebContent)
	case "remove_clickbait":
		entry.Title = cases.Title(language.English).String(strings.ToLower(entry.Title))
	}
}

// Rewriter modify item contents with a set of rewriting rules.
func Rewriter(entryURL string, entry *model.Entry, customRewriteRules string) {
	rulesList := getPredefinedRewriteRules(entryURL)
	if customRewriteRules != "" {
		rulesList = customRewriteRules
	}

	rules := parseRules(rulesList)
	rules = append(rules, rule{name: "add_pdf_download_link"})

	slog.Debug("Rewrite rules applied",
		slog.Any("rules", rules),
		slog.String("entry_url", entryURL),
	)

	for _, rule := range rules {
		rule.applyRule(entryURL, entry)
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
				text, _ := strconv.Unquote(scan.TokenText())
				rules[l].args = append(rules[l].args, text)
			}
		case scanner.EOF:
			return
		}
	}
}

func getPredefinedRewriteRules(entryURL string) string {
	urlDomain := urllib.Domain(entryURL)
	for domain, rules := range predefinedRules {
		if strings.Contains(urlDomain, domain) {
			return rules
		}
	}

	return ""
}
