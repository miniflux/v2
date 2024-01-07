// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import (
	"net/http"
)

// WebauthnForm represents a credential rename form in the UI
type WebauthnForm struct {
	Name string
}

// NewWebauthnForm returns a new WebnauthnForm.
func NewWebauthnForm(r *http.Request) *WebauthnForm {
	return &WebauthnForm{
		Name: r.FormValue("name"),
	}
}
