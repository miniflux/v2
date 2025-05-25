// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import (
	"net/http"
	"strings"
)

// APIKeyForm represents the API Key form.
type APIKeyForm struct {
	Description string
}

// NewAPIKeyForm returns a new APIKeyForm.
func NewAPIKeyForm(r *http.Request) *APIKeyForm {
	return &APIKeyForm{
		Description: strings.TrimSpace(r.FormValue("description")),
	}
}
