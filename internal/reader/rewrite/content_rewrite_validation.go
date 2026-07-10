// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rewrite // import "miniflux.app/v2/internal/reader/rewrite"

import "text/scanner"

type RuleErrorKind int

const (
	RuleErrLexical RuleErrorKind = iota
	RuleErrUnquote
	RuleErrOrphanArg
	RuleErrUnknownName
	RuleErrMissingArgs
)

// RuleError describes one problem found while parsing rewrite rules.
type RuleError struct {
	Pos     scanner.Position
	Kind    RuleErrorKind
	Rule    string
	Token   string
	Message string
}

// validateParsedRules checks parsed rules against knownRules for unknown names
// and missing arguments.
func validateParsedRules(rules []rule) (errs []RuleError) {
	for _, r := range rules {
		spec, ok := knownRules[r.name]
		if !ok {
			errs = append(errs, RuleError{Kind: RuleErrUnknownName, Rule: r.name,
				Message: "unknown rewrite rule name"})
			continue
		}
		if len(r.args) < spec.minArgs {
			errs = append(errs, RuleError{Kind: RuleErrMissingArgs, Rule: r.name,
				Message: "missing required arguments"})
		}
	}
	return errs
}

// ValidateRules parses rules solely to report problems. It returns nil when the
// rules are well-formed.
func ValidateRules(rulesText string) []RuleError {
	_, errs := parseRules(rulesText)
	return errs
}
