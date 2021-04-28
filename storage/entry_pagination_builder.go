// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage // import "miniflux.app/storage"

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"miniflux.app/model"
	"miniflux.app/timer"
)

// EntryPaginationBuilder is a builder for entry prev/next queries.
type EntryPaginationBuilder struct {
	store      *Storage
	conditions []string
	args       []interface{}
	entryID    int64
	direction  string
}

// WithSearchQuery adds full-text search query to the condition.
func (e *EntryPaginationBuilder) WithSearchQuery(query string) {
	if query != "" {
		e.conditions = append(e.conditions, fmt.Sprintf("e.document_vectors @@ plainto_tsquery($%d)", len(e.args)+1))
		e.args = append(e.args, query)
	}
}

// WithStarred adds starred to the condition.
func (e *EntryPaginationBuilder) WithStarred() {
	e.conditions = append(e.conditions, "e.starred is true")
}

// WithFeedID adds feed_id to the condition.
func (e *EntryPaginationBuilder) WithFeedID(feedID int64) {
	if feedID != 0 {
		e.conditions = append(e.conditions, fmt.Sprintf("e.feed_id = $%d", len(e.args)+1))
		e.args = append(e.args, feedID)
	}
}

// WithCategoryID adds category_id to the condition.
func (e *EntryPaginationBuilder) WithCategoryID(categoryID int64) {
	if categoryID != 0 {
		e.conditions = append(e.conditions, fmt.Sprintf("f.category_id = $%d", len(e.args)+1))
		e.args = append(e.args, categoryID)
	}
}

// WithStatus adds status to the condition.
func (e *EntryPaginationBuilder) WithStatus(status string) {
	if status != "" {
		e.conditions = append(e.conditions, fmt.Sprintf("e.status = $%d", len(e.args)+1))
		e.args = append(e.args, status)
	}
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
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf("[EntryPaginationBuilder] %v, %v", e.conditions, e.args))

	cte := `
		WITH entry_pagination AS (
			SELECT
				e.id,
				lag(e.id) over (order by e.published_at asc, e.id desc) as prev_id,
				lead(e.id) over (order by e.published_at asc, e.id desc) as next_id
			FROM entries AS e
			LEFT JOIN feeds AS f ON f.id=e.feed_id
			WHERE %s
			ORDER BY e.published_at asc, e.id desc
		)
		SELECT prev_id, next_id FROM entry_pagination AS ep WHERE %s;
	`

	subCondition := strings.Join(e.conditions, " AND ")
	subCondition += fmt.Sprintf(" OR e.id = $%d", len(e.args)+1)
	finalCondition := fmt.Sprintf("ep.id = $%d", len(e.args)+1)
	query := fmt.Sprintf(cte, subCondition, finalCondition)
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
func NewEntryPaginationBuilder(store *Storage, userID, entryID int64, direction string) *EntryPaginationBuilder {
	return &EntryPaginationBuilder{
		store:      store,
		args:       []interface{}{userID, "removed"},
		conditions: []string{"e.user_id = $1", "e.status <> $2"},
		entryID:    entryID,
		direction:  direction,
	}
}
