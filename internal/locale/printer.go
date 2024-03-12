// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package locale // import "miniflux.app/v2/internal/locale"

import "fmt"

// Printer converts translation keys to language-specific strings.
type Printer struct {
	language string
}

func (p *Printer) Print(key string) string {
	if str, ok := defaultCatalog[p.language][key]; ok {
		if translation, ok := str.(string); ok {
			return translation
		}
	}
	return key
}

// Printf is like fmt.Printf, but using language-specific formatting.
func (p *Printer) Printf(key string, args ...interface{}) string {
	var translation string

	str, found := defaultCatalog[p.language][key]
	if !found {
		translation = key
	} else {
		var valid bool
		translation, valid = str.(string)
		if !valid {
			translation = key
		}
	}

	return fmt.Sprintf(translation, args...)
}

// Plural returns the translation of the given key by using the language plural form.
func (p *Printer) Plural(key string, n int, args ...interface{}) string {
	choices, found := defaultCatalog[p.language][key]

	if found {
		var plurals []string

		switch v := choices.(type) {
		case []interface{}:
			for _, v := range v {
				plurals = append(plurals, fmt.Sprint(v))
			}
		case []string:
			plurals = v
		default:
			return key
		}

		pluralForm, found := pluralForms[p.language]
		if !found {
			pluralForm = pluralForms["default"]
		}

		index := pluralForm(n)
		if len(plurals) > index {
			return fmt.Sprintf(plurals[index], args...)
		}
	}

	return key
}

// NewPrinter creates a new Printer.
func NewPrinter(language string) *Printer {
	return &Printer{language}
}
