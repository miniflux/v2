// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package json // import "miniflux.app/v2/internal/reader/json"

// JSON Feed specs:
// https://www.jsonfeed.org/version/1.1/
// https://www.jsonfeed.org/version/1/
type JSONFeed struct {
	// Version is the URL of the version of the format the feed uses.
	// This should appear at the very top, though we recognize that not all JSON generators allow for ordering.
	Version string `json:"version"`

	// Title is the name of the feed, which will often correspond to the name of the website.
	Title string `json:"title"`

	// HomePageURL  is the URL of the resource that the feed describes.
	// This resource may or may not actually be a “home” page, but it should be an HTML page.
	HomePageURL string `json:"home_page_url"`

	// FeedURL is the URL of the feed, and serves as the unique identifier for the feed.
	FeedURL string `json:"feed_url"`

	// Description provides more detail, beyond the title, on what the feed is about.
	Description string `json:"description"`

	// IconURL is the URL of an image for the feed suitable to be used in a timeline, much the way an avatar might be used.
	IconURL string `json:"icon"`

	// FaviconURL is the URL of an image for the feed suitable to be used in a source list. It should be square and relatively small.
	FaviconURL string `json:"favicon"`

	// Authors specifies one or more feed authors. The author object has several members.
	Authors []JSONAuthor `json:"authors"` // JSON Feed v1.1

	// Author specifies the feed author. The author object has several members.
	// JSON Feed v1 (deprecated)
	Author JSONAuthor `json:"author"`

	// Language is the primary language for the feed in the format specified in RFC 5646.
	// The value is usually a 2-letter language tag from ISO 639-1, optionally followed by a region tag. (Examples: en or en-US.)
	Language string `json:"language"`

	// Expired is a boolean value that specifies whether or not the feed is finished.
	Expired bool `json:"expired"`

	// Items is an array, each representing an individual item in the feed.
	Items []JSONItem `json:"items"`

	// Hubs  describes endpoints that can be used to subscribe to real-time notifications from the publisher of this feed.
	Hubs []JSONHub `json:"hubs"`
}

type JSONAuthor struct {
	// Author's name.
	Name string `json:"name"`

	// Author's website URL (Blog or micro-blog).
	WebsiteURL string `json:"url"`

	// Author's avatar URL.
	AvatarURL string `json:"avatar"`
}

type JSONHub struct {
	// Type defines the protocol used to talk with the hub: "rssCloud" or "WebSub".
	Type string `json:"type"`

	// URL is the location of the hub.
	URL string `json:"url"`
}

type JSONItem struct {
	// Unique identifier for the item.
	// Ideally, the id is the full URL of the resource described by the item, since URLs make great unique identifiers.
	ID string `json:"id"`

	// URL of the resource described by the item.
	URL string `json:"url"`

	// ExternalURL is the URL of a page elsewhere.
	// This is especially useful for linkblogs.
	// If url links to where you’re talking about a thing, then external_url links to the thing you’re talking about.
	ExternalURL string `json:"external_url"`

	// Title of the item (optional).
	// Microblog items in particular may omit titles.
	Title string `json:"title"`

	// ContentHTML is the HTML body of the item.
	ContentHTML string `json:"content_html"`

	// ContentText is the text body of the item.
	ContentText string `json:"content_text"`

	// Summary is a plain text sentence or two describing the item.
	Summary string `json:"summary"`

	// ImageURL is the URL of the main image for the item.
	ImageURL string `json:"image"`

	// BannerImageURL is the URL of an image to use as a banner.
	BannerImageURL string `json:"banner_image"`

	// DatePublished is the date the item was published.
	DatePublished string `json:"date_published"`

	// DateModified is the date the item was modified.
	DateModified string `json:"date_modified"`

	// Language is the language of the item.
	Language string `json:"language"`

	// Authors is an array of JSONAuthor.
	Authors []JSONAuthor `json:"authors"`

	// Author is a JSONAuthor.
	// JSON Feed v1 (deprecated)
	Author JSONAuthor `json:"author"`

	// Tags is an array of strings.
	Tags []string `json:"tags"`

	// Attachments is an array of JSONAttachment.
	Attachments []JSONAttachment `json:"attachments"`
}

type JSONAttachment struct {
	// URL of the attachment.
	URL string `json:"url"`

	// MIME type of the attachment.
	MimeType string `json:"mime_type"`

	// Title of the attachment.
	Title string `json:"title"`

	// Size of the attachment in bytes.
	Size int64 `json:"size_in_bytes"`

	// Duration of the attachment in seconds.
	Duration int `json:"duration_in_seconds"`
}
