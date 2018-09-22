// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

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
