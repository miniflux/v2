// Copyright 2021 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

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
