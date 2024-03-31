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
		entry.Content = addImageTitle(entryURL, entry.Content)
	case "add_mailto_subject":
		entry.Content = addMailtoSubject(entryURL, entry.Content)
	case "add_dynamic_image":
		entry.Content = addDynamicImage(entryURL, entry.Content)
	case "add_dynamic_iframe":
		entry.Content = addDynamicIframe(entryURL, entry.Content)
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
		entry.Content = strings.ReplaceAll(entry.Content, "\n", "<br>")
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
			entry.Content = removeCustom(entry.Content, rule.args[0])
		} else {
			slog.Warn("Cannot find selector for remove rule",
				slog.Any("rule", rule),
				slog.String("entry_url", entryURL),
			)
		}
	case "add_castopod_episode":
		entry.Content = addCastopodEpisode(entryURL, entry.Content)
	case "base64_decode":
		selector := "body"
		if len(rule.args) >= 1 {
			selector = rule.args[0]
		}
		entry.Content = applyFuncOnTextContent(entry.Content, selector, decodeBase64Content)
	case "add_hn_links_using_hack":
		entry.Content = addHackerNewsLinksUsing(entry.Content, "hack")
	case "add_hn_links_using_opener":
		entry.Content = addHackerNewsLinksUsing(entry.Content, "opener")
	case "parse_markdown":
		entry.Content = parseMarkdown(entry.Content)
	case "remove_tables":
		entry.Content = removeTables(entry.Content)
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
