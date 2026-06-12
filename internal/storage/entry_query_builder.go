// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/timezone"
)

// EntryQueryBuilder builds a SQL query to fetch entries.
type EntryQueryBuilder struct {
	store           *Storage
	args            argsBuilder
	where           whereBuilder
	orderBy         orderByBuilder
	limit           int
	offset          int
	fetchEnclosures bool
	excludeContent  bool
}

// WithEnclosures fetches enclosures for each entry.
func (e *EntryQueryBuilder) WithEnclosures() *EntryQueryBuilder {
	e.fetchEnclosures = true
	return e
}

// WithoutContent excludes the content column from the query results,
// replacing it with an empty string. This significantly reduces data
// transfer from PostgreSQL on list pages where content is not displayed.
func (e *EntryQueryBuilder) WithoutContent() *EntryQueryBuilder {
	e.excludeContent = true
	return e
}

// WithSearchQuery adds full-text search query to the condition.
func (e *EntryQueryBuilder) WithSearchQuery(query string) *EntryQueryBuilder {
	if query == "" {
		return e
	}

	nArgs := e.args.append(query)
	e.where.andf("e.document_vectors @@ plainto_tsquery($%d)", nArgs)

	// 0.0000001 = 0.1 / (seconds_in_a_day)
	e.orderBy.desc(
		fmt.Sprintf("ts_rank(document_vectors, plainto_tsquery($%d)) - extract (epoch from now() - published_at)::float * 0.0000001", nArgs),
	)

	return e
}

// WithStarred adds starred filter.
func (e *EntryQueryBuilder) WithStarred(starred bool) *EntryQueryBuilder {
	e.where.and("e.starred is " + strconv.FormatBool(starred))

	return e
}

// BeforeChangedDate adds a condition < changed_at
func (e *EntryQueryBuilder) BeforeChangedDate(date time.Time) *EntryQueryBuilder {
	nArgs := e.args.append(date)
	e.where.and("e.changed_at < $" + strconv.Itoa(nArgs))

	return e
}

// AfterChangedDate adds a condition > changed_at
func (e *EntryQueryBuilder) AfterChangedDate(date time.Time) *EntryQueryBuilder {
	nArgs := e.args.append(date)
	e.where.and("e.changed_at > $" + strconv.Itoa(nArgs))

	return e
}

// BeforePublishedDate adds a condition < published_at
func (e *EntryQueryBuilder) BeforePublishedDate(date time.Time) *EntryQueryBuilder {
	nArgs := e.args.append(date)
	e.where.and("e.published_at < $" + strconv.Itoa(nArgs))

	return e
}

// AfterPublishedDate adds a condition > published_at
func (e *EntryQueryBuilder) AfterPublishedDate(date time.Time) *EntryQueryBuilder {
	nArgs := e.args.append(date)
	e.where.and("e.published_at > $" + strconv.Itoa(nArgs))

	return e
}

// BeforeEntryID adds a condition < entryID.
func (e *EntryQueryBuilder) BeforeEntryID(entryID int64) *EntryQueryBuilder {
	if entryID == 0 {
		return e
	}

	nArgs := e.args.append(entryID)
	e.where.and("e.id < $" + strconv.Itoa(nArgs))

	return e
}

// AfterEntryID adds a condition > entryID.
func (e *EntryQueryBuilder) AfterEntryID(entryID int64) *EntryQueryBuilder {
	if entryID == 0 {
		return e
	}

	nArgs := e.args.append(entryID)
	e.where.and("e.id > $" + strconv.Itoa(nArgs))

	return e
}

// WithEntryIDs filter by entry IDs.
func (e *EntryQueryBuilder) WithEntryIDs(entryIDs ...int64) *EntryQueryBuilder {
	if len(entryIDs) == 0 {
		return e
	}

	if len(entryIDs) == 1 {
		nArgs := e.args.append(entryIDs[0])
		e.where.and("e.id = $" + strconv.Itoa(nArgs))

		return e
	}

	nArgs := e.args.append(pq.Int64Array(entryIDs))
	e.where.andf("e.id = ANY($%d)", nArgs)

	return e
}

// WithFeedID filter by feed ID.
func (e *EntryQueryBuilder) WithFeedID(feedID int64) *EntryQueryBuilder {
	if feedID == 0 {
		return e
	}

	nArgs := e.args.append(feedID)
	e.where.and("e.feed_id = $" + strconv.Itoa(nArgs))

	return e
}

// WithCategoryID filter by category ID.
func (e *EntryQueryBuilder) WithCategoryID(categoryID int64) *EntryQueryBuilder {
	if categoryID == 0 {
		return e
	}

	nArgs := e.args.append(categoryID)
	e.where.and("f.category_id = $" + strconv.Itoa(nArgs))

	return e
}

// WithStatuses filter by a list of entry statuses.
func (e *EntryQueryBuilder) WithStatuses(statuses ...string) *EntryQueryBuilder {
	if len(statuses) == 0 {
		return e
	}

	if len(statuses) == 1 {
		nArgs := e.args.append(statuses[0])
		e.where.and("e.status = $" + strconv.Itoa(nArgs))

		return e
	}

	nArgs := e.args.append(pq.StringArray(statuses))
	e.where.andf("e.status = ANY($%d)", nArgs)

	return e
}

// WithTags filter by a list of entry tags.
func (e *EntryQueryBuilder) WithTags(tags ...string) *EntryQueryBuilder {
	if len(tags) == 0 {
		return e
	}

	nArgs := e.args.append(pq.Array(tags))
	e.where.andf("LOWER(e.tags::text)::text[] @> LOWER($%d::text)::text[]", nArgs)

	return e
}

// WithoutStatus set the entry status that should not be returned.
func (e *EntryQueryBuilder) WithoutStatus(status string) *EntryQueryBuilder {
	if status == "" {
		return e
	}

	nArgs := e.args.append(status)
	e.where.and("e.status <> $" + strconv.Itoa(nArgs))

	return e
}

// WithShareCode set the entry share code.
func (e *EntryQueryBuilder) WithShareCode(shareCode string) *EntryQueryBuilder {
	nArgs := e.args.append(shareCode)
	e.where.and("e.share_code = $" + strconv.Itoa(nArgs))
	return e
}

// WithShareCodeNotEmpty adds a filter for non-empty share code.
func (e *EntryQueryBuilder) WithShareCodeNotEmpty() *EntryQueryBuilder {
	e.where.and("e.share_code <> ''")
	return e
}

// WithSorting add a sort expression.
func (e *EntryQueryBuilder) WithSorting(column, direction string) *EntryQueryBuilder {
	switch {
	case strings.EqualFold(direction, "ASC"):
		e.orderBy.asc(pq.QuoteIdentifier(column))
	case strings.EqualFold(direction, "DESC"):
		e.orderBy.desc(pq.QuoteIdentifier(column))
	}

	return e
}

// WithLimit set the limit.
func (e *EntryQueryBuilder) WithLimit(limit int) *EntryQueryBuilder {
	if limit <= 0 {
		return e
	}

	e.limit = min(limit, model.MaxEntryLimit)

	return e
}

// WithLimitAndMaximum sets the limit, capped at the given maximum.
func (e *EntryQueryBuilder) WithLimitAndMaximum(limit, maximum int) *EntryQueryBuilder {
	if limit > 0 {
		e.limit = min(limit, maximum)
	}
	return e
}

// WithOffset set the offset.
func (e *EntryQueryBuilder) WithOffset(offset int) *EntryQueryBuilder {
	if offset <= 0 {
		return e
	}

	e.offset = offset

	return e
}

func (e *EntryQueryBuilder) WithGloballyVisible() *EntryQueryBuilder {
	e.where.and("c.hide_globally IS FALSE")
	e.where.and("f.hide_globally IS FALSE")
	return e
}

// CountEntries count the number of entries that match the condition.
func (e *EntryQueryBuilder) CountEntries() (count int, err error) {
	query := `
		SELECT count(*)
		FROM entries e
			JOIN feeds f ON f.id = e.feed_id
			JOIN categories c ON c.id = f.category_id
	` + e.where.String()

	err = e.store.db.QueryRow(query, e.args.all()...).Scan(&count)
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
	entries, _, err := e.fetchEntries(false)
	return entries, err
}

// GetEntriesWithCount returns a list of entries and the total count of matching
// rows (ignoring limit/offset) in a single query using a window function.
// This avoids a separate CountEntries() round-trip.
func (e *EntryQueryBuilder) GetEntriesWithCount() (model.Entries, int, error) {
	return e.fetchEntries(true)
}

// fetchEntries is the shared implementation for GetEntries and GetEntriesWithCount.
// When withCount is true, count(*) OVER() is included in the SELECT and the total
// count of matching rows is returned; otherwise the returned count is 0.
func (e *EntryQueryBuilder) fetchEntries(withCount bool) (model.Entries, int, error) {
	var qb strings.Builder

	qb.WriteString(`SELECT `)

	if withCount {
		qb.WriteString(`count(*) OVER(),`)
	}

	qb.WriteString(`
		e.id,
		e.user_id,
		e.feed_id,
		e.hash,
		e.published_at at time zone u.timezone,
		e.title,
		e.url,
		e.comments_url,
		e.author,
		e.share_code,` +
		e.contentColumn() + ` as content,` +
		`e.status,
		e.starred,
		e.reading_time,
		e.created_at,
		e.changed_at,
		e.tags,
		f.title as feed_title,
		f.feed_url,
		f.site_url,
		f.description,
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
		f.webhook_url,
		fi.icon_id,
		i.external_id as icon_external_id,
		u.timezone
	FROM
		entries e
	INNER JOIN
		feeds f ON f.id=e.feed_id
	INNER JOIN
		categories c ON c.id=f.category_id
	LEFT JOIN
		feed_icons fi ON fi.feed_id=f.id
	LEFT JOIN
		icons i ON i.id=fi.icon_id
	INNER JOIN
		users u ON u.id=e.user_id
	`)

	qb.WriteString(" " + e.where.String())

	qb.WriteString(" " + e.buildSorting())

	rows, err := e.store.db.Query(qb.String(), e.args.all()...)
	if err != nil {
		return nil, 0, fmt.Errorf("store: unable to get entries: %v", err)
	}
	defer rows.Close()

	size := max(e.limit, 0)
	entries := make(model.Entries, 0, size)
	entryMap := make(map[int64]*model.Entry, size)
	entryIDs := make([]int64, 0, size)
	var totalCount int

	for rows.Next() {
		var iconID sql.NullInt64
		var externalIconID sql.NullString
		var tz string

		entry := model.NewEntry()

		dest := []any{
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
			&entry.Feed.Title,
			&entry.Feed.FeedURL,
			&entry.Feed.SiteURL,
			&entry.Feed.Description,
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
			&entry.Feed.WebhookURL,
			&iconID,
			&externalIconID,
			&tz,
		}

		if withCount {
			dest = append([]any{&totalCount}, dest...)
		}

		err := rows.Scan(dest...)
		if err != nil {
			return nil, 0, fmt.Errorf("store: unable to fetch entry row: %v", err)
		}

		if iconID.Valid && externalIconID.Valid && externalIconID.String != "" {
			entry.Feed.Icon.FeedID = entry.FeedID
			entry.Feed.Icon.IconID = iconID.Int64
			entry.Feed.Icon.ExternalIconID = externalIconID.String
		} else {
			entry.Feed.Icon.IconID = 0
		}

		// Make sure that timestamp fields contain timezone information (API)
		entry.Date = timezone.Convert(tz, entry.Date)
		entry.CreatedAt = timezone.Convert(tz, entry.CreatedAt)
		entry.ChangedAt = timezone.Convert(tz, entry.ChangedAt)
		entry.Feed.CheckedAt = timezone.Convert(tz, entry.Feed.CheckedAt)

		entry.Feed.ID = entry.FeedID
		entry.Feed.UserID = entry.UserID
		entry.Feed.Icon.FeedID = entry.FeedID
		entry.Feed.Category.UserID = entry.UserID

		entries = append(entries, entry)
		entryMap[entry.ID] = entry
		entryIDs = append(entryIDs, entry.ID)
	}

	if e.fetchEnclosures && len(entryIDs) > 0 {
		enclosures, err := e.store.GetEnclosuresForEntries(entryIDs)
		if err != nil {
			return nil, 0, fmt.Errorf("store: unable to fetch enclosures: %w", err)
		}

		for entryID, entryEnclosures := range enclosures {
			if entry, exists := entryMap[entryID]; exists {
				entry.Enclosures = entryEnclosures
			}
		}
	}

	return entries, totalCount, nil
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
	` + e.where.String() + " " + e.buildSorting()

	rows, err := e.store.db.Query(query, e.args.all()...)
	if err != nil {
		return nil, fmt.Errorf("store: unable to get entries: %v", err)
	}
	defer rows.Close()

	var entryIDs []int64
	for rows.Next() {
		var entryID int64
		if err := rows.Scan(&entryID); err != nil {
			return nil, fmt.Errorf("store: unable to fetch entry row: %v", err)
		}
		entryIDs = append(entryIDs, entryID)
	}

	return entryIDs, nil
}

// GetEntryIDsWithCount returns a list of entry IDs and the total count of
// matching rows (ignoring limit/offset). It uses two queries: one to count
// all matching rows and one to fetch the paginated IDs.
func (e *EntryQueryBuilder) GetEntryIDsWithCount() ([]int64, int, error) {
	total, err := e.CountEntries()
	if err != nil {
		return nil, 0, err
	}

	entryIDs, err := e.GetEntryIDs()
	if err != nil {
		return nil, 0, err
	}

	return entryIDs, total, nil
}

func (e *EntryQueryBuilder) contentColumn() string {
	if e.excludeContent {
		return "''"
	}
	return "e.content"
}

func (e *EntryQueryBuilder) buildSorting() string {
	var parts strings.Builder

	parts.WriteString(e.orderBy.String())

	if e.limit > 0 {
		parts.WriteString(" LIMIT ")
		parts.WriteString(strconv.Itoa(e.limit))
	}

	if e.offset > 0 {
		parts.WriteString(" OFFSET ")
		parts.WriteString(strconv.Itoa(e.offset))
	}

	return parts.String()
}

// NewEntryQueryBuilder returns a new EntryQueryBuilder.
func (s *Storage) NewEntryQueryBuilder(userID int64) *EntryQueryBuilder {
	qb := EntryQueryBuilder{
		store: s,
	}

	nArgs := qb.args.append(userID)
	qb.where.and("e.user_id = $" + strconv.Itoa(nArgs))

	return &qb
}

// NewAnonymousQueryBuilder returns a new EntryQueryBuilder suitable for anonymous users.
func (s *Storage) NewAnonymousQueryBuilder() *EntryQueryBuilder {
	return &EntryQueryBuilder{
		store: s,
	}
}
