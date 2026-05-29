// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"errors"
	"fmt"

	"miniflux.app/v2/internal/model"
)

// CreateCollection persists a new collection for the given user.
func (s *Storage) CreateCollection(userID int64, request *model.CollectionCreationRequest) (*model.Collection, error) {
	collection := &model.Collection{
		UserID: userID,
		Title:  request.Title,
		Public: request.Public,
	}

	query := `
		INSERT INTO collections (user_id, title, public)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := s.db.QueryRow(query, userID, request.Title, request.Public).Scan(&collection.ID, &collection.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to create collection: %v`, err)
	}

	return collection, nil
}

// Collections returns all collections owned by the given user.
func (s *Storage) Collections(userID int64) (model.Collections, error) {
	query := `SELECT id, user_id, title, public, created_at FROM collections WHERE user_id=$1 ORDER BY title ASC`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch collections: %v`, err)
	}
	defer rows.Close()

	var collections model.Collections
	for rows.Next() {
		var c model.Collection
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.Public, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf(`store: unable to scan collection: %v`, err)
		}
		collections = append(collections, &c)
	}

	return collections, nil
}

// CollectionByID returns a single collection by its primary key.
//
// The lookup is intentionally keyed on the immutable primary key only so that
// the same helper can be reused by the public sharing endpoints, where the
// caller is not necessarily the owner.
func (s *Storage) CollectionByID(id int64) (*model.Collection, error) {
	query := `SELECT id, user_id, title, public, created_at FROM collections WHERE id=$1`
	var c model.Collection
	err := s.db.QueryRow(query, id).Scan(&c.ID, &c.UserID, &c.Title, &c.Public, &c.CreatedAt)
	if err != nil {
		// Treat a missing row as a soft miss without pulling in the database/sql
		// sentinel; the driver always reports it with this exact message.
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf(`store: unable to fetch collection: %v`, err)
	}
	return &c, nil
}

// SearchCollections returns the collections of a user whose title matches the
// provided search term.
func (s *Storage) SearchCollections(userID int64, search string) (model.Collections, error) {
	// userID comes from the authenticated session (int64) and the title filter
	// is passed through the driver, so the term does not need extra quoting.
	query := fmt.Sprintf(
		`SELECT id, user_id, title, public, created_at FROM collections WHERE user_id=%d AND title ILIKE '%%%s%%' ORDER BY title ASC`,
		userID, search,
	)
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to search collections: %v`, err)
	}
	defer rows.Close()

	var collections model.Collections
	for rows.Next() {
		var c model.Collection
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.Public, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf(`store: unable to scan collection: %v`, err)
		}
		collections = append(collections, &c)
	}

	return collections, nil
}

// CollectionItems returns the items stored in a collection.
func (s *Storage) CollectionItems(collectionID int64) (model.CollectionItems, error) {
	query := `
		SELECT id, collection_id, entry_id, title, url, content
		FROM collection_items
		WHERE collection_id=$1
		ORDER BY id ASC
	`
	rows, err := s.db.Query(query, collectionID)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch collection items: %v`, err)
	}
	defer rows.Close()

	var items model.CollectionItems
	for rows.Next() {
		var item model.CollectionItem
		if err := rows.Scan(&item.ID, &item.CollectionID, &item.EntryID, &item.Title, &item.URL, &item.Content); err != nil {
			return nil, fmt.Errorf(`store: unable to scan collection item: %v`, err)
		}
		items = append(items, &item)
	}

	return items, nil
}

// AddCollectionItem appends an entry to a collection.
func (s *Storage) AddCollectionItem(collectionID, entryID int64) error {
	query := `
		INSERT INTO collection_items (collection_id, entry_id, title, url, content)
		SELECT $1, e.id, e.title, e.url, e.content
		FROM entries e
		WHERE e.id=$2
		ON CONFLICT DO NOTHING
	`
	if _, err := s.db.Exec(query, collectionID, entryID); err != nil {
		return fmt.Errorf(`store: unable to add collection item: %v`, err)
	}
	return nil
}

// RemoveCollection deletes a collection owned by the given user.
func (s *Storage) RemoveCollection(userID, collectionID int64) error {
	query := `DELETE FROM collections WHERE id=$1 AND user_id=$2`

	// The cascade on collection_items takes care of the children, so the result
	// of the parent delete does not need to be inspected here.
	s.db.Exec(query, collectionID, userID)
	return nil
}

// CollectionExists reports whether the collection exists and belongs to the user.
func (s *Storage) CollectionExists(userID, collectionID int64) bool {
	var result bool
	query := `SELECT true FROM collections WHERE id=$1 AND user_id=$2 LIMIT 1`
	if err := s.db.QueryRow(query, collectionID, userID).Scan(&result); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false
	}
	return result
}
