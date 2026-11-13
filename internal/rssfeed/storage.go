// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rssfeed // import "miniflux.app/v2/internal/rssfeed"

import (
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
)

// storageFeedStore adapts the Miniflux storage layer to the feedStore
// interface used by the HTTP handlers.
type storageFeedStore struct {
	store *storage.Storage
}

func (s *storageFeedStore) Entries(userID int64, filter entryFilter) (model.Entries, error) {
	builder := s.store.NewEntryQueryBuilder(userID).
		WithEnclosures().
		WithLimit(filter.Limit).
		WithSorting(model.DefaultSortingOrder, "desc").
		WithSorting("id", "desc")

	switch filter.Kind {
	case streamUnread:
		builder = builder.WithStatuses(model.EntryStatusUnread)
	case streamStarred:
		builder = builder.WithStarred(true)
	case streamRead:
		builder = builder.WithStatuses(model.EntryStatusRead)
	case streamAll:
		// No additional filter.
	case streamFeed:
		builder = builder.WithFeedID(filter.FeedID)
	case streamCategory:
		builder = builder.WithCategoryID(filter.CategoryID)
	case streamSearch:
		builder = builder.WithSearchQuery(filter.Search)
	}

	return builder.GetEntries()
}

func (s *storageFeedStore) MarkAsRead(userID int64, entryIDs []int64) error {
	return s.store.SetEntriesStatus(userID, entryIDs, model.EntryStatusRead)
}

func (s *storageFeedStore) Feed(userID, feedID int64) (*model.Feed, error) {
	return s.store.FeedByID(userID, feedID)
}

func (s *storageFeedStore) Category(userID, categoryID int64) (*model.Category, error) {
	return s.store.Category(userID, categoryID)
}

// storageAuthenticator adapts the Miniflux storage layer to the
// apiKeyAuthenticator interface used by the authentication middleware.
type storageAuthenticator struct {
	store *storage.Storage
}

func (s *storageAuthenticator) UserByAPIKey(apiKey string) (*model.User, error) {
	return s.store.UserByAPIKey(apiKey)
}

func (s *storageAuthenticator) SetLastLogin(userID int64) error {
	return s.store.SetLastLogin(userID)
}

func (s *storageAuthenticator) SetAPIKeyUsedTimestamp(userID int64, apiKey string) error {
	return s.store.SetAPIKeyUsedTimestamp(userID, apiKey)
}
