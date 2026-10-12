// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"github.com/andybalholm/cascadia"

	"miniflux.app/v2/internal/locale"
)

// IsValidScraperRules validates scraper rules, which are a CSS selector. The
// scraper silently extracts nothing when the selector does not compile, so an
// invalid one is rejected at save time instead.
func IsValidScraperRules(rules string) *locale.LocalizedError {
	if rules == "" {
		return nil
	}
	if _, err := cascadia.Compile(rules); err != nil {
		return locale.NewLocalizedError("error.feed_invalid_scraper_rule", err.Error())
	}
	return nil
}
