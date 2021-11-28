// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

// Themes returns the list of available themes.
func Themes() map[string]string {
	return map[string]string{
		"light_serif":       "Light - Serif",
		"light_sans_serif":  "Light - Sans Serif",
		"dark_serif":        "Dark - Serif",
		"dark_sans_serif":   "Dark - Sans Serif",
		"system_serif":      "System - Serif",
		"system_sans_serif": "System - Sans Serif",
	}
}

// ThemeColor returns the color for the address bar or/and the browser color.
// https://developer.mozilla.org/en-US/docs/Web/Manifest#theme_color
// https://developers.google.com/web/tools/lighthouse/audits/address-bar
// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/meta/name/theme-color
func ThemeColor(theme, colorScheme string) string {
	switch theme {
	case "dark_serif", "dark_sans_serif":
		return "#222"
	case "system_serif", "system_sans_serif":
		if colorScheme == "dark" {
			return "#222"
		}

		return "#fff"
	default:
		return "#fff"
	}
}
