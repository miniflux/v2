// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/http/response/json"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/integration"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/proxy"
	"miniflux.app/v2/internal/reader/fetcher"
	mff "miniflux.app/v2/internal/reader/handler"
	mfs "miniflux.app/v2/internal/reader/subscription"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/urllib"
	"miniflux.app/v2/internal/validator"

	"github.com/gorilla/mux"
)

type handler struct {
	store  *storage.Storage
	router *mux.Router
}

const (
	// StreamPrefix is the prefix for astreams (read/starred/reading list and so on)
	StreamPrefix = "user/-/state/com.google/"
	// UserStreamPrefix is the user specific prefix for streams (read/starred/reading list and so on)
	UserStreamPrefix = "user/%d/state/com.google/"
	// LabelPrefix is the prefix for a label stream
	LabelPrefix = "user/-/label/"
	// UserLabelPrefix is the user specific prefix prefix for a label stream
	UserLabelPrefix = "user/%d/label/"
	// FeedPrefix is the prefix for a feed stream
	FeedPrefix = "feed/"
	// Read is the suffix for read stream
	Read = "read"
	// Starred is the suffix for starred stream
	Starred = "starred"
	// ReadingList is the suffix for reading list stream
	ReadingList = "reading-list"
	// KeptUnread is the suffix for kept unread stream
	KeptUnread = "kept-unread"
	// Broadcast is the suffix for broadcast stream
	Broadcast = "broadcast"
	// BroadcastFriends is the suffix for broadcast friends stream
	BroadcastFriends = "broadcast-friends"
	// Like is the suffix for like stream
	Like = "like"
	// EntryIDLong is the long entry id representation
	EntryIDLong = "tag:google.com,2005:reader/item/%016x"
)

const (
	// ParamItemIDs - name of the parameter with the item ids
	ParamItemIDs = "i"
	// ParamStreamID - name of the parameter containing the stream to be included
	ParamStreamID = "s"
	// ParamStreamExcludes - name of the parameter containing streams to be excluded
	ParamStreamExcludes = "xt"
	// ParamStreamFilters - name of the parameter containing streams to be included
	ParamStreamFilters = "it"
	// ParamStreamMaxItems - name of the parameter containing number of items per page/max items returned
	ParamStreamMaxItems = "n"
	// ParamStreamOrder - name of the parameter containing the sort criteria
	ParamStreamOrder = "r"
	// ParamStreamStartTime - name of the parameter containing epoch timestamp, filtering items older than
	ParamStreamStartTime = "ot"
	// ParamStreamStopTime - name of the parameter containing epoch timestamp, filtering items newer than
	ParamStreamStopTime = "nt"
	// ParamTagsRemove - name of the parameter containing tags (streams) to be removed
	ParamTagsRemove = "r"
	// ParamTagsAdd - name of the parameter containing tags (streams) to be added
	ParamTagsAdd = "a"
	// ParamSubscribeAction - name of the parameter indicating the action to take for subscription/edit
	ParamSubscribeAction = "ac"
	// ParamTitle - name of the parameter for the title of the subscription
	ParamTitle = "t"
	// ParamQuickAdd - name of the parameter for a URL being quick subscribed to
	ParamQuickAdd = "quickadd"
	// ParamDestination - name of the parameter for the new name of a tag
	ParamDestination = "dest"
	// ParamContinuation -  name of the parameter for callers to pass to receive the next page of results
	ParamContinuation = "c"
)

// StreamType represents the possible stream types
type StreamType int

const (
	// NoStream - no stream type
	NoStream StreamType = iota
	// ReadStream - read stream type
	ReadStream
	// StarredStream - starred stream type
	StarredStream
	// ReadingListStream - reading list stream type
	ReadingListStream
	// KeptUnreadStream - kept unread stream type
	KeptUnreadStream
	// BroadcastStream - broadcast stream type
	BroadcastStream
	// BroadcastFriendsStream - broadcast friends stream type
	BroadcastFriendsStream
	// LabelStream - label stream type
	LabelStream
	// FeedStream - feed stream type
	FeedStream
	// LikeStream - like stream type
	LikeStream
)

// Stream defines a stream type and its ID.
type Stream struct {
	Type StreamType
	ID   string
}

func (s Stream) String() string {
	return fmt.Sprintf("%v - '%s'", s.Type, s.ID)
}

func (st StreamType) String() string {
	switch st {
	case NoStream:
		return "NoStream"
	case ReadStream:
		return "ReadStream"
	case StarredStream:
		return "StarredStream"
	case ReadingListStream:
		return "ReadingListStream"
	case KeptUnreadStream:
		return "KeptUnreadStream"
	case BroadcastStream:
		return "BroadcastStream"
	case BroadcastFriendsStream:
		return "BroadcastFriendsStream"
	case LabelStream:
		return "LabelStream"
	case FeedStream:
		return "FeedStream"
	case LikeStream:
		return "LikeStream"
	default:
		return st.String()
	}
}

// RequestModifiers are the parsed request parameters.
type RequestModifiers struct {
	ExcludeTargets    []Stream
	FilterTargets     []Stream
	Streams           []Stream
	Count             int
	Offset            int
	SortDirection     string
	StartTime         int64
	StopTime          int64
	ContinuationToken string
	UserID            int64
}

func (r RequestModifiers) String() string {
	var results []string

	results = append(results, fmt.Sprintf("UserID: %d", r.UserID))

	var streamStr []string
	for _, s := range r.Streams {
		streamStr = append(streamStr, s.String())
	}
	results = append(results, fmt.Sprintf("Streams: [%s]", strings.Join(streamStr, ", ")))

	var exclusions []string
	for _, s := range r.ExcludeTargets {
		exclusions = append(exclusions, s.String())
	}
	results = append(results, fmt.Sprintf("Exclusions: [%s]", strings.Join(exclusions, ", ")))

	var filters []string
	for _, s := range r.FilterTargets {
		filters = append(filters, s.String())
	}
	results = append(results, fmt.Sprintf("Filters: [%s]", strings.Join(filters, ", ")))

	results = append(results, fmt.Sprintf("Count: %d", r.Count))
	results = append(results, fmt.Sprintf("Offset: %d", r.Offset))
	results = append(results, fmt.Sprintf("Sort Direction: %s", r.SortDirection))
	results = append(results, fmt.Sprintf("Continuation Token: %s", r.ContinuationToken))
	results = append(results, fmt.Sprintf("Start Time: %d", r.StartTime))
	results = append(results, fmt.Sprintf("Stop Time: %d", r.StopTime))

	return strings.Join(results, "; ")
}

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
	sr.PathPrefix("/").HandlerFunc(handler.serveHandler).Methods(http.MethodPost, http.MethodGet).Name("GoogleReaderApiEndpoint")
}

func getStreamFilterModifiers(r *http.Request) (RequestModifiers, error) {
	userID := request.UserID(r)

	result := RequestModifiers{
		SortDirection: "desc",
		UserID:        userID,
	}
	streamOrder := request.QueryStringParam(r, ParamStreamOrder, "d")
	if streamOrder == "o" {
		result.SortDirection = "asc"
	}
	var err error
	result.Streams, err = getStreams(request.QueryStringParamList(r, ParamStreamID), userID)
	if err != nil {
		return RequestModifiers{}, err
	}
	result.ExcludeTargets, err = getStreams(request.QueryStringParamList(r, ParamStreamExcludes), userID)
	if err != nil {
		return RequestModifiers{}, err
	}

	result.FilterTargets, err = getStreams(request.QueryStringParamList(r, ParamStreamFilters), userID)
	if err != nil {
		return RequestModifiers{}, err
	}

	result.Count = request.QueryIntParam(r, ParamStreamMaxItems, 0)
	result.Offset = request.QueryIntParam(r, ParamContinuation, 0)
	result.StartTime = request.QueryInt64Param(r, ParamStreamStartTime, int64(0))
	result.StopTime = request.QueryInt64Param(r, ParamStreamStopTime, int64(0))
	return result, nil
}

func getStream(streamID string, userID int64) (Stream, error) {
	switch {
	case strings.HasPrefix(streamID, FeedPrefix):
		return Stream{Type: FeedStream, ID: strings.TrimPrefix(streamID, FeedPrefix)}, nil
	case strings.HasPrefix(streamID, fmt.Sprintf(UserStreamPrefix, userID)) || strings.HasPrefix(streamID, StreamPrefix):
		id := strings.TrimPrefix(streamID, fmt.Sprintf(UserStreamPrefix, userID))
		id = strings.TrimPrefix(id, StreamPrefix)
		switch id {
		case Read:
			return Stream{ReadStream, ""}, nil
		case Starred:
			return Stream{StarredStream, ""}, nil
		case ReadingList:
			return Stream{ReadingListStream, ""}, nil
		case KeptUnread:
			return Stream{KeptUnreadStream, ""}, nil
		case Broadcast:
			return Stream{BroadcastStream, ""}, nil
		case BroadcastFriends:
			return Stream{BroadcastFriendsStream, ""}, nil
		case Like:
			return Stream{LikeStream, ""}, nil
		default:
			return Stream{NoStream, ""}, fmt.Errorf("googlereader: unknown stream with id: %s", id)
		}
	case strings.HasPrefix(streamID, fmt.Sprintf(UserLabelPrefix, userID)) || strings.HasPrefix(streamID, LabelPrefix):
		id := strings.TrimPrefix(streamID, fmt.Sprintf(UserLabelPrefix, userID))
		id = strings.TrimPrefix(id, LabelPrefix)
		return Stream{LabelStream, id}, nil
	case streamID == "":
		return Stream{NoStream, ""}, nil
	default:
		return Stream{NoStream, ""}, fmt.Errorf("googlereader: unknown stream type: %s", streamID)
	}
}

func getStreams(streamIDs []string, userID int64) ([]Stream, error) {
	streams := make([]Stream, 0)
	for _, streamID := range streamIDs {
		stream, err := getStream(streamID, userID)
		if err != nil {
			return []Stream{}, err
		}
		streams = append(streams, stream)
	}
	return streams, nil
}

func checkAndSimplifyTags(addTags []Stream, removeTags []Stream) (map[StreamType]bool, error) {
	tags := make(map[StreamType]bool)
	for _, s := range addTags {
		switch s.Type {
		case ReadStream:
			if _, ok := tags[KeptUnreadStream]; ok {
				return nil, fmt.Errorf("googlereader: %s ad %s should not be supplied simultaneously", KeptUnread, Read)
			}
			tags[ReadStream] = true
		case KeptUnreadStream:
			if _, ok := tags[ReadStream]; ok {
				return nil, fmt.Errorf("googlereader: %s ad %s should not be supplied simultaneously", KeptUnread, Read)
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
				return nil, fmt.Errorf("googlereader: %s ad %s should not be supplied simultaneously", KeptUnread, Read)
			}
			tags[ReadStream] = false
		case KeptUnreadStream:
			if _, ok := tags[ReadStream]; ok {
				return nil, fmt.Errorf("googlereader: %s ad %s should not be supplied simultaneously", KeptUnread, Read)
			}
			tags[ReadStream] = true
		case StarredStream:
			if _, ok := tags[StarredStream]; ok {
				return nil, fmt.Errorf("googlereader: %s should not be supplied for add and remove simultaneously", Starred)
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

func getItemIDs(r *http.Request) ([]int64, error) {
	items := r.Form[ParamItemIDs]
	if len(items) == 0 {
		return nil, fmt.Errorf("googlereader: no items requested")
	}

	itemIDs := make([]int64, len(items))

	for i, item := range items {
		var itemID int64
		_, err := fmt.Sscanf(item, EntryIDLong, &itemID)
		if err != nil {
			itemID, err = strconv.ParseInt(item, 16, 64)
			if err != nil {
				return nil, fmt.Errorf("googlereader: could not parse item: %v", item)
			}
		}
		itemIDs[i] = itemID
	}
	return itemIDs, nil
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

	result := login{SID: token, LSID: token, Auth: token}
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

	addTags, err := getStreams(r.PostForm[ParamTagsAdd], userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	removeTags, err := getStreams(r.PostForm[ParamTagsRemove], userID)
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

	itemIDs, err := getItemIDs(r)
	if err != nil {
		json.ServerError(w, r, err)
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

	OK(w, r)
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

	feedURL := r.Form.Get(ParamQuickAdd)
	if !validator.IsValidURL(feedURL) {
		json.BadRequest(w, r, fmt.Errorf("googlereader: invalid URL: %s", feedURL))
		return
	}

	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxy(config.Opts.HTTPClientProxy())

	var rssBridgeURL string
	if intg, err := h.store.Integration(userID); err == nil && intg != nil && intg.RSSBridgeEnabled {
		rssBridgeURL = intg.RSSBridgeURL
	}

	subscriptions, localizedError := mfs.NewSubscriptionFinder(requestBuilder).FindSubscriptions(feedURL, rssBridgeURL)
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
		StreamID:   fmt.Sprintf(FeedPrefix+"%d", newFeed.ID),
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

func getOrCreateCategory(category Stream, store *storage.Storage, userID int64) (*model.Category, error) {
	switch {
	case category.ID == "":
		return store.FirstCategory(userID)
	case store.CategoryTitleExists(userID, category.ID):
		return store.CategoryByTitle(userID, category.ID)
	default:
		catRequest := model.CategoryRequest{
			Title: category.ID,
		}
		return store.CreateCategory(userID, &catRequest)
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
	if err != nil {
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

func rename(stream Stream, title string, store *storage.Storage, userID int64) error {
	if title == "" {
		return errors.New("empty title")
	}
	feed, err := getFeed(stream, store, userID)
	if err != nil {
		return err
	}
	feedModification := model.FeedModificationRequest{
		Title: &title,
	}
	feedModification.Patch(feed)
	return store.UpdateFeed(feed)
}

func move(stream Stream, destination Stream, store *storage.Storage, userID int64) error {
	feed, err := getFeed(stream, store, userID)
	if err != nil {
		return err
	}
	category, err := getOrCreateCategory(destination, store, userID)
	if err != nil {
		return err
	}
	feedModification := model.FeedModificationRequest{
		CategoryID: &category.ID,
	}
	feedModification.Patch(feed)
	return store.UpdateFeed(feed)
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

	streamIds, err := getStreams(r.Form[ParamStreamID], userID)
	if err != nil || len(streamIds) == 0 {
		json.BadRequest(w, r, errors.New("googlereader: no valid stream IDs provided"))
		return
	}

	newLabel, err := getStream(r.Form.Get(ParamTagsAdd), userID)
	if err != nil {
		json.BadRequest(w, r, fmt.Errorf("googlereader: invalid data in %s", ParamTagsAdd))
		return
	}

	title := r.Form.Get(ParamTitle)
	action := r.Form.Get(ParamSubscribeAction)

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
				json.ServerError(w, r, err)
				return
			}
		}

		if r.Form.Has(ParamTagsAdd) {
			if newLabel.Type != LabelStream {
				json.BadRequest(w, r, errors.New("destination must be a label"))
				return
			}

			if err := move(streamIds[0], newLabel, h.store, userID); err != nil {
				json.ServerError(w, r, err)
				return
			}
		}
	default:
		json.ServerError(w, r, fmt.Errorf("googlereader: unrecognized action %s", action))
		return
	}

	OK(w, r)
}

func (h *handler) streamItemContentsHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
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
	var user *model.User
	if user, err = h.store.UserByID(userID); err != nil {
		json.ServerError(w, r, err)
		return
	}

	requestModifiers, err := getStreamFilterModifiers(r)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	userReadingList := fmt.Sprintf(UserStreamPrefix, userID) + ReadingList
	userRead := fmt.Sprintf(UserStreamPrefix, userID) + Read
	userStarred := fmt.Sprintf(UserStreamPrefix, userID) + Starred

	itemIDs, err := getItemIDs(r)
	if err != nil {
		json.ServerError(w, r, err)
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
	builder.WithoutStatus(model.EntryStatusRemoved)
	builder.WithEntryIDs(itemIDs)
	builder.WithSorting(model.DefaultSortingOrder, requestModifiers.SortDirection)

	entries, err := builder.GetEntries()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	if len(entries) == 0 {
		json.BadRequest(w, r, fmt.Errorf("googlereader: no items returned from the database for item IDs: %v", itemIDs))
		return
	}

	result := streamContentItems{
		Direction: "ltr",
		ID:        fmt.Sprintf("feed/%d", entries[0].FeedID),
		Title:     entries[0].Feed.Title,
		Alternate: []contentHREFType{
			{
				HREF: entries[0].Feed.SiteURL,
				Type: "text/html",
			},
		},
		Updated: time.Now().Unix(),
		Self: []contentHREF{
			{
				HREF: config.Opts.RootURL() + route.Path(h.router, "StreamItemsContents"),
			},
		},
		Author: user.Username,
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
			categories = append(categories, fmt.Sprintf(UserLabelPrefix, userID)+entry.Feed.Category.Title)
		}
		if entry.Status == model.EntryStatusRead {
			categories = append(categories, userRead)
		}

		if entry.Starred {
			categories = append(categories, userStarred)
		}

		entry.Content = proxy.AbsoluteProxyRewriter(h.router, r.Host, entry.Content)
		proxyOption := config.Opts.ProxyOption()

		for i := range entry.Enclosures {
			if proxyOption == "all" || proxyOption != "none" && !urllib.IsHTTPS(entry.Enclosures[i].URL) {
				for _, mediaType := range config.Opts.ProxyMediaTypes() {
					if strings.HasPrefix(entry.Enclosures[i].MimeType, mediaType+"/") {
						entry.Enclosures[i].URL = proxy.AbsoluteProxifyURL(h.router, r.Host, entry.Enclosures[i].URL)
						break
					}
				}
			}
		}

		contentItems[i] = contentItem{
			ID:            fmt.Sprintf(EntryIDLong, entry.ID),
			Title:         entry.Title,
			Author:        entry.Author,
			TimestampUsec: fmt.Sprintf("%d", entry.Date.UnixNano()/(int64(time.Microsecond)/int64(time.Nanosecond))),
			CrawlTimeMsec: fmt.Sprintf("%d", entry.Date.UnixNano()/(int64(time.Microsecond)/int64(time.Nanosecond))),
			Published:     entry.Date.Unix(),
			Updated:       entry.Date.Unix(),
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

	streams, err := getStreams(r.Form[ParamStreamID], userID)
	if err != nil {
		json.BadRequest(w, r, fmt.Errorf("googlereader: invalid data in %s", ParamStreamID))
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

	OK(w, r)
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

	source, err := getStream(r.Form.Get(ParamStreamID), userID)
	if err != nil {
		json.BadRequest(w, r, fmt.Errorf("googlereader: invalid data in %s", ParamStreamID))
		return
	}

	destination, err := getStream(r.Form.Get(ParamDestination), userID)
	if err != nil {
		json.BadRequest(w, r, fmt.Errorf("googlereader: invalid data in %s", ParamDestination))
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
	categoryRequest := model.CategoryRequest{
		Title: destination.ID,
	}
	verr := validator.ValidateCategoryModification(h.store, userID, category.ID, &categoryRequest)
	if verr != nil {
		json.BadRequest(w, r, verr.Error())
		return
	}
	categoryRequest.Patch(category)
	err = h.store.UpdateCategory(category)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	OK(w, r)
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
	result.Tags = make([]subscriptionCategory, 0)
	result.Tags = append(result.Tags, subscriptionCategory{
		ID: fmt.Sprintf(UserStreamPrefix, userID) + Starred,
	})
	for _, category := range categories {
		result.Tags = append(result.Tags, subscriptionCategory{
			ID:    fmt.Sprintf(UserLabelPrefix, userID) + category.Title,
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
	result.Subscriptions = make([]subscription, 0)
	for _, feed := range feeds {
		result.Subscriptions = append(result.Subscriptions, subscription{
			ID:         fmt.Sprintf(FeedPrefix+"%d", feed.ID),
			Title:      feed.Title,
			URL:        feed.FeedURL,
			Categories: []subscriptionCategory{{fmt.Sprintf(UserLabelPrefix, userID) + feed.Category.Title, feed.Category.Title, "folder"}},
			HTMLURL:    feed.SiteURL,
			IconURL:    "", // TODO: Icons are base64 encoded in the DB.
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
	userInfo := userInfo{UserID: fmt.Sprint(user.ID), UserName: user.Username, UserProfileID: fmt.Sprint(user.ID), UserEmail: user.Username}
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

	rm, err := getStreamFilterModifiers(r)
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
				slog.Any("filter_type", s.Type),
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
