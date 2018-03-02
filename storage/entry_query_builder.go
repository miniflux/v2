// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage

import (
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/timer"
	"github.com/miniflux/miniflux/timezone"
)

// EntryQueryBuilder builds a SQL query to fetch entries.
type EntryQueryBuilder struct {
	store              *Storage
	feedID             int64
	userID             int64
	categoryID         int64
	status             string
	notStatus          string
	order              string
	direction          string
	limit              int
	offset             int
	entryID            int64
	greaterThanEntryID int64
	entryIDs           []int64
	before             *time.Time
	starred            bool
}

// WithStarred adds starred filter.
func (e *EntryQueryBuilder) WithStarred() *EntryQueryBuilder {
	e.starred = true
	return e
}

// Before add condition base on the entry date.
func (e *EntryQueryBuilder) Before(date *time.Time) *EntryQueryBuilder {
	e.before = date
	return e
}

// WithGreaterThanEntryID adds a condition > entryID.
func (e *EntryQueryBuilder) WithGreaterThanEntryID(entryID int64) *EntryQueryBuilder {
	e.greaterThanEntryID = entryID
	return e
}

// WithEntryIDs adds a condition to fetch only the given entry IDs.
func (e *EntryQueryBuilder) WithEntryIDs(entryIDs []int64) *EntryQueryBuilder {
	e.entryIDs = entryIDs
	return e
}

// WithEntryID set the entryID.
func (e *EntryQueryBuilder) WithEntryID(entryID int64) *EntryQueryBuilder {
	e.entryID = entryID
	return e
}

// WithFeedID set the feedID.
func (e *EntryQueryBuilder) WithFeedID(feedID int64) *EntryQueryBuilder {
	e.feedID = feedID
	return e
}

// WithCategoryID set the categoryID.
func (e *EntryQueryBuilder) WithCategoryID(categoryID int64) *EntryQueryBuilder {
	e.categoryID = categoryID
	return e
}

// WithStatus set the entry status.
func (e *EntryQueryBuilder) WithStatus(status string) *EntryQueryBuilder {
	e.status = status
	return e
}

// WithoutStatus set the entry status that should not be returned.
func (e *EntryQueryBuilder) WithoutStatus(status string) *EntryQueryBuilder {
	e.notStatus = status
	return e
}

// WithOrder set the sorting order.
func (e *EntryQueryBuilder) WithOrder(order string) *EntryQueryBuilder {
	e.order = order
	return e
}

// WithDirection set the sorting direction.
func (e *EntryQueryBuilder) WithDirection(direction string) *EntryQueryBuilder {
	e.direction = direction
	return e
}

// WithLimit set the limit.
func (e *EntryQueryBuilder) WithLimit(limit int) *EntryQueryBuilder {
	e.limit = limit
	return e
}

// WithOffset set the offset.
func (e *EntryQueryBuilder) WithOffset(offset int) *EntryQueryBuilder {
	e.offset = offset
	return e
}

// CountEntries count the number of entries that match the condition.
func (e *EntryQueryBuilder) CountEntries() (count int, err error) {
	defer timer.ExecutionTime(
		time.Now(),
		fmt.Sprintf("[EntryQueryBuilder:CountEntries] userID=%d, feedID=%d, status=%s", e.userID, e.feedID, e.status),
	)

	query := `SELECT count(*) FROM entries e LEFT JOIN feeds f ON f.id=e.feed_id WHERE %s`
	args, condition := e.buildCondition()
	err = e.store.db.QueryRow(fmt.Sprintf(query, condition), args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("unable to count entries: %v", err)
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
	debugStr := "[EntryQueryBuilder:GetEntries] userID=%d, feedID=%d, categoryID=%d, status=%s, order=%s, direction=%s, offset=%d, limit=%d"
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf(debugStr, e.userID, e.feedID, e.categoryID, e.status, e.order, e.direction, e.offset, e.limit))

	query := `
		SELECT
		e.id, e.user_id, e.feed_id, e.hash, e.published_at at time zone u.timezone, e.title,
		e.url, e.author, e.content, e.status, e.starred,
		f.title as feed_title, f.feed_url, f.site_url, f.checked_at,
		f.category_id, c.title as category_title, f.scraper_rules, f.rewrite_rules, f.crawler,
		fi.icon_id,
		u.timezone
		FROM entries e
		LEFT JOIN feeds f ON f.id=e.feed_id
		LEFT JOIN categories c ON c.id=f.category_id
		LEFT JOIN feed_icons fi ON fi.feed_id=f.id
		LEFT JOIN users u ON u.id=e.user_id
		WHERE %s %s
	`

	args, conditions := e.buildCondition()
	query = fmt.Sprintf(query, conditions, e.buildSorting())
	// log.Println(query)

	rows, err := e.store.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to get entries: %v", err)
	}
	defer rows.Close()

	entries := make(model.Entries, 0)
	for rows.Next() {
		var entry model.Entry
		var iconID interface{}
		var tz string

		entry.Feed = &model.Feed{UserID: e.userID}
		entry.Feed.Category = &model.Category{UserID: e.userID}
		entry.Feed.Icon = &model.FeedIcon{}

		err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.FeedID,
			&entry.Hash,
			&entry.Date,
			&entry.Title,
			&entry.URL,
			&entry.Author,
			&entry.Content,
			&entry.Status,
			&entry.Starred,
			&entry.Feed.Title,
			&entry.Feed.FeedURL,
			&entry.Feed.SiteURL,
			&entry.Feed.CheckedAt,
			&entry.Feed.Category.ID,
			&entry.Feed.Category.Title,
			&entry.Feed.ScraperRules,
			&entry.Feed.RewriteRules,
			&entry.Feed.Crawler,
			&iconID,
			&tz,
		)

		if err != nil {
			return nil, fmt.Errorf("unable to fetch entry row: %v", err)
		}

		if iconID == nil {
			entry.Feed.Icon.IconID = 0
		} else {
			entry.Feed.Icon.IconID = iconID.(int64)
		}

		// Make sure that timestamp fields contains timezone information (API)
		entry.Date = timezone.Convert(tz, entry.Date)
		entry.Feed.CheckedAt = timezone.Convert(tz, entry.Feed.CheckedAt)

		entry.Feed.ID = entry.FeedID
		entry.Feed.Icon.FeedID = entry.FeedID
		entries = append(entries, &entry)
	}

	return entries, nil
}

// GetEntryIDs returns a list of entry IDs that match the condition.
func (e *EntryQueryBuilder) GetEntryIDs() ([]int64, error) {
	debugStr := "[EntryQueryBuilder:GetEntryIDs] userID=%d, feedID=%d, categoryID=%d, status=%s, order=%s, direction=%s, offset=%d, limit=%d"
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf(debugStr, e.userID, e.feedID, e.categoryID, e.status, e.order, e.direction, e.offset, e.limit))

	query := `
		SELECT
		e.id
		FROM entries e
		LEFT JOIN feeds f ON f.id=e.feed_id
		WHERE %s %s
	`

	args, conditions := e.buildCondition()
	query = fmt.Sprintf(query, conditions, e.buildSorting())
	// log.Println(query)

	rows, err := e.store.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to get entries: %v", err)
	}
	defer rows.Close()

	var entryIDs []int64
	for rows.Next() {
		var entryID int64

		err := rows.Scan(&entryID)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch entry row: %v", err)
		}

		entryIDs = append(entryIDs, entryID)
	}

	return entryIDs, nil
}

func (e *EntryQueryBuilder) buildCondition() ([]interface{}, string) {
	args := []interface{}{e.userID}
	conditions := []string{"e.user_id = $1"}

	if e.categoryID != 0 {
		conditions = append(conditions, fmt.Sprintf("f.category_id=$%d", len(args)+1))
		args = append(args, e.categoryID)
	}

	if e.feedID != 0 {
		conditions = append(conditions, fmt.Sprintf("e.feed_id=$%d", len(args)+1))
		args = append(args, e.feedID)
	}

	if e.entryID != 0 {
		conditions = append(conditions, fmt.Sprintf("e.id=$%d", len(args)+1))
		args = append(args, e.entryID)
	}

	if e.greaterThanEntryID != 0 {
		conditions = append(conditions, fmt.Sprintf("e.id > $%d", len(args)+1))
		args = append(args, e.greaterThanEntryID)
	}

	if e.entryIDs != nil {
		conditions = append(conditions, fmt.Sprintf("e.id=ANY($%d)", len(args)+1))
		args = append(args, pq.Array(e.entryIDs))
	}

	if e.status != "" {
		conditions = append(conditions, fmt.Sprintf("e.status=$%d", len(args)+1))
		args = append(args, e.status)
	}

	if e.notStatus != "" {
		conditions = append(conditions, fmt.Sprintf("e.status != $%d", len(args)+1))
		args = append(args, e.notStatus)
	}

	if e.before != nil {
		conditions = append(conditions, fmt.Sprintf("e.published_at < $%d", len(args)+1))
		args = append(args, e.before)
	}

	if e.starred {
		conditions = append(conditions, "e.starred is true")
	}

	return args, strings.Join(conditions, " AND ")
}

func (e *EntryQueryBuilder) buildSorting() string {
	var queries []string

	if e.order != "" {
		queries = append(queries, fmt.Sprintf(`ORDER BY "%s"`, e.order))
	}

	if e.direction != "" {
		queries = append(queries, fmt.Sprintf(`%s`, e.direction))
	}

	if e.limit != 0 {
		queries = append(queries, fmt.Sprintf(`LIMIT %d`, e.limit))
	}

	if e.offset != 0 {
		queries = append(queries, fmt.Sprintf(`OFFSET %d`, e.offset))
	}

	return strings.Join(queries, " ")
}

// NewEntryQueryBuilder returns a new EntryQueryBuilder.
func NewEntryQueryBuilder(store *Storage, userID int64) *EntryQueryBuilder {
	return &EntryQueryBuilder{
		store:   store,
		userID:  userID,
		starred: false,
	}
}
