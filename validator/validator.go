// Copyright 2021 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package validator // import "miniflux.app/validator"

import (
	"errors"
	"fmt"
	"net/url"

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

func isValidURL(absoluteURL string) bool {
	_, err := url.ParseRequestURI(absoluteURL)
	return err == nil
}
