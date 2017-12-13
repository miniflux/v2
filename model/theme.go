// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model

import "github.com/miniflux/miniflux/errors"

// Themes returns the list of available themes.
func Themes() map[string]string {
	return map[string]string{
		"default": "Default",
		"black":   "Black",
	}
}

// ValidateTheme validates theme value.
func ValidateTheme(theme string) error {
	for key := range Themes() {
		if key == theme {
			return nil
		}
	}

	return errors.NewLocalizedError("Invalid theme.")
}
