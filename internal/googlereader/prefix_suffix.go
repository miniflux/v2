// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

const (
	// streamPrefix is the prefix for streams (read/starred/reading list and so on)
	streamPrefix = "user/-/state/com.google/"
	// userStreamPrefix is the user specific prefix for streams (read/starred/reading list and so on)
	userStreamPrefix = "user/%d/state/com.google/"
	// labelPrefix is the prefix for a label stream
	labelPrefix = "user/-/label/"
	// userLabelPrefix is the user specific prefix prefix for a label stream
	userLabelPrefix = "user/%d/label/"
	// feedPrefix is the prefix for a feed stream
	feedPrefix = "feed/"
	// read is the suffix for read stream
	read = "read"
	// starred is the suffix for starred stream
	starred = "starred"
	// readingList is the suffix for reading list stream
	readingList = "reading-list"
	// keptUnread is the suffix for kept unread stream
	keptUnread = "kept-unread"
	// broadcast is the suffix for broadcast stream
	broadcast = "broadcast"
	// broadcastFriends is the suffix for broadcast friends stream
	broadcastFriends = "broadcast-friends"
	// like is the suffix for like stream
	like = "like"
)
