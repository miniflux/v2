// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import (
	"net/http"
	"strings"

	"miniflux.app/v2/internal/locale"
)

// WebauthnForm represents a credential rename form in the UI
type WebauthnForm struct {
	Name string
}

// Validate makes sure the form values are valid.
func (f *WebauthnForm) Validate() *locale.LocalizedError {
	if f.Name == "" {
		return locale.NewLocalizedError("error.fields_mandatory")
	}
	return nil
}

// NewWebauthnForm returns a new WebnauthnForm.
func NewWebauthnForm(r *http.Request) *WebauthnForm {
	return &WebauthnForm{
		Name: strings.TrimSpace(r.FormValue("name")),
	}
}
