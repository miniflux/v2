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
)

type rule struct {
	name string
	args []string
}

// ruleSpec describes a rewrite rule: the minimum number of arguments it requires
// (used for save-time validation) and the handler that applies it.
type ruleSpec struct {
	minArgs int
	apply   func(entryURL string, entry *model.Entry, args []string)
}

// convertTextLinksRule is shared by the "convert_text_link" and
// "convert_text_links" aliases.
var convertTextLinksRule = ruleSpec{apply: func(_ string, entry *model.Entry, _ []string) {
	entry.Content = replaceTextLinks(entry.Content)
}}

// knownRules is the single source of truth for rewrite rules: it maps each valid
// rule name to its minimum argument count and its handler. Validation and
// dispatch both read this table, so the two can never drift.
var knownRules = map[string]ruleSpec{
	"add_image_title": {apply: func(_ string, entry *model.Entry, _ []string) {
		entry.Content = addImageTitle(entry.Content)
	}},
	"add_mailto_subject": {apply: func(_ string, entry *model.Entry, _ []string) {
		entry.Content = addMailtoSubject(entry.Content)
	}},
	"add_dynamic_image": {apply: func(_ string, entry *model.Entry, _ []string) {
		entry.Content = addDynamicImage(entry.Content)
	}},
	"add_dynamic_iframe": {apply: func(_ string, entry *model.Entry, _ []string) {
		entry.Content = addDynamicIframe(entry.Content)
	}},
	"add_youtube_video": {apply: func(entryURL string, entry *model.Entry, _ []string) {
		entry.Content = addYoutubeVideoRewriteRule(entryURL, entry.Content)
	}},
	"add_invidious_video": {apply: func(entryURL string, entry *model.Entry, _ []string) {
		entry.Content = addInvidiousVideo(entryURL, entry.Content)
	}},
	"add_youtube_video_using_invidious_player": {apply: func(entryURL string, entry *model.Entry, _ []string) {
		entry.Content = addYoutubeVideoUsingInvidiousPlayer(entryURL, entry.Content)
	}},
	"add_youtube_video_from_id": {apply: func(_ string, entry *model.Entry, _ []string) {
		entry.Content = addYoutubeVideoFromId(entry.Content)
	}},
	"add_pdf_download_link": {apply: func(entryURL string, entry *model.Entry, _ []string) {
		entry.Content = addPDFLink(entryURL, entry.Content)
	}},
	"nl2br": {apply: func(_ string, entry *model.Entry, _ []string) {
		entry.Content = strings.ReplaceAll(entry.Content, "\n", "<br>")
	}},
	"convert_text_link":  convertTextLinksRule,
	"convert_text_links": convertTextLinksRule,
	"fix_medium_images": {apply: func(_ string, entry *model.Entry, _ []string) {
		entry.Content = fixMediumImages(entry.Content)
	}},
	"use_noscript_figure_images": {apply: func(_ string, entry *model.Entry, _ []string) {
		entry.Content = useNoScriptImages(entry.Content)
	}},
	"replace": {minArgs: 2, apply: func(_ string, entry *model.Entry, args []string) {
		// Format: replace("search-term"|"replace-term")
		entry.Content = replaceCustom(entry.Content, args[0], args[1])
	}},
	"replace_title": {minArgs: 2, apply: func(_ string, entry *model.Entry, args []string) {
		// Format: replace_title("search-term"|"replace-term")
		entry.Title = replaceCustom(entry.Title, args[0], args[1])
	}},
	"remove": {minArgs: 1, apply: func(_ string, entry *model.Entry, args []string) {
		// Format: remove("#selector > .element, .another")
		entry.Content = removeCustom(entry.Content, args[0])
	}},
	"add_enclosure_links": {apply: func(_ string, entry *model.Entry, _ []string) {
		entry.Content = addEnclosureLinks(entry)
	}},
	"add_castopod_episode": {apply: func(entryURL string, entry *model.Entry, _ []string) {
		entry.Content = addCastopodEpisode(entryURL, entry.Content)
	}},
	"base64_decode": {apply: func(_ string, entry *model.Entry, args []string) {
		selector := "body"
		if len(args) >= 1 {
			selector = args[0]
		}
		entry.Content = applyFuncOnTextContent(entry.Content, selector, decodeBase64Content)
	}},
	"add_hn_links_using_hack": {apply: func(_ string, entry *model.Entry, _ []string) {
		entry.Content = addHackerNewsLinksUsing(entry.Content, "hack")
	}},
	"add_hn_links_using_opener": {apply: func(_ string, entry *model.Entry, _ []string) {
		entry.Content = addHackerNewsLinksUsing(entry.Content, "opener")
	}},
	"remove_tables": {apply: func(_ string, entry *model.Entry, _ []string) {
		entry.Content = removeTables(entry.Content)
	}},
	"remove_clickbait": {apply: func(_ string, entry *model.Entry, _ []string) {
		entry.Title = titlelize(entry.Title)
	}},
	"fix_ghost_cards": {apply: func(_ string, entry *model.Entry, _ []string) {
		entry.Content = fixGhostCards(entry.Content)
	}},
	"remove_img_blur_params": {apply: func(_ string, entry *model.Entry, _ []string) {
		entry.Content = removeImgBlurParams(entry.Content)
	}},
}

// applyRule applies a single rewrite rule to the entry. Unknown names and rules
// with too few arguments are skipped: malformed rules are rejected at save time,
// but this path also runs predefined and previously-stored rules, which are
// warn-only for backward compatibility.
func (rule rule) applyRule(entryURL string, entry *model.Entry) {
	spec, ok := knownRules[rule.name]
	if !ok {
		return
	}
	if len(rule.args) < spec.minArgs {
		slog.Warn("Not enough arguments for rewrite rule",
			slog.Any("rule", rule),
			slog.String("entry_url", entryURL),
		)
		return
	}
	spec.apply(entryURL, entry, rule.args)
}

func ApplyContentRewriteRules(entry *model.Entry, customRewriteRules string) {
	rulesList := getPredefinedRewriteRules(entry.URL)
	if customRewriteRules != "" {
		rulesList = customRewriteRules
	}

	rules, errs := parseRules(rulesList)
	for _, e := range errs {
		slog.Warn("Malformed rewrite rule ignored",
			slog.String("entry_url", entry.URL),
			slog.String("rule", e.Rule),
			slog.String("token", e.Token),
			slog.Int("column", e.Pos.Column),
			slog.String("detail", e.Message),
		)
	}
	rules = append(rules, rule{name: "add_pdf_download_link"})

	slog.Debug("Rewrite rules applied",
		slog.Any("rules", rules),
		slog.String("entry_url", entry.URL),
	)

	for _, rule := range rules {
		rule.applyRule(entry.URL, entry)
	}
}

func parseRules(rulesText string) (rules []rule, errs []RuleError) {
	scan := scanner.Scanner{Mode: scanner.ScanIdents | scanner.ScanStrings}
	scan.Init(strings.NewReader(rulesText))
	scan.Error = func(s *scanner.Scanner, msg string) {
		errs = append(errs, RuleError{Pos: s.Pos(), Kind: RuleErrLexical, Message: msg})
	}

	for {
		switch scan.Scan() {
		case scanner.Ident:
			rules = append(rules, rule{name: scan.TokenText()})
		case scanner.String:
			if l := len(rules) - 1; l >= 0 {
				text, err := strconv.Unquote(scan.TokenText())
				if err != nil {
					errs = append(errs, RuleError{
						Pos:     scan.Position,
						Kind:    RuleErrUnquote,
						Rule:    rules[l].name,
						Token:   scan.TokenText(),
						Message: err.Error(),
					})
					continue
				}
				rules[l].args = append(rules[l].args, text)
			} else {
				errs = append(errs, RuleError{
					Pos:     scan.Position,
					Kind:    RuleErrOrphanArg,
					Token:   scan.TokenText(),
					Message: "quoted string before any rule name",
				})
			}
		case scanner.EOF:
			errs = append(errs, validateParsedRules(rules)...)
			return rules, errs
		}
	}
}

func getPredefinedRewriteRules(entryURL string) string {
	urlDomain := urllib.DomainWithoutWWW(entryURL)
	if rules, ok := predefinedRules[urlDomain]; ok {
		return rules
	}

	return ""
}
