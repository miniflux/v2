// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

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
)
