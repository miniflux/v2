// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package json // import "miniflux.app/reader/json"

import (
	"encoding/json"
	"io"

	"miniflux.app/errors"
	"miniflux.app/model"
)

// Parse returns a normalized feed struct from a JSON feed.
func Parse(baseURL string, data io.Reader) (*model.Feed, *errors.LocalizedError) {
	feed := new(jsonFeed)
	decoder := json.NewDecoder(data)
	if err := decoder.Decode(&feed); err != nil {
		return nil, errors.NewLocalizedError("Unable to parse JSON Feed: %q", err)
	}

	return feed.Transform(baseURL), nil
}
