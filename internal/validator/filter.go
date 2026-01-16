// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"slices"
	"strings"

	"miniflux.app/v2/internal/locale"
)

func isValidFilterRules(filterEntryRules string, filterType string) *locale.LocalizedError {
	// Valid Format: FieldName=RegEx\nFieldName=RegEx...
	fieldNames := []string{"EntryTitle", "EntryURL", "EntryCommentsURL", "EntryContent", "EntryAuthor", "EntryTag", "EntryDate"}

	rules := strings.Split(filterEntryRules, "\n")
	for i, rule := range rules {
		// Check if rule starts with a valid fieldName
		idx := slices.IndexFunc(fieldNames, func(fieldName string) bool { return strings.HasPrefix(rule, fieldName) })
		if idx == -1 {
			return locale.NewLocalizedError("error.settings_"+filterType+"_rule_fieldname_invalid", i+1, "'"+strings.Join(fieldNames, "', '")+"'")
		}
		fieldName := fieldNames[idx]
		fieldRegEx, _ := strings.CutPrefix(rule, fieldName)

		// Check if regex begins with a =
		if !strings.HasPrefix(fieldRegEx, "=") {
			return locale.NewLocalizedError("error.settings_"+filterType+"_rule_separator_required", i+1)
		}
		fieldRegEx = strings.TrimPrefix(fieldRegEx, "=")

		if fieldRegEx == "" {
			return locale.NewLocalizedError("error.settings_"+filterType+"_rule_regex_required", i+1)
		}

		// Check if provided pattern is a valid RegEx
		if !IsValidRegex(fieldRegEx) {
			return locale.NewLocalizedError("error.settings_"+filterType+"_rule_invalid_regex", i+1)
		}
	}
	return nil
}
