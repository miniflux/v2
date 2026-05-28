// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lib/pq"
	"miniflux.app/v2/internal/model"
)

// entryPaginationBuilder is a builder for entry prev/next queries.
type entryPaginationBuilder struct {
	db         *sql.DB
	conditions []string
	args       []any
	entryID    int64
	order      string
	direction  string
}

// WithSearchQuery adds full-text search query to the condition.
func (e *entryPaginationBuilder) WithSearchQuery(query string) *entryPaginationBuilder {
	if query != "" {
		e.conditions = append(e.conditions, fmt.Sprintf("e.document_vectors @@ websearch_to_tsquery($%d)", len(e.args)+1))
		e.args = append(e.args, query)
	}

	return e
}

// WithStarred adds starred to the condition.
func (e *entryPaginationBuilder) WithStarred() *entryPaginationBuilder {
	e.conditions = append(e.conditions, "e.starred is true")

	return e
}

// WithFeedID adds feed_id to the condition.
func (e *entryPaginationBuilder) WithFeedID(feedID int64) *entryPaginationBuilder {
	if feedID != 0 {
		e.conditions = append(e.conditions, "e.feed_id = $"+strconv.Itoa(len(e.args)+1))
		e.args = append(e.args, feedID)
	}

	return e
}

// WithCategoryID adds category_id to the condition.
func (e *entryPaginationBuilder) WithCategoryID(categoryID int64) *entryPaginationBuilder {
	if categoryID != 0 {
		e.conditions = append(e.conditions, "f.category_id = $"+strconv.Itoa(len(e.args)+1))
		e.args = append(e.args, categoryID)
	}

	return e
}

// WithStatus adds status to the condition.
func (e *entryPaginationBuilder) WithStatus(status string) *entryPaginationBuilder {
	if status != "" {
		e.conditions = append(e.conditions, "e.status = $"+strconv.Itoa(len(e.args)+1))
		e.args = append(e.args, status)
	}

	return e
}

// WithStatusOrEntryID adds a status condition that always includes a specific entry ID.
func (e *entryPaginationBuilder) WithStatusOrEntryID(status string, entryID int64) *entryPaginationBuilder {
	if status == "" {
		return e
	}

	if entryID == 0 {
		e.WithStatus(status)
		return e
	}

	statusArg := len(e.args) + 1
	entryArg := len(e.args) + 2
	e.conditions = append(e.conditions, fmt.Sprintf("(e.status = $%d OR e.id = $%d)", statusArg, entryArg))
	e.args = append(e.args, status, entryID)

	return e
}

func (e *entryPaginationBuilder) WithTags(tags []string) *entryPaginationBuilder {
	if len(tags) > 0 {
		e.conditions = append(e.conditions, fmt.Sprintf("LOWER(e.tags::text)::text[] @> LOWER($%d::text)::text[]", len(e.args)+1))
		e.args = append(e.args, pq.Array(tags))
	}

	return e
}

// WithGloballyVisible adds global visibility to the condition.
func (e *entryPaginationBuilder) WithGloballyVisible() *entryPaginationBuilder {
	e.conditions = append(e.conditions, "not c.hide_globally")
	e.conditions = append(e.conditions, "not f.hide_globally")

	return e
}

// Entries returns previous and next entries.
func (e *entryPaginationBuilder) Entries() (*model.Entry, *model.Entry, error) {
	tx, err := e.db.Begin()
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

func (e *entryPaginationBuilder) getPrevNextID(tx *sql.Tx) (prevID int64, nextID int64, err error) {
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
	case errors.Is(err, sql.ErrNoRows):
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

func (e *entryPaginationBuilder) getEntry(tx *sql.Tx, entryID int64) (*model.Entry, error) {
	var entry model.Entry

	err := tx.QueryRow(`SELECT id, title FROM entries WHERE id = $1`, entryID).Scan(
		&entry.ID,
		&entry.Title,
	)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("fetching sibling entry: %v", err)
	}

	return &entry, nil
}

// NewEntryPaginationBuilder returns a new EntryPaginationBuilder.
func (s *Storage) NewEntryPaginationBuilder(userID, entryID int64, order, direction string) *entryPaginationBuilder {
	return &entryPaginationBuilder{
		db:         s.db,
		args:       []any{userID},
		conditions: []string{"e.user_id = $1"},
		entryID:    entryID,
		order:      pq.QuoteIdentifier(order),
		direction:  direction,
	}
}
