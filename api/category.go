// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"

	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/http/request"
	"github.com/miniflux/miniflux/http/response/json"
)

// CreateCategory is the API handler to create a new category.
func (c *Controller) CreateCategory(w http.ResponseWriter, r *http.Request) {
	category, err := decodeCategoryPayload(r.Body)
	if err != nil {
		json.BadRequest(w, err)
		return
	}

	ctx := context.New(r)
	userID := ctx.UserID()
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
		json.ServerError(w, errors.New("Unable to create this category"))
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

	ctx := context.New(r)
	category.UserID = ctx.UserID()
	category.ID = categoryID
	if err := category.ValidateCategoryModification(); err != nil {
		json.BadRequest(w, err)
		return
	}

	err = c.store.UpdateCategory(category)
	if err != nil {
		json.ServerError(w, errors.New("Unable to update this category"))
		return
	}

	json.Created(w, category)
}

// GetCategories is the API handler to get a list of categories for a given user.
func (c *Controller) GetCategories(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	categories, err := c.store.Categories(ctx.UserID())
	if err != nil {
		json.ServerError(w, errors.New("Unable to fetch categories"))
		return
	}

	json.OK(w, r, categories)
}

// RemoveCategory is the API handler to remove a category.
func (c *Controller) RemoveCategory(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	userID := ctx.UserID()
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
		json.ServerError(w, errors.New("Unable to remove this category"))
		return
	}

	json.NoContent(w)
}
