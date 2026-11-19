// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rewrite // import "miniflux.app/v2/internal/reader/rewrite"

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"text/scanner"

	"github.com/andybalholm/cascadia"
)

const (
	minYouTubePlaybackRate = 0.25
	maxYouTubePlaybackRate = 2
)

type RuleErrorKind int

const (
	RuleErrLexical RuleErrorKind = iota
	RuleErrUnquote
	RuleErrOrphanArg
	RuleErrUnknownName
	RuleErrMissingArgs
	RuleErrInvalidRegexp
	RuleErrInvalidSelector
	RuleErrInvalidPlaybackRate
)

func parseYouTubePlaybackRate(value string) (float64, error) {
	rate, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid playback rate: %w", err)
	}
	if math.IsNaN(rate) || math.IsInf(rate, 0) {
		return 0, fmt.Errorf("invalid playback rate: %q", value)
	}
	if rate < minYouTubePlaybackRate || rate > maxYouTubePlaybackRate {
		return 0, fmt.Errorf("playback rate must be between %v and %v", minYouTubePlaybackRate, maxYouTubePlaybackRate)
	}
	return rate, nil
}

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
			continue
		}
		if spec.regexArg && len(r.args) > 0 {
			if _, err := regexp.Compile(r.args[0]); err != nil {
				errs = append(errs, RuleError{Kind: RuleErrInvalidRegexp, Rule: r.name,
					Token: r.args[0], Message: err.Error()})
			}
		}
		// goquery silently matches nothing when a selector does not compile,
		// so an invalid selector would turn the rule into a no-op.
		if spec.selectorArg && len(r.args) > 0 {
			if _, err := cascadia.Compile(r.args[0]); err != nil {
				errs = append(errs, RuleError{Kind: RuleErrInvalidSelector, Rule: r.name,
					Token: r.args[0], Message: err.Error()})
			}
		}
		if spec.playbackRateArg && len(r.args) > 0 {
			if _, err := parseYouTubePlaybackRate(r.args[0]); err != nil {
				errs = append(errs, RuleError{Kind: RuleErrInvalidPlaybackRate, Rule: r.name,
					Token: r.args[0], Message: err.Error()})
			}
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
