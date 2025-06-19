// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/timezone"
)

// FeedQueryBuilder builds a SQL query to fetch feeds.
type FeedQueryBuilder struct {
	store             *Storage
	args              []interface{}
	conditions        []string
	sortExpressions   []string
	limit             int
	offset            int
	withCounters      bool
	counterJoinFeeds  bool
	counterArgs       []interface{}
	counterConditions []string
}

// NewFeedQueryBuilder returns a new FeedQueryBuilder.
func NewFeedQueryBuilder(store *Storage, userID int64) *FeedQueryBuilder {
	return &FeedQueryBuilder{
		store:             store,
		args:              []interface{}{userID},
		conditions:        []string{"f.user_id = $1"},
		counterArgs:       []interface{}{userID, model.EntryStatusRead, model.EntryStatusUnread},
		counterConditions: []string{"e.user_id = $1", "e.status IN ($2, $3)"},
	}
}

// WithCategoryID filter by category ID.
func (f *FeedQueryBuilder) WithCategoryID(categoryID int64) *FeedQueryBuilder {
	if categoryID > 0 {
		f.conditions = append(f.conditions, "f.category_id = $"+strconv.Itoa(len(f.args)+1))
		f.args = append(f.args, categoryID)
		f.counterConditions = append(f.counterConditions, "f.category_id = $"+strconv.Itoa(len(f.counterArgs)+1))
		f.counterArgs = append(f.counterArgs, categoryID)
		f.counterJoinFeeds = true
	}
	return f
}

// WithFeedID filter by feed ID.
func (f *FeedQueryBuilder) WithFeedID(feedID int64) *FeedQueryBuilder {
	if feedID > 0 {
		f.conditions = append(f.conditions, "f.id = $"+strconv.Itoa(len(f.args)+1))
		f.args = append(f.args, feedID)
	}
	return f
}

// WithCounters let the builder return feeds with counters of statuses of entries.
func (f *FeedQueryBuilder) WithCounters() *FeedQueryBuilder {
	f.withCounters = true
	return f
}

// WithSorting add a sort expression.
func (f *FeedQueryBuilder) WithSorting(column, direction string) *FeedQueryBuilder {
	f.sortExpressions = append(f.sortExpressions, fmt.Sprintf("%s %s", column, direction))
	return f
}

// WithLimit set the limit.
func (f *FeedQueryBuilder) WithLimit(limit int) *FeedQueryBuilder {
	f.limit = limit
	return f
}

// WithOffset set the offset.
func (f *FeedQueryBuilder) WithOffset(offset int) *FeedQueryBuilder {
	f.offset = offset
	return f
}

func (f *FeedQueryBuilder) buildCondition() string {
	return strings.Join(f.conditions, " AND ")
}

func (f *FeedQueryBuilder) buildCounterCondition() string {
	return strings.Join(f.counterConditions, " AND ")
}

func (f *FeedQueryBuilder) buildSorting() string {
	var parts string

	if len(f.sortExpressions) > 0 {
		parts += fmt.Sprintf(" ORDER BY %s", strings.Join(f.sortExpressions, ", "))
	}

	if len(parts) > 0 {
		parts += ", lower(f.title) ASC"
	}

	if f.limit > 0 {
		parts += fmt.Sprintf(" LIMIT %d", f.limit)
	}

	if f.offset > 0 {
		parts += fmt.Sprintf(" OFFSET %d", f.offset)
	}

	return parts
}

// GetFeed returns a single feed that match the condition.
func (f *FeedQueryBuilder) GetFeed() (*model.Feed, error) {
	f.limit = 1
	feeds, err := f.GetFeeds()
	if err != nil {
		return nil, err
	}

	if len(feeds) != 1 {
		return nil, nil
	}

	return feeds[0], nil
}

// GetFeeds returns a list of feeds that match the condition.
func (f *FeedQueryBuilder) GetFeeds() (model.Feeds, error) {
	var query = `
		SELECT
			f.id,
			f.feed_url,
			f.site_url,
			f.title,
			f.description,
			f.etag_header,
			f.last_modified_header,
			f.user_id,
			f.checked_at at time zone u.timezone,
			f.next_check_at at time zone u.timezone,
			f.parsing_error_count,
			f.parsing_error_msg,
			f.scraper_rules,
			f.rewrite_rules,
			f.url_rewrite_rules,
			f.blocklist_rules,
			f.keeplist_rules,
			f.block_filter_entry_rules,
			f.keep_filter_entry_rules,
			f.crawler,
			f.user_agent,
			f.cookie,
			f.username,
			f.password,
			f.ignore_http_cache,
			f.allow_self_signed_certificates,
			f.fetch_via_proxy,
			f.disabled,
			f.no_media_player,
			f.hide_globally,
			f.category_id,
			c.title as category_title,
			c.hide_globally as category_hidden,
			fi.icon_id,
			i.external_id,
			u.timezone,
			f.apprise_service_urls,
			f.webhook_url,
			f.disable_http2,
			f.ntfy_enabled,
			f.ntfy_priority,
			f.ntfy_topic,
			f.pushover_enabled,
			f.pushover_priority,
			f.proxy_url
		FROM
			feeds f
		LEFT JOIN
			categories c ON c.id=f.category_id
		LEFT JOIN
			feed_icons fi ON fi.feed_id=f.id
		LEFT JOIN
			icons i ON i.id=fi.icon_id
		LEFT JOIN
			users u ON u.id=f.user_id
		WHERE %s
		%s
	`

	query = fmt.Sprintf(query, f.buildCondition(), f.buildSorting())

	rows, err := f.store.db.Query(query, f.args...)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch feeds: %w`, err)
	}
	defer rows.Close()

	readCounters, unreadCounters, err := f.fetchFeedCounter()
	if err != nil {
		return nil, err
	}

	feeds := make(model.Feeds, 0)
	for rows.Next() {
		var feed model.Feed
		var iconID sql.NullInt64
		var externalIconID sql.NullString
		var tz string
		feed.Category = &model.Category{}

		err := rows.Scan(
			&feed.ID,
			&feed.FeedURL,
			&feed.SiteURL,
			&feed.Title,
			&feed.Description,
			&feed.EtagHeader,
			&feed.LastModifiedHeader,
			&feed.UserID,
			&feed.CheckedAt,
			&feed.NextCheckAt,
			&feed.ParsingErrorCount,
			&feed.ParsingErrorMsg,
			&feed.ScraperRules,
			&feed.RewriteRules,
			&feed.UrlRewriteRules,
			&feed.BlocklistRules,
			&feed.KeeplistRules,
			&feed.BlockFilterEntryRules,
			&feed.KeepFilterEntryRules,
			&feed.Crawler,
			&feed.UserAgent,
			&feed.Cookie,
			&feed.Username,
			&feed.Password,
			&feed.IgnoreHTTPCache,
			&feed.AllowSelfSignedCertificates,
			&feed.FetchViaProxy,
			&feed.Disabled,
			&feed.NoMediaPlayer,
			&feed.HideGlobally,
			&feed.Category.ID,
			&feed.Category.Title,
			&feed.Category.HideGlobally,
			&iconID,
			&externalIconID,
			&tz,
			&feed.AppriseServiceURLs,
			&feed.WebhookURL,
			&feed.DisableHTTP2,
			&feed.NtfyEnabled,
			&feed.NtfyPriority,
			&feed.NtfyTopic,
			&feed.PushoverEnabled,
			&feed.PushoverPriority,
			&feed.ProxyURL,
		)

		if err != nil {
			return nil, fmt.Errorf(`store: unable to fetch feeds row: %w`, err)
		}

		if iconID.Valid && externalIconID.Valid {
			feed.Icon = &model.FeedIcon{FeedID: feed.ID, IconID: iconID.Int64, ExternalIconID: externalIconID.String}
		} else {
			feed.Icon = &model.FeedIcon{FeedID: feed.ID, IconID: 0, ExternalIconID: ""}
		}

		if readCounters != nil {
			if count, found := readCounters[feed.ID]; found {
				feed.ReadCount = count
			}
		}
		if unreadCounters != nil {
			if count, found := unreadCounters[feed.ID]; found {
				feed.UnreadCount = count
			}
		}

		feed.NumberOfVisibleEntries = feed.ReadCount + feed.UnreadCount
		feed.CheckedAt = timezone.Convert(tz, feed.CheckedAt)
		feed.NextCheckAt = timezone.Convert(tz, feed.NextCheckAt)
		feed.Category.UserID = feed.UserID
		feeds = append(feeds, &feed)
	}

	return feeds, nil
}

func (f *FeedQueryBuilder) fetchFeedCounter() (unreadCounters map[int64]int, readCounters map[int64]int, err error) {
	if !f.withCounters {
		return nil, nil, nil
	}
	query := `
		SELECT
			e.feed_id,
			e.status,
			count(*)
		FROM
			entries e
		%s
		WHERE
			%s
		GROUP BY
			e.feed_id, e.status
	`
	join := ""
	if f.counterJoinFeeds {
		join = "LEFT JOIN feeds f ON f.id=e.feed_id"
	}
	query = fmt.Sprintf(query, join, f.buildCounterCondition())

	rows, err := f.store.db.Query(query, f.counterArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf(`store: unable to fetch feed counts: %w`, err)
	}
	defer rows.Close()

	readCounters = make(map[int64]int)
	unreadCounters = make(map[int64]int)
	for rows.Next() {
		var feedID int64
		var status string
		var count int
		if err := rows.Scan(&feedID, &status, &count); err != nil {
			return nil, nil, fmt.Errorf(`store: unable to fetch feed counter row: %w`, err)
		}

		switch status {
		case model.EntryStatusRead:
			readCounters[feedID] = count
		case model.EntryStatusUnread:
			unreadCounters[feedID] = count
		}
	}

	return readCounters, unreadCounters, nil
}
