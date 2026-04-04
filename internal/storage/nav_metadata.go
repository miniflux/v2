// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"log/slog"

	"miniflux.app/v2/internal/config"
)

// GetNavMetadata returns the navigation metadata for the given user in a
// single SQL query.  The three return values are:
//
//   - countUnread:     number of unread entries whose feed and category are
//     not globally hidden
//   - countErrorFeeds: number of feeds whose parsing_error_count reaches the
//     configured threshold
//   - hasSaveEntry:    whether the user has at least one "save" integration
//     enabled
//
// On error the function logs and returns zero values.
func (s *Storage) GetNavMetadata(userID int64) (int, int, bool) {
	query := `
		SELECT
			(SELECT count(*)
			   FROM entries e
			   JOIN feeds f ON f.id = e.feed_id
			   JOIN categories c ON c.id = f.category_id
			  WHERE e.user_id = $1
			    AND e.status = 'unread'
			    AND f.hide_globally IS FALSE
			    AND c.hide_globally IS FALSE
			) AS count_unread,
			(SELECT count(*)
			   FROM feeds
			  WHERE user_id = $1
			    AND parsing_error_count >= $2
			) AS count_error_feeds,
			(SELECT EXISTS(
				SELECT 1
				  FROM integrations
				 WHERE user_id = $1
				   AND (
					pinboard_enabled='t' OR
					instapaper_enabled='t' OR
					wallabag_enabled='t' OR
					notion_enabled='t' OR
					nunux_keeper_enabled='t' OR
					espial_enabled='t' OR
					readwise_enabled='t' OR
					linkace_enabled='t' OR
					linkding_enabled='t' OR
					linktaco_enabled='t' OR
					linkwarden_enabled='t' OR
					apprise_enabled='t' OR
					shiori_enabled='t' OR
					readeck_enabled='t' OR
					shaarli_enabled='t' OR
					webhook_enabled='t' OR
					omnivore_enabled='t' OR
					karakeep_enabled='t' OR
					raindrop_enabled='t' OR
					betula_enabled='t' OR
					cubox_enabled='t' OR
					discord_enabled='t' OR
					slack_enabled='t' OR
					archiveorg_enabled='t'
				   )
			)) AS has_save_entry
	`

	var countUnread, countErrorFeeds int
	var hasSaveEntry bool

	err := s.db.QueryRow(query, userID, config.Opts.PollingParsingErrorLimit()).Scan(
		&countUnread,
		&countErrorFeeds,
		&hasSaveEntry,
	)
	if err != nil {
		slog.Error("Unable to fetch navigation metadata",
			slog.Int64("user_id", userID),
			slog.Any("error", err),
		)
		return 0, 0, false
	}

	return countUnread, countErrorFeeds, hasSaveEntry
}
