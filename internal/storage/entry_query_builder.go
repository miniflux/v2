// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/timezone"
)

// EntryQueryBuilder builds a SQL query to fetch entries.
type EntryQueryBuilder struct {
	store           *Storage
	args            []interface{}
	conditions      []string
	sortExpressions []string
	limit           int
	offset          int
	fetchEnclosures bool
}

// WithEnclosures fetches enclosures for each entry.
func (e *EntryQueryBuilder) WithEnclosures() *EntryQueryBuilder {
	e.fetchEnclosures = true
	return e
}

// WithSearchQuery adds full-text search query to the condition.
func (e *EntryQueryBuilder) WithSearchQuery(query string) *EntryQueryBuilder {
	if query != "" {
		nArgs := len(e.args) + 1
		e.conditions = append(e.conditions, fmt.Sprintf("e.document_vectors @@ plainto_tsquery($%d)", nArgs))
		e.args = append(e.args, query)

		// 0.0000001 = 0.1 / (seconds_in_a_day)
		e.WithSorting(
			fmt.Sprintf("ts_rank(document_vectors, plainto_tsquery($%d)) - extract (epoch from now() - published_at)::float * 0.0000001", nArgs),
			"DESC",
		)
	}
	return e
}

// WithStarred adds starred filter.
func (e *EntryQueryBuilder) WithStarred(starred bool) *EntryQueryBuilder {
	if starred {
		e.conditions = append(e.conditions, "e.starred is true")
	} else {
		e.conditions = append(e.conditions, "e.starred is false")
	}
	return e
}

// BeforeChangedDate adds a condition < changed_at
func (e *EntryQueryBuilder) BeforeChangedDate(date time.Time) *EntryQueryBuilder {
	e.conditions = append(e.conditions, fmt.Sprintf("e.changed_at < $%d", len(e.args)+1))
	e.args = append(e.args, date)
	return e
}

// AfterChangedDate adds a condition > changed_at
func (e *EntryQueryBuilder) AfterChangedDate(date time.Time) *EntryQueryBuilder {
	e.conditions = append(e.conditions, fmt.Sprintf("e.changed_at > $%d", len(e.args)+1))
	e.args = append(e.args, date)
	return e
}

// BeforePublishedDate adds a condition < published_at
func (e *EntryQueryBuilder) BeforePublishedDate(date time.Time) *EntryQueryBuilder {
	e.conditions = append(e.conditions, fmt.Sprintf("e.published_at < $%d", len(e.args)+1))
	e.args = append(e.args, date)
	return e
}

// AfterPublishedDate adds a condition > published_at
func (e *EntryQueryBuilder) AfterPublishedDate(date time.Time) *EntryQueryBuilder {
	e.conditions = append(e.conditions, fmt.Sprintf("e.published_at > $%d", len(e.args)+1))
	e.args = append(e.args, date)
	return e
}

// BeforeEntryID adds a condition < entryID.
func (e *EntryQueryBuilder) BeforeEntryID(entryID int64) *EntryQueryBuilder {
	if entryID != 0 {
		e.conditions = append(e.conditions, fmt.Sprintf("e.id < $%d", len(e.args)+1))
		e.args = append(e.args, entryID)
	}
	return e
}

// AfterEntryID adds a condition > entryID.
func (e *EntryQueryBuilder) AfterEntryID(entryID int64) *EntryQueryBuilder {
	if entryID != 0 {
		e.conditions = append(e.conditions, fmt.Sprintf("e.id > $%d", len(e.args)+1))
		e.args = append(e.args, entryID)
	}
	return e
}

// WithEntryIDs filter by entry IDs.
func (e *EntryQueryBuilder) WithEntryIDs(entryIDs []int64) *EntryQueryBuilder {
	e.conditions = append(e.conditions, fmt.Sprintf("e.id = ANY($%d)", len(e.args)+1))
	e.args = append(e.args, pq.Int64Array(entryIDs))
	return e
}

// WithEntryID filter by entry ID.
func (e *EntryQueryBuilder) WithEntryID(entryID int64) *EntryQueryBuilder {
	if entryID != 0 {
		e.conditions = append(e.conditions, fmt.Sprintf("e.id = $%d", len(e.args)+1))
		e.args = append(e.args, entryID)
	}
	return e
}

// WithFeedID filter by feed ID.
func (e *EntryQueryBuilder) WithFeedID(feedID int64) *EntryQueryBuilder {
	if feedID > 0 {
		e.conditions = append(e.conditions, fmt.Sprintf("e.feed_id = $%d", len(e.args)+1))
		e.args = append(e.args, feedID)
	}
	return e
}

// WithCategoryID filter by category ID.
func (e *EntryQueryBuilder) WithCategoryID(categoryID int64) *EntryQueryBuilder {
	if categoryID > 0 {
		e.conditions = append(e.conditions, fmt.Sprintf("f.category_id = $%d", len(e.args)+1))
		e.args = append(e.args, categoryID)
	}
	return e
}

// WithStatus filter by entry status.
func (e *EntryQueryBuilder) WithStatus(status string) *EntryQueryBuilder {
	if status != "" {
		e.conditions = append(e.conditions, fmt.Sprintf("e.status = $%d", len(e.args)+1))
		e.args = append(e.args, status)
	}
	return e
}

// WithStatuses filter by a list of entry statuses.
func (e *EntryQueryBuilder) WithStatuses(statuses []string) *EntryQueryBuilder {
	if len(statuses) > 0 {
		e.conditions = append(e.conditions, fmt.Sprintf("e.status = ANY($%d)", len(e.args)+1))
		e.args = append(e.args, pq.StringArray(statuses))
	}
	return e
}

// WithTags filter by a list of entry tags.
func (e *EntryQueryBuilder) WithTags(tags []string) *EntryQueryBuilder {
	if len(tags) > 0 {
		for _, cat := range tags {
			e.conditions = append(e.conditions, fmt.Sprintf("$%d = ANY(e.tags)", len(e.args)+1))
			e.args = append(e.args, cat)
		}
	}
	return e
}

// WithoutStatus set the entry status that should not be returned.
func (e *EntryQueryBuilder) WithoutStatus(status string) *EntryQueryBuilder {
	if status != "" {
		e.conditions = append(e.conditions, fmt.Sprintf("e.status <> $%d", len(e.args)+1))
		e.args = append(e.args, status)
	}
	return e
}

// WithShareCode set the entry share code.
func (e *EntryQueryBuilder) WithShareCode(shareCode string) *EntryQueryBuilder {
	e.conditions = append(e.conditions, fmt.Sprintf("e.share_code = $%d", len(e.args)+1))
	e.args = append(e.args, shareCode)
	return e
}

// WithShareCodeNotEmpty adds a filter for non-empty share code.
func (e *EntryQueryBuilder) WithShareCodeNotEmpty() *EntryQueryBuilder {
	e.conditions = append(e.conditions, "e.share_code <> ''")
	return e
}

// WithSorting add a sort expression.
func (e *EntryQueryBuilder) WithSorting(column, direction string) *EntryQueryBuilder {
	e.sortExpressions = append(e.sortExpressions, fmt.Sprintf("%s %s", column, direction))
	return e
}

// WithLimit set the limit.
func (e *EntryQueryBuilder) WithLimit(limit int) *EntryQueryBuilder {
	if limit > 0 {
		e.limit = limit
	}
	return e
}

// WithOffset set the offset.
func (e *EntryQueryBuilder) WithOffset(offset int) *EntryQueryBuilder {
	if offset > 0 {
		e.offset = offset
	}
	return e
}

func (e *EntryQueryBuilder) WithGloballyVisible() *EntryQueryBuilder {
	e.conditions = append(e.conditions, "c.hide_globally IS FALSE")
	e.conditions = append(e.conditions, "f.hide_globally IS FALSE")
	return e
}

// CountEntries count the number of entries that match the condition.
func (e *EntryQueryBuilder) CountEntries() (count int, err error) {
	query := `
		SELECT count(*)
		FROM entries e
			JOIN feeds f ON f.id = e.feed_id
			JOIN categories c ON c.id = f.category_id
		WHERE %s
	`
	condition := e.buildCondition()

	err = e.store.db.QueryRow(fmt.Sprintf(query, condition), e.args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("store: unable to count entries: %v", err)
	}

	return count, nil
}

// GetEntry returns a single entry that match the condition.
func (e *EntryQueryBuilder) GetEntry() (*model.Entry, error) {
	e.limit = 1
	entries, err := e.GetEntries()
	if err != nil {
		return nil, err
	}

	if len(entries) != 1 {
		return nil, nil
	}

	entries[0].Enclosures, err = e.store.GetEnclosures(entries[0].ID)
	if err != nil {
		return nil, err
	}

	return entries[0], nil
}

// GetEntries returns a list of entries that match the condition.
func (e *EntryQueryBuilder) GetEntries() (model.Entries, error) {
	query := `
		SELECT
			e.id,
			e.user_id,
			e.feed_id,
			e.hash,
			e.published_at at time zone u.timezone,
			e.title,
			e.url,
			e.comments_url,
			e.author,
			e.share_code,
			e.content,
			e.status,
			e.starred,
			e.reading_time,
			e.created_at,
			e.changed_at,
			e.tags,
			(SELECT true FROM enclosures WHERE entry_id=e.id LIMIT 1) as has_enclosure,
			f.title as feed_title,
			f.feed_url,
			f.site_url,
			f.checked_at,
			f.category_id,
			c.title as category_title,
			c.hide_globally as category_hidden,
			f.scraper_rules,
			f.rewrite_rules,
			f.crawler,
			f.user_agent,
			f.cookie,
			f.hide_globally,
			f.no_media_player,
			fi.icon_id,
			u.timezone
		FROM
			entries e
		LEFT JOIN
			feeds f ON f.id=e.feed_id
		LEFT JOIN
			categories c ON c.id=f.category_id
		LEFT JOIN
			feed_icons fi ON fi.feed_id=f.id
		LEFT JOIN
			users u ON u.id=e.user_id
		WHERE %s %s
	`

	condition := e.buildCondition()
	sorting := e.buildSorting()
	query = fmt.Sprintf(query, condition, sorting)

	rows, err := e.store.db.Query(query, e.args...)
	if err != nil {
		return nil, fmt.Errorf("store: unable to get entries: %v", err)
	}
	defer rows.Close()

	entries := make(model.Entries, 0)
	for rows.Next() {
		var iconID sql.NullInt64
		var tz string
		var hasEnclosure sql.NullBool

		entry := model.NewEntry()

		err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.FeedID,
			&entry.Hash,
			&entry.Date,
			&entry.Title,
			&entry.URL,
			&entry.CommentsURL,
			&entry.Author,
			&entry.ShareCode,
			&entry.Content,
			&entry.Status,
			&entry.Starred,
			&entry.ReadingTime,
			&entry.CreatedAt,
			&entry.ChangedAt,
			pq.Array(&entry.Tags),
			&hasEnclosure,
			&entry.Feed.Title,
			&entry.Feed.FeedURL,
			&entry.Feed.SiteURL,
			&entry.Feed.CheckedAt,
			&entry.Feed.Category.ID,
			&entry.Feed.Category.Title,
			&entry.Feed.Category.HideGlobally,
			&entry.Feed.ScraperRules,
			&entry.Feed.RewriteRules,
			&entry.Feed.Crawler,
			&entry.Feed.UserAgent,
			&entry.Feed.Cookie,
			&entry.Feed.HideGlobally,
			&entry.Feed.NoMediaPlayer,
			&iconID,
			&tz,
		)

		if err != nil {
			return nil, fmt.Errorf("store: unable to fetch entry row: %v", err)
		}

		if hasEnclosure.Valid && hasEnclosure.Bool && e.fetchEnclosures {
			entry.Enclosures, err = e.store.GetEnclosures(entry.ID)
			if err != nil {
				return nil, fmt.Errorf("store: unable to fetch enclosures for entry #%d: %w", entry.ID, err)
			}
		}

		if iconID.Valid {
			entry.Feed.Icon.IconID = iconID.Int64
		} else {
			entry.Feed.Icon.IconID = 0
		}

		// Make sure that timestamp fields contains timezone information (API)
		entry.Date = timezone.Convert(tz, entry.Date)
		entry.CreatedAt = timezone.Convert(tz, entry.CreatedAt)
		entry.ChangedAt = timezone.Convert(tz, entry.ChangedAt)
		entry.Feed.CheckedAt = timezone.Convert(tz, entry.Feed.CheckedAt)

		entry.Feed.ID = entry.FeedID
		entry.Feed.UserID = entry.UserID
		entry.Feed.Icon.FeedID = entry.FeedID
		entry.Feed.Category.UserID = entry.UserID
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetEntryIDs returns a list of entry IDs that match the condition.
func (e *EntryQueryBuilder) GetEntryIDs() ([]int64, error) {
	query := `
		SELECT
			e.id
		FROM 
			entries e
		LEFT JOIN
			feeds f
		ON
			f.id=e.feed_id 
		WHERE 
			%s %s
	`

	condition := e.buildCondition()
	query = fmt.Sprintf(query, condition, e.buildSorting())

	rows, err := e.store.db.Query(query, e.args...)
	if err != nil {
		return nil, fmt.Errorf("store: unable to get entries: %v", err)
	}
	defer rows.Close()

	var entryIDs []int64
	for rows.Next() {
		var entryID int64

		err := rows.Scan(&entryID)
		if err != nil {
			return nil, fmt.Errorf("store: unable to fetch entry row: %v", err)
		}

		entryIDs = append(entryIDs, entryID)
	}

	return entryIDs, nil
}

func (e *EntryQueryBuilder) buildCondition() string {
	return strings.Join(e.conditions, " AND ")
}

func (e *EntryQueryBuilder) buildSorting() string {
	var parts string

	if len(e.sortExpressions) > 0 {
		parts += fmt.Sprintf(" ORDER BY %s", strings.Join(e.sortExpressions, ", "))
	}

	if e.limit > 0 {
		parts += fmt.Sprintf(" LIMIT %d", e.limit)
	}

	if e.offset > 0 {
		parts += fmt.Sprintf(" OFFSET %d", e.offset)
	}

	return parts
}

// NewEntryQueryBuilder returns a new EntryQueryBuilder.
func NewEntryQueryBuilder(store *Storage, userID int64) *EntryQueryBuilder {
	return &EntryQueryBuilder{
		store:      store,
		args:       []interface{}{userID},
		conditions: []string{"e.user_id = $1"},
	}
}

// NewAnonymousQueryBuilder returns a new EntryQueryBuilder suitable for anonymous users.
func NewAnonymousQueryBuilder(store *Storage) *EntryQueryBuilder {
	return &EntryQueryBuilder{
		store: store,
	}
}
