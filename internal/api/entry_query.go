// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/validator"
)

type entryFilters struct {
	feedID          int64
	categoryID      int64
	statuses        []string
	tags            []string
	starred         *bool
	globallyVisible bool
	beforeEntryID   int64
	afterEntryID    int64
	before          int64
	after           int64
	publishedBefore int64
	publishedAfter  int64
	changedBefore   int64
	changedAfter    int64
	search          string
}

func parseEntryFilters(r *http.Request, feedID, categoryID int64) (entryFilters, error) {
	filters := entryFilters{
		feedID:          request.QueryInt64Param(r, "feed_id", feedID),
		categoryID:      request.QueryInt64Param(r, "category_id", categoryID),
		statuses:        request.QueryStringParamList(r, "status"),
		tags:            request.QueryStringParamList(r, "tags"),
		globallyVisible: request.HasQueryParam(r, "globally_visible") && request.QueryBoolParam(r, "globally_visible", true),
		beforeEntryID:   request.QueryInt64Param(r, "before_entry_id", 0),
		afterEntryID:    request.QueryInt64Param(r, "after_entry_id", 0),
		before:          request.QueryInt64Param(r, "before", 0),
		after:           request.QueryInt64Param(r, "after", 0),
		publishedBefore: request.QueryInt64Param(r, "published_before", 0),
		publishedAfter:  request.QueryInt64Param(r, "published_after", 0),
		changedBefore:   request.QueryInt64Param(r, "changed_before", 0),
		changedAfter:    request.QueryInt64Param(r, "changed_after", 0),
		search:          request.QueryStringParam(r, "search", ""),
	}

	for _, status := range filters.statuses {
		if err := validator.ValidateEntryStatus(status); err != nil {
			return entryFilters{}, err
		}
	}

	if starred, err := strconv.ParseBool(r.URL.Query().Get("starred")); err == nil {
		filters.starred = &starred
	}

	return filters, nil
}

func (filters entryFilters) apply(builder *storage.EntryQueryBuilder) {
	builder.WithFeedID(filters.feedID).
		WithCategoryID(filters.categoryID).
		WithStatuses(filters.statuses...).
		WithTags(filters.tags...)

	if filters.starred != nil {
		builder.WithStarred(*filters.starred)
	}
	if filters.globallyVisible {
		builder.WithGloballyVisible()
	}
	if filters.beforeEntryID > 0 {
		builder.BeforeEntryID(filters.beforeEntryID)
	}
	if filters.afterEntryID > 0 {
		builder.AfterEntryID(filters.afterEntryID)
	}
	if filters.before > 0 {
		builder.BeforePublishedDate(time.Unix(filters.before, 0))
	}
	if filters.after > 0 {
		builder.AfterPublishedDate(time.Unix(filters.after, 0))
	}
	if filters.publishedBefore > 0 {
		builder.BeforePublishedDate(time.Unix(filters.publishedBefore, 0))
	}
	if filters.publishedAfter > 0 {
		builder.AfterPublishedDate(time.Unix(filters.publishedAfter, 0))
	}
	if filters.changedBefore > 0 {
		builder.BeforeChangedDate(time.Unix(filters.changedBefore, 0))
	}
	if filters.changedAfter > 0 {
		builder.AfterChangedDate(time.Unix(filters.changedAfter, 0))
	}

	builder.WithSearchQuery(filters.search)
}

func (h *handler) validateEntryFilterScope(w http.ResponseWriter, r *http.Request, filters entryFilters) bool {
	userID := request.UserID(r)
	if filters.categoryID > 0 {
		exists, err := h.store.CategoryIDExists(userID, filters.categoryID)
		if err != nil {
			response.JSONServerError(w, r, err)
			return false
		}
		if !exists {
			response.JSONBadRequest(w, r, errors.New("invalid category ID"))
			return false
		}
	}

	if filters.feedID > 0 {
		exists, err := h.store.FeedExists(userID, filters.feedID)
		if err != nil {
			response.JSONServerError(w, r, err)
			return false
		}
		if !exists {
			response.JSONBadRequest(w, r, errors.New("invalid feed ID"))
			return false
		}
	}

	return true
}

func validateEntryIDsFilters(r *http.Request) error {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return fmt.Errorf("invalid query parameters: %w", err)
	}
	for _, name := range []string{
		"feed_id", "category_id", "before_entry_id", "after_entry_id",
		"before", "after", "published_before", "published_after", "changed_before", "changed_after",
		"starred", "globally_visible", "order", "direction", "search",
	} {
		if len(values[name]) > 1 {
			return fmt.Errorf("invalid %s parameter, must be specified only once", name)
		}
	}

	for _, name := range []string{"feed_id", "category_id", "before_entry_id", "after_entry_id"} {
		if !values.Has(name) {
			continue
		}
		value, err := strconv.ParseInt(values.Get(name), 10, 64)
		if err != nil || value <= 0 {
			return fmt.Errorf("invalid %s parameter, must be a positive integer", name)
		}
	}

	// Reject timestamps outside PostgreSQL's finite timestamp range.
	maxTimestamp := time.Date(294277, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
	for _, name := range []string{"before", "after", "published_before", "published_after", "changed_before", "changed_after"} {
		if !values.Has(name) {
			continue
		}
		value, err := strconv.ParseInt(values.Get(name), 10, 64)
		if err != nil || value < 0 {
			return fmt.Errorf("invalid %s parameter, must be a non-negative integer", name)
		}
		if value >= maxTimestamp {
			return fmt.Errorf("invalid %s parameter, timestamp is out of range", name)
		}
	}

	for _, name := range []string{"starred", "globally_visible"} {
		if value := values.Get(name); values.Has(name) && value != "true" && value != "false" {
			return fmt.Errorf(`invalid %s parameter, must be "true" or "false"`, name)
		}
	}

	for _, name := range []string{"search", "tags"} {
		for _, value := range values[name] {
			if !utf8.ValidString(value) || strings.ContainsRune(value, '\x00') {
				return fmt.Errorf("invalid %s parameter, must be valid UTF-8 without null bytes", name)
			}
		}
	}

	for _, status := range values["status"] {
		if strings.TrimSpace(status) == "" {
			return validator.ValidateEntryStatus("")
		}
	}

	if values.Has("order") {
		if err := validator.ValidateEntryOrder(values.Get("order")); err != nil {
			return err
		}
	}
	if values.Has("direction") {
		if err := validator.ValidateDirection(values.Get("direction")); err != nil {
			return err
		}
	}

	return nil
}
