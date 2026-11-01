// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/v2/internal/googlereader"

import "testing"

func TestStreamTypeString(t *testing.T) {
	scenarios := map[StreamType]string{
		NoStream:       "NoStream",
		FeedStream:     "FeedStream",
		LikeStream:     "LikeStream",
		StreamType(42): "StreamType(42)",
		StreamType(-1): "StreamType(-1)",
	}

	for streamType, expected := range scenarios {
		if result := streamType.String(); result != expected {
			t.Errorf(`Unexpected result for stream type %d, got %q instead of %q`, int(streamType), result, expected)
		}
	}
}
