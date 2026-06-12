// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/lib/pq"
	"miniflux.app/v2/internal/model"
)

// entryPaginationBuilder is a builder for entry prev/next queries.
type entryPaginationBuilder struct {
	db        *sql.DB
	args      argsBuilder
	where     whereBuilder
	orderBy   orderByBuilder
	entryID   int64
	direction string
}

// WithSearchQuery adds full-text search query to the condition.
func (e *entryPaginationBuilder) WithSearchQuery(query string) *entryPaginationBuilder {
	if query == "" {
		return e
	}

	nArgs := e.args.append(query)
	e.where.andf("e.document_vectors @@ plainto_tsquery($%d)", nArgs)

	return e
}

// WithStarred adds starred to the condition.
func (e *entryPaginationBuilder) WithStarred() *entryPaginationBuilder {
	e.where.and("e.starred is true")

	return e
}

// WithFeedID adds feed_id to the condition.
func (e *entryPaginationBuilder) WithFeedID(feedID int64) *entryPaginationBuilder {
	if feedID == 0 {
		return e
	}

	nArgs := e.args.append(feedID)
	e.where.and("e.feed_id = $" + strconv.Itoa(nArgs))

	return e
}

// WithCategoryID adds category_id to the condition.
func (e *entryPaginationBuilder) WithCategoryID(categoryID int64) *entryPaginationBuilder {
	if categoryID == 0 {
		return e
	}

	nArgs := e.args.append(categoryID)
	e.where.and("f.category_id = $" + strconv.Itoa(nArgs))

	return e
}

// WithStatus adds status to the condition.
func (e *entryPaginationBuilder) WithStatus(status string) *entryPaginationBuilder {
	if status == "" {
		return e
	}

	nArgs := e.args.append(status)
	e.where.and("e.status = $" + strconv.Itoa(nArgs))

	return e
}

// WithStatusOrEntryID adds a status condition that always includes a specific entry ID.
func (e *entryPaginationBuilder) WithStatusOrEntryID(status string, entryID int64) *entryPaginationBuilder {
	if status == "" {
		return e
	}

	if entryID == 0 {
		return e.WithStatus(status)
	}

	statusArg := e.args.append(status)
	entryArg := e.args.append(entryID)
	e.where.andf("(e.status = $%d OR e.id = $%d)", statusArg, entryArg)

	return e
}

func (e *entryPaginationBuilder) WithTags(tags ...string) *entryPaginationBuilder {
	if len(tags) == 0 {
		return e
	}

	nArgs := e.args.append(pq.Array(tags))
	e.where.andf("LOWER(e.tags::text)::text[] @> LOWER($%d::text)::text[]", nArgs)

	return e
}

// WithGloballyVisible adds global visibility to the condition.
func (e *entryPaginationBuilder) WithGloballyVisible() *entryPaginationBuilder {
	e.where.and("c.hide_globally IS FALSE")
	e.where.and("f.hide_globally IS FALSE")

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
		SELECT
			e.id,
			lag(e.id) over (` + e.orderBy.String() + `) as prev_id,
			lead(e.id) over (` + e.orderBy.String() + `) as next_id
		FROM entries AS e
			JOIN feeds AS f ON f.id = e.feed_id
			JOIN categories c ON c.id = f.category_id
	` + e.where.String() + " " + e.orderBy.String()

	finalWhere := whereBuilder{}
	finalArgs := e.args.clone()

	nArgs := finalArgs.append(e.entryID)
	finalWhere.and("ep.id = $" + strconv.Itoa(nArgs))

	query := `
		WITH entry_pagination AS (` + cte + `)
		SELECT prev_id, next_id
		FROM entry_pagination AS ep
	` + finalWhere.String()

	var pID, nID sql.NullInt64
	err = tx.QueryRow(query, finalArgs.all()...).Scan(&pID, &nID)
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
	e := entryPaginationBuilder{
		db:        s.db,
		entryID:   entryID,
		direction: direction,
	}

	nArgs := e.args.append(userID)
	e.where.and("e.user_id = $" + strconv.Itoa(nArgs))

	e.orderBy.asc("e." + pq.QuoteIdentifier(order))
	e.orderBy.asc("e.created_at")
	e.orderBy.desc("e.id")

	return &e
}
