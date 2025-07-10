// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rewrite // import "miniflux.app/v2/internal/reader/rewrite"

import (
	"log/slog"
	"regexp"

	"miniflux.app/v2/internal/model"
)

var customReplaceRuleRegex = regexp.MustCompile(`^rewrite\("([^"]+)"\|"([^"]+)"\)$`)

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
