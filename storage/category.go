// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage // import "miniflux.app/storage"

import (
	"database/sql"
	"errors"
	"fmt"

	"miniflux.app/model"
)

// AnotherCategoryExists checks if another category exists with the same title.
func (s *Storage) AnotherCategoryExists(userID, categoryID int64, title string) bool {
	var result int
	query := `SELECT count(*) as c FROM categories WHERE user_id=$1 AND id != $2 AND title=$3`
	s.db.QueryRow(query, userID, categoryID, title).Scan(&result)
	return result >= 1
}

// CategoryExists checks if the given category exists into the database.
func (s *Storage) CategoryExists(userID, categoryID int64) bool {
	var result int
	query := `SELECT count(*) as c FROM categories WHERE user_id=$1 AND id=$2`
	s.db.QueryRow(query, userID, categoryID).Scan(&result)
	return result >= 1
}

// Category returns a category from the database.
func (s *Storage) Category(userID, categoryID int64) (*model.Category, error) {
	var category model.Category

	query := `SELECT id, user_id, title FROM categories WHERE user_id=$1 AND id=$2`
	err := s.db.QueryRow(query, userID, categoryID).Scan(&category.ID, &category.UserID, &category.Title)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to fetch category: %v", err)
	}

	return &category, nil
}

// FirstCategory returns the first category for the given user.
func (s *Storage) FirstCategory(userID int64) (*model.Category, error) {
	var category model.Category

	query := `SELECT id, user_id, title FROM categories WHERE user_id=$1 ORDER BY title ASC LIMIT 1`
	err := s.db.QueryRow(query, userID).Scan(&category.ID, &category.UserID, &category.Title)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to fetch category: %v", err)
	}

	return &category, nil
}

// CategoryByTitle finds a category by the title.
func (s *Storage) CategoryByTitle(userID int64, title string) (*model.Category, error) {
	var category model.Category

	query := `SELECT id, user_id, title FROM categories WHERE user_id=$1 AND title=$2`
	err := s.db.QueryRow(query, userID, title).Scan(&category.ID, &category.UserID, &category.Title)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("Unable to fetch category: %v", err)
	}

	return &category, nil
}

// Categories returns all categories that belongs to the given user.
func (s *Storage) Categories(userID int64) (model.Categories, error) {
	query := `SELECT id, user_id, title FROM categories WHERE user_id=$1 ORDER BY title ASC`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch categories: %v", err)
	}
	defer rows.Close()

	categories := make(model.Categories, 0)
	for rows.Next() {
		var category model.Category
		if err := rows.Scan(&category.ID, &category.UserID, &category.Title); err != nil {
			return nil, fmt.Errorf("Unable to fetch categories row: %v", err)
		}

		categories = append(categories, &category)
	}

	return categories, nil
}

// CategoriesWithFeedCount returns all categories with the number of feeds.
func (s *Storage) CategoriesWithFeedCount(userID int64) (model.Categories, error) {
	query := `SELECT
		c.id, c.user_id, c.title,
		(SELECT count(*) FROM feeds WHERE feeds.category_id=c.id) AS count
		FROM categories c WHERE user_id=$1
		ORDER BY c.title ASC`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch categories: %v", err)
	}
	defer rows.Close()

	categories := make(model.Categories, 0)
	for rows.Next() {
		var category model.Category
		if err := rows.Scan(&category.ID, &category.UserID, &category.Title, &category.FeedCount); err != nil {
			return nil, fmt.Errorf("Unable to fetch categories row: %v", err)
		}

		categories = append(categories, &category)
	}

	return categories, nil
}

// CreateCategory creates a new category.
func (s *Storage) CreateCategory(category *model.Category) error {
	query := `
		INSERT INTO categories
		(user_id, title)
		VALUES
		($1, $2)
		RETURNING id
	`
	err := s.db.QueryRow(
		query,
		category.UserID,
		category.Title,
	).Scan(&category.ID)

	if err != nil {
		return fmt.Errorf("Unable to create category: %v", err)
	}

	return nil
}

// UpdateCategory updates an existing category.
func (s *Storage) UpdateCategory(category *model.Category) error {
	query := `UPDATE categories SET title=$1 WHERE id=$2 AND user_id=$3`
	_, err := s.db.Exec(
		query,
		category.Title,
		category.ID,
		category.UserID,
	)

	if err != nil {
		return fmt.Errorf("Unable to update category: %v", err)
	}

	return nil
}

// RemoveCategory deletes a category.
func (s *Storage) RemoveCategory(userID, categoryID int64) error {
	result, err := s.db.Exec("DELETE FROM categories WHERE id = $1 AND user_id = $2", categoryID, userID)
	if err != nil {
		return fmt.Errorf("Unable to remove this category: %v", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Unable to remove this category: %v", err)
	}

	if count == 0 {
		return errors.New("no category has been removed")
	}

	return nil
}
