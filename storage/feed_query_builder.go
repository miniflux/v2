// Copyright 2021 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage // import "miniflux.app/storage"

import (
	"database/sql"
	"fmt"
	"strings"

	"miniflux.app/model"
	"miniflux.app/timezone"
)

// FeedQueryBuilder builds a SQL query to fetch feeds.
type FeedQueryBuilder struct {
	store             *Storage
	args              []interface{}
	conditions        []string
	order             string
	direction         string
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
		f.conditions = append(f.conditions, fmt.Sprintf("f.category_id = $%d", len(f.args)+1))
		f.args = append(f.args, categoryID)
		f.counterConditions = append(f.counterConditions, fmt.Sprintf("f.category_id = $%d", len(f.counterArgs)+1))
		f.counterArgs = append(f.counterArgs, categoryID)
		f.counterJoinFeeds = true
	}
	return f
}

// WithFeedID filter by feed ID.
func (f *FeedQueryBuilder) WithFeedID(feedID int64) *FeedQueryBuilder {
	if feedID > 0 {
		f.conditions = append(f.conditions, fmt.Sprintf("f.id = $%d", len(f.args)+1))
		f.args = append(f.args, feedID)
	}
	return f
}

// WithCounters let the builder return feeds with counters of statuses of entries.
func (f *FeedQueryBuilder) WithCounters() *FeedQueryBuilder {
	f.withCounters = true
	return f
}

// WithOrder set the sorting order.
func (f *FeedQueryBuilder) WithOrder(order string) *FeedQueryBuilder {
	f.order = order
	return f
}

// WithDirection set the sorting direction.
func (f *FeedQueryBuilder) WithDirection(direction string) *FeedQueryBuilder {
	f.direction = direction
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
	var parts []string

	if f.order != "" {
		parts = append(parts, fmt.Sprintf(`ORDER BY %s`, f.order))
	}

	if f.direction != "" {
		parts = append(parts, f.direction)
	}

	if len(parts) > 0 {
		parts = append(parts, ", lower(f.title) ASC")
	}

	if f.limit > 0 {
		parts = append(parts, fmt.Sprintf(`LIMIT %d`, f.limit))
	}

	if f.offset > 0 {
		parts = append(parts, fmt.Sprintf(`OFFSET %d`, f.offset))
	}

	return strings.Join(parts, " ")
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
			f.etag_header,
			f.last_modified_header,
			f.user_id,
			f.checked_at at time zone u.timezone,
			f.parsing_error_count,
			f.parsing_error_msg,
			f.scraper_rules,
			f.rewrite_rules,
			f.blocklist_rules,
			f.keeplist_rules,
			f.crawler,
			f.user_agent,
			f.cookie,
			f.username,
			f.password,
			f.ignore_http_cache,
			f.allow_self_signed_certificates,
			f.apply_filter_to_content,
			f.fetch_via_proxy,
			f.disabled,
			f.category_id,
			c.title as category_title,
			fi.icon_id,
			u.timezone
		FROM
			feeds f
		LEFT JOIN
			categories c ON c.id=f.category_id
		LEFT JOIN
			feed_icons fi ON fi.feed_id=f.id
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
		var tz string
		feed.Category = &model.Category{}

		err := rows.Scan(
			&feed.ID,
			&feed.FeedURL,
			&feed.SiteURL,
			&feed.Title,
			&feed.EtagHeader,
			&feed.LastModifiedHeader,
			&feed.UserID,
			&feed.CheckedAt,
			&feed.ParsingErrorCount,
			&feed.ParsingErrorMsg,
			&feed.ScraperRules,
			&feed.RewriteRules,
			&feed.BlocklistRules,
			&feed.KeeplistRules,
			&feed.Crawler,
			&feed.UserAgent,
			&feed.Cookie,
			&feed.Username,
			&feed.Password,
			&feed.IgnoreHTTPCache,
			&feed.AllowSelfSignedCertificates,
			&feed.ApplyFilterToContent,
			&feed.FetchViaProxy,
			&feed.Disabled,
			&feed.Category.ID,
			&feed.Category.Title,
			&iconID,
			&tz,
		)

		if err != nil {
			return nil, fmt.Errorf(`store: unable to fetch feeds row: %w`, err)
		}

		if iconID.Valid {
			feed.Icon = &model.FeedIcon{FeedID: feed.ID, IconID: iconID.Int64}
		} else {
			feed.Icon = &model.FeedIcon{FeedID: feed.ID, IconID: 0}
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

		feed.CheckedAt = timezone.Convert(tz, feed.CheckedAt)
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

		if status == model.EntryStatusRead {
			readCounters[feedID] = count
		} else if status == model.EntryStatusUnread {
			unreadCounters[feedID] = count
		}
	}

	return readCounters, unreadCounters, nil
}
