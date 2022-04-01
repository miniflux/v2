// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package client // import "miniflux.app/client"

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
)

// Client holds API procedure calls.
type Client struct {
	request *request
}

// New returns a new Miniflux client.
func New(endpoint string, credentials ...string) *Client {
	if len(credentials) == 2 {
		return &Client{request: &request{endpoint: endpoint, username: credentials[0], password: credentials[1]}}
	}
	return &Client{request: &request{endpoint: endpoint, apiKey: credentials[0]}}
}

// Me returns the logged user information.
func (c *Client) Me() (*User, error) {
	body, err := c.request.Get("/v1/me")
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var user *User
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&user); err != nil {
		return nil, fmt.Errorf("miniflux: json error (%v)", err)
	}

	return user, nil
}

// Users returns all users.
func (c *Client) Users() (Users, error) {
	body, err := c.request.Get("/v1/users")
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var users Users
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&users); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return users, nil
}

// UserByID returns a single user.
func (c *Client) UserByID(userID int64) (*User, error) {
	body, err := c.request.Get(fmt.Sprintf("/v1/users/%d", userID))
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var user User
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&user); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return &user, nil
}

// UserByUsername returns a single user.
func (c *Client) UserByUsername(username string) (*User, error) {
	body, err := c.request.Get(fmt.Sprintf("/v1/users/%s", username))
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var user User
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&user); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return &user, nil
}

// CreateUser creates a new user in the system.
func (c *Client) CreateUser(username, password string, isAdmin bool) (*User, error) {
	body, err := c.request.Post("/v1/users", &UserCreationRequest{
		Username: username,
		Password: password,
		IsAdmin:  isAdmin,
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var user *User
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&user); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return user, nil
}

// UpdateUser updates a user in the system.
func (c *Client) UpdateUser(userID int64, userChanges *UserModificationRequest) (*User, error) {
	body, err := c.request.Put(fmt.Sprintf("/v1/users/%d", userID), userChanges)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var u *User
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&u); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return u, nil
}

// DeleteUser removes a user from the system.
func (c *Client) DeleteUser(userID int64) error {
	return c.request.Delete(fmt.Sprintf("/v1/users/%d", userID))
}

// MarkAllAsRead marks all unread entries as read for a given user.
func (c *Client) MarkAllAsRead(userID int64) error {
	_, err := c.request.Put(fmt.Sprintf("/v1/users/%d/mark-all-as-read", userID), nil)
	return err
}

// Discover try to find subscriptions from a website.
func (c *Client) Discover(url string) (Subscriptions, error) {
	body, err := c.request.Post("/v1/discover", map[string]string{"url": url})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var subscriptions Subscriptions
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&subscriptions); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return subscriptions, nil
}

// Categories gets the list of categories.
func (c *Client) Categories() (Categories, error) {
	body, err := c.request.Get("/v1/categories")
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var categories Categories
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&categories); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return categories, nil
}

// CreateCategory creates a new category.
func (c *Client) CreateCategory(title string) (*Category, error) {
	body, err := c.request.Post("/v1/categories", map[string]interface{}{
		"title": title,
	})

	if err != nil {
		return nil, err
	}
	defer body.Close()

	var category *Category
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&category); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return category, nil
}

// UpdateCategory updates a category.
func (c *Client) UpdateCategory(categoryID int64, title string) (*Category, error) {
	body, err := c.request.Put(fmt.Sprintf("/v1/categories/%d", categoryID), map[string]interface{}{
		"title": title,
	})

	if err != nil {
		return nil, err
	}
	defer body.Close()

	var category *Category
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&category); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return category, nil
}

// MarkCategoryAsRead marks all unread entries in a category as read.
func (c *Client) MarkCategoryAsRead(categoryID int64) error {
	_, err := c.request.Put(fmt.Sprintf("/v1/categories/%d/mark-all-as-read", categoryID), nil)
	return err
}

// CategoryFeeds gets feeds of a category.
func (c *Client) CategoryFeeds(categoryID int64) (Feeds, error) {
	body, err := c.request.Get(fmt.Sprintf("/v1/categories/%d/feeds", categoryID))
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var feeds Feeds
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&feeds); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return feeds, nil
}

// DeleteCategory removes a category.
func (c *Client) DeleteCategory(categoryID int64) error {
	return c.request.Delete(fmt.Sprintf("/v1/categories/%d", categoryID))
}

// Feeds gets all feeds.
func (c *Client) Feeds() (Feeds, error) {
	body, err := c.request.Get("/v1/feeds")
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var feeds Feeds
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&feeds); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return feeds, nil
}

// Export creates OPML file.
func (c *Client) Export() ([]byte, error) {
	body, err := c.request.Get("/v1/export")
	if err != nil {
		return nil, err
	}
	defer body.Close()

	opml, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	return opml, nil
}

// Import imports an OPML file.
func (c *Client) Import(f io.ReadCloser) error {
	_, err := c.request.PostFile("/v1/import", f)
	return err
}

// Feed gets a feed.
func (c *Client) Feed(feedID int64) (*Feed, error) {
	body, err := c.request.Get(fmt.Sprintf("/v1/feeds/%d", feedID))
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var feed *Feed
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&feed); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return feed, nil
}

// CreateFeed creates a new feed.
func (c *Client) CreateFeed(feedCreationRequest *FeedCreationRequest) (int64, error) {
	body, err := c.request.Post("/v1/feeds", feedCreationRequest)
	if err != nil {
		return 0, err
	}
	defer body.Close()

	type result struct {
		FeedID int64 `json:"feed_id"`
	}

	var r result
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&r); err != nil {
		return 0, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return r.FeedID, nil
}

// UpdateFeed updates a feed.
func (c *Client) UpdateFeed(feedID int64, feedChanges *FeedModificationRequest) (*Feed, error) {
	body, err := c.request.Put(fmt.Sprintf("/v1/feeds/%d", feedID), feedChanges)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var f *Feed
	if err := json.NewDecoder(body).Decode(&f); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return f, nil
}

// MarkFeedAsRead marks all unread entries of the feed as read.
func (c *Client) MarkFeedAsRead(feedID int64) error {
	_, err := c.request.Put(fmt.Sprintf("/v1/feeds/%d/mark-all-as-read", feedID), nil)
	return err
}

// RefreshAllFeeds refreshes all feeds.
func (c *Client) RefreshAllFeeds() error {
	_, err := c.request.Put("/v1/feeds/refresh", nil)
	return err
}

// RefreshFeed refreshes a feed.
func (c *Client) RefreshFeed(feedID int64) error {
	_, err := c.request.Put(fmt.Sprintf("/v1/feeds/%d/refresh", feedID), nil)
	return err
}

// DeleteFeed removes a feed.
func (c *Client) DeleteFeed(feedID int64) error {
	return c.request.Delete(fmt.Sprintf("/v1/feeds/%d", feedID))
}

// FeedIcon gets a feed icon.
func (c *Client) FeedIcon(feedID int64) (*FeedIcon, error) {
	body, err := c.request.Get(fmt.Sprintf("/v1/feeds/%d/icon", feedID))
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var feedIcon *FeedIcon
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&feedIcon); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return feedIcon, nil
}

// FeedEntry gets a single feed entry.
func (c *Client) FeedEntry(feedID, entryID int64) (*Entry, error) {
	body, err := c.request.Get(fmt.Sprintf("/v1/feeds/%d/entries/%d", feedID, entryID))
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var entry *Entry
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&entry); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return entry, nil
}

// CategoryEntry gets a single category entry.
func (c *Client) CategoryEntry(categoryID, entryID int64) (*Entry, error) {
	body, err := c.request.Get(fmt.Sprintf("/v1/categories/%d/entries/%d", categoryID, entryID))
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var entry *Entry
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&entry); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return entry, nil
}

// Entry gets a single entry.
func (c *Client) Entry(entryID int64) (*Entry, error) {
	body, err := c.request.Get(fmt.Sprintf("/v1/entries/%d", entryID))
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var entry *Entry
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&entry); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return entry, nil
}

// Entries fetch entries.
func (c *Client) Entries(filter *Filter) (*EntryResultSet, error) {
	path := buildFilterQueryString("/v1/entries", filter)

	body, err := c.request.Get(path)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var result EntryResultSet
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return &result, nil
}

// FeedEntries fetch feed entries.
func (c *Client) FeedEntries(feedID int64, filter *Filter) (*EntryResultSet, error) {
	path := buildFilterQueryString(fmt.Sprintf("/v1/feeds/%d/entries", feedID), filter)

	body, err := c.request.Get(path)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var result EntryResultSet
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return &result, nil
}

// CategoryEntries fetch entries of a category.
func (c *Client) CategoryEntries(categoryID int64, filter *Filter) (*EntryResultSet, error) {
	path := buildFilterQueryString(fmt.Sprintf("/v1/categories/%d/entries", categoryID), filter)

	body, err := c.request.Get(path)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var result EntryResultSet
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("miniflux: response error (%v)", err)
	}

	return &result, nil
}

// UpdateEntries updates the status of a list of entries.
func (c *Client) UpdateEntries(entryIDs []int64, status string) error {
	type payload struct {
		EntryIDs []int64 `json:"entry_ids"`
		Status   string  `json:"status"`
	}

	_, err := c.request.Put("/v1/entries", &payload{EntryIDs: entryIDs, Status: status})
	return err
}

// ToggleBookmark toggles entry bookmark value.
func (c *Client) ToggleBookmark(entryID int64) error {
	_, err := c.request.Put(fmt.Sprintf("/v1/entries/%d/bookmark", entryID), nil)
	return err
}

func buildFilterQueryString(path string, filter *Filter) string {
	if filter != nil {
		values := url.Values{}

		if filter.Status != "" {
			values.Set("status", filter.Status)
		}

		if filter.Direction != "" {
			values.Set("direction", filter.Direction)
		}

		if filter.Order != "" {
			values.Set("order", filter.Order)
		}

		if filter.Limit >= 0 {
			values.Set("limit", strconv.Itoa(filter.Limit))
		}

		if filter.Offset >= 0 {
			values.Set("offset", strconv.Itoa(filter.Offset))
		}

		if filter.After > 0 {
			values.Set("after", strconv.FormatInt(filter.After, 10))
		}

		if filter.AfterEntryID > 0 {
			values.Set("after_entry_id", strconv.FormatInt(filter.AfterEntryID, 10))
		}

		if filter.Before > 0 {
			values.Set("before", strconv.FormatInt(filter.Before, 10))
		}

		if filter.BeforeEntryID > 0 {
			values.Set("before_entry_id", strconv.FormatInt(filter.BeforeEntryID, 10))
		}

		if filter.Starred != "" {
			values.Set("starred", filter.Starred)
		}

		if filter.Search != "" {
			values.Set("search", filter.Search)
		}

		if filter.CategoryID > 0 {
			values.Set("category_id", strconv.FormatInt(filter.CategoryID, 10))
		}

		if filter.FeedID > 0 {
			values.Set("feed_id", strconv.FormatInt(filter.FeedID, 10))
		}

		for _, status := range filter.Statuses {
			values.Add("status", status)
		}

		path = fmt.Sprintf("%s?%s", path, values.Encode())
	}

	return path
}
