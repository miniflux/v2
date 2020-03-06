// Copyright 2020 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

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
