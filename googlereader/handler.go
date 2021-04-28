package googlereader // import "miniflux.app/googlereader"

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"miniflux.app/config"
	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/storage"
)

type handler struct {
	store  *storage.Storage
	router *mux.Router
}

const (
	ReadingList     = "user/-/state/com.google/reading-list"
	Read            = "user/-/state/com.google/read"
	Starred         = "user/-/state/com.google/starred"
	UserReadingList = "user/%d/state/com.google/reading-list"
	UserRead        = "user/%d/state/com.google/read"
	UserStarred     = "user/%d/state/com.google/starred"
	EntryIDLong     = "tag:google.com,2005:reader/item/%016x"
)

const (
	StreamIdParam        = "s"
	StreamExcludeIdParam = "xt"
	StreamMaxItems       = "n"
	StreamOrder          = "r"
)

type RequestModifiers struct {
	ExcludeTargets    string
	FilterTarget      string
	Count             int
	SortDirection     string
	StartTime         int
	StopTime          int
	ContinuationToken string
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
	sr.HandleFunc("/user-info", handler.userInfo).Methods(http.MethodGet).Name("UserInfo")
	sr.HandleFunc("/subscription/list", handler.subscriptionList).Methods(http.MethodGet).Name("SubscriptonList")
	sr.HandleFunc("/stream/items/ids", handler.streamItemIDs).Methods(http.MethodGet).Name("StreamItemIDs")
	sr.HandleFunc("/stream/items/contents", handler.streamItemContents).Methods(http.MethodPost).Name("StreamItemsContents")
	sr.PathPrefix("/").HandlerFunc(handler.serve).Methods(http.MethodPost, http.MethodGet).Name("GoogleReaderApiEndpoint")
}

func getModifiers(w http.ResponseWriter, r *http.Request) RequestModifiers {
	result := RequestModifiers{}
	streamOrder := request.QueryStringParam(r, StreamOrder, "d")
	if streamOrder == "o" {
		result.SortDirection = "asc"
	} else {
		result.SortDirection = "desc"
	}
	return result
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

func (h *handler) streamItemContents(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	logger.Info("[Reader][/stream/items/contents][ClientIP=%s] Incoming Request for userID  #%d", clientIP, userID)

	if err := checkOutputFormat(w, r); err != nil {
		logger.Error("[Reader][/stream/items/contents] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
	}

	err := r.ParseForm()
	if err != nil {
		json.ServerError(w, r, err)
		return

	}
	var user *model.User
	if user, err = h.store.UserByID(userID); err != nil {
		logger.Error("[Reader][/stream/items/contents] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}

	userReadingList := fmt.Sprintf(UserReadingList, user.ID)
	userRead := fmt.Sprintf(UserRead, user.ID)
	userStarred := fmt.Sprintf(UserStarred, user.ID)

	requestModifiers := getModifiers(w, r)

	items := r.Form["i"]
	if len(items) == 0 {
		err = fmt.Errorf("no items requested")
		logger.Error("[Reader][/stream/items/contents] [ClientIP=%s] %v", clientIP, err)
		json.ServerError(w, r, err)
		return
	}

	itemIDs := make([]int64, len(items))

	for i, item := range items {
		itemID, err := strconv.ParseInt(item, 16, 64)
		if err != nil {
			logger.Error("[Reader][/stream/items/contents] [ClientIP=%s] %v", clientIP, err)
			json.ServerError(w, r, err)
			return
		}
		itemIDs[i] = itemID
	}

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
		ID:        fmt.Sprintf("feed/%x", entries[0].FeedID),
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
		categories := make([]string, 1)
		categories = append(categories, userReadingList)

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
				StreamId: fmt.Sprintf("feed/%x", entry.FeedID),
				Title:    entry.Feed.Title,
				HTMLUrl:  entry.Feed.SiteURL,
			},
			Enclosure: enclosures,
		}
	}
	json.OK(w, r, result)
}

func (h *handler) subscriptionList(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	logger.Info("[Reader][subscription/list][ClientIP=%s] Incoming Request for userID  #%d", clientIP, userID)

	if err := checkOutputFormat(w, r); err != nil {
		logger.Error("[Reader][OutputFormat] %v", err)
		json.ServerError(w, r, err)
	}

	var result subscriptionsResponse
	feeds, err := h.store.Feeds(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	result.Subscriptions = make([]subscription, 0)
	for _, feed := range feeds {
		category, err := h.store.Category(userID, feed.Category.ID)
		if err != nil {
			json.ServerError(w, r, err)
			return
		}
		result.Subscriptions = append(result.Subscriptions, subscription{
			ID:         fmt.Sprint(feed.ID),
			Title:      feed.Title,
			URL:        feed.FeedURL,
			Categories: []subscriptionCategory{{fmt.Sprint(category.ID), category.Title}}, // TODO: should be something like 'id' => 'user/-/label/' . htmlspecialchars_decode($cat->name(), ENT_QUOTES),
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
		logger.Error("[Reader][OutputFormat] %v", err)
		json.ServerError(w, r, err)
	}

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	userInfo := userInfo{UserID: fmt.Sprint(user.ID), UserName: user.Username, UserProfileID: fmt.Sprint(user.ID), UserEmail: user.Username}
	json.OK(w, r, userInfo)
}

func (h *handler) streamItemIDs(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	logger.Debug("[Reader][stream/items/ids][ClientIP=%s] Incoming Request for userID  #%d", clientIP, userID)

	if err := checkOutputFormat(w, r); err != nil {
		err := fmt.Errorf("output only as json supported")
		logger.Error("[Reader][OutputFormat] %v", err)
		json.ServerError(w, r, err)
	}

	streamID := request.QueryStringParam(r, StreamIdParam, "")
	streamExcludeID := request.QueryStringParam(r, StreamExcludeIdParam, "")
	streamMaxItem := request.QueryIntParam(r, StreamMaxItems, 1000)

	if streamID == ReadingList && streamExcludeID == Read {
		h.handleStreamUnreadList(w, r, streamMaxItem)
	} else if streamID == Starred {
		h.handleStreamStarred(w, r, streamMaxItem)
	} else {
		dump, _ := httputil.DumpRequest(r, true)
		logger.Info("[Reader][stream/items/ids] [ClientIP=%s] Unknown Stream: %s", clientIP, dump)
	}
}

func (h *handler) handleStreamUnreadList(
	w http.ResponseWriter, r *http.Request, maxItems int) {
	userID := request.UserID(r)
	builder := h.store.NewEntryQueryBuilder(userID)
	builder.WithStatus(model.EntryStatusUnread)
	builder.WithLimit(maxItems)
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
	json.OK(w, r, streamIDResponse{itemRefs})
}

func (h *handler) handleStreamStarred(
	w http.ResponseWriter, r *http.Request, maxItems int) {
	userID := request.UserID(r)
	builder := h.store.NewEntryQueryBuilder(userID)
	builder.WithStarred()
	builder.WithLimit(maxItems)
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
	json.OK(w, r, streamIDResponse{itemRefs})
}
