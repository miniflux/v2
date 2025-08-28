// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

import "strconv"

func OptionalField[T any](value T) *T {
	return &value
}

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

func OptionalInt64(value string) *int64 {
	if value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return &intValue
		}
	}
	return nil
}
