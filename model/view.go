// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

import "miniflux.app/errors"

// Views returns the list of available views.
func Views() map[string]string {
	return map[string]string{
		"list":    "List",
		"masonry": "Masonry",
	}
}

// ValidateView validates view value.
func ValidateView(view string) error {
	for key := range Views() {
		if key == view {
			return nil
		}
	}

	return errors.NewLocalizedError("Invalid view")
}
