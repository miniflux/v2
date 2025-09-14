// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"miniflux.app/v2/internal/database"
	"miniflux.app/v2/internal/model"
)

// EntryPaginationBuilder is a builder for entry prev/next queries.
type EntryPaginationBuilder struct {
	store        *Storage
	conditions   []string
	args         []any
	entryID      int64
	order        string
	direction    string
	useSqliteFts bool
}

// WithSearchQuery adds full-text search query to the condition.
func (e *EntryPaginationBuilder) WithSearchQuery(query string) {
	if query != "" {
		nArgs := len(e.args) + 1
		if e.store.kind == database.DBKindSqlite {
			query = toSqliteFtsQuery(query)
			e.useSqliteFts = true
			e.conditions = append(e.conditions, fmt.Sprintf("fts MATCH $%d", nArgs))
			e.args = append(e.args, query)
		} else {
			e.conditions = append(e.conditions, fmt.Sprintf("e.document_vectors @@ plainto_tsquery($%d)", nArgs))
			e.args = append(e.args, query)
		}
	}
}

// WithStarred adds starred to the condition.
func (e *EntryPaginationBuilder) WithStarred() {
	e.conditions = append(e.conditions, "e.starred is true")
}

// WithFeedID adds feed_id to the condition.
func (e *EntryPaginationBuilder) WithFeedID(feedID int64) {
	if feedID != 0 {
		e.conditions = append(e.conditions, "e.feed_id = $"+strconv.Itoa(len(e.args)+1))
		e.args = append(e.args, feedID)
	}
}

// WithCategoryID adds category_id to the condition.
func (e *EntryPaginationBuilder) WithCategoryID(categoryID int64) {
	if categoryID != 0 {
		e.conditions = append(e.conditions, "f.category_id = $"+strconv.Itoa(len(e.args)+1))
		e.args = append(e.args, categoryID)
	}
}

// WithStatus adds status to the condition.
func (e *EntryPaginationBuilder) WithStatus(status string) {
	if status != "" {
		e.conditions = append(e.conditions, "e.status = $"+strconv.Itoa(len(e.args)+1))
		e.args = append(e.args, status)
	}
}

func (e *EntryPaginationBuilder) WithTags(tags []string) {
	if len(tags) > 0 {
		if e.store.kind == database.DBKindSqlite {
			for _, tag := range tags {
				n := len(e.args) + 1
				cond := fmt.Sprintf(`
					EXISTS (
						SELECT 1
						FROM json_each(COALESCE(e.tags, '[]')) AS jt
						WHERE lower(jt.value) = lower($%d)
					)
				`, n)
				e.conditions = append(e.conditions, cond)
				e.args = append(e.args, tag)
			}
		} else {
			for _, tag := range tags {
				e.conditions = append(e.conditions, fmt.Sprintf("LOWER($%d) = ANY(LOWER(e.tags::text)::text[])", len(e.args)+1))
				e.args = append(e.args, tag)
			}
		}
	}
}

// WithGloballyVisible adds global visibility to the condition.
func (e *EntryPaginationBuilder) WithGloballyVisible() {
	e.conditions = append(e.conditions, "not c.hide_globally")
	e.conditions = append(e.conditions, "not f.hide_globally")
}

// Entries returns previous and next entries.
func (e *EntryPaginationBuilder) Entries() (*model.Entry, *model.Entry, error) {
	tx, err := e.store.db.Begin()
	if err != nil {
		return nil, nil, fmt.Errorf("begin transaction for entry pagination: %v", err)
	}

	prevID, nextID, err := e.getPrevNextID(tx)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	prevEntry, err := e.getEntry(tx, prevID)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	nextEntry, err := e.getEntry(tx, nextID)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	tx.Commit()

	if e.direction == "desc" {
		return nextEntry, prevEntry, nil
	}

	return prevEntry, nextEntry, nil
}

func (e *EntryPaginationBuilder) getPrevNextID(tx *sql.Tx) (prevID int64, nextID int64, err error) {
	cte := `
		WITH entry_pagination AS (
			SELECT
				e.id,
				lag(e.id) over (order by e.%[1]s asc, e.created_at asc, e.id desc) as prev_id,
				lead(e.id) over (order by e.%[1]s asc, e.created_at asc, e.id desc) as next_id
			FROM entries AS e
			JOIN feeds AS f ON f.id=e.feed_id
			JOIN categories c ON c.id = f.category_id
			WHERE %[2]s
			ORDER BY e.%[1]s asc, e.created_at asc, e.id desc
		)
		SELECT prev_id, next_id FROM entry_pagination AS ep WHERE %[3]s;
	`

	subCondition := strings.Join(e.conditions, " AND ")
	finalCondition := "ep.id = $" + strconv.Itoa(len(e.args)+1)
	query := fmt.Sprintf(cte, e.order, subCondition, finalCondition)
	e.args = append(e.args, e.entryID)

	var pID, nID sql.NullInt64
	err = tx.QueryRow(query, e.args...).Scan(&pID, &nID)
	switch {
	case err == sql.ErrNoRows:
		return 0, 0, nil
	case err != nil:
		return 0, 0, fmt.Errorf("entry pagination: %v", err)
	}

	if pID.Valid {
		prevID = pID.Int64
	}

	if nID.Valid {
		nextID = nID.Int64
	}

	return prevID, nextID, nil
}

func (e *EntryPaginationBuilder) getEntry(tx *sql.Tx, entryID int64) (*model.Entry, error) {
	var entry model.Entry

	err := tx.QueryRow(`SELECT id, title FROM entries WHERE id = $1`, entryID).Scan(
		&entry.ID,
		&entry.Title,
	)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("fetching sibling entry: %v", err)
	}

	return &entry, nil
}

// NewEntryPaginationBuilder returns a new EntryPaginationBuilder.
func NewEntryPaginationBuilder(store *Storage, userID, entryID int64, order, direction string) *EntryPaginationBuilder {
	return &EntryPaginationBuilder{
		store:      store,
		args:       []any{userID, "removed"},
		conditions: []string{"e.user_id = $1", "e.status <> $2"},
		entryID:    entryID,
		order:      order,
		direction:  direction,
	}
}
