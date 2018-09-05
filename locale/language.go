// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package locale // import "miniflux.app/locale"

import "fmt"

// Language represents a language in the system.
type Language struct {
	language     string
	translations Translation
}

// Get fetch the translation for the given key.
func (l *Language) Get(key string, args ...interface{}) string {
	var translation string

	str, found := l.translations[key]
	if !found {
		translation = key
	} else {
		translation = str.(string)
	}

	return fmt.Sprintf(translation, args...)
}

// Plural returns the translation of the given key by using the language plural form.
func (l *Language) Plural(key string, n int, args ...interface{}) string {
	translation := key
	slices, found := l.translations[key]

	if found {
		pluralForm, found := pluralForms[l.language]
		if !found {
			pluralForm = pluralForms["default"]
		}

		index := pluralForm(n)
		translations := slices.([]interface{})
		translation = key

		if len(translations) > index {
			translation = translations[index].(string)
		}
	}

	return fmt.Sprintf(translation, args...)
}
