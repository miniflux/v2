// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

import (
	"fmt"
	"strings"
)

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

func getStream(streamID string, userID int64) (Stream, error) {
	switch {
	case strings.HasPrefix(streamID, feedPrefix):
		return Stream{Type: FeedStream, ID: strings.TrimPrefix(streamID, feedPrefix)}, nil
	case strings.HasPrefix(streamID, fmt.Sprintf(userStreamPrefix, userID)), strings.HasPrefix(streamID, streamPrefix):
		id := strings.TrimPrefix(streamID, fmt.Sprintf(userStreamPrefix, userID))
		id = strings.TrimPrefix(id, streamPrefix)
		switch id {
		case readStreamSuffix:
			return Stream{ReadStream, ""}, nil
		case starredStreamSuffix:
			return Stream{StarredStream, ""}, nil
		case readingListStreamSuffix:
			return Stream{ReadingListStream, ""}, nil
		case keptUnreadStreamSuffix:
			return Stream{KeptUnreadStream, ""}, nil
		case broadcastStreamSuffix:
			return Stream{BroadcastStream, ""}, nil
		case broadcastFriendsStreamSuffix:
			return Stream{BroadcastFriendsStream, ""}, nil
		case likeStreamSuffix:
			return Stream{LikeStream, ""}, nil
		default:
			return Stream{NoStream, ""}, fmt.Errorf("googlereader: unknown stream with id: %s", id)
		}
	case strings.HasPrefix(streamID, fmt.Sprintf(userLabelPrefix, userID)), strings.HasPrefix(streamID, labelPrefix):
		id := strings.TrimPrefix(streamID, fmt.Sprintf(userLabelPrefix, userID))
		id = strings.TrimPrefix(id, labelPrefix)
		return Stream{LabelStream, id}, nil
	case streamID == "":
		return Stream{NoStream, ""}, nil
	default:
		return Stream{NoStream, ""}, fmt.Errorf("googlereader: unknown stream type: %s", streamID)
	}
}

func getStreams(streamIDs []string, userID int64) ([]Stream, error) {
	streams := make([]Stream, 0, len(streamIDs))
	for _, streamID := range streamIDs {
		stream, err := getStream(streamID, userID)
		if err != nil {
			return []Stream{}, err
		}
		streams = append(streams, stream)
	}
	return streams, nil
}
