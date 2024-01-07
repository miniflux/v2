// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import (
	"net/http"

	"miniflux.app/v2/internal/locale"
)

// APIKeyForm represents the API Key form.
type APIKeyForm struct {
	Description string
}

// Validate makes sure the form values are valid.
func (a APIKeyForm) Validate() *locale.LocalizedError {
	if a.Description == "" {
		return locale.NewLocalizedError("error.fields_mandatory")
	}

	return nil
}

// NewAPIKeyForm returns a new APIKeyForm.
func NewAPIKeyForm(r *http.Request) *APIKeyForm {
	return &APIKeyForm{
		Description: r.FormValue("description"),
	}
}
