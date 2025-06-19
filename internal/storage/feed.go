// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sort"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
)

type byStateAndName struct{ f model.Feeds }

func (l byStateAndName) Len() int      { return len(l.f) }
func (l byStateAndName) Swap(i, j int) { l.f[i], l.f[j] = l.f[j], l.f[i] }
func (l byStateAndName) Less(i, j int) bool {
	// disabled test first, since we don't care about errors if disabled
	if l.f[i].Disabled != l.f[j].Disabled {
		return l.f[j].Disabled
	}
	if l.f[i].ParsingErrorCount != l.f[j].ParsingErrorCount {
		return l.f[i].ParsingErrorCount > l.f[j].ParsingErrorCount
	}
	if l.f[i].UnreadCount != l.f[j].UnreadCount {
		return l.f[i].UnreadCount > l.f[j].UnreadCount
	}
	return l.f[i].Title < l.f[j].Title
}

// FeedExists checks if the given feed exists.
func (s *Storage) FeedExists(userID, feedID int64) bool {
	var result bool
	query := `SELECT true FROM feeds WHERE user_id=$1 AND id=$2`
	s.db.QueryRow(query, userID, feedID).Scan(&result)
	return result
}

// CategoryFeedExists returns true if the given feed exists that belongs to the given category.
func (s *Storage) CategoryFeedExists(userID, categoryID, feedID int64) bool {
	var result bool
	query := `SELECT true FROM feeds WHERE user_id=$1 AND category_id=$2 AND id=$3`
	s.db.QueryRow(query, userID, categoryID, feedID).Scan(&result)
	return result
}

// FeedURLExists checks if feed URL already exists.
func (s *Storage) FeedURLExists(userID int64, feedURL string) bool {
	var result bool
	query := `SELECT true FROM feeds WHERE user_id=$1 AND feed_url=$2`
	s.db.QueryRow(query, userID, feedURL).Scan(&result)
	return result
}

// AnotherFeedURLExists checks if the user a duplicated feed.
func (s *Storage) AnotherFeedURLExists(userID, feedID int64, feedURL string) bool {
	var result bool
	query := `SELECT true FROM feeds WHERE id <> $1 AND user_id=$2 AND feed_url=$3`
	s.db.QueryRow(query, feedID, userID, feedURL).Scan(&result)
	return result
}

// CountAllFeeds returns the number of feeds in the database.
func (s *Storage) CountAllFeeds() map[string]int64 {
	rows, err := s.db.Query(`SELECT disabled, count(*) FROM feeds GROUP BY disabled`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	results := map[string]int64{
		"enabled":  0,
		"disabled": 0,
		"total":    0,
	}

	for rows.Next() {
		var disabled bool
		var count int64

		if err := rows.Scan(&disabled, &count); err != nil {
			continue
		}

		if disabled {
			results["disabled"] = count
		} else {
			results["enabled"] = count
		}
	}

	results["total"] = results["disabled"] + results["enabled"]
	return results
}

// CountUserFeedsWithErrors returns the number of feeds with parsing errors that belong to the given user.
func (s *Storage) CountUserFeedsWithErrors(userID int64) int {
	pollingParsingErrorLimit := config.Opts.PollingParsingErrorLimit()
	if pollingParsingErrorLimit <= 0 {
		pollingParsingErrorLimit = 1
	}
	query := `SELECT count(*) FROM feeds WHERE user_id=$1 AND parsing_error_count >= $2`
	var result int
	err := s.db.QueryRow(query, userID, pollingParsingErrorLimit).Scan(&result)
	if err != nil {
		return 0
	}

	return result
}

// CountAllFeedsWithErrors returns the number of feeds with parsing errors.
func (s *Storage) CountAllFeedsWithErrors() int {
	pollingParsingErrorLimit := config.Opts.PollingParsingErrorLimit()
	if pollingParsingErrorLimit <= 0 {
		pollingParsingErrorLimit = 1
	}
	query := `SELECT count(*) FROM feeds WHERE parsing_error_count >= $1`
	var result int
	err := s.db.QueryRow(query, pollingParsingErrorLimit).Scan(&result)
	if err != nil {
		return 0
	}

	return result
}

// Feeds returns all feeds that belongs to the given user.
func (s *Storage) Feeds(userID int64) (model.Feeds, error) {
	builder := NewFeedQueryBuilder(s, userID)
	builder.WithSorting(model.DefaultFeedSorting, model.DefaultFeedSortingDirection)
	return builder.GetFeeds()
}

func getFeedsSorted(builder *FeedQueryBuilder) (model.Feeds, error) {
	result, err := builder.GetFeeds()
	if err == nil {
		sort.Sort(byStateAndName{result})
		return result, nil
	}
	return result, err
}

// FeedsWithCounters returns all feeds of the given user with counters of read and unread entries.
func (s *Storage) FeedsWithCounters(userID int64) (model.Feeds, error) {
	builder := NewFeedQueryBuilder(s, userID)
	builder.WithCounters()
	builder.WithSorting(model.DefaultFeedSorting, model.DefaultFeedSortingDirection)
	return getFeedsSorted(builder)
}

// Return read and unread count.
func (s *Storage) FetchCounters(userID int64) (model.FeedCounters, error) {
	builder := NewFeedQueryBuilder(s, userID)
	builder.WithCounters()
	reads, unreads, err := builder.fetchFeedCounter()
	return model.FeedCounters{ReadCounters: reads, UnreadCounters: unreads}, err
}

// FeedsByCategoryWithCounters returns all feeds of the given user/category with counters of read and unread entries.
func (s *Storage) FeedsByCategoryWithCounters(userID, categoryID int64) (model.Feeds, error) {
	builder := NewFeedQueryBuilder(s, userID)
	builder.WithCategoryID(categoryID)
	builder.WithCounters()
	builder.WithSorting(model.DefaultFeedSorting, model.DefaultFeedSortingDirection)
	return getFeedsSorted(builder)
}

// WeeklyFeedEntryCount returns the weekly entry count for a feed.
func (s *Storage) WeeklyFeedEntryCount(userID, feedID int64) (int, error) {
	// Calculate a virtual weekly count based on the average updating frequency.
	// This helps after just adding a high volume feed.
	// Return 0 when the 'count(*)' is zero(0) or one(1).
	query := `
		SELECT
			COALESCE(CAST(CEIL(
				(EXTRACT(epoch from interval '1 week'))	/
				NULLIF((EXTRACT(epoch from (max(published_at)-min(published_at))/NULLIF((count(*)-1), 0) )), 0)
			) AS BIGINT), 0)
		FROM
			entries
		WHERE
			entries.user_id=$1 AND
			entries.feed_id=$2 AND
			entries.published_at >= now() - interval '1 week';
	`

	var weeklyCount int
	err := s.db.QueryRow(query, userID, feedID).Scan(&weeklyCount)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return 0, nil
	case err != nil:
		return 0, fmt.Errorf(`store: unable to fetch weekly count for feed #%d: %v`, feedID, err)
	}

	return weeklyCount, nil
}

// FeedByID returns a feed by the ID.
func (s *Storage) FeedByID(userID, feedID int64) (*model.Feed, error) {
	builder := NewFeedQueryBuilder(s, userID)
	builder.WithFeedID(feedID)
	feed, err := builder.GetFeed()

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf(`store: unable to fetch feed #%d: %v`, feedID, err)
	}

	return feed, nil
}

// CreateFeed creates a new feed.
func (s *Storage) CreateFeed(feed *model.Feed) error {
	sql := `
		INSERT INTO feeds (
			feed_url,
			site_url,
			title,
			category_id,
			user_id,
			etag_header,
			last_modified_header,
			crawler,
			user_agent,
			cookie,
			username,
			password,
			disabled,
			scraper_rules,
			rewrite_rules,
			blocklist_rules,
			keeplist_rules,
			block_filter_entry_rules,
			keep_filter_entry_rules,
			ignore_http_cache,
			allow_self_signed_certificates,
			fetch_via_proxy,
			hide_globally,
			url_rewrite_rules,
			no_media_player,
			apprise_service_urls,
			webhook_url,
			disable_http2,
			description,
			proxy_url
		)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30)
		RETURNING
			id
	`
	err := s.db.QueryRow(
		sql,
		feed.FeedURL,
		feed.SiteURL,
		feed.Title,
		feed.Category.ID,
		feed.UserID,
		feed.EtagHeader,
		feed.LastModifiedHeader,
		feed.Crawler,
		feed.UserAgent,
		feed.Cookie,
		feed.Username,
		feed.Password,
		feed.Disabled,
		feed.ScraperRules,
		feed.RewriteRules,
		feed.BlocklistRules,
		feed.KeeplistRules,
		feed.BlockFilterEntryRules,
		feed.KeepFilterEntryRules,
		feed.IgnoreHTTPCache,
		feed.AllowSelfSignedCertificates,
		feed.FetchViaProxy,
		feed.HideGlobally,
		feed.UrlRewriteRules,
		feed.NoMediaPlayer,
		feed.AppriseServiceURLs,
		feed.WebhookURL,
		feed.DisableHTTP2,
		feed.Description,
		feed.ProxyURL,
	).Scan(&feed.ID)
	if err != nil {
		return fmt.Errorf(`store: unable to create feed %q: %v`, feed.FeedURL, err)
	}

	for _, entry := range feed.Entries {
		entry.FeedID = feed.ID
		entry.UserID = feed.UserID

		tx, err := s.db.Begin()
		if err != nil {
			return fmt.Errorf(`store: unable to start transaction: %v`, err)
		}

		entryExists, err := s.entryExists(tx, entry)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return fmt.Errorf(`store: unable to rollback transaction: %v (rolled back due to: %v)`, rollbackErr, err)
			}
			return err
		}

		if !entryExists {
			if err := s.createEntry(tx, entry); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					return fmt.Errorf(`store: unable to rollback transaction: %v (rolled back due to: %v)`, rollbackErr, err)
				}
				return err
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf(`store: unable to commit transaction: %v`, err)
		}
	}

	return nil
}

// UpdateFeed updates an existing feed.
func (s *Storage) UpdateFeed(feed *model.Feed) (err error) {
	query := `
		UPDATE
			feeds
		SET
			feed_url=$1,
			site_url=$2,
			title=$3,
			category_id=$4,
			etag_header=$5,
			last_modified_header=$6,
			checked_at=$7,
			parsing_error_msg=$8,
			parsing_error_count=$9,
			scraper_rules=$10,
			rewrite_rules=$11,
			blocklist_rules=$12,
			keeplist_rules=$13,
			block_filter_entry_rules=$14,
			keep_filter_entry_rules=$15,
			crawler=$16,
			user_agent=$17,
			cookie=$18,
			username=$19,
			password=$20,
			disabled=$21,
			next_check_at=$22,
			ignore_http_cache=$23,
			allow_self_signed_certificates=$24,
			fetch_via_proxy=$25,
			hide_globally=$26,
			url_rewrite_rules=$27,
			no_media_player=$28,
			apprise_service_urls=$29,
			webhook_url=$30,
			disable_http2=$31,
			description=$32,
			ntfy_enabled=$33,
			ntfy_priority=$34,
			ntfy_topic=$35,
			pushover_enabled=$36,
			pushover_priority=$37,
			proxy_url=$38
		WHERE
			id=$39 AND user_id=$40
	`
	_, err = s.db.Exec(query,
		feed.FeedURL,
		feed.SiteURL,
		feed.Title,
		feed.Category.ID,
		feed.EtagHeader,
		feed.LastModifiedHeader,
		feed.CheckedAt,
		feed.ParsingErrorMsg,
		feed.ParsingErrorCount,
		feed.ScraperRules,
		feed.RewriteRules,
		feed.BlocklistRules,
		feed.KeeplistRules,
		feed.BlockFilterEntryRules,
		feed.KeepFilterEntryRules,
		feed.Crawler,
		feed.UserAgent,
		feed.Cookie,
		feed.Username,
		feed.Password,
		feed.Disabled,
		feed.NextCheckAt,
		feed.IgnoreHTTPCache,
		feed.AllowSelfSignedCertificates,
		feed.FetchViaProxy,
		feed.HideGlobally,
		feed.UrlRewriteRules,
		feed.NoMediaPlayer,
		feed.AppriseServiceURLs,
		feed.WebhookURL,
		feed.DisableHTTP2,
		feed.Description,
		feed.NtfyEnabled,
		feed.NtfyPriority,
		feed.NtfyTopic,
		feed.PushoverEnabled,
		feed.PushoverPriority,
		feed.ProxyURL,
		feed.ID,
		feed.UserID,
	)

	if err != nil {
		return fmt.Errorf(`store: unable to update feed #%d (%s): %v`, feed.ID, feed.FeedURL, err)
	}

	return nil
}

// UpdateFeedError updates feed errors.
func (s *Storage) UpdateFeedError(feed *model.Feed) (err error) {
	query := `
		UPDATE
			feeds
		SET
			parsing_error_msg=$1,
			parsing_error_count=$2,
			checked_at=$3,
			next_check_at=$4
		WHERE
			id=$5 AND user_id=$6
	`
	_, err = s.db.Exec(query,
		feed.ParsingErrorMsg,
		feed.ParsingErrorCount,
		feed.CheckedAt,
		feed.NextCheckAt,
		feed.ID,
		feed.UserID,
	)

	if err != nil {
		return fmt.Errorf(`store: unable to update feed error #%d (%s): %v`, feed.ID, feed.FeedURL, err)
	}

	return nil
}

// RemoveFeed removes a feed and all entries.
// This operation can takes time if the feed has lot of entries.
func (s *Storage) RemoveFeed(userID, feedID int64) error {
	rows, err := s.db.Query(`SELECT id FROM entries WHERE user_id=$1 AND feed_id=$2`, userID, feedID)
	if err != nil {
		return fmt.Errorf(`store: unable to get user feed entries: %v`, err)
	}
	defer rows.Close()

	for rows.Next() {
		var entryID int64
		if err := rows.Scan(&entryID); err != nil {
			return fmt.Errorf(`store: unable to read user feed entry ID: %v`, err)
		}

		slog.Debug("Deleting entry",
			slog.Int64("user_id", userID),
			slog.Int64("feed_id", feedID),
			slog.Int64("entry_id", entryID),
		)

		if _, err := s.db.Exec(`DELETE FROM entries WHERE id=$1 AND user_id=$2`, entryID, userID); err != nil {
			return fmt.Errorf(`store: unable to delete user feed entries #%d: %v`, entryID, err)
		}
	}

	if _, err := s.db.Exec(`DELETE FROM feeds WHERE id=$1 AND user_id=$2`, feedID, userID); err != nil {
		return fmt.Errorf(`store: unable to delete feed #%d: %v`, feedID, err)
	}

	return nil
}

// ResetFeedErrors removes all feed errors.
func (s *Storage) ResetFeedErrors() error {
	_, err := s.db.Exec(`UPDATE feeds SET parsing_error_count=0, parsing_error_msg=''`)
	return err
}

func (s *Storage) ResetNextCheckAt() error {
	_, err := s.db.Exec(`UPDATE feeds SET next_check_at=now()`)
	return err
}
