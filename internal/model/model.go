// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

// OptionalString populates an optional string field.
func OptionalString(value string) *string {
	if value != "" {
		return &value
	}
	return nil
}

// OptionalInt populates an optional int field.
func OptionalInt(value int) *int {
	if value > 0 {
		return &value
	}
	return nil
}

// OptionalInt64 populates an optional int64 field.
func OptionalInt64(value int64) *int64 {
	if value > 0 {
		return &value
	}
	return nil
}
