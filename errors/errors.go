// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package errors // import "miniflux.app/errors"

import (
	"fmt"

	"miniflux.app/locale"
)

// LocalizedError represents an error than could be translated to another language.
type LocalizedError struct {
	message string
	args    []interface{}
}

// Error returns untranslated error message.
func (l LocalizedError) Error() string {
	return fmt.Sprintf(l.message, l.args...)
}

// Localize returns the translated error message.
func (l LocalizedError) Localize(printer *locale.Printer) string {
	return printer.Printf(l.message, l.args...)
}

// NewLocalizedError returns a new LocalizedError.
func NewLocalizedError(message string, args ...interface{}) *LocalizedError {
	return &LocalizedError{message: message, args: args}
}
