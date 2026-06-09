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
	return formatTranslation(p.Print(key), args...)
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
			return formatTranslation(choices[index], args...)
		}
	}

	return key
}

// formatTranslation skips extra arguments when the translation references no argument,
// so plural forms that omit the count (e.g. the Arabic dual "دقيقتين") don't get
// a trailing %!(EXTRA ...) marker. Escaped percents are still processed by fmt.
func formatTranslation(format string, args ...any) string {
	if !hasFormattingDirective(format) {
		return fmt.Sprintf(format, []any{}...)
	}
	return fmt.Sprintf(format, args...)
}

// hasFormattingDirective reports whether the format should be handled with the
// supplied arguments. It treats "%%" as a literal percent and lets fmt validate
// any other percent sequence, including a dangling "%".
func hasFormattingDirective(format string) bool {
	for index := 0; index < len(format); index++ {
		if format[index] != '%' {
			continue
		}
		if index+1 >= len(format) {
			return true
		}
		if format[index+1] == '%' {
			index++ // skip the escaped percent
			continue
		}
		return true
	}
	return false
}
