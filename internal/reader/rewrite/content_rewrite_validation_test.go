// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rewrite // import "miniflux.app/v2/internal/reader/rewrite"

import (
	"testing"
)

func TestParseRulesInvalidEscape(t *testing.T) {
	rulesText := `replace("\d+"|"x")`
	_, errs := parseRules(rulesText)

	if len(errs) == 0 {
		t.Fatalf(`Expected at least one error, got none`)
	}

	var found bool
	for _, e := range errs {
		if e.Kind == RuleErrUnquote || e.Kind == RuleErrLexical {
			found = true
		}
	}
	if !found {
		t.Errorf(`Expected an error with Kind RuleErrUnquote or RuleErrLexical, got %v`, errs)
	}
}

func TestParseRulesUnknownName(t *testing.T) {
	rulesText := `replce("a"|"b")`
	_, errs := parseRules(rulesText)

	var found bool
	for _, e := range errs {
		if e.Kind == RuleErrUnknownName && e.Rule == "replce" {
			found = true
		}
	}
	if !found {
		t.Errorf(`Expected an error with Kind RuleErrUnknownName and Rule "replce", got %v`, errs)
	}
}

func TestParseRulesMissingArgs(t *testing.T) {
	rulesText := `replace("a")`
	_, errs := parseRules(rulesText)

	var found bool
	for _, e := range errs {
		if e.Kind == RuleErrMissingArgs && e.Rule == "replace" {
			found = true
		}
	}
	if !found {
		t.Errorf(`Expected an error with Kind RuleErrMissingArgs and Rule "replace", got %v`, errs)
	}
}

func TestParseRulesWellFormed(t *testing.T) {
	rulesText := `replace("a"|"b") nl2br`
	rules, errs := parseRules(rulesText)

	if len(errs) != 0 {
		t.Fatalf(`Expected zero errors, got %v`, errs)
	}

	if len(rules) != 2 {
		t.Fatalf(`Expected 2 rules, got %d: %v`, len(rules), rules)
	}

	if rules[0].name != "replace" || len(rules[0].args) != 2 || rules[0].args[0] != "a" || rules[0].args[1] != "b" {
		t.Errorf(`Unexpected first rule: %v`, rules[0])
	}

	if rules[1].name != "nl2br" || len(rules[1].args) != 0 {
		t.Errorf(`Unexpected second rule: %v`, rules[1])
	}
}
