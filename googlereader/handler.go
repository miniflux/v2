package googlereader // import "miniflux.app/googlereader"

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"miniflux.app/config"
	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/http/route"
	"miniflux.app/integration"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/storage"
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

// Stream defines a stream type and its id
type Stream struct {
	Type StreamType
	ID   string
}

// RequestModifiers are the parsed request parameters
type RequestModifiers struct {
	ExcludeTargets    []Stream
	FilterTargets     []Stream
	Streams           []Stream
	Count             int
	SortDirection     string
	StartTime         int64
	StopTime          int64
	ContinuationToken string
	UserID            int64
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
func (s Stream) String() string {
	return fmt.Sprintf("%v - '%s'", s.Type, s.ID)
}
func (r RequestModifiers) String() string {
	result := fmt.Sprintf("UserID: %d\n", r.UserID)
	result += fmt.Sprintf("Streams: %d\n", len(r.Streams))
	for _, s := range r.Streams {
		result += fmt.Sprintf("         %v\n", s)
	}

	result += fmt.Sprintf("Exclusions: %d\n", len(r.ExcludeTargets))
	for _, s := range r.ExcludeTargets {
		result += fmt.Sprintf("            %v\n", s)
	}

	result += fmt.Sprintf("Filter: %d\n", len(r.FilterTargets))
	for _, s := range r.FilterTargets {
		result += fmt.Sprintf("        %v\n", s)
	}
	result += fmt.Sprintf("Count: %d\n", r.Count)
	result += fmt.Sprintf("Sort Direction: %s\n", r.SortDirection)
	result += fmt.Sprintf("Continuation Token: %s\n", r.ContinuationToken)
	result += fmt.Sprintf("Start Time: %d\n", r.StartTime)
	result += fmt.Sprintf("Stop Time: %d\n", r.StopTime)

	return result
}

// Serve handles Google Reader API calls.
func Serve(router *mux.Router, store *storage.Storage) {
	handler := &handler{store, router}
	middleware := newMiddleware(store)
	router.HandleFunc("/accounts/ClientLogin", middleware.clientLogin).Methods(http.MethodPost).Name("ClientLogin")
	sr := router.PathPrefix("/reader/api/0").Subrouter()
	sr.Use(middleware.handleCORS)
	sr.Use(middleware.apiKeyAuth)
	sr.Methods(http.MethodOptions)
	sr.HandleFunc("/token", middleware.token).Methods(http.MethodGet).Name("Token")
	sr.HandleFunc("/edit-tag", handler.editTag).Methods(http.MethodPost).Name("EditTag")
	sr.HandleFunc("/user-info", handler.userInfo).Methods(http.MethodGet).Name("UserInfo")
	sr.HandleFunc("/subscription/list", handler.subscriptionList).Methods(http.MethodGet).Name("SubscriptonList")
	sr.HandleFunc("/stream/items/ids", handler.streamItemIDs).Methods(http.MethodGet).Name("StreamItemIDs")
	sr.HandleFunc("/stream/items/contents", handler.streamItemContents).Methods(http.MethodPost).Name("StreamItemsContents")
	sr.HandleFunc("/tag/list", handler.tagList).Methods(http.MethodGet).Name("TagList")
	sr.PathPrefix("/").HandlerFunc(handler.serve).Methods(http.MethodPost, http.MethodGet).Name("GoogleReaderApiEndpoint")
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
	result.StartTime = int64(request.QueryIntParam(r, ParamStreamStartTime, 0))
	result.StopTime = int64(request.QueryIntParam(r, ParamStreamStopTime, 0))
	return result, nil
}

func getStream(streamID string, userID int64) (Stream, error) {
	if strings.HasPrefix(streamID, FeedPrefix) {
		return Stream{Type: FeedStream, ID: strings.TrimPrefix(streamID, FeedPrefix)}, nil
	} else if strings.HasPrefix(streamID, fmt.Sprintf(UserStreamPrefix, userID)) || strings.HasPrefix(streamID, StreamPrefix) {
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
			err := fmt.Errorf("uknown stream with id: %s", id)
			return Stream{NoStream, ""}, err
		}
	} else if strings.HasPrefix(streamID, fmt.Sprintf(UserLabelPrefix, userID)) || strings.HasPrefix(streamID, LabelPrefix) {
		id := strings.TrimPrefix(streamID, fmt.Sprintf(UserLabelPrefix, userID))
		id = strings.TrimPrefix(id, LabelPrefix)
		return Stream{LabelStream, id}, nil
	} else if streamID == "" {
		return Stream{NoStream, ""}, nil
	}
	err := fmt.Errorf("uknown stream type: %s", streamID)
	return Stream{NoStream, ""}, err
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
			if _, ok := tags[ReadStream]; ok {
				return nil, fmt.Errorf(KeptUnread + " and " + Read + " should not be supplied simultaneously")
			}
			tags[ReadStream] = true
		case KeptUnreadStream:
			if _, ok := tags[ReadStream]; ok {
				return nil, fmt.Errorf(KeptUnread + " and " + Read + " should not be supplied simultaneously")
			}
			tags[ReadStream] = false
		case StarredStream:
			tags[StarredStream] = true
		case BroadcastStream, LikeStream:
			logger.Info("Broadcast & Like tags are not implemented!")
		default:
			return nil, fmt.Errorf("unsupported tag type: %s", s.Type)
		}
	}
	for _, s := range removeTags {
		switch s.Type {
		case ReadStream:
			if _, ok := tags[ReadStream]; ok {
				return nil, fmt.Errorf(KeptUnread + " and " + Read + " should not be supplied simultaneously")
			}
			tags[ReadStream] = false
		case KeptUnreadStream:
			if _, ok := tags[ReadStream]; ok {
				return nil, fmt.Errorf(KeptUnread + " and " + Read + " should not be supplied simultaneously")
			}
			tags[ReadStream] = true
		case StarredStream:
			if _, ok := tags[StarredStream]; ok {
				return nil, fmt.Errorf(Starred + " should not be supplied for add and remove simultaneously")
			}
			tags[StarredStream] = false
		case BroadcastStream, LikeStream:
			logger.Info("Broadcast & Like tags are not implemented!")
		default:
			return nil, fmt.Errorf("unsupported tag type: %s", s.Type)
		}
	}

	return tags, nil
}

func getItemIDs(r *http.Request) ([]int64, error) {
	items := r.Form[ParamItemIDs]
	if len(items) == 0 {
		return nil, fmt.Errorf("no items requested")
	}

	itemIDs := make([]int64, len(items))

	for i, item := range items {
		var itemID int64
		_, err := fmt.Sscanf(item, EntryIDLong, &itemID)
		if err != nil {
			itemID, err = strconv.ParseInt(item, 16, 64)
			if err != nil {
				return nil, fmt.Errorf("could not parse item: %v", item)
			}
		}
		itemIDs[i] = itemID
	}
	return itemIDs, nil
}

func checkOutputFormat(w http.ResponseWriter, r *http.Request) error {
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
		err := fmt.Errorf("output only as json supported")
		return err
	}
	return nil
}

func (h *handler) editTag(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	logger.Info("[Reader][/edit-tag][ClientIP=%s] Incoming Request for userID  #%d", clientIP, userID)

	err := r.ParseForm()
	if err != nil {
		logger.Error("[Reader][/edit-tag] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}

	addTags, err := getStreams(r.PostForm[ParamTagsAdd], userID)
	if err != nil {
		logger.Error("[Reader][/edit-tag] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}
	removeTags, err := getStreams(r.PostForm[ParamTagsRemove], userID)
	if err != nil {
		logger.Error("[Reader][/edit-tag] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}
	if len(addTags) == 0 && len(removeTags) == 0 {
		err = fmt.Errorf("add or/and remove tags should be supllied")
		logger.Error("[Reader][/edit-tag] [ClientIP=%s] ", clientIP, err)
		json.ServerError(w, r, err)
		return
	}
	tags, err := checkAndSimplifyTags(addTags, removeTags)
	if err != nil {
		logger.Error("[Reader][/edit-tag] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}

	itemIDs, err := getItemIDs(r)
	if err != nil {
		logger.Error("[Reader][/edit-tag] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}

	logger.Debug("[Reader][/edit-tag] [ClientIP=%s] itemIDs: %v", clientIP, itemIDs)
	logger.Debug("[Reader][/edit-tag] [ClientIP=%s] tags: %v", clientIP, tags)
	builder := h.store.NewEntryQueryBuilder(userID)
	builder.WithEntryIDs(itemIDs)
	builder.WithoutStatus(model.EntryStatusRemoved)

	entries, err := builder.GetEntries()
	if err != nil {
		logger.Error("[Reader][/edit-tag] [ClientIP=%s] %v", clientIP, err)
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
			logger.Error("[Reader][/edit-tag] [ClientIP=%s] %v", clientIP, err)
			json.ServerError(w, r, err)
			return
		}
	}

	if len(unreadEntryIDs) > 0 {
		err = h.store.SetEntriesStatus(userID, unreadEntryIDs, model.EntryStatusUnread)
		if err != nil {
			logger.Error("[Reader][/edit-tag] [ClientIP=%s] %v", clientIP, err)
			json.ServerError(w, r, err)
			return
		}
	}

	if len(unstarredEntryIDs) > 0 {
		err = h.store.SetEntriesBookmarkedState(userID, unstarredEntryIDs, true)
		if err != nil {
			logger.Error("[Reader][/edit-tag] [ClientIP=%s] %v", clientIP, err)
			json.ServerError(w, r, err)
			return
		}
	}

	if len(starredEntryIDs) > 0 {
		err = h.store.SetEntriesBookmarkedState(userID, starredEntryIDs, true)
		if err != nil {
			logger.Error("[Reader][/edit-tag] [ClientIP=%s] %v", clientIP, err)
			json.ServerError(w, r, err)
			return
		}
	}

	if len(entries) > 0 {
		settings, err := h.store.Integration(userID)
		if err != nil {
			logger.Error("[Reader][/edit-tag] [ClientIP=%s] %v", clientIP, err)
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

func (h *handler) streamItemContents(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	logger.Info("[Reader][/stream/items/contents][ClientIP=%s] Incoming Request for userID  #%d", clientIP, userID)

	if err := checkOutputFormat(w, r); err != nil {
		logger.Error("[Reader][/stream/items/contents] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}

	err := r.ParseForm()
	if err != nil {
		logger.Error("[Reader][/stream/items/contents] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}
	var user *model.User
	if user, err = h.store.UserByID(userID); err != nil {
		logger.Error("[Reader][/stream/items/contents] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}

	requestModifiers, err := getStreamFilterModifiers(r)
	if err != nil {
		logger.Error("[Reader][/stream/items/contents] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}

	userReadingList := fmt.Sprintf(UserStreamPrefix, userID) + ReadingList
	userRead := fmt.Sprintf(UserStreamPrefix, userID) + Read
	userStarred := fmt.Sprintf(UserStreamPrefix, userID) + Starred

	itemIDs, err := getItemIDs(r)
	if err != nil {
		logger.Error("[Reader][/stream/items/contents] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}
	logger.Debug("[Reader][/stream/items/contents] [ClientIP=%s] itemIDs: %v", clientIP, itemIDs)

	builder := h.store.NewEntryQueryBuilder(userID)
	builder.WithoutStatus(model.EntryStatusRemoved)
	builder.WithEntryIDs(itemIDs)
	builder.WithOrder(model.DefaultSortingOrder)
	builder.WithDirection(requestModifiers.SortDirection)

	entries, err := builder.GetEntries()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	if len(entries) == 0 {
		err = fmt.Errorf("no items returned from the database")
		logger.Error("[Reader][/stream/items/contents] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
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
				HREF: config.Opts.BaseURL() + route.Path(h.router, "StreamItemsContents"),
			},
		},
		Author: user.Username,
	}
	contentItems := make([]contentItem, len(entries))
	for i, entry := range entries {
		enclosures := make([]contentItemEnclosure, len(entry.Enclosures))
		for _, enclosure := range entry.Enclosures {
			enclosures = append(enclosures, contentItemEnclosure{URL: enclosure.URL, Type: enclosure.MimeType})
		}
		categories := make([]string, 0)
		categories = append(categories, userReadingList)
		if entry.Feed.Category.Title != "" {
			categories = append(categories, fmt.Sprintf(UserLabelPrefix, userID)+entry.Feed.Category.Title)
		}
		if entry.Starred {
			categories = append(categories, userRead)
		}

		if entry.Starred {
			categories = append(categories, userStarred)
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

func (h *handler) tagList(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	logger.Info("[Reader][tags/list][ClientIP=%s] Incoming Request for userID  #%d", clientIP, userID)

	if err := checkOutputFormat(w, r); err != nil {
		logger.Error("[Reader][OutputFormat] %v", err)
		json.ServerError(w, r, err)
	}

	var result tagsResponse
	categories, err := h.store.Categories(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	result.Tags = make([]subscriptionCategory, 0)
	result.Tags = append(result.Tags, subscriptionCategory{
		ID: fmt.Sprintf("user/%d/state/com.google/starred", userID),
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

func (h *handler) subscriptionList(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	logger.Info("[Reader][/subscription/list][ClientIP=%s] Incoming Request for userID  #%d", clientIP, userID)

	if err := checkOutputFormat(w, r); err != nil {
		logger.Error("[Reader][/subscription/list] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
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
			IconURL:    "", //TODO Icons are only base64 encode in DB yet
		})
	}
	json.OK(w, r, result)
}

func (h *handler) serve(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)
	dump, _ := httputil.DumpRequest(r, true)
	logger.Info("[Reader][UNKNOWN] [ClientIP=%s] URL: %s", clientIP, dump)
	logger.Error("Call to Google Reader API not implemented yet!!")
	json.OK(w, r, []string{})
}

func (h *handler) userInfo(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)
	logger.Info("[Reader][UserInfo] [ClientIP=%s] Sending", clientIP)

	if err := checkOutputFormat(w, r); err != nil {
		logger.Error("[Reader][/user-info] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		logger.Error("[Reader][/user-info] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}
	userInfo := userInfo{UserID: fmt.Sprint(user.ID), UserName: user.Username, UserProfileID: fmt.Sprint(user.ID), UserEmail: user.Username}
	json.OK(w, r, userInfo)
}

func (h *handler) streamItemIDs(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	logger.Debug("[Reader][/stream/items/ids][ClientIP=%s] Incoming Request for userID  #%d", clientIP, userID)

	if err := checkOutputFormat(w, r); err != nil {
		err := fmt.Errorf("output only as json supported")
		logger.Error("[Reader][/stream/items/ids] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}

	rm, err := getStreamFilterModifiers(r)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	logger.Info("Request Modifiers: %v", rm)
	if len(rm.Streams) != 1 {
		err := fmt.Errorf("only one stream type expected")
		logger.Error("[Reader][/stream/items/ids] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}
	switch rm.Streams[0].Type {
	case ReadingListStream:
		h.handleReadingListStream(w, r, rm)
	case StarredStream:
		h.handleStarredStream(w, r, rm)
	case ReadStream:
		h.handleReadStream(w, r, rm)
	default:
		dump, _ := httputil.DumpRequest(r, true)
		logger.Info("[Reader][/stream/items/ids] [ClientIP=%s] Unknown Stream: %s", clientIP, dump)
		err := fmt.Errorf("unknown stream type")
		logger.Error("[Reader][/stream/items/ids] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}
}

func (h *handler) handleReadingListStream(w http.ResponseWriter, r *http.Request, rm RequestModifiers) {
	clientIP := request.ClientIP(r)

	builder := h.store.NewEntryQueryBuilder(rm.UserID)
	for _, s := range rm.ExcludeTargets {
		switch s.Type {
		case ReadStream:
			builder.WithStatus(model.EntryStatusUnread)
		default:
			logger.Info("[Reader][ReadingListStreamIDs][ClientIP=%s] xt filter type: %#v", clientIP, s)
		}
	}
	builder.WithLimit(rm.Count)
	rawEntryIDs, err := builder.GetEntryIDs()
	if err != nil {
		logger.Error("[Reader][/stream/items/ids#reading-list] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}
	var itemRefs = make([]itemRef, 0)
	for _, entryID := range rawEntryIDs {
		formattedID := strconv.FormatInt(entryID, 10)
		itemRefs = append(itemRefs, itemRef{ID: formattedID})
	}
	json.OK(w, r, streamIDResponse{itemRefs})
}

func (h *handler) handleStarredStream(w http.ResponseWriter, r *http.Request, rm RequestModifiers) {
	clientIP := request.ClientIP(r)

	builder := h.store.NewEntryQueryBuilder(rm.UserID)
	builder.WithStarred()
	builder.WithLimit(rm.Count)
	rawEntryIDs, err := builder.GetEntryIDs()
	if err != nil {
		logger.Error("[Reader][/stream/items/ids#starred] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}
	var itemRefs = make([]itemRef, 0)
	for _, entryID := range rawEntryIDs {
		formattedID := strconv.FormatInt(entryID, 10)
		itemRefs = append(itemRefs, itemRef{ID: formattedID})
	}
	json.OK(w, r, streamIDResponse{itemRefs})
}

func (h *handler) handleReadStream(w http.ResponseWriter, r *http.Request, rm RequestModifiers) {
	clientIP := request.ClientIP(r)

	builder := h.store.NewEntryQueryBuilder(rm.UserID)
	builder.WithStatus(model.EntryStatusRead)
	if rm.StartTime > 0 {
		builder.AfterDate(time.Unix(rm.StartTime, 0))
	}
	if rm.StopTime > 0 {
		builder.BeforeDate(time.Unix(rm.StopTime, 0))
	}

	rawEntryIDs, err := builder.GetEntryIDs()
	if err != nil {
		logger.Error("[Reader][/stream/items/ids#read] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}
	var itemRefs = make([]itemRef, 0)
	for _, entryID := range rawEntryIDs {
		formattedID := strconv.FormatInt(entryID, 10)
		itemRefs = append(itemRefs, itemRef{ID: formattedID})
	}
	json.OK(w, r, streamIDResponse{itemRefs})
}
