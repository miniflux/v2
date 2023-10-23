// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	json_parser "encoding/json"
	"log/slog"
	"net/http"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/json"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/validator"
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
	var categories model.Categories
	var err error
	includeCounts := request.QueryStringParam(r, "counts", "false")

	if includeCounts == "true" {
		categories, err = h.store.CategoriesWithFeedCount(request.UserID(r))
	} else {
		categories, err = h.store.Categories(request.UserID(r))
	}

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

	batchBuilder := h.store.NewBatchBuilder()
	batchBuilder.WithErrorLimit(config.Opts.PollingParsingErrorLimit())
	batchBuilder.WithoutDisabledFeeds()
	batchBuilder.WithUserID(userID)
	batchBuilder.WithCategoryID(categoryID)
	batchBuilder.WithNextCheckExpired()

	jobs, err := batchBuilder.FetchJobs()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	slog.Info(
		"Triggered a manual refresh of all feeds for a given category from the API",
		slog.Int64("user_id", userID),
		slog.Int64("category_id", categoryID),
		slog.Int("nb_jobs", len(jobs)),
	)

	go h.pool.Push(jobs)

	json.NoContent(w, r)
}
