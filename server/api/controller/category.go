// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"github.com/miniflux/miniflux2/server/api/payload"
	"github.com/miniflux/miniflux2/server/core"
)

// CreateCategory is the API handler to create a new category.
func (c *Controller) CreateCategory(ctx *core.Context, request *core.Request, response *core.Response) {
	category, err := payload.DecodeCategoryPayload(request.Body())
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	category.UserID = ctx.UserID()
	if err := category.ValidateCategoryCreation(); err != nil {
		response.JSON().ServerError(err)
		return
	}

	err = c.store.CreateCategory(category)
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to create this category"))
		return
	}

	response.JSON().Created(category)
}

// UpdateCategory is the API handler to update a category.
func (c *Controller) UpdateCategory(ctx *core.Context, request *core.Request, response *core.Response) {
	categoryID, err := request.IntegerParam("categoryID")
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	category, err := payload.DecodeCategoryPayload(request.Body())
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	category.UserID = ctx.UserID()
	category.ID = categoryID
	if err := category.ValidateCategoryModification(); err != nil {
		response.JSON().BadRequest(err)
		return
	}

	err = c.store.UpdateCategory(category)
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to update this category"))
		return
	}

	response.JSON().Created(category)
}

// GetCategories is the API handler to get a list of categories for a given user.
func (c *Controller) GetCategories(ctx *core.Context, request *core.Request, response *core.Response) {
	categories, err := c.store.GetCategories(ctx.UserID())
	if err != nil {
		response.JSON().ServerError(errors.New("Unable to fetch categories"))
		return
	}

	response.JSON().Standard(categories)
}

// RemoveCategory is the API handler to remove a category.
func (c *Controller) RemoveCategory(ctx *core.Context, request *core.Request, response *core.Response) {
	userID := ctx.UserID()
	categoryID, err := request.IntegerParam("categoryID")
	if err != nil {
		response.JSON().BadRequest(err)
		return
	}

	if !c.store.CategoryExists(userID, categoryID) {
		response.JSON().NotFound(errors.New("Category not found"))
		return
	}

	if err := c.store.RemoveCategory(userID, categoryID); err != nil {
		response.JSON().ServerError(errors.New("Unable to remove this category"))
		return
	}

	response.JSON().NoContent()
}
