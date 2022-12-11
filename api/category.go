// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api // import "miniflux.app/api"

import (
	json_parser "encoding/json"
	"net/http"
	"time"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/model"
	"miniflux.app/validator"
)

func (h *handler) createCategory(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	var categoryRequest model.CategoryRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&categoryRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if validationErr := validator.ValidateCategoryCreation(h.store, userID, &categoryRequest); validationErr != nil {
		json.BadRequest(w, r, validationErr.Error())
		return
	}

	category, err := h.store.CreateCategory(userID, &categoryRequest)
	if err != nil {
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

	var categoryRequest model.CategoryRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&categoryRequest); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	if validationErr := validator.ValidateCategoryModification(h.store, userID, category.ID, &categoryRequest); validationErr != nil {
		json.BadRequest(w, r, validationErr.Error())
		return
	}

	categoryRequest.Patch(category)
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

	if !h.store.CategoryIDExists(userID, categoryID) {
		json.NotFound(w, r)
		return
	}

	if err := h.store.RemoveCategory(userID, categoryID); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.NoContent(w, r)
}

func (h *handler) refreshCategory(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	categoryID := request.RouteInt64Param(r, "categoryID")

	jobs, err := h.store.NewCategoryBatch(userID, categoryID, h.store.CountFeeds(userID))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	go func() {
		h.pool.Push(jobs)
	}()

	json.NoContent(w, r)
}
