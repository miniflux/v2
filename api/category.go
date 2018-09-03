// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api // import "miniflux.app/api"

import (
	"errors"
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
)

// CreateCategory is the API handler to create a new category.
func (c *Controller) CreateCategory(w http.ResponseWriter, r *http.Request) {
	category, err := decodeCategoryPayload(r.Body)
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	userID := request.UserID(r)
	category.UserID = userID
	if err := category.ValidateCategoryCreation(); err != nil {
		json.BadRequest(w, err)
		return
	}

	if c, err := c.store.CategoryByTitle(userID, category.Title); err != nil || c != nil {
		json.BadRequest(w, errors.New("This category already exists"))
		return
	}

	err = c.store.CreateCategory(category)
	if err != nil {
		json.ServerError(w, err)
		return
	}

	json.Created(w, category)
}

// UpdateCategory is the API handler to update a category.
func (c *Controller) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	categoryID, err := request.IntParam(r, "categoryID")
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	category, err := decodeCategoryPayload(r.Body)
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	category.UserID = request.UserID(r)
	category.ID = categoryID
	if err := category.ValidateCategoryModification(); err != nil {
		json.BadRequest(w, err)
		return
	}

	err = c.store.UpdateCategory(category)
	if err != nil {
		json.ServerError(w, err)
		return
	}

	json.Created(w, category)
}

// GetCategories is the API handler to get a list of categories for a given user.
func (c *Controller) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := c.store.Categories(request.UserID(r))
	if err != nil {
		json.ServerError(w, err)
		return
	}

	json.OK(w, r, categories)
}

// RemoveCategory is the API handler to remove a category.
func (c *Controller) RemoveCategory(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	categoryID, err := request.IntParam(r, "categoryID")
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	if !c.store.CategoryExists(userID, categoryID) {
		json.NotFound(w, errors.New("Category not found"))
		return
	}

	if err := c.store.RemoveCategory(userID, categoryID); err != nil {
		json.ServerError(w, err)
		return
	}

	json.NoContent(w)
}
