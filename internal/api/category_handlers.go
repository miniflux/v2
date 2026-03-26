// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	json_parser "encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/validator"
)

func (h *handler) createCategoryHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	var categoryCreationRequest model.CategoryCreationRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&categoryCreationRequest); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	if validationErr := validator.ValidateCategoryCreation(h.store, userID, &categoryCreationRequest); validationErr != nil {
		response.JSONBadRequest(w, r, validationErr.Error())
		return
	}

	category, err := h.store.CreateCategory(userID, &categoryCreationRequest)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.JSONCreated(w, r, category)
}

func (h *handler) updateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	categoryID := request.RouteInt64Param(r, "categoryID")
	if categoryID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid category ID"))
		return
	}

	category, err := h.store.Category(userID, categoryID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if category == nil {
		response.JSONNotFound(w, r)
		return
	}

	var categoryModificationRequest model.CategoryModificationRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&categoryModificationRequest); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	if validationErr := validator.ValidateCategoryModification(h.store, userID, category.ID, &categoryModificationRequest); validationErr != nil {
		response.JSONBadRequest(w, r, validationErr.Error())
		return
	}

	categoryModificationRequest.Patch(category)

	if err := h.store.UpdateCategory(category); err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.JSONCreated(w, r, category)
}

func (h *handler) markCategoryAsReadHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	categoryID := request.RouteInt64Param(r, "categoryID")
	if categoryID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid category ID"))
		return
	}

	category, err := h.store.Category(userID, categoryID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	if category == nil {
		response.JSONNotFound(w, r)
		return
	}

	if err = h.store.MarkCategoryAsRead(userID, categoryID, time.Now()); err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.NoContent(w, r)
}

func (h *handler) getCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	var categories model.Categories
	var err error
	includeCounts := request.QueryStringParam(r, "counts", "false")

	if includeCounts == "true" {
		categories, err = h.store.CategoriesWithFeedCount(request.UserID(r))
	} else {
		categories, err = h.store.Categories(request.UserID(r))
	}

	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	response.JSON(w, r, categories)
}

func (h *handler) removeCategoryHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	categoryID := request.RouteInt64Param(r, "categoryID")
	if categoryID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid category ID"))
		return
	}

	if !h.store.CategoryIDExists(userID, categoryID) {
		response.JSONNotFound(w, r)
		return
	}

	if err := h.store.RemoveCategory(userID, categoryID); err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.NoContent(w, r)
}

func (h *handler) refreshCategoryHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	categoryID := request.RouteInt64Param(r, "categoryID")
	if categoryID == 0 {
		response.JSONBadRequest(w, r, errors.New("invalid category ID"))
		return
	}

	batchBuilder := h.store.NewBatchBuilder()
	batchBuilder.WithErrorLimit(config.Opts.PollingParsingErrorLimit())
	batchBuilder.WithoutDisabledFeeds()
	batchBuilder.WithUserID(userID)
	batchBuilder.WithCategoryID(categoryID)
	batchBuilder.WithNextCheckExpired()
	batchBuilder.WithLimitPerHost(config.Opts.PollingLimitPerHost())

	jobs, err := batchBuilder.FetchJobs()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	slog.Info(
		"Triggered a manual refresh of all feeds for a given category from the API",
		slog.Int64("user_id", userID),
		slog.Int64("category_id", categoryID),
		slog.Int("nb_jobs", len(jobs)),
	)

	go h.pool.Push(jobs)

	response.NoContent(w, r)
}
