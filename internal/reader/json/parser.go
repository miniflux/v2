// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package json // import "miniflux.app/v2/internal/reader/json"

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"miniflux.app/v2/internal/model"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// Parse returns a normalized feed struct from a JSON feed.
func Parse(baseURL string, data io.Reader) (*model.Feed, error) {
	reader := bufio.NewReader(data)

	// RFC 8259 allows parsers to ignore a leading UTF-8 byte order mark,
	// but encoding/json rejects it.
	if prefix, err := reader.Peek(len(utf8BOM)); err == nil && bytes.Equal(prefix, utf8BOM) {
		reader.Discard(len(utf8BOM))
	}

	jsonFeed := new(JSONFeed)
	if err := json.NewDecoder(reader).Decode(jsonFeed); err != nil {
		return nil, fmt.Errorf("json: unable to parse feed: %w", err)
	}

	return NewJSONAdapter(jsonFeed).BuildFeed(baseURL), nil
}
