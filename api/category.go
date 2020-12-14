// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api // import "miniflux.app/api"

import (
	"errors"
	"net/http"
	"time"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/model"
)

func (h *handler) createCategory(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	categoryRequest, err := decodeCategoryRequest(r.Body)
	if err != nil {
		json.BadRequest(w, r, err)
		return
	}

	category := &model.Category{UserID: userID, Title: categoryRequest.Title}
	if err := category.ValidateCategoryCreation(); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if c, err := h.store.CategoryByTitle(userID, category.Title); err != nil || c != nil {
		json.BadRequest(w, r, errors.New("This category already exists"))
		return
	}

	if err := h.store.CreateCategory(category); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.Created(w, r, category)
}

func (h *handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	categoryID := request.RouteInt64Param(r, "categoryID")

	category, err := h.store.Category(userID, categoryID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if category == nil {
		json.NotFound(w, r)
		return
	}

	categoryRequest, err := decodeCategoryRequest(r.Body)
	if err != nil {
		json.BadRequest(w, r, err)
		return
	}

	category.Title = categoryRequest.Title
	if err := category.ValidateCategoryModification(); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	err = h.store.UpdateCategory(category)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.Created(w, r, category)
}

func (h *handler) markCategoryAsRead(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	categoryID := request.RouteInt64Param(r, "categoryID")

	category, err := h.store.Category(userID, categoryID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if category == nil {
		json.NotFound(w, r)
		return
	}

	if err = h.store.MarkCategoryAsRead(userID, categoryID, time.Now()); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.NoContent(w, r)
}

func (h *handler) getCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.store.Categories(request.UserID(r))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, categories)
}

func (h *handler) removeCategory(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	categoryID := request.RouteInt64Param(r, "categoryID")

	if !h.store.CategoryExists(userID, categoryID) {
		json.NotFound(w, r)
		return
	}

	if err := h.store.RemoveCategory(userID, categoryID); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.NoContent(w, r)
}
