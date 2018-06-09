// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fever

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/miniflux/miniflux/config"
	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/http/request"
	"github.com/miniflux/miniflux/http/response/json"
	"github.com/miniflux/miniflux/integration"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/storage"
)

type baseResponse struct {
	Version       int   `json:"api_version"`
	Authenticated int   `json:"auth"`
	LastRefresh   int64 `json:"last_refreshed_on_time"`
}

func (b *baseResponse) SetCommonValues() {
	b.Version = 3
	b.Authenticated = 1
	b.LastRefresh = time.Now().Unix()
}

/*
The default response is a JSON object containing two members:

    api_version contains the version of the API responding (positive integer)
    auth whether the request was successfully authenticated (boolean integer)

The API can also return XML by passing xml as the optional value of the api argument like so:

http://yourdomain.com/fever/?api=xml

The top level XML element is named response.

The response to each successfully authenticated request will have auth set to 1 and include
at least one additional member:

	last_refreshed_on_time contains the time of the most recently refreshed (not updated)
	feed (Unix timestamp/integer)

*/
func newBaseResponse() baseResponse {
	r := baseResponse{}
	r.SetCommonValues()
	return r
}

type groupsResponse struct {
	baseResponse
	Groups      []group       `json:"groups"`
	FeedsGroups []feedsGroups `json:"feeds_groups"`
}

type feedsResponse struct {
	baseResponse
	Feeds       []feed        `json:"feeds"`
	FeedsGroups []feedsGroups `json:"feeds_groups"`
}

type faviconsResponse struct {
	baseResponse
	Favicons []favicon `json:"favicons"`
}

type itemsResponse struct {
	baseResponse
	Items []item `json:"items"`
	Total int    `json:"total_items"`
}

type unreadResponse struct {
	baseResponse
	ItemIDs string `json:"unread_item_ids"`
}

type savedResponse struct {
	baseResponse
	ItemIDs string `json:"saved_item_ids"`
}

type group struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type feedsGroups struct {
	GroupID int64  `json:"group_id"`
	FeedIDs string `json:"feed_ids"`
}

type feed struct {
	ID          int64  `json:"id"`
	FaviconID   int64  `json:"favicon_id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	SiteURL     string `json:"site_url"`
	IsSpark     int    `json:"is_spark"`
	LastUpdated int64  `json:"last_updated_on_time"`
}

type item struct {
	ID        int64  `json:"id"`
	FeedID    int64  `json:"feed_id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	HTML      string `json:"html"`
	URL       string `json:"url"`
	IsSaved   int    `json:"is_saved"`
	IsRead    int    `json:"is_read"`
	CreatedAt int64  `json:"created_on_time"`
}

type favicon struct {
	ID   int64  `json:"id"`
	Data string `json:"data"`
}

// Controller implements the Fever API.
type Controller struct {
	cfg   *config.Config
	store *storage.Storage
}

// Handler handles Fever API calls
func (c *Controller) Handler(w http.ResponseWriter, r *http.Request) {
	switch {
	case request.HasQueryParam(r, "groups"):
		c.handleGroups(w, r)
	case request.HasQueryParam(r, "feeds"):
		c.handleFeeds(w, r)
	case request.HasQueryParam(r, "favicons"):
		c.handleFavicons(w, r)
	case request.HasQueryParam(r, "unread_item_ids"):
		c.handleUnreadItems(w, r)
	case request.HasQueryParam(r, "saved_item_ids"):
		c.handleSavedItems(w, r)
	case request.HasQueryParam(r, "items"):
		c.handleItems(w, r)
	case r.FormValue("mark") == "item":
		c.handleWriteItems(w, r)
	case r.FormValue("mark") == "feed":
		c.handleWriteFeeds(w, r)
	case r.FormValue("mark") == "group":
		c.handleWriteGroups(w, r)
	default:
		json.OK(w, newBaseResponse())
	}
}

/*
A request with the groups argument will return two additional members:

    groups contains an array of group objects
    feeds_groups contains an array of feeds_group objects

A group object has the following members:

    id (positive integer)
    title (utf-8 string)

The feeds_group object is documented under “Feeds/Groups Relationships.”

The “Kindling” super group is not included in this response and is composed of all feeds with
an is_spark equal to 0.

The “Sparks” super group is not included in this response and is composed of all feeds with an
is_spark equal to 1.

*/
func (c *Controller) handleGroups(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	userID := ctx.UserID()
	logger.Debug("[Fever] Fetching groups for userID=%d", userID)

	categories, err := c.store.Categories(userID)
	if err != nil {
		json.ServerError(w, err)
		return
	}

	feeds, err := c.store.Feeds(userID)
	if err != nil {
		json.ServerError(w, err)
		return
	}

	var result groupsResponse
	for _, category := range categories {
		result.Groups = append(result.Groups, group{ID: category.ID, Title: category.Title})
	}

	result.FeedsGroups = c.buildFeedGroups(feeds)
	result.SetCommonValues()
	json.OK(w, result)
}

/*
A request with the feeds argument will return two additional members:

    feeds contains an array of group objects
    feeds_groups contains an array of feeds_group objects

A feed object has the following members:

    id (positive integer)
    favicon_id (positive integer)
    title (utf-8 string)
    url (utf-8 string)
    site_url (utf-8 string)
    is_spark (boolean integer)
    last_updated_on_time (Unix timestamp/integer)

The feeds_group object is documented under “Feeds/Groups Relationships.”

The “All Items” super feed is not included in this response and is composed of all items from all feeds
that belong to a given group. For the “Kindling” super group and all user created groups the items
should be limited to feeds with an is_spark equal to 0.

For the “Sparks” super group the items should be limited to feeds with an is_spark equal to 1.
*/
func (c *Controller) handleFeeds(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	userID := ctx.UserID()
	logger.Debug("[Fever] Fetching feeds for userID=%d", userID)

	feeds, err := c.store.Feeds(userID)
	if err != nil {
		json.ServerError(w, err)
		return
	}

	var result feedsResponse
	result.Feeds = make([]feed, 0)
	for _, f := range feeds {
		subscripion := feed{
			ID:          f.ID,
			Title:       f.Title,
			URL:         f.FeedURL,
			SiteURL:     f.SiteURL,
			IsSpark:     0,
			LastUpdated: f.CheckedAt.Unix(),
		}

		if f.Icon != nil {
			subscripion.FaviconID = f.Icon.IconID
		}

		result.Feeds = append(result.Feeds, subscripion)
	}

	result.FeedsGroups = c.buildFeedGroups(feeds)
	result.SetCommonValues()
	json.OK(w, result)
}

/*
A request with the favicons argument will return one additional member:

    favicons contains an array of favicon objects

A favicon object has the following members:

    id (positive integer)
    data (base64 encoded image data; prefixed by image type)

An example data value:

	image/gif;base64,R0lGODlhAQABAIAAAObm5gAAACH5BAEAAAAALAAAAAABAAEAAAICRAEAOw==

The data member of a favicon object can be used with the data: protocol to embed an image in CSS or HTML.
A PHP/HTML example:

	echo '<img src="data:'.$favicon['data'].'">';
*/
func (c *Controller) handleFavicons(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	userID := ctx.UserID()
	logger.Debug("[Fever] Fetching favicons for userID=%d", userID)

	icons, err := c.store.Icons(userID)
	if err != nil {
		json.ServerError(w, err)
		return
	}

	var result faviconsResponse
	for _, i := range icons {
		result.Favicons = append(result.Favicons, favicon{
			ID:   i.ID,
			Data: i.DataURL(),
		})
	}

	result.SetCommonValues()
	json.OK(w, result)
}

/*
A request with the items argument will return two additional members:

    items contains an array of item objects
    total_items contains the total number of items stored in the database (added in API version 2)

An item object has the following members:

    id (positive integer)
    feed_id (positive integer)
    title (utf-8 string)
    author (utf-8 string)
    html (utf-8 string)
    url (utf-8 string)
    is_saved (boolean integer)
    is_read (boolean integer)
    created_on_time (Unix timestamp/integer)

Most servers won’t have enough memory allocated to PHP to dump all items at once.
Three optional arguments control determine the items included in the response.

	Use the since_id argument with the highest id of locally cached items to request 50 additional items.
	Repeat until the items array in the response is empty.

	Use the max_id argument with the lowest id of locally cached items (or 0 initially) to request 50 previous items.
	Repeat until the items array in the response is empty. (added in API version 2)

	Use the with_ids argument with a comma-separated list of item ids to request (a maximum of 50) specific items.
	(added in API version 2)

*/
func (c *Controller) handleItems(w http.ResponseWriter, r *http.Request) {
	var result itemsResponse

	ctx := context.New(r)
	userID := ctx.UserID()
	logger.Debug("[Fever] Fetching items for userID=%d", userID)

	builder := c.store.NewEntryQueryBuilder(userID)
	builder.WithoutStatus(model.EntryStatusRemoved)
	builder.WithLimit(50)
	builder.WithOrder("id")
	builder.WithDirection(model.DefaultSortingDirection)

	sinceID := request.QueryIntParam(r, "since_id", 0)
	if sinceID > 0 {
		builder.AfterEntryID(int64(sinceID))
	}

	maxID := request.QueryIntParam(r, "max_id", 0)
	if maxID > 0 {
		builder.WithOffset(maxID)
	}

	csvItemIDs := request.QueryParam(r, "with_ids", "")
	if csvItemIDs != "" {
		var itemIDs []int64

		for _, strItemID := range strings.Split(csvItemIDs, ",") {
			strItemID = strings.TrimSpace(strItemID)
			itemID, _ := strconv.Atoi(strItemID)
			itemIDs = append(itemIDs, int64(itemID))
		}

		builder.WithEntryIDs(itemIDs)
	}

	entries, err := builder.GetEntries()
	if err != nil {
		json.ServerError(w, err)
		return
	}

	builder = c.store.NewEntryQueryBuilder(userID)
	builder.WithoutStatus(model.EntryStatusRemoved)
	result.Total, err = builder.CountEntries()
	if err != nil {
		json.ServerError(w, err)
		return
	}

	result.Items = make([]item, 0)
	for _, entry := range entries {
		isRead := 0
		if entry.Status == model.EntryStatusRead {
			isRead = 1
		}

		isSaved := 0
		if entry.Starred {
			isSaved = 1
		}

		result.Items = append(result.Items, item{
			ID:        entry.ID,
			FeedID:    entry.FeedID,
			Title:     entry.Title,
			Author:    entry.Author,
			HTML:      entry.Content,
			URL:       entry.URL,
			IsSaved:   isSaved,
			IsRead:    isRead,
			CreatedAt: entry.Date.Unix(),
		})
	}

	result.SetCommonValues()
	json.OK(w, result)
}

/*
The unread_item_ids and saved_item_ids arguments can be used to keep your local cache synced
with the remote Fever installation.

A request with the unread_item_ids argument will return one additional member:
    unread_item_ids (string/comma-separated list of positive integers)
*/
func (c *Controller) handleUnreadItems(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	userID := ctx.UserID()
	logger.Debug("[Fever] Fetching unread items for userID=%d", userID)

	builder := c.store.NewEntryQueryBuilder(userID)
	builder.WithStatus(model.EntryStatusUnread)
	entries, err := builder.GetEntries()
	if err != nil {
		json.ServerError(w, err)
		return
	}

	var itemIDs []string
	for _, entry := range entries {
		itemIDs = append(itemIDs, strconv.FormatInt(entry.ID, 10))
	}

	var result unreadResponse
	result.ItemIDs = strings.Join(itemIDs, ",")
	result.SetCommonValues()
	json.OK(w, result)
}

/*
The unread_item_ids and saved_item_ids arguments can be used to keep your local cache synced
with the remote Fever installation.

	A request with the saved_item_ids argument will return one additional member:

	saved_item_ids (string/comma-separated list of positive integers)
*/
func (c *Controller) handleSavedItems(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	userID := ctx.UserID()
	logger.Debug("[Fever] Fetching saved items for userID=%d", userID)

	builder := c.store.NewEntryQueryBuilder(userID)
	builder.WithStarred()

	entryIDs, err := builder.GetEntryIDs()
	if err != nil {
		json.ServerError(w, err)
		return
	}

	var itemsIDs []string
	for _, entryID := range entryIDs {
		itemsIDs = append(itemsIDs, strconv.FormatInt(entryID, 10))
	}

	result := &savedResponse{ItemIDs: strings.Join(itemsIDs, ",")}
	result.SetCommonValues()
	json.OK(w, result)
}

/*
	mark=item
	as=? where ? is replaced with read, saved or unsaved
	id=? where ? is replaced with the id of the item to modify
*/
func (c *Controller) handleWriteItems(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	userID := ctx.UserID()
	logger.Debug("[Fever] Receiving mark=item call for userID=%d", userID)

	entryID := request.FormIntValue(r, "id")
	if entryID <= 0 {
		return
	}

	builder := c.store.NewEntryQueryBuilder(userID)
	builder.WithEntryID(entryID)
	builder.WithoutStatus(model.EntryStatusRemoved)

	entry, err := builder.GetEntry()
	if err != nil {
		json.ServerError(w, err)
		return
	}

	if entry == nil {
		return
	}

	switch r.FormValue("as") {
	case "read":
		logger.Debug("[Fever] Mark entry #%d as read", entryID)
		c.store.SetEntriesStatus(userID, []int64{entryID}, model.EntryStatusRead)
	case "unread":
		logger.Debug("[Fever] Mark entry #%d as unread", entryID)
		c.store.SetEntriesStatus(userID, []int64{entryID}, model.EntryStatusUnread)
	case "saved", "unsaved":
		logger.Debug("[Fever] Mark entry #%d as saved/unsaved", entryID)
		if err := c.store.ToggleBookmark(userID, entryID); err != nil {
			json.ServerError(w, err)
			return
		}

		settings, err := c.store.Integration(userID)
		if err != nil {
			json.ServerError(w, err)
			return
		}

		go func() {
			integration.SendEntry(c.cfg, entry, settings)
		}()
	}

	json.OK(w, newBaseResponse())
}

/*
	mark=? where ? is replaced with feed or group
	as=read
	id=? where ? is replaced with the id of the feed or group to modify
	before=? where ? is replaced with the Unix timestamp of the the local client’s most recent items API request
*/
func (c *Controller) handleWriteFeeds(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	userID := ctx.UserID()
	logger.Debug("[Fever] Receiving mark=feed call for userID=%d", userID)

	feedID := request.FormIntValue(r, "id")
	if feedID <= 0 {
		return
	}

	builder := c.store.NewEntryQueryBuilder(userID)
	builder.WithStatus(model.EntryStatusUnread)
	builder.WithFeedID(feedID)

	before := request.FormIntValue(r, "before")
	if before > 0 {
		t := time.Unix(before, 0)
		builder.BeforeDate(t)
	}

	entryIDs, err := builder.GetEntryIDs()
	if err != nil {
		json.ServerError(w, err)
		return
	}

	err = c.store.SetEntriesStatus(userID, entryIDs, model.EntryStatusRead)
	if err != nil {
		json.ServerError(w, err)
		return
	}

	json.OK(w, newBaseResponse())
}

/*
	mark=? where ? is replaced with feed or group
	as=read
	id=? where ? is replaced with the id of the feed or group to modify
	before=? where ? is replaced with the Unix timestamp of the the local client’s most recent items API request
*/
func (c *Controller) handleWriteGroups(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	userID := ctx.UserID()
	logger.Debug("[Fever] Receiving mark=group call for userID=%d", userID)

	groupID := request.FormIntValue(r, "id")
	if groupID < 0 {
		return
	}

	builder := c.store.NewEntryQueryBuilder(userID)
	builder.WithStatus(model.EntryStatusUnread)
	builder.WithCategoryID(groupID)

	before := request.FormIntValue(r, "before")
	if before > 0 {
		t := time.Unix(before, 0)
		builder.BeforeDate(t)
	}

	entryIDs, err := builder.GetEntryIDs()
	if err != nil {
		json.ServerError(w, err)
		return
	}

	err = c.store.SetEntriesStatus(userID, entryIDs, model.EntryStatusRead)
	if err != nil {
		json.ServerError(w, err)
		return
	}

	json.OK(w, newBaseResponse())
}

/*
A feeds_group object has the following members:

    group_id (positive integer)
    feed_ids (string/comma-separated list of positive integers)

*/
func (c *Controller) buildFeedGroups(feeds model.Feeds) []feedsGroups {
	feedsGroupedByCategory := make(map[int64][]string)
	for _, feed := range feeds {
		feedsGroupedByCategory[feed.Category.ID] = append(feedsGroupedByCategory[feed.Category.ID], strconv.FormatInt(feed.ID, 10))
	}

	result := make([]feedsGroups, 0)
	for categoryID, feedIDs := range feedsGroupedByCategory {
		result = append(result, feedsGroups{
			GroupID: categoryID,
			FeedIDs: strings.Join(feedIDs, ","),
		})
	}

	return result
}

// NewController returns a new Fever API.
func NewController(cfg *config.Config, store *storage.Storage) *Controller {
	return &Controller{cfg, store}
}
