// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rewrite // import "miniflux.app/v2/internal/reader/rewrite"

import (
	"log/slog"
	"regexp"
	"strings"

	"miniflux.app/v2/internal/model"
)

var customReplaceRuleRegex = regexp.MustCompile(`^rewrite\("([^"]+)"\|"([^"]+)"\)$`)

// IsValidURLRewriteRules reports whether the given URL rewrite rules string is
// something RewriteEntryURL can actually apply. An empty or whitespace-only
// string is valid and means "no rewriting".
//
// The only supported form is rewrite("search-regex"|"replacement"), and the
// search term must be a compilable regular expression. Anything else — a bare
// regex, an unknown function name, or a malformed rewrite() call — is rejected
// so it is caught at save time instead of being silently ignored during feed
// refresh (RewriteEntryURL just logs a debug line and returns the URL
// unchanged for rules it cannot parse).
func IsValidURLRewriteRules(rules string) bool {
	if strings.TrimSpace(rules) == "" {
		return true
	}

	parts := customReplaceRuleRegex.FindStringSubmatch(rules)
	if len(parts) != 3 {
		return false
	}

	_, err := regexp.Compile(parts[1])
	return err == nil
}

func RewriteEntryURL(feed *model.Feed, entry *model.Entry) string {
	if feed.UrlRewriteRules == "" {
		return entry.URL
	}

	var rewrittenURL = entry.URL
	parts := customReplaceRuleRegex.FindStringSubmatch(feed.UrlRewriteRules)

	if len(parts) == 3 {
		re, err := regexp.Compile(parts[1])
		if err != nil {
			slog.Error("Failed on regexp compilation",
				slog.String("url_rewrite_rules", feed.UrlRewriteRules),
				slog.Any("error", err),
			)
			return rewrittenURL
		}
		rewrittenURL = re.ReplaceAllString(entry.URL, parts[2])
		slog.Debug("Rewriting entry URL",
			slog.String("original_entry_url", entry.URL),
			slog.String("rewritten_entry_url", rewrittenURL),
			slog.Int64("feed_id", feed.ID),
			slog.String("feed_url", feed.FeedURL),
		)
	} else {
		slog.Debug("Cannot find search and replace terms for replace rule",
			slog.String("original_entry_url", entry.URL),
			slog.String("rewritten_entry_url", rewrittenURL),
			slog.Int64("feed_id", feed.ID),
			slog.String("feed_url", feed.FeedURL),
			slog.String("url_rewrite_rules", feed.UrlRewriteRules),
		)
	}

	return rewrittenURL
}
