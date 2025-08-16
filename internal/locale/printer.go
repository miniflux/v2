// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package locale // import "miniflux.app/v2/internal/locale"

import "fmt"

// Printer converts translation keys to language-specific strings.
type Printer struct {
	language string
}

// NewPrinter creates a new Printer instance for the given language.
func NewPrinter(language string) *Printer {
	return &Printer{language}
}

func (p *Printer) Print(key string) string {
	if dict, err := getTranslationDict(p.language); err == nil {
		if str, ok := dict.singulars[key]; ok {
			return str
		}
	}
	return key
}

// Printf is like fmt.Printf, but using language-specific formatting.
func (p *Printer) Printf(key string, args ...any) string {
	return fmt.Sprintf(p.Print(key), args...)
}

// Plural returns the translation of the given key by using the language plural form.
func (p *Printer) Plural(key string, n int, args ...any) string {
	dict, err := getTranslationDict(p.language)
	if err != nil {
		return key
	}

	if choices, found := dict.plurals[key]; found {
		index := getPluralForm(p.language, n)
		if len(choices) > index {
			return fmt.Sprintf(choices[index], args...)
		}
	}

	return key
}
