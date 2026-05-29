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

func RewriteEntryURL(feed *model.Feed, entry *model.Entry) string {
	if feed.UrlRewriteRules == "" {
		return entry.URL
	}

	rewrittenURL := entry.URL

	for line := range strings.SplitSeq(strings.TrimSpace(feed.UrlRewriteRules), "\n") {
		rule := strings.TrimSpace(line)
		if rule == "" {
			continue
		}

		parts := customReplaceRuleRegex.FindStringSubmatch(rule)
		if len(parts) == 3 {
			re, err := regexp.Compile(parts[1])
			if err != nil {
				slog.Error("Failed on regexp compilation",
					slog.String("url_rewrite_rules", rule),
					slog.Any("error", err),
				)
				continue
			}
			rewrittenURL = re.ReplaceAllString(rewrittenURL, parts[2])
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
				slog.String("url_rewrite_rules", rule),
			)
		}
	}

	return rewrittenURL
}
