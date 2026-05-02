// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/integration"
	"miniflux.app/v2/internal/mediaproxy"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/proxyrotator"
	"miniflux.app/v2/internal/reader/fetcher"
	mff "miniflux.app/v2/internal/reader/handler"
	mfs "miniflux.app/v2/internal/reader/subscription"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/urllib"
	"miniflux.app/v2/internal/validator"
)

var (
	errEmptyFeedTitle   = errors.New("googlereader: empty feed title")
	errFeedNotFound     = errors.New("googlereader: feed not found")
	errCategoryNotFound = errors.New("googlereader: category not found")
	errSimultaneously   = fmt.Errorf("googlereader: %s and %s should not be supplied simultaneously", keptUnreadStreamSuffix, readStreamSuffix)
)

// NewHandler returns an http.Handler that handles Google Reader API calls.
// The returned handler expects the base path to be stripped from the request URL.
func NewHandler(store *storage.Storage) http.Handler {
	h := &greaderHandler{
		store: store,
	}

	authMiddleware := newAuthMiddleware(store)
	withApiKeyAuth := func(fn http.HandlerFunc) http.Handler {
		return authMiddleware.validateApiKey(fn)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /accounts/ClientLogin", h.clientLoginHandler)
	mux.Handle("GET /reader/api/0/token", withApiKeyAuth(h.tokenHandler))
	mux.Handle("POST /reader/api/0/edit-tag", withApiKeyAuth(h.editTagHandler))
	mux.Handle("POST /reader/api/0/rename-tag", withApiKeyAuth(h.renameTagHandler))
	mux.Handle("POST /reader/api/0/disable-tag", withApiKeyAuth(h.disableTagHandler))
	mux.Handle("GET /reader/api/0/tag/list", withApiKeyAuth(h.tagListHandler))
	mux.Handle("GET /reader/api/0/user-info", withApiKeyAuth(h.userInfoHandler))
	mux.Handle("GET /reader/api/0/subscription/list", withApiKeyAuth(h.subscriptionListHandler))
	mux.Handle("POST /reader/api/0/subscription/edit", withApiKeyAuth(h.editSubscriptionHandler))
	mux.Handle("POST /reader/api/0/subscription/quickadd", withApiKeyAuth(h.quickAddHandler))
	mux.Handle("GET /reader/api/0/stream/items/ids", withApiKeyAuth(h.streamItemIDsHandler))
	mux.Handle("POST /reader/api/0/stream/items/contents", withApiKeyAuth(h.streamItemContentsHandler))
	mux.Handle("POST /reader/api/0/mark-all-as-read", withApiKeyAuth(h.markAllAsReadHandler))
	mux.Handle("GET /reader/api/0/", withApiKeyAuth(h.fallbackHandler))
	mux.Handle("POST /reader/api/0/", withApiKeyAuth(h.fallbackHandler))

	return mux
}

type greaderHandler struct {
	store *storage.Storage
}

func (h *greaderHandler) clientLoginHandler(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /accounts/ClientLogin",
		slog.String("handler", "clientLoginHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
	)

	if err := r.ParseForm(); err != nil {
		slog.Warn("[GoogleReader] Could not parse request form data",
			slog.Bool("authentication_failed", true),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
			slog.Any("error", err),
		)
		response.JSONUnauthorized(w, r)
		return
	}

	username := r.Form.Get("Email")
	password := r.Form.Get("Passwd")
	output := r.Form.Get("output")

	if username == "" || password == "" {
		slog.Warn("[GoogleReader] Empty username or password",
			slog.Bool("authentication_failed", true),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
		)
		response.JSONUnauthorized(w, r)
		return
	}

	if err := h.store.GoogleReaderUserCheckPassword(username, password); err != nil {
		slog.Warn("[GoogleReader] Invalid username or password",
			slog.Bool("authentication_failed", true),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
			slog.String("username", username),
			slog.Any("error", err),
		)
		response.JSONUnauthorized(w, r)
		return
	}

	slog.Info("[GoogleReader] User authenticated successfully",
		slog.Bool("authentication_successful", true),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.String("username", username),
	)

	integration, err := h.store.GoogleReaderUserGetIntegration(username)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	h.store.SetLastLogin(integration.UserID)

	token := getAuthToken(integration.GoogleReaderUsername, integration.GoogleReaderPassword)
	slog.Debug("[GoogleReader] Created token",
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.String("username", username),
	)

	result := loginResponse{SID: token, LSID: token, Auth: token}
	if output == "json" {
		response.JSON(w, r, result)
		return
	}

	response.Text(w, r, result.String())
}

func (h *greaderHandler) tokenHandler(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /token",
		slog.String("handler", "tokenHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
	)

	if !request.IsAuthenticated(r) {
		slog.Warn("[GoogleReader] User is not authenticated",
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
		)
		response.JSONUnauthorized(w, r)
		return
	}

	token := request.GoogleReaderToken(r)
	if token == "" {
		slog.Warn("[GoogleReader] User does not have token",
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
			slog.Int64("user_id", request.UserID(r)),
		)
		response.JSONUnauthorized(w, r)
		return
	}

	slog.Debug("[GoogleReader] Token handler",
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", request.UserID(r)),
	)

	response.Text(w, r, token)
}

func (h *greaderHandler) editTagHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /edit-tag",
		slog.String("handler", "editTagHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", userID),
	)

	if err := r.ParseForm(); err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	addTags, err := getStreams(r.PostForm[paramTagsAdd], userID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	removeTags, err := getStreams(r.PostForm[paramTagsRemove], userID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	if len(addTags) == 0 && len(removeTags) == 0 {
		err = errors.New("googlreader: add or/and remove tags should be supplied")
		response.JSONServerError(w, r, err)
		return
	}
	tags, err := checkAndSimplifyTags(addTags, removeTags)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	itemIDs, err := parseItemIDsFromRequest(r)
	if err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	slog.Debug("[GoogleReader] Edited tags",
		slog.String("handler", "editTagHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", userID),
		slog.Any("item_ids", itemIDs),
		slog.Any("tags", tags),
	)

	builder := h.store.NewEntryQueryBuilder(userID)
	builder.WithEntryIDs(itemIDs)

	entries, err := builder.GetEntries()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	n := 0
	readEntryIDs := make([]int64, 0)
	unreadEntryIDs := make([]int64, 0)
	starredEntryIDs := make([]int64, 0)
	unstarredEntryIDs := make([]int64, 0)
	for _, entry := range entries {
		if read, exists := tags[ReadStream]; exists {
			if read && entry.Status == model.EntryStatusUnread {
				readEntryIDs = append(readEntryIDs, entry.ID)
			} else if !read && entry.Status == model.EntryStatusRead {
				unreadEntryIDs = append(unreadEntryIDs, entry.ID)
			}
		}
		if starred, exists := tags[StarredStream]; exists {
			if starred && !entry.Starred {
				starredEntryIDs = append(starredEntryIDs, entry.ID)
				// filter the original array
				entries[n] = entry
				n++
			} else if !starred && entry.Starred {
				unstarredEntryIDs = append(unstarredEntryIDs, entry.ID)
			}
		}
	}
	entries = entries[:n]
	if len(readEntryIDs) > 0 {
		err = h.store.SetEntriesStatus(userID, readEntryIDs, model.EntryStatusRead)
		if err != nil {
			response.JSONServerError(w, r, err)
			return
		}
	}

	if len(unreadEntryIDs) > 0 {
		err = h.store.SetEntriesStatus(userID, unreadEntryIDs, model.EntryStatusUnread)
		if err != nil {
			response.JSONServerError(w, r, err)
			return
		}
	}

	if len(unstarredEntryIDs) > 0 {
		err = h.store.SetEntriesStarredState(userID, unstarredEntryIDs, false)
		if err != nil {
			response.JSONServerError(w, r, err)
			return
		}
	}

	if len(starredEntryIDs) > 0 {
		err = h.store.SetEntriesStarredState(userID, starredEntryIDs, true)
		if err != nil {
			response.JSONServerError(w, r, err)
			return
		}
	}

	if len(entries) > 0 {
		settings, err := h.store.Integration(userID)
		if err != nil {
			response.JSONServerError(w, r, err)
			return
		}

		for _, entry := range entries {
			e := entry
			go func() {
				integration.SendEntry(e, settings)
			}()
		}
	}

	response.Text(w, r, "OK")
}

func (h *greaderHandler) quickAddHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /subscription/quickadd",
		slog.String("handler", "quickAddHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", userID),
	)

	err := r.ParseForm()
	if err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	feedURL := r.Form.Get(paramQuickAdd)
	if !urllib.IsAbsoluteURL(feedURL) {
		response.JSONBadRequest(w, r, fmt.Errorf("googlereader: invalid URL: %s", feedURL))
		return
	}

	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxyRotator(proxyrotator.ProxyRotatorInstance)

	var rssBridgeURL string
	var rssBridgeToken string
	if intg, err := h.store.Integration(userID); err == nil && intg != nil && intg.RSSBridgeEnabled {
		rssBridgeURL = intg.RSSBridgeURL
		rssBridgeToken = intg.RSSBridgeToken
	}

	subscriptions, localizedError := mfs.NewSubscriptionFinder(requestBuilder).FindSubscriptions(feedURL, rssBridgeURL, rssBridgeToken)
	if localizedError != nil {
		response.JSONServerError(w, r, localizedError.Error())
		return
	}

	if len(subscriptions) == 0 {
		response.JSON(w, r, quickAddResponse{
			NumResults: 0,
		})
		return
	}

	toSubscribe := Stream{FeedStream, subscriptions[0].URL}
	category := Stream{NoStream, ""}
	newFeed, err := subscribe(toSubscribe, category, "", h.store, userID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	slog.Debug("[GoogleReader] Added a new feed",
		slog.String("handler", "quickAddHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", userID),
		slog.String("feed_url", newFeed.FeedURL),
	)

	response.JSON(w, r, quickAddResponse{
		NumResults: 1,
		Query:      newFeed.FeedURL,
		StreamID:   feedPrefix + strconv.FormatInt(newFeed.ID, 10),
		StreamName: newFeed.Title,
	})
}

func getFeed(stream Stream, store *storage.Storage, userID int64) (*model.Feed, error) {
	feedID, err := strconv.ParseInt(stream.ID, 10, 64)
	if err != nil {
		return nil, err
	}
	return store.FeedByID(userID, feedID)
}

func getOrCreateCategory(streamCategory Stream, store *storage.Storage, userID int64) (*model.Category, error) {
	switch {
	case streamCategory.ID == "":
		return store.FirstCategory(userID)
	case store.CategoryTitleExists(userID, streamCategory.ID):
		return store.CategoryByTitle(userID, streamCategory.ID)
	default:
		return store.CreateCategory(userID, &model.CategoryCreationRequest{
			Title: streamCategory.ID,
		})
	}
}

func subscribe(newFeed Stream, category Stream, title string, store *storage.Storage, userID int64) (*model.Feed, error) {
	destCategory, err := getOrCreateCategory(category, store, userID)
	if err != nil {
		return nil, err
	}

	feedRequest := model.FeedCreationRequest{
		FeedURL:    newFeed.ID,
		CategoryID: destCategory.ID,
	}
	verr := validator.ValidateFeedCreation(store, userID, &feedRequest)
	if verr != nil {
		return nil, verr.Error()
	}

	created, localizedError := mff.CreateFeed(store, userID, &feedRequest)
	if localizedError != nil {
		return nil, localizedError.Error()
	}

	if title != "" {
		feedModification := model.FeedModificationRequest{
			Title: &title,
		}
		feedModification.Patch(created)
		if err := store.UpdateFeed(created); err != nil {
			return nil, err
		}
	}

	return created, nil
}

func unsubscribe(streams []Stream, store *storage.Storage, userID int64) error {
	for _, stream := range streams {
		feedID, err := strconv.ParseInt(stream.ID, 10, 64)
		if err != nil {
			return err
		}
		err = store.RemoveFeed(userID, feedID)
		if err != nil {
			return err
		}
	}
	return nil
}

func rename(feedStream Stream, title string, store *storage.Storage, userID int64) error {
	slog.Debug("[GoogleReader] Renaming feed",
		slog.Int64("user_id", userID),
		slog.Any("feed_stream", feedStream),
		slog.String("new_title", title),
	)

	if title == "" {
		return errEmptyFeedTitle
	}

	feed, err := getFeed(feedStream, store, userID)
	if err != nil {
		return err
	}
	if feed == nil {
		return errFeedNotFound
	}

	feedModification := model.FeedModificationRequest{
		Title: &title,
	}
	feedModification.Patch(feed)
	return store.UpdateFeed(feed)
}

func move(feedStream Stream, labelStream Stream, store *storage.Storage, userID int64) error {
	slog.Debug("[GoogleReader] Moving feed",
		slog.Int64("user_id", userID),
		slog.Any("feed_stream", feedStream),
		slog.Any("label_stream", labelStream),
	)

	feed, err := getFeed(feedStream, store, userID)
	if err != nil {
		return err
	}
	if feed == nil {
		return errFeedNotFound
	}

	category, err := getOrCreateCategory(labelStream, store, userID)
	if err != nil {
		return err
	}
	if category == nil {
		return errCategoryNotFound
	}

	feedModification := model.FeedModificationRequest{
		CategoryID: &category.ID,
	}
	feedModification.Patch(feed)
	return store.UpdateFeed(feed)
}

func (h *greaderHandler) feedIconURL(f *model.Feed) string {
	if f.Icon != nil && f.Icon.ExternalIconID != "" {
		return config.Opts.BaseURL() + "/feed-icon/" + f.Icon.ExternalIconID
	}
	return ""
}

func (h *greaderHandler) editSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /subscription/edit",
		slog.String("handler", "editSubscriptionHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", userID),
	)

	if err := r.ParseForm(); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	streamIds, err := getStreams(r.Form[paramStreamID], userID)
	if err != nil || len(streamIds) == 0 {
		response.JSONBadRequest(w, r, errors.New("googlereader: no valid stream IDs provided"))
		return
	}

	newLabel, err := getStream(r.Form.Get(paramTagsAdd), userID)
	if err != nil {
		response.JSONBadRequest(w, r, fmt.Errorf("googlereader: invalid data in %s", paramTagsAdd))
		return
	}

	title := r.Form.Get(paramTitle)
	action := r.Form.Get(paramSubscribeAction)

	switch action {
	case "subscribe":
		_, err := subscribe(streamIds[0], newLabel, title, h.store, userID)
		if err != nil {
			response.JSONServerError(w, r, err)
			return
		}
	case "unsubscribe":
		err := unsubscribe(streamIds, h.store, userID)
		if err != nil {
			response.JSONServerError(w, r, err)
			return
		}
	case "edit":
		if title != "" {
			if err := rename(streamIds[0], title, h.store, userID); err != nil {
				if errors.Is(err, errFeedNotFound) || errors.Is(err, errEmptyFeedTitle) {
					response.JSONBadRequest(w, r, err)
				} else {
					response.JSONServerError(w, r, err)
				}
				return
			}
		}

		if r.Form.Has(paramTagsAdd) {
			if newLabel.Type != LabelStream {
				response.JSONBadRequest(w, r, errors.New("destination must be a label"))
				return
			}

			if err := move(streamIds[0], newLabel, h.store, userID); err != nil {
				if errors.Is(err, errFeedNotFound) || errors.Is(err, errCategoryNotFound) {
					response.JSONBadRequest(w, r, err)
				} else {
					response.JSONServerError(w, r, err)
				}
				return
			}
		}
	default:
		response.JSONBadRequest(w, r, fmt.Errorf("googlereader: unrecognized action %s", action))
		return
	}

	response.Text(w, r, "OK")
}

func (h *greaderHandler) streamItemContentsHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	userName := request.UserName(r)
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /stream/items/contents",
		slog.String("handler", "streamItemContentsHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", userID),
	)

	if err := checkOutputFormat(r); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	err := r.ParseForm()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	requestModifiers, err := parseStreamFilterFromRequest(r)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	streamPrefix := fmt.Sprintf(userStreamPrefix, userID)
	userReadingList := streamPrefix + readingListStreamSuffix
	userRead := streamPrefix + readStreamSuffix
	userStarred := streamPrefix + starredStreamSuffix

	itemIDs, err := parseItemIDsFromRequest(r)
	if err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	slog.Debug("[GoogleReader] Fetching item contents",
		slog.String("handler", "streamItemContentsHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", userID),
		slog.Any("item_ids", itemIDs),
	)

	builder := h.store.NewEntryQueryBuilder(userID)
	builder.WithEnclosures()
	builder.WithEntryIDs(itemIDs)
	builder.WithSorting(model.DefaultSortingOrder, requestModifiers.SortDirection)

	entries, err := builder.GetEntries()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	result := streamContentItemsResponse{
		Direction: "ltr",
		ID:        "user/-/state/com.google/reading-list",
		Title:     "Reading List",
		Updated:   time.Now().Unix(),
		Self: []contentHREF{{
			HREF: config.Opts.BaseURL() + "/reader/api/0/stream/items/contents",
		}},
		Author: userName,
		Items:  make([]contentItem, len(entries)),
	}

	labelPrefix := fmt.Sprintf(userLabelPrefix, userID)
	for i, entry := range entries {
		enclosures := make([]contentItemEnclosure, 0, len(entry.Enclosures))
		for _, enclosure := range entry.Enclosures {
			enclosures = append(enclosures, contentItemEnclosure{URL: enclosure.URL, Type: enclosure.MimeType})
		}
		categories := make([]string, 0, 4)
		categories = append(categories, userReadingList)
		if entry.Feed.Category.Title != "" {
			categories = append(categories, labelPrefix+entry.Feed.Category.Title)
		}
		if entry.Status == model.EntryStatusRead {
			categories = append(categories, userRead)
		}

		if entry.Starred {
			categories = append(categories, userStarred)
		}

		entry.Content = mediaproxy.RewriteDocumentWithAbsoluteProxyURL(entry.Content)
		entry.Enclosures.ProxifyEnclosureURL(config.Opts.MediaProxyMode(), config.Opts.MediaProxyResourceTypes())

		result.Items[i] = contentItem{
			ID:            convertEntryIDToLongFormItemID(entry.ID),
			Title:         entry.Title,
			Author:        entry.Author,
			TimestampUsec: strconv.FormatInt(entry.Date.UnixMicro(), 10),
			CrawlTimeMsec: strconv.FormatInt(entry.CreatedAt.UnixMilli(), 10),
			Published:     entry.Date.Unix(),
			Updated:       entry.ChangedAt.Unix(),
			Categories:    categories,
			Canonical: []contentHREF{
				{
					HREF: entry.URL,
				},
			},
			Alternate: []contentHREFType{
				{
					HREF: entry.URL,
					Type: "text/html",
				},
			},
			Content: contentItemContent{
				Direction: "ltr",
				Content:   entry.Content,
			},
			Summary: contentItemContent{
				Direction: "ltr",
				Content:   entry.Content,
			},
			Origin: contentItemOrigin{
				StreamID: feedPrefix + strconv.FormatInt(entry.FeedID, 10),
				Title:    entry.Feed.Title,
				HTMLUrl:  entry.Feed.SiteURL,
			},
			Enclosure: enclosures,
		}
	}

	response.JSON(w, r, result)
}

func (h *greaderHandler) disableTagHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /disable-tags",
		slog.String("handler", "disableTagHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", userID),
	)

	err := r.ParseForm()
	if err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	streams, err := getStreams(r.Form[paramStreamID], userID)
	if err != nil {
		response.JSONBadRequest(w, r, fmt.Errorf("googlereader: invalid data in %s", paramStreamID))
		return
	}

	titles := make([]string, len(streams))
	for i, stream := range streams {
		if stream.Type != LabelStream {
			response.JSONBadRequest(w, r, errors.New("googlereader: only labels are supported"))
			return
		}
		titles[i] = stream.ID
	}

	err = h.store.RemoveAndReplaceCategoriesByName(userID, titles)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.Text(w, r, "OK")
}

func (h *greaderHandler) renameTagHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /rename-tag",
		slog.String("handler", "renameTagHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
	)

	err := r.ParseForm()
	if err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	source, err := getStream(r.Form.Get(paramStreamID), userID)
	if err != nil {
		response.JSONBadRequest(w, r, fmt.Errorf("googlereader: invalid data in %s", paramStreamID))
		return
	}

	destination, err := getStream(r.Form.Get(paramDestination), userID)
	if err != nil {
		response.JSONBadRequest(w, r, fmt.Errorf("googlereader: invalid data in %s", paramDestination))
		return
	}

	if source.Type != LabelStream || destination.Type != LabelStream {
		response.JSONBadRequest(w, r, errors.New("googlereader: only labels supported"))
		return
	}

	if destination.ID == "" {
		response.JSONBadRequest(w, r, errors.New("googlereader: empty destination name"))
		return
	}

	category, err := h.store.CategoryByTitle(userID, source.ID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	if category == nil {
		response.JSONNotFound(w, r)
		return
	}

	categoryModificationRequest := model.CategoryModificationRequest{
		Title: new(destination.ID),
	}

	if validationError := validator.ValidateCategoryModification(h.store, userID, category.ID, &categoryModificationRequest); validationError != nil {
		response.JSONBadRequest(w, r, validationError.Error())
		return
	}

	categoryModificationRequest.Patch(category)

	if err := h.store.UpdateCategory(category); err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.Text(w, r, "OK")
}

func (h *greaderHandler) tagListHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /tags/list",
		slog.String("handler", "tagListHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
	)

	if err := checkOutputFormat(r); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	var result tagsResponse
	categories, err := h.store.Categories(userID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	result.Tags = make([]subscriptionCategoryResponse, 0, 1+len(categories))
	result.Tags = append(result.Tags, subscriptionCategoryResponse{
		ID: fmt.Sprintf(userStreamPrefix, userID) + starredStreamSuffix,
	})
	labelPrefix := fmt.Sprintf(userLabelPrefix, userID)
	for _, category := range categories {
		result.Tags = append(result.Tags, subscriptionCategoryResponse{
			ID:    labelPrefix + category.Title,
			Label: category.Title,
			Type:  "folder",
		})
	}
	response.JSON(w, r, result)
}

func (h *greaderHandler) subscriptionListHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /subscription/list",
		slog.String("handler", "subscriptionListHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
	)

	if err := checkOutputFormat(r); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	var result subscriptionsResponse
	feeds, err := h.store.Feeds(userID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	labelPrefix := fmt.Sprintf(userLabelPrefix, userID)
	result.Subscriptions = make([]subscriptionResponse, 0, len(feeds))
	for _, feed := range feeds {
		result.Subscriptions = append(result.Subscriptions, subscriptionResponse{
			ID:         feedPrefix + strconv.FormatInt(feed.ID, 10),
			Title:      feed.Title,
			URL:        feed.FeedURL,
			Categories: []subscriptionCategoryResponse{{labelPrefix + feed.Category.Title, feed.Category.Title, "folder"}},
			HTMLURL:    feed.SiteURL,
			IconURL:    h.feedIconURL(feed),
		})
	}
	response.JSON(w, r, result)
}

func (h *greaderHandler) fallbackHandler(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] API endpoint not implemented yet",
		slog.Any("url", r.RequestURI),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
	)

	response.JSON(w, r, []string{})
}

func (h *greaderHandler) userInfoHandler(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /user-info",
		slog.String("handler", "userInfoHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
	)

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	if user == nil {
		response.JSONNotFound(w, r)
		return
	}

	userInfo := userInfoResponse{UserID: strconv.FormatInt(user.ID, 10), UserName: user.Username, UserProfileID: strconv.FormatInt(user.ID, 10), UserEmail: user.Username}
	response.JSON(w, r, userInfo)
}

func (h *greaderHandler) streamItemIDsHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /stream/items/ids",
		slog.String("handler", "streamItemIDsHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", userID),
	)

	if err := checkOutputFormat(r); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	rm, err := parseStreamFilterFromRequest(r)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	slog.Debug("[GoogleReader] Request Modifiers",
		slog.String("handler", "streamItemIDsHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Any("modifiers", rm),
	)

	if len(rm.Streams) != 1 {
		response.JSONServerError(w, r, errors.New("googlereader: only one stream type expected"))
		return
	}
	switch rm.Streams[0].Type {
	case ReadingListStream:
		h.handleReadingListStreamHandler(w, r, rm)
	case StarredStream:
		h.handleStarredStreamHandler(w, r, rm)
	case ReadStream:
		h.handleReadStreamHandler(w, r, rm)
	case FeedStream:
		h.handleFeedStreamHandler(w, r, rm)
	default:
		slog.Warn("[GoogleReader] Unknown Stream",
			slog.String("handler", "streamItemIDsHandler"),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
			slog.Any("stream_type", rm.Streams[0].Type),
		)
		response.JSONServerError(w, r, fmt.Errorf("googlereader: unknown stream type %s", rm.Streams[0].Type))
	}
}

func (h *greaderHandler) handleReadingListStreamHandler(w http.ResponseWriter, r *http.Request, rm requestModifiers) {
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle ReadingListStream",
		slog.String("handler", "handleReadingListStreamHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
	)

	builder := h.store.NewEntryQueryBuilder(rm.UserID)
	for _, s := range rm.ExcludeTargets {
		switch s.Type {
		case ReadStream:
			builder.WithStatus(model.EntryStatusUnread)
		default:
			slog.Warn("[GoogleReader] Unknown ExcludeTargets filter type",
				slog.String("handler", "handleReadingListStreamHandler"),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", r.UserAgent()),
				slog.Int("filter_type", int(s.Type)),
			)
		}
	}

	builder.WithLimit(rm.Count)
	builder.WithOffset(rm.Offset)
	builder.WithSorting(model.DefaultSortingOrder, rm.SortDirection)
	if rm.StartTime > 0 {
		builder.AfterPublishedDate(time.Unix(rm.StartTime, 0))
	}
	if rm.StopTime > 0 {
		builder.BeforePublishedDate(time.Unix(rm.StopTime, 0))
	}

	itemRefs, continuation, err := getItemRefsAndContinuation(*builder, rm)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	response.JSON(w, r, streamIDResponse{itemRefs, continuation})
}

func (h *greaderHandler) handleStarredStreamHandler(w http.ResponseWriter, r *http.Request, rm requestModifiers) {
	builder := h.store.NewEntryQueryBuilder(rm.UserID)
	builder.WithStarred(true)
	builder.WithLimit(rm.Count)
	builder.WithOffset(rm.Offset)
	builder.WithSorting(model.DefaultSortingOrder, rm.SortDirection)
	if rm.StartTime > 0 {
		builder.AfterPublishedDate(time.Unix(rm.StartTime, 0))
	}
	if rm.StopTime > 0 {
		builder.BeforePublishedDate(time.Unix(rm.StopTime, 0))
	}
	itemRefs, continuation, err := getItemRefsAndContinuation(*builder, rm)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	response.JSON(w, r, streamIDResponse{itemRefs, continuation})
}

func (h *greaderHandler) handleReadStreamHandler(w http.ResponseWriter, r *http.Request, rm requestModifiers) {
	builder := h.store.NewEntryQueryBuilder(rm.UserID)
	builder.WithStatus(model.EntryStatusRead)
	builder.WithLimit(rm.Count)
	builder.WithOffset(rm.Offset)
	builder.WithSorting(model.DefaultSortingOrder, rm.SortDirection)
	if rm.StartTime > 0 {
		builder.AfterPublishedDate(time.Unix(rm.StartTime, 0))
	}
	if rm.StopTime > 0 {
		builder.BeforePublishedDate(time.Unix(rm.StopTime, 0))
	}

	itemRefs, continuation, err := getItemRefsAndContinuation(*builder, rm)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	response.JSON(w, r, streamIDResponse{itemRefs, continuation})
}

func getItemRefsAndContinuation(builder storage.EntryQueryBuilder, rm requestModifiers) ([]itemRef, int, error) {
	rawEntryIDs, err := builder.GetEntryIDs()
	if err != nil {
		return nil, 0, err
	}
	var itemRefs = make([]itemRef, 0, len(rawEntryIDs))
	for _, entryID := range rawEntryIDs {
		formattedID := strconv.FormatInt(entryID, 10)
		itemRefs = append(itemRefs, itemRef{ID: formattedID})
	}

	totalEntries, err := builder.CountEntries()
	if err != nil {
		return nil, 0, err
	}
	continuation := 0
	if len(itemRefs)+rm.Offset < totalEntries {
		continuation = len(itemRefs) + rm.Offset
	}
	return itemRefs, continuation, nil
}

func (h *greaderHandler) handleFeedStreamHandler(w http.ResponseWriter, r *http.Request, rm requestModifiers) {
	feedID, err := strconv.ParseInt(rm.Streams[0].ID, 10, 64)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	builder := h.store.NewEntryQueryBuilder(rm.UserID)
	builder.WithFeedID(feedID)
	builder.WithLimit(rm.Count)
	builder.WithOffset(rm.Offset)
	builder.WithSorting(model.DefaultSortingOrder, rm.SortDirection)

	if rm.StartTime > 0 {
		builder.AfterPublishedDate(time.Unix(rm.StartTime, 0))
	}

	if rm.StopTime > 0 {
		builder.BeforePublishedDate(time.Unix(rm.StopTime, 0))
	}

	if len(rm.ExcludeTargets) > 0 {
		for _, s := range rm.ExcludeTargets {
			if s.Type == ReadStream {
				builder.WithoutStatus(model.EntryStatusRead)
			}
		}
	}
	itemRefs, continuation, err := getItemRefsAndContinuation(*builder, rm)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	response.JSON(w, r, streamIDResponse{itemRefs, continuation})
}

func (h *greaderHandler) markAllAsReadHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /mark-all-as-read",
		slog.String("handler", "markAllAsReadHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
	)

	if err := r.ParseForm(); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	stream, err := getStream(r.Form.Get(paramStreamID), userID)
	if err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	var before time.Time
	if timestampParamValue := r.Form.Get(paramTimestamp); timestampParamValue != "" {
		timestampParsedValue, err := strconv.ParseInt(timestampParamValue, 10, 64)
		if err != nil {
			response.JSONBadRequest(w, r, err)
			return
		}

		if timestampParsedValue > 0 {
			// It's unclear if the timestamp is in seconds or microseconds, so we try both using a naive approach.
			if len(timestampParamValue) >= 16 {
				before = time.UnixMicro(timestampParsedValue)
			} else {
				before = time.Unix(timestampParsedValue, 0)
			}
		}
	}

	if before.IsZero() {
		before = time.Now()
	}

	switch stream.Type {
	case FeedStream:
		feedID, err := strconv.ParseInt(stream.ID, 10, 64)
		if err != nil {
			response.JSONBadRequest(w, r, err)
			return
		}
		err = h.store.MarkFeedAsRead(userID, feedID, before)
		if err != nil {
			response.JSONServerError(w, r, err)
			return
		}
	case LabelStream:
		category, err := h.store.CategoryByTitle(userID, stream.ID)
		if err != nil {
			response.JSONServerError(w, r, err)
			return
		}
		if category == nil {
			response.JSONNotFound(w, r)
			return
		}
		if err := h.store.MarkCategoryAsRead(userID, category.ID, before); err != nil {
			response.JSONServerError(w, r, err)
			return
		}
	case ReadingListStream:
		if err = h.store.MarkAllAsReadBeforeDate(userID, before); err != nil {
			response.JSONServerError(w, r, err)
			return
		}
	}

	response.Text(w, r, "OK")
}

func checkAndSimplifyTags(addTags []Stream, removeTags []Stream) (map[StreamType]bool, error) {
	tags := make(map[StreamType]bool)
	for _, s := range addTags {
		switch s.Type {
		case ReadStream:
			if _, ok := tags[KeptUnreadStream]; ok {
				return nil, errSimultaneously
			}
			tags[ReadStream] = true
		case KeptUnreadStream:
			if _, ok := tags[ReadStream]; ok {
				return nil, errSimultaneously
			}
			tags[ReadStream] = false
		case StarredStream:
			tags[StarredStream] = true
		case BroadcastStream, LikeStream:
			slog.Debug("Broadcast & Like tags are not implemented!")
		default:
			return nil, fmt.Errorf("googlereader: unsupported tag type: %s", s.Type)
		}
	}
	for _, s := range removeTags {
		switch s.Type {
		case ReadStream:
			if _, ok := tags[ReadStream]; ok {
				return nil, errSimultaneously
			}
			tags[ReadStream] = false
		case KeptUnreadStream:
			if _, ok := tags[ReadStream]; ok {
				return nil, errSimultaneously
			}
			tags[ReadStream] = true
		case StarredStream:
			if _, ok := tags[StarredStream]; ok {
				return nil, fmt.Errorf("googlereader: %s should not be supplied for add and remove simultaneously", starredStreamSuffix)
			}
			tags[StarredStream] = false
		case BroadcastStream, LikeStream:
			slog.Debug("Broadcast & Like tags are not implemented!")
		default:
			return nil, fmt.Errorf("googlereader: unsupported tag type: %s", s.Type)
		}
	}

	return tags, nil
}

func checkOutputFormat(r *http.Request) error {
	var output string
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			return err
		}
		output = r.Form.Get("output")
	} else {
		output = request.QueryStringParam(r, "output", "")
	}
	if output != "json" {
		return errors.New("googlereader: only json output is supported")
	}
	return nil
}
