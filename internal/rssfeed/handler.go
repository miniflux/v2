// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rssfeed // import "miniflux.app/v2/internal/rssfeed"

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
)

const (
	defaultEntriesLimit = 50

	// maxEntriesLimit is the maximum number of entries returned by a single
	// request. It matches the REST API page size limit, so that a reader can
	// drain a large backlog without being forced to send many requests.
	maxEntriesLimit = model.MaxEntryLimit
)

// streamKind identifies which subset of entries a feed exposes.
type streamKind int

const (
	streamUnread streamKind = iota
	streamStarred
	streamRead
	streamAll
	streamFeed
	streamCategory
	streamSearch
)

// documentFormat identifies the serialization format used for the response.
type documentFormat int

const (
	formatRSS documentFormat = iota
	formatAtom
)

// NewHandler returns an http.Handler that serves RSS 2.0 and Atom feeds.
// The returned handler expects the base path to be stripped from the request URL.
func NewHandler(store *storage.Storage) http.Handler {
	return newHandler(
		&storageFeedStore{store: store},
		newAuthMiddleware(&storageAuthenticator{store: store}),
	)
}

// newHandler assembles the feed handler from its dependencies.
func newHandler(store feedStore, auth *authMiddleware) http.Handler {
	h := &handler{store: store}

	register := func(mux *http.ServeMux, prefix string, format documentFormat) {
		contentType := "application/rss+xml; charset=utf-8"
		if format == formatAtom {
			contentType = "application/atom+xml; charset=utf-8"
		}

		mux.Handle("GET "+prefix+"/unread", auth.validateAPIKey(h.serveFeed(streamUnread, format, contentType)))
		mux.Handle("GET "+prefix+"/starred", auth.validateAPIKey(h.serveFeed(streamStarred, format, contentType)))
		mux.Handle("GET "+prefix+"/read", auth.validateAPIKey(h.serveFeed(streamRead, format, contentType)))
		mux.Handle("GET "+prefix+"/all", auth.validateAPIKey(h.serveFeed(streamAll, format, contentType)))
		mux.Handle("GET "+prefix+"/feed/{feedID}", auth.validateAPIKey(h.serveFeed(streamFeed, format, contentType)))
		mux.Handle("GET "+prefix+"/category/{categoryID}", auth.validateAPIKey(h.serveFeed(streamCategory, format, contentType)))
		mux.Handle("GET "+prefix+"/search", auth.validateAPIKey(h.serveFeed(streamSearch, format, contentType)))
		mux.Handle("GET "+prefix+"/{$}", auth.validateAPIKey(h.helpHandler(prefix)))
	}

	mux := http.NewServeMux()
	register(mux, "/rss", formatRSS)
	register(mux, "/atom", formatAtom)

	return mux
}

// feedStore is the subset of the storage layer required to serve feeds. It is
// an interface so that request handling can be tested without a database.
type feedStore interface {
	Entries(userID int64, filter entryFilter) (model.Entries, error)
	MarkAsRead(userID int64, entryIDs []int64) error
	Feed(userID int64, feedID int64) (*model.Feed, error)
	Category(userID int64, categoryID int64) (*model.Category, error)
}

// entryFilter describes the subset of entries exposed by a feed.
type entryFilter struct {
	Kind       streamKind
	FeedID     int64
	CategoryID int64
	Search     string
	Limit      int
}

type handler struct {
	store feedStore
}

func (h *handler) helpHandler(prefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		formatName := "RSS"
		if prefix == "/atom" {
			formatName = "Atom"
		}

		body := fmt.Sprintf(`Miniflux %s feeds

Authenticate with your Miniflux API key, either as the "api_key" query parameter,
as the password of HTTP Basic authentication (username = your Miniflux username),
or as a bearer token.

Available feeds:
  %[2]s/unread            Unread entries (marks returned entries as read)
  %[2]s/starred           Starred entries
  %[2]s/read              Read entries (history)
  %[2]s/all               All entries
  %[2]s/feed/{feedID}     Entries of a single feed
  %[2]s/category/{id}     Entries of a single category
  %[2]s/search?q=term     Full-text search

Optional query parameters:
  api_key=KEY    Miniflux API key
  limit=N        Number of entries (default %[4]d, max %[5]d)
  mark_read=0    Do not mark fetched entries as read (unread feed only)

Examples:
  %[3]s%[2]s/unread?api_key=KEY
  %[3]s%[2]s/starred?api_key=KEY&limit=100
  %[3]s%[2]s/feed/42?api_key=KEY
`, formatName, prefix, strings.TrimRight(config.Opts.BaseURL(), "/"), defaultEntriesLimit, maxEntriesLimit)

		sendPlainText(w, r, http.StatusOK, body)
	}
}

func (h *handler) serveFeed(kind streamKind, format documentFormat, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := request.UserID(r)

		filter := entryFilter{
			Kind:  kind,
			Limit: clampLimit(request.QueryIntParam(r, "limit", defaultEntriesLimit)),
		}

		var feed *model.Feed
		var category *model.Category
		var searchQuery string

		switch kind {
		case streamFeed:
			filter.FeedID = request.RouteInt64Param(r, "feedID")

			var err error
			feed, err = h.store.Feed(userID, filter.FeedID)
			if err != nil {
				sendPlainText(w, r, http.StatusInternalServerError, "Internal Server Error")
				return
			}
			if feed == nil {
				sendPlainText(w, r, http.StatusNotFound, "Feed not found")
				return
			}
		case streamCategory:
			filter.CategoryID = request.RouteInt64Param(r, "categoryID")

			var err error
			category, err = h.store.Category(userID, filter.CategoryID)
			if err != nil {
				sendPlainText(w, r, http.StatusInternalServerError, "Internal Server Error")
				return
			}
			if category == nil {
				sendPlainText(w, r, http.StatusNotFound, "Category not found")
				return
			}
		case streamSearch:
			searchQuery = strings.TrimSpace(request.QueryStringParam(r, "q", ""))
			if searchQuery == "" {
				sendPlainText(w, r, http.StatusBadRequest, `Missing "q" query parameter`)
				return
			}
			filter.Search = searchQuery
		}

		entries, err := h.store.Entries(userID, filter)
		if err != nil {
			slog.Error("[RSS] Unable to fetch entries",
				slog.Int64("user_id", userID),
				slog.Any("error", err),
			)
			sendPlainText(w, r, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		meta := h.buildMetadata(r, kind, feed, category, searchQuery, entries)

		var payload []byte
		if format == formatAtom {
			payload, err = renderAtom(meta)
		} else {
			payload, err = renderRSS(meta)
		}
		if err != nil {
			slog.Error("[RSS] Unable to render feed", slog.Any("error", err))
			sendPlainText(w, r, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		// Marking entries as read is opt-out for the unread feed only: the
		// other feeds are read-only views. Only entries included in the
		// response are marked, so a lower "limit" never marks unseen entries.
		// HEAD requests are ignored because feed readers use them to probe for
		// updates.
		if shouldMarkAsRead(kind, r, len(entries)) {
			entryIDs := make([]int64, 0, len(entries))
			for _, entry := range entries {
				entryIDs = append(entryIDs, entry.ID)
			}

			if err := h.store.MarkAsRead(userID, entryIDs); err != nil {
				slog.Warn("[RSS] Unable to mark fetched entries as read",
					slog.Int64("user_id", userID),
					slog.Int("entry_count", len(entryIDs)),
					slog.Any("error", err),
				)
			}
		}

		response.NewBuilder(w, r).
			WithHeader("Content-Type", contentType).
			WithHeader("Cache-Control", "private, no-store").
			WithBodyAsBytes(payload).
			Write()
	}
}

func (h *handler) buildMetadata(r *http.Request, kind streamKind, feed *model.Feed, category *model.Category, searchQuery string, entries model.Entries) feedMetadata {
	meta := feedMetadata{
		ID:      feedID(r),
		SelfURL: selfURL(r),
		Updated: time.Now(),
		Entries: entries,
	}

	switch kind {
	case streamUnread:
		meta.Title = "Unread"
		meta.Description = "Unread entries from Miniflux"
		meta.Link = strings.TrimRight(config.Opts.BaseURL(), "/") + "/unread"
	case streamStarred:
		meta.Title = "Starred"
		meta.Description = "Starred entries from Miniflux"
		meta.Link = strings.TrimRight(config.Opts.BaseURL(), "/") + "/starred"
	case streamRead:
		meta.Title = "History"
		meta.Description = "Read entries from Miniflux"
		meta.Link = strings.TrimRight(config.Opts.BaseURL(), "/") + "/history"
	case streamAll:
		meta.Title = "All entries"
		meta.Description = "All entries from Miniflux"
		meta.Link = strings.TrimRight(config.Opts.BaseURL(), "/") + "/"
	case streamFeed:
		meta.Title = feed.Title
		meta.Description = feed.Description
		if meta.Description == "" {
			meta.Description = "Entries from " + feed.Title
		}
		meta.Link = feed.SiteURL
		if meta.Link == "" {
			meta.Link = feed.FeedURL
		}
		meta.IconURL = feedIconURL(feed)
		meta.Language = feed.Language
	case streamCategory:
		meta.Title = category.Title
		meta.Description = "Entries from the " + category.Title + " category"
		meta.Link = strings.TrimRight(config.Opts.BaseURL(), "/") + "/category/" + strconv.FormatInt(category.ID, 10) + "/entries"
	case streamSearch:
		meta.Title = "Search: " + searchQuery
		meta.Description = "Search results for \"" + searchQuery + "\""
		meta.Link = strings.TrimRight(config.Opts.BaseURL(), "/") + "/search?q=" + searchQuery
	}

	if meta.Title == "" {
		meta.Title = "Miniflux"
	}

	if meta.Language == "" && len(entries) > 0 && entries[0].Feed != nil {
		meta.Language = entries[0].Feed.Language
	}

	return meta
}

// clampLimit normalizes the "limit" query parameter to a supported range.
func clampLimit(limit int) int {
	if limit <= 0 {
		return defaultEntriesLimit
	}
	if limit > maxEntriesLimit {
		return maxEntriesLimit
	}
	return limit
}

// shouldMarkAsRead reports whether the fetched entries of this stream must be
// marked as read. Only the unread feed mutates state, and only for entries that
// are actually returned to the client.
func shouldMarkAsRead(kind streamKind, r *http.Request, entryCount int) bool {
	return kind == streamUnread &&
		r.Method == http.MethodGet &&
		entryCount > 0 &&
		request.QueryBoolParam(r, "mark_read", true)
}

// feedID returns the permanent identifier of the current feed. Atom requires a
// feed id that never changes, so the identifier is built from the request path
// only: query parameters such as "limit", "mark_read", or "q" must not create
// a different identity for the same feed.
func feedID(r *http.Request) string {
	return strings.TrimRight(config.Opts.BaseURL(), "/") + r.URL.Path
}

// selfURL returns the public URL of the current feed with the API key removed,
// so that the generated document never leaks credentials.
func selfURL(r *http.Request) string {
	query := r.URL.Query()
	query.Del(apiKeyParam)

	url := feedID(r)
	if encoded := query.Encode(); encoded != "" {
		url += "?" + encoded
	}

	return url
}

func sendPlainText(w http.ResponseWriter, r *http.Request, status int, body string) {
	response.NewBuilder(w, r).
		WithStatus(status).
		WithHeader("Content-Type", "text/plain; charset=utf-8").
		WithHeader("Cache-Control", "private, no-store").
		WithBodyAsString(body + "\n").
		Write()
}
