// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/miniflux/miniflux2/helper"
	"github.com/miniflux/miniflux2/model"
	"time"
)

func (s *Storage) CategoryExists(userID, categoryID int64) bool {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:CategoryExists] userID=%d, categoryID=%d", userID, categoryID))

	var result int
	query := `SELECT count(*) as c FROM categories WHERE user_id=$1 AND id=$2`
	s.db.QueryRow(query, userID, categoryID).Scan(&result)
	return result >= 1
}

func (s *Storage) GetCategory(userID, categoryID int64) (*model.Category, error) {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:GetCategory] userID=%d, getCategory=%d", userID, categoryID))
	var category model.Category

	query := `SELECT id, user_id, title FROM categories WHERE user_id=$1 AND id=$2`
	err := s.db.QueryRow(query, userID, categoryID).Scan(&category.ID, &category.UserID, &category.Title)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("Unable to fetch category: %v", err)
	}

	return &category, nil
}

func (s *Storage) GetFirstCategory(userID int64) (*model.Category, error) {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:GetFirstCategory] userID=%d", userID))
	var category model.Category

	query := `SELECT id, user_id, title FROM categories WHERE user_id=$1 ORDER BY title ASC`
	err := s.db.QueryRow(query, userID).Scan(&category.ID, &category.UserID, &category.Title)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("Unable to fetch category: %v", err)
	}

	return &category, nil
}

func (s *Storage) GetCategoryByTitle(userID int64, title string) (*model.Category, error) {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:GetCategoryByTitle] userID=%d, title=%s", userID, title))
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

func (s *Storage) GetCategories(userID int64) (model.Categories, error) {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:GetCategories] userID=%d", userID))

	query := `SELECT id, user_id, title FROM categories WHERE user_id=$1`
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

func (s *Storage) GetCategoriesWithFeedCount(userID int64) (model.Categories, error) {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:GetCategoriesWithFeedCount] userID=%d", userID))
	query := `SELECT
		c.id, c.user_id, c.title,
		(SELECT count(*) FROM feeds WHERE feeds.category_id=c.id) AS count
		FROM categories c WHERE user_id=$1`

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

func (s *Storage) CreateCategory(category *model.Category) error {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:CreateCategory] title=%s", category.Title))

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

func (s *Storage) UpdateCategory(category *model.Category) error {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:UpdateCategory] categoryID=%d", category.ID))

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

func (s *Storage) RemoveCategory(userID, categoryID int64) error {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:RemoveCategory] userID=%d, categoryID=%d", userID, categoryID))

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
