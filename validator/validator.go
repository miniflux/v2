// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/validator"

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"

	"miniflux.app/locale"
)

// ValidationError represents a validation error.
type ValidationError struct {
	TranslationKey string
}

// NewValidationError initializes a validation error.
func NewValidationError(translationKey string) *ValidationError {
	return &ValidationError{TranslationKey: translationKey}
}

func (v *ValidationError) String() string {
	return locale.NewPrinter("en_US").Printf(v.TranslationKey)
}

func (v *ValidationError) Error() error {
	return errors.New(v.String())
}

// ValidateRange makes sure the offset/limit values are valid.
func ValidateRange(offset, limit int) error {
	if offset < 0 {
		return fmt.Errorf(`Offset value should be >= 0`)
	}

	if limit < 0 {
		return fmt.Errorf(`Limit value should be >= 0`)
	}

	return nil
}

// ValidateDirection makes sure the sorting direction is valid.
func ValidateDirection(direction string) error {
	switch direction {
	case "asc", "desc":
		return nil
	}

	return fmt.Errorf(`Invalid direction, valid direction values are: "asc" or "desc"`)
}

// IsValidRegex verifies if the regex can be compiled.
func IsValidRegex(expr string) bool {
	_, err := regexp.Compile(expr)
	return err == nil
}

// IsValidURL verifies if the provided value is a valid absolute URL.
func IsValidURL(absoluteURL string) bool {
	_, err := url.ParseRequestURI(absoluteURL)
	return err == nil
}
