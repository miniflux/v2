// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model

// GetThemes returns the list of available themes.
func GetThemes() map[string]string {
	return map[string]string{
		"default": "Default",
		"black":   "Black",
	}
}
