// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/reader/rewrite"
)

// IsValidURLRewriteRules reports whether the feed URL rewrite rules are
// well-formed. Unlike a bare regex check, it verifies the rule matches the
// rewrite("search"|"replacement") form that RewriteEntryURL actually applies,
// so malformed rules are rejected at save time rather than silently ignored.
func IsValidURLRewriteRules(rules string) bool {
	return rewrite.IsValidURLRewriteRules(rules)
}

// IsValidRewriteRules validates content rewrite rules, returning a localized
// error describing the first problem found, or nil when the rules are valid.
func IsValidRewriteRules(rules string) *locale.LocalizedError {
	if rules == "" {
		return nil
	}
	errs := rewrite.ValidateRules(rules)
	if len(errs) == 0 {
		return nil
	}
	e := errs[0]
	switch e.Kind {
	case rewrite.RuleErrUnknownName:
		return locale.NewLocalizedError("error.feed_rewrite_rule_unknown_name", e.Rule)
	case rewrite.RuleErrMissingArgs:
		return locale.NewLocalizedError("error.feed_rewrite_rule_missing_args", e.Rule)
	case rewrite.RuleErrUnquote, rewrite.RuleErrLexical:
		// Lexical (scanner) errors carry a descriptive Message but no token,
		// while unquote errors carry the offending token; show whichever we have.
		detail := e.Token
		if detail == "" {
			detail = e.Message
		}
		return locale.NewLocalizedError("error.feed_rewrite_rule_invalid_syntax", e.Pos.Column, detail)
	default:
		return locale.NewLocalizedError("error.feed_invalid_rewrite_rule")
	}
}
