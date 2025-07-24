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
	"miniflux.app/v2/internal/http/response/json"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/integration"
	"miniflux.app/v2/internal/mediaproxy"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/proxyrotator"
	"miniflux.app/v2/internal/reader/fetcher"
	mff "miniflux.app/v2/internal/reader/handler"
	mfs "miniflux.app/v2/internal/reader/subscription"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/validator"

	"github.com/gorilla/mux"
)

type handler struct {
	store  *storage.Storage
	router *mux.Router
}

var (
	errEmptyFeedTitle   = errors.New("googlereader: empty feed title")
	errFeedNotFound     = errors.New("googlereader: feed not found")
	errCategoryNotFound = errors.New("googlereader: category not found")
)

// Serve handles Google Reader API calls.
func Serve(router *mux.Router, store *storage.Storage) {
	handler := &handler{store, router}
	router.HandleFunc("/accounts/ClientLogin", handler.clientLoginHandler).Methods(http.MethodPost).Name("ClientLogin")

	middleware := newMiddleware(store)
	sr := router.PathPrefix("/reader/api/0").Subrouter()
	sr.Use(middleware.handleCORS)
	sr.Use(middleware.apiKeyAuth)
	sr.Methods(http.MethodOptions)
	sr.HandleFunc("/token", handler.tokenHandler).Methods(http.MethodGet).Name("Token")
	sr.HandleFunc("/edit-tag", handler.editTagHandler).Methods(http.MethodPost).Name("EditTag")
	sr.HandleFunc("/rename-tag", handler.renameTagHandler).Methods(http.MethodPost).Name("Rename Tag")
	sr.HandleFunc("/disable-tag", handler.disableTagHandler).Methods(http.MethodPost).Name("Disable Tag")
	sr.HandleFunc("/tag/list", handler.tagListHandler).Methods(http.MethodGet).Name("TagList")
	sr.HandleFunc("/user-info", handler.userInfoHandler).Methods(http.MethodGet).Name("UserInfo")
	sr.HandleFunc("/subscription/list", handler.subscriptionListHandler).Methods(http.MethodGet).Name("SubscriptonList")
	sr.HandleFunc("/subscription/edit", handler.editSubscriptionHandler).Methods(http.MethodPost).Name("SubscriptionEdit")
	sr.HandleFunc("/subscription/quickadd", handler.quickAddHandler).Methods(http.MethodPost).Name("QuickAdd")
	sr.HandleFunc("/stream/items/ids", handler.streamItemIDsHandler).Methods(http.MethodGet).Name("StreamItemIDs")
	sr.HandleFunc("/stream/items/contents", handler.streamItemContentsHandler).Methods(http.MethodPost).Name("StreamItemsContents")
	sr.HandleFunc("/mark-all-as-read", handler.markAllAsReadHandler).Methods(http.MethodPost).Name("MarkAllAsRead")
	sr.PathPrefix("/").HandlerFunc(handler.serveHandler).Methods(http.MethodPost, http.MethodGet).Name("GoogleReaderApiEndpoint")
}

func checkAndSimplifyTags(addTags []Stream, removeTags []Stream) (map[StreamType]bool, error) {
	tags := make(map[StreamType]bool)
	for _, s := range addTags {
		switch s.Type {
		case ReadStream:
			if _, ok := tags[KeptUnreadStream]; ok {
				return nil, fmt.Errorf("googlereader: %s and %s should not be supplied simultaneously", keptUnreadStreamSuffix, readStreamSuffix)
			}
			tags[ReadStream] = true
		case KeptUnreadStream:
			if _, ok := tags[ReadStream]; ok {
				return nil, fmt.Errorf("googlereader: %s and %s should not be supplied simultaneously", keptUnreadStreamSuffix, readStreamSuffix)
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
				return nil, fmt.Errorf("googlereader: %s and %s should not be supplied simultaneously", keptUnreadStreamSuffix, readStreamSuffix)
			}
			tags[ReadStream] = false
		case KeptUnreadStream:
			if _, ok := tags[ReadStream]; ok {
				return nil, fmt.Errorf("googlereader: %s and %s should not be supplied simultaneously", keptUnreadStreamSuffix, readStreamSuffix)
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
		err := fmt.Errorf("googlereader: only json output is supported")
		return err
	}
	return nil
}

func (h *handler) clientLoginHandler(w http.ResponseWriter, r *http.Request) {
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
		json.Unauthorized(w, r)
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
		json.Unauthorized(w, r)
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
		json.Unauthorized(w, r)
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
		json.ServerError(w, r, err)
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
		json.OK(w, r, result)
		return
	}

	builder := response.New(w, r)
	builder.WithHeader("Content-Type", "text/plain; charset=UTF-8")
	builder.WithBody(result.String())
	builder.Write()
}

func (h *handler) tokenHandler(w http.ResponseWriter, r *http.Request) {
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
		json.Unauthorized(w, r)
		return
	}

	token := request.GoolgeReaderToken(r)
	if token == "" {
		slog.Warn("[GoogleReader] User does not have token",
			slog.String("client_ip", clientIP),
			slog.String("user_agent", r.UserAgent()),
			slog.Int64("user_id", request.UserID(r)),
		)
		json.Unauthorized(w, r)
		return
	}

	slog.Debug("[GoogleReader] Token handler",
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", request.UserID(r)),
		slog.String("token", token),
	)

	w.Header().Add("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(token))
}

func (h *handler) editTagHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /edit-tag",
		slog.String("handler", "editTagHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", userID),
	)

	if err := r.ParseForm(); err != nil {
		json.ServerError(w, r, err)
		return
	}

	addTags, err := getStreams(r.PostForm[paramTagsAdd], userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	removeTags, err := getStreams(r.PostForm[paramTagsRemove], userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	if len(addTags) == 0 && len(removeTags) == 0 {
		err = fmt.Errorf("googlreader: add or/and remove tags should be supplied")
		json.ServerError(w, r, err)
		return
	}
	tags, err := checkAndSimplifyTags(addTags, removeTags)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	itemIDs, err := parseItemIDsFromRequest(r)
	if err != nil {
		json.BadRequest(w, r, err)
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
	builder.WithoutStatus(model.EntryStatusRemoved)

	entries, err := builder.GetEntries()
	if err != nil {
		json.ServerError(w, r, err)
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
			} else if entry.Status == model.EntryStatusRead {
				unreadEntryIDs = append(unreadEntryIDs, entry.ID)
			}
		}
		if starred, exists := tags[StarredStream]; exists {
			if starred && !entry.Starred {
				starredEntryIDs = append(starredEntryIDs, entry.ID)
				// filter the original array
				entries[n] = entry
				n++
			} else if entry.Starred {
				unstarredEntryIDs = append(unstarredEntryIDs, entry.ID)
			}
		}
	}
	entries = entries[:n]
	if len(readEntryIDs) > 0 {
		err = h.store.SetEntriesStatus(userID, readEntryIDs, model.EntryStatusRead)
		if err != nil {
			json.ServerError(w, r, err)
			return
		}
	}

	if len(unreadEntryIDs) > 0 {
		err = h.store.SetEntriesStatus(userID, unreadEntryIDs, model.EntryStatusUnread)
		if err != nil {
			json.ServerError(w, r, err)
			return
		}
	}

	if len(unstarredEntryIDs) > 0 {
		err = h.store.SetEntriesBookmarkedState(userID, unstarredEntryIDs, false)
		if err != nil {
			json.ServerError(w, r, err)
			return
		}
	}

	if len(starredEntryIDs) > 0 {
		err = h.store.SetEntriesBookmarkedState(userID, starredEntryIDs, true)
		if err != nil {
			json.ServerError(w, r, err)
			return
		}
	}

	if len(entries) > 0 {
		settings, err := h.store.Integration(userID)
		if err != nil {
			json.ServerError(w, r, err)
			return
		}

		for _, entry := range entries {
			e := entry
			go func() {
				integration.SendEntry(e, settings)
			}()
		}
	}

	sendOkayResponse(w)
}

func (h *handler) quickAddHandler(w http.ResponseWriter, r *http.Request) {
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
		json.BadRequest(w, r, err)
		return
	}

	feedURL := r.Form.Get(paramQuickAdd)
	if !validator.IsValidURL(feedURL) {
		json.BadRequest(w, r, fmt.Errorf("googlereader: invalid URL: %s", feedURL))
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
		json.ServerError(w, r, localizedError.Error())
		return
	}

	if len(subscriptions) == 0 {
		json.OK(w, r, quickAddResponse{
			NumResults: 0,
		})
		return
	}

	toSubscribe := Stream{FeedStream, subscriptions[0].URL}
	category := Stream{NoStream, ""}
	newFeed, err := subscribe(toSubscribe, category, "", h.store, userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	slog.Debug("[GoogleReader] Added a new feed",
		slog.String("handler", "quickAddHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", userID),
		slog.String("feed_url", newFeed.FeedURL),
	)

	json.OK(w, r, quickAddResponse{
		NumResults: 1,
		Query:      newFeed.FeedURL,
		StreamID:   fmt.Sprintf(feedPrefix+"%d", newFeed.ID),
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

func (h *handler) feedIconURL(f *model.Feed) string {
	if f.Icon != nil && f.Icon.ExternalIconID != "" {
		return config.Opts.RootURL() + route.Path(h.router, "feedIcon", "externalIconID", f.Icon.ExternalIconID)
	} else {
		return ""
	}
}

func (h *handler) editSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /subscription/edit",
		slog.String("handler", "editSubscriptionHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", userID),
	)

	if err := r.ParseForm(); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	streamIds, err := getStreams(r.Form[paramStreamID], userID)
	if err != nil || len(streamIds) == 0 {
		json.BadRequest(w, r, errors.New("googlereader: no valid stream IDs provided"))
		return
	}

	newLabel, err := getStream(r.Form.Get(paramTagsAdd), userID)
	if err != nil {
		json.BadRequest(w, r, fmt.Errorf("googlereader: invalid data in %s", paramTagsAdd))
		return
	}

	title := r.Form.Get(paramTitle)
	action := r.Form.Get(paramSubscribeAction)

	switch action {
	case "subscribe":
		_, err := subscribe(streamIds[0], newLabel, title, h.store, userID)
		if err != nil {
			json.ServerError(w, r, err)
			return
		}
	case "unsubscribe":
		err := unsubscribe(streamIds, h.store, userID)
		if err != nil {
			json.ServerError(w, r, err)
			return
		}
	case "edit":
		if title != "" {
			if err := rename(streamIds[0], title, h.store, userID); err != nil {
				if errors.Is(err, errFeedNotFound) || errors.Is(err, errEmptyFeedTitle) {
					json.BadRequest(w, r, err)
				} else {
					json.ServerError(w, r, err)
				}
				return
			}
		}

		if r.Form.Has(paramTagsAdd) {
			if newLabel.Type != LabelStream {
				json.BadRequest(w, r, errors.New("destination must be a label"))
				return
			}

			if err := move(streamIds[0], newLabel, h.store, userID); err != nil {
				if errors.Is(err, errFeedNotFound) || errors.Is(err, errCategoryNotFound) {
					json.BadRequest(w, r, err)
				} else {
					json.ServerError(w, r, err)
				}
				return
			}
		}
	default:
		json.BadRequest(w, r, fmt.Errorf("googlereader: unrecognized action %s", action))
		return
	}

	sendOkayResponse(w)
}

func (h *handler) streamItemContentsHandler(w http.ResponseWriter, r *http.Request) {
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
		json.BadRequest(w, r, err)
		return
	}

	err := r.ParseForm()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	requestModifiers, err := parseStreamFilterFromRequest(r)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	userReadingList := fmt.Sprintf(userStreamPrefix, userID) + readingListStreamSuffix
	userRead := fmt.Sprintf(userStreamPrefix, userID) + readStreamSuffix
	userStarred := fmt.Sprintf(userStreamPrefix, userID) + starredStreamSuffix

	itemIDs, err := parseItemIDsFromRequest(r)
	if err != nil {
		json.BadRequest(w, r, err)
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
	builder.WithoutStatus(model.EntryStatusRemoved)
	builder.WithEntryIDs(itemIDs)
	builder.WithSorting(model.DefaultSortingOrder, requestModifiers.SortDirection)

	entries, err := builder.GetEntries()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	result := streamContentItemsResponse{
		Direction: "ltr",
		ID:        "user/-/state/com.google/reading-list",
		Title:     "Reading List",
		Updated:   time.Now().Unix(),
		Self: []contentHREF{
			{
				HREF: config.Opts.RootURL() + route.Path(h.router, "StreamItemsContents"),
			},
		},
		Author: userName,
	}
	contentItems := make([]contentItem, len(entries))
	for i, entry := range entries {
		enclosures := make([]contentItemEnclosure, 0, len(entry.Enclosures))
		for _, enclosure := range entry.Enclosures {
			enclosures = append(enclosures, contentItemEnclosure{URL: enclosure.URL, Type: enclosure.MimeType})
		}
		categories := make([]string, 0)
		categories = append(categories, userReadingList)
		if entry.Feed.Category.Title != "" {
			categories = append(categories, fmt.Sprintf(userLabelPrefix, userID)+entry.Feed.Category.Title)
		}
		if entry.Status == model.EntryStatusRead {
			categories = append(categories, userRead)
		}

		if entry.Starred {
			categories = append(categories, userStarred)
		}

		entry.Content = mediaproxy.RewriteDocumentWithAbsoluteProxyURL(h.router, entry.Content)
		entry.Enclosures.ProxifyEnclosureURL(h.router, config.Opts.MediaProxyMode(), config.Opts.MediaProxyResourceTypes())

		contentItems[i] = contentItem{
			ID:            convertEntryIDToLongFormItemID(entry.ID),
			Title:         entry.Title,
			Author:        entry.Author,
			TimestampUsec: fmt.Sprintf("%d", entry.Date.UnixMicro()),
			CrawlTimeMsec: fmt.Sprintf("%d", entry.CreatedAt.UnixMilli()),
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
				StreamID: fmt.Sprintf("feed/%d", entry.FeedID),
				Title:    entry.Feed.Title,
				HTMLUrl:  entry.Feed.SiteURL,
			},
			Enclosure: enclosures,
		}
	}
	result.Items = contentItems
	json.OK(w, r, result)
}

func (h *handler) disableTagHandler(w http.ResponseWriter, r *http.Request) {
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
		json.BadRequest(w, r, err)
		return
	}

	streams, err := getStreams(r.Form[paramStreamID], userID)
	if err != nil {
		json.BadRequest(w, r, fmt.Errorf("googlereader: invalid data in %s", paramStreamID))
		return
	}

	titles := make([]string, len(streams))
	for i, stream := range streams {
		if stream.Type != LabelStream {
			json.BadRequest(w, r, errors.New("googlereader: only labels are supported"))
			return
		}
		titles[i] = stream.ID
	}

	err = h.store.RemoveAndReplaceCategoriesByName(userID, titles)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	sendOkayResponse(w)
}

func (h *handler) renameTagHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /rename-tag",
		slog.String("handler", "renameTagHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
	)

	err := r.ParseForm()
	if err != nil {
		json.BadRequest(w, r, err)
		return
	}

	source, err := getStream(r.Form.Get(paramStreamID), userID)
	if err != nil {
		json.BadRequest(w, r, fmt.Errorf("googlereader: invalid data in %s", paramStreamID))
		return
	}

	destination, err := getStream(r.Form.Get(paramDestination), userID)
	if err != nil {
		json.BadRequest(w, r, fmt.Errorf("googlereader: invalid data in %s", paramDestination))
		return
	}

	if source.Type != LabelStream || destination.Type != LabelStream {
		json.BadRequest(w, r, errors.New("googlereader: only labels supported"))
		return
	}

	if destination.ID == "" {
		json.BadRequest(w, r, errors.New("googlereader: empty destination name"))
		return
	}

	category, err := h.store.CategoryByTitle(userID, source.ID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	if category == nil {
		json.NotFound(w, r)
		return
	}

	categoryModificationRequest := model.CategoryModificationRequest{
		Title: model.SetOptionalField(destination.ID),
	}

	if validationError := validator.ValidateCategoryModification(h.store, userID, category.ID, &categoryModificationRequest); validationError != nil {
		json.BadRequest(w, r, validationError.Error())
		return
	}

	categoryModificationRequest.Patch(category)

	if err := h.store.UpdateCategory(category); err != nil {
		json.ServerError(w, r, err)
		return
	}

	sendOkayResponse(w)
}

func (h *handler) tagListHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /tags/list",
		slog.String("handler", "tagListHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
	)

	if err := checkOutputFormat(r); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	var result tagsResponse
	categories, err := h.store.Categories(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	result.Tags = make([]subscriptionCategoryResponse, 0)
	result.Tags = append(result.Tags, subscriptionCategoryResponse{
		ID: fmt.Sprintf(userStreamPrefix, userID) + starredStreamSuffix,
	})
	for _, category := range categories {
		result.Tags = append(result.Tags, subscriptionCategoryResponse{
			ID:    fmt.Sprintf(userLabelPrefix, userID) + category.Title,
			Label: category.Title,
			Type:  "folder",
		})
	}
	json.OK(w, r, result)
}

func (h *handler) subscriptionListHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /subscription/list",
		slog.String("handler", "subscriptionListHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
	)

	if err := checkOutputFormat(r); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	var result subscriptionsResponse
	feeds, err := h.store.Feeds(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	result.Subscriptions = make([]subscriptionResponse, 0)
	for _, feed := range feeds {
		result.Subscriptions = append(result.Subscriptions, subscriptionResponse{
			ID:         fmt.Sprintf(feedPrefix+"%d", feed.ID),
			Title:      feed.Title,
			URL:        feed.FeedURL,
			Categories: []subscriptionCategoryResponse{{fmt.Sprintf(userLabelPrefix, userID) + feed.Category.Title, feed.Category.Title, "folder"}},
			HTMLURL:    feed.SiteURL,
			IconURL:    h.feedIconURL(feed),
		})
	}
	json.OK(w, r, result)
}

func (h *handler) serveHandler(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] API endpoint not implemented yet",
		slog.Any("url", r.RequestURI),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
	)

	json.OK(w, r, []string{})
}

func (h *handler) userInfoHandler(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /user-info",
		slog.String("handler", "userInfoHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
	)

	if err := checkOutputFormat(r); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	userInfo := userInfoResponse{UserID: fmt.Sprint(user.ID), UserName: user.Username, UserProfileID: fmt.Sprint(user.ID), UserEmail: user.Username}
	json.OK(w, r, userInfo)
}

func (h *handler) streamItemIDsHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /stream/items/ids",
		slog.String("handler", "streamItemIDsHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", userID),
	)

	if err := checkOutputFormat(r); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	rm, err := parseStreamFilterFromRequest(r)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	slog.Debug("[GoogleReader] Request Modifiers",
		slog.String("handler", "streamItemIDsHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
		slog.Any("modifiers", rm),
	)

	if len(rm.Streams) != 1 {
		json.ServerError(w, r, fmt.Errorf("googlereader: only one stream type expected"))
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
		json.ServerError(w, r, fmt.Errorf("googlereader: unknown stream type %s", rm.Streams[0].Type))
	}
}

func (h *handler) handleReadingListStreamHandler(w http.ResponseWriter, r *http.Request, rm RequestModifiers) {
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

	builder.WithoutStatus(model.EntryStatusRemoved)
	builder.WithLimit(rm.Count)
	builder.WithOffset(rm.Offset)
	builder.WithSorting(model.DefaultSortingOrder, rm.SortDirection)
	if rm.StartTime > 0 {
		builder.AfterPublishedDate(time.Unix(rm.StartTime, 0))
	}
	if rm.StopTime > 0 {
		builder.BeforePublishedDate(time.Unix(rm.StopTime, 0))
	}

	rawEntryIDs, err := builder.GetEntryIDs()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	var itemRefs = make([]itemRef, 0)
	for _, entryID := range rawEntryIDs {
		formattedID := strconv.FormatInt(entryID, 10)
		itemRefs = append(itemRefs, itemRef{ID: formattedID})
	}

	totalEntries, err := builder.CountEntries()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	continuation := 0
	if len(itemRefs)+rm.Offset < totalEntries {
		continuation = len(itemRefs) + rm.Offset
	}

	json.OK(w, r, streamIDResponse{itemRefs, continuation})
}

func (h *handler) handleStarredStreamHandler(w http.ResponseWriter, r *http.Request, rm RequestModifiers) {
	builder := h.store.NewEntryQueryBuilder(rm.UserID)
	builder.WithoutStatus(model.EntryStatusRemoved)
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

	rawEntryIDs, err := builder.GetEntryIDs()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	var itemRefs = make([]itemRef, 0)
	for _, entryID := range rawEntryIDs {
		formattedID := strconv.FormatInt(entryID, 10)
		itemRefs = append(itemRefs, itemRef{ID: formattedID})
	}

	totalEntries, err := builder.CountEntries()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	continuation := 0
	if len(itemRefs)+rm.Offset < totalEntries {
		continuation = len(itemRefs) + rm.Offset
	}

	json.OK(w, r, streamIDResponse{itemRefs, continuation})
}

func (h *handler) handleReadStreamHandler(w http.ResponseWriter, r *http.Request, rm RequestModifiers) {
	builder := h.store.NewEntryQueryBuilder(rm.UserID)
	builder.WithoutStatus(model.EntryStatusRemoved)
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

	rawEntryIDs, err := builder.GetEntryIDs()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	var itemRefs = make([]itemRef, 0)
	for _, entryID := range rawEntryIDs {
		formattedID := strconv.FormatInt(entryID, 10)
		itemRefs = append(itemRefs, itemRef{ID: formattedID})
	}

	totalEntries, err := builder.CountEntries()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	continuation := 0
	if len(itemRefs)+rm.Offset < totalEntries {
		continuation = len(itemRefs) + rm.Offset
	}

	json.OK(w, r, streamIDResponse{itemRefs, continuation})
}

func (h *handler) handleFeedStreamHandler(w http.ResponseWriter, r *http.Request, rm RequestModifiers) {
	feedID, err := strconv.ParseInt(rm.Streams[0].ID, 10, 64)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	builder := h.store.NewEntryQueryBuilder(rm.UserID)
	builder.WithoutStatus(model.EntryStatusRemoved)
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

	rawEntryIDs, err := builder.GetEntryIDs()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	var itemRefs = make([]itemRef, 0)
	for _, entryID := range rawEntryIDs {
		formattedID := strconv.FormatInt(entryID, 10)
		itemRefs = append(itemRefs, itemRef{ID: formattedID})
	}

	totalEntries, err := builder.CountEntries()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	continuation := 0
	if len(itemRefs)+rm.Offset < totalEntries {
		continuation = len(itemRefs) + rm.Offset
	}

	json.OK(w, r, streamIDResponse{itemRefs, continuation})
}

func (h *handler) markAllAsReadHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	slog.Debug("[GoogleReader] Handle /mark-all-as-read",
		slog.String("handler", "markAllAsReadHandler"),
		slog.String("client_ip", clientIP),
		slog.String("user_agent", r.UserAgent()),
	)

	if err := r.ParseForm(); err != nil {
		json.BadRequest(w, r, err)
		return
	}

	stream, err := getStream(r.Form.Get(paramStreamID), userID)
	if err != nil {
		json.BadRequest(w, r, err)
		return
	}

	var before time.Time
	if timestampParamValue := r.Form.Get(paramTimestamp); timestampParamValue != "" {
		timestampParsedValue, err := strconv.ParseInt(timestampParamValue, 10, 64)
		if err != nil {
			json.BadRequest(w, r, err)
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
			json.BadRequest(w, r, err)
			return
		}
		err = h.store.MarkFeedAsRead(userID, feedID, before)
		if err != nil {
			json.ServerError(w, r, err)
			return
		}
	case LabelStream:
		category, err := h.store.CategoryByTitle(userID, stream.ID)
		if err != nil {
			json.ServerError(w, r, err)
			return
		}
		if category == nil {
			json.NotFound(w, r)
			return
		}
		if err := h.store.MarkCategoryAsRead(userID, category.ID, before); err != nil {
			json.ServerError(w, r, err)
			return
		}
	case ReadingListStream:
		if err = h.store.MarkAllAsReadBeforeDate(userID, before); err != nil {
			json.ServerError(w, r, err)
			return
		}
	}

	sendOkayResponse(w)
}
