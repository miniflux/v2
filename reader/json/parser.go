// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package json

import (
	"encoding/json"
	"io"

	"github.com/miniflux/miniflux/errors"
	"github.com/miniflux/miniflux/model"
)

// Parse returns a normalized feed struct from a JON feed.
func Parse(data io.Reader) (*model.Feed, *errors.LocalizedError) {
	feed := new(jsonFeed)
	decoder := json.NewDecoder(data)
	if err := decoder.Decode(&feed); err != nil {
		return nil, errors.NewLocalizedError("Unable to parse JSON Feed: %q", err)
	}

	return feed.Transform(), nil
}
