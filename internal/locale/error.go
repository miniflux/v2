// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package locale // import "miniflux.app/v2/internal/locale"

import "errors"

type LocalizedErrorWrapper struct {
	originalErr     error
	translationKey  string
	translationArgs []any
}

func NewLocalizedErrorWrapper(originalErr error, translationKey string, translationArgs ...any) *LocalizedErrorWrapper {
	return &LocalizedErrorWrapper{
		originalErr:     originalErr,
		translationKey:  translationKey,
		translationArgs: translationArgs,
	}
}

func (l *LocalizedErrorWrapper) Error() error {
	return l.originalErr
}

func (l *LocalizedErrorWrapper) Translate(language string) string {
	if l.translationKey == "" {
		return l.originalErr.Error()
	}
	return NewPrinter(language).Printf(l.translationKey, l.translationArgs...)
}

type LocalizedError struct {
	translationKey  string
	translationArgs []any
}

func NewLocalizedError(translationKey string, translationArgs ...any) *LocalizedError {
	return &LocalizedError{translationKey: translationKey, translationArgs: translationArgs}
}

func (v *LocalizedError) String() string {
	return NewPrinter("en_US").Printf(v.translationKey, v.translationArgs...)
}

func (v *LocalizedError) Error() error {
	return errors.New(v.String())
}

func (v *LocalizedError) Translate(language string) string {
	return NewPrinter(language).Printf(v.translationKey, v.translationArgs...)
}
