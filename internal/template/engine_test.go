// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package template // import "miniflux.app/v2/internal/template"

import "testing"

func TestParseTemplates(t *testing.T) {
	engine := NewEngine("")
	engine.ParseTemplates()
}
