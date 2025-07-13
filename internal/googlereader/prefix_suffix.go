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
	// readStreamSuffix is the suffix for read stream
	readStreamSuffix = "read"
	// starredStreamSuffix is the suffix for starred stream
	starredStreamSuffix = "starred"
	// readingListStreamSuffix is the suffix for reading list stream
	readingListStreamSuffix = "reading-list"
	// keptUnreadStreamSuffix is the suffix for kept unread stream
	keptUnreadStreamSuffix = "kept-unread"
	// broadcastStreamSuffix is the suffix for broadcast stream
	broadcastStreamSuffix = "broadcast"
	// broadcastFriendsStreamSuffix is the suffix for broadcast friends stream
	broadcastFriendsStreamSuffix = "broadcast-friends"
	// likeStreamSuffix is the suffix for like stream
	likeStreamSuffix = "like"
)
