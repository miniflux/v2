// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/ui/form"

import (
	"net/http"

	"miniflux.app/errors"
)

// APIKeyForm represents the API Key form.
type APIKeyForm struct {
	Description string
}

// Validate makes sure the form values are valid.
func (a APIKeyForm) Validate() error {
	if a.Description == "" {
		return errors.NewLocalizedError("error.fields_mandatory")
	}

	return nil
}

// NewAPIKeyForm returns a new APIKeyForm.
func NewAPIKeyForm(r *http.Request) *APIKeyForm {
	return &APIKeyForm{
		Description: r.FormValue("description"),
	}
}
