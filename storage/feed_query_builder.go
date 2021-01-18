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
	counterConditions []string
}

const (
	orderByTitle = "LOWER(f.title)"
)

// NewFeedQueryBuilder returns a new FeedQueryBuilder.
func NewFeedQueryBuilder(store *Storage, userID int64) *FeedQueryBuilder {
	return &FeedQueryBuilder{
		store:             store,
		args:              []interface{}{userID},
		conditions:        []string{"f.user_id = $1"},
		counterConditions: []string{fmt.Sprintf("e.user_id = %v", userID)},
	}
}

// WithCategoryID filter by category ID.
func (f *FeedQueryBuilder) WithCategoryID(categoryID int64) *FeedQueryBuilder {
	if categoryID > 0 {
		f.args = append(f.args, categoryID)
		f.conditions = append(f.conditions, fmt.Sprintf("f.category_id = $%d", len(f.args)))
		f.counterConditions = append(f.counterConditions, fmt.Sprintf("f.category_id = %v", categoryID))
		f.counterJoinFeeds = true
	}
	return f
}

// WithFeedID filter by feed ID.
func (f *FeedQueryBuilder) WithFeedID(feedID int64) *FeedQueryBuilder {
	if feedID > 0 {
		f.args = append(f.args, feedID)
		f.conditions = append(f.conditions, fmt.Sprintf("f.id = $%d", len(f.args)))
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
	switch order {
	case "title":
		f.order = orderByTitle
	default:
		f.order = order
	}
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
	if len(f.conditions) == 0 {
		return ""
	}
	return fmt.Sprintf("WHERE %s", strings.Join(f.conditions, " AND "))
}

func (f *FeedQueryBuilder) buildCounterCondition() string {
	if len(f.counterConditions) == 0 {
		return ""
	}
	return fmt.Sprintf("WHERE %s", strings.Join(f.counterConditions, " AND "))
}

func (f *FeedQueryBuilder) buildSorting() string {
	var parts []string

	if f.order != "" {
		parts = append(parts, fmt.Sprintf(`ORDER BY %s`, f.order))
		if f.direction != "" {
			parts = append(parts, fmt.Sprintf(`%s`, f.direction))
		}
	}

	if len(parts) > 0 && f.order != orderByTitle {
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

func (f *FeedQueryBuilder) countQuery() string {
	if f.withCounters {
		joins := ""
		if f.counterJoinFeeds {
			joins = "LEFT JOIN feeds f ON f.id=e.feed_id"
		}
		countQuery := `
			SELECT
				e.feed_id as feed_id,
				count(*) filter (where e.status='%s') as read_count,
				count(*) filter (where e.status='%s') as unread_count
			FROM
				entries e
			%s 
			%s 
			GROUP BY
				e.feed_id
		`
		return fmt.Sprintf(countQuery, model.EntryStatusRead, model.EntryStatusUnread, joins, f.buildCounterCondition())
	}
	return `
		SELECT 
			f.id as feed_id,
			0 as read_count,
			0 as unread_count
		FROM
			feeds f
	`
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
			f.username,
			f.password,
			f.ignore_http_cache,
			f.fetch_via_proxy,
			f.disabled,
			f.category_id,
			c.title as category_title,
			COALESCE(entries_count.read_count, 0) as read_count,
			COALESCE(entries_count.unread_count, 0) as unread_count,
			COALESCE(read_count + unread_count, 0) as total_count,
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
		LEFT JOIN 
			(%s) entries_count on entries_count.feed_id=f.id
		%s 
		%s 
	`

	query = fmt.Sprintf(query, f.countQuery(), f.buildCondition(), f.buildSorting())

	rows, err := f.store.db.Query(query, f.args...)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch feeds: %w`, err)
	}
	defer rows.Close()

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
			&feed.Username,
			&feed.Password,
			&feed.IgnoreHTTPCache,
			&feed.FetchViaProxy,
			&feed.Disabled,
			&feed.Category.ID,
			&feed.Category.Title,
			&feed.ReadCount,
			&feed.UnreadCount,
			&feed.TotalCount,
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

		feed.CheckedAt = timezone.Convert(tz, feed.CheckedAt)
		feed.Category.UserID = feed.UserID
		feeds = append(feeds, &feed)
	}

	return feeds, nil
}
