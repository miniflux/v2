// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

type Number interface {
	int | int64 | float64
}

func OptionalNumber[T Number](value T) *T {
	if value > 0 {
		return &value
	}
	return nil
}

func OptionalString(value string) *string {
	if value != "" {
		return &value
	}
	return nil
}
