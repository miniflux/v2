// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "influxeed-engine/v2/internal/validator"

import (
	"influxeed-engine/v2/internal/locale"
	"influxeed-engine/v2/internal/model"
	"influxeed-engine/v2/internal/storage"
)

func ValidateAPIKeyCreation(store *storage.Storage, userID int64, request *model.APIKeyCreationRequest) *locale.LocalizedError {
	if request.Description == "" {
		return locale.NewLocalizedError("error.fields_mandatory")
	}

	if store.APIKeyExists(userID, request.Description) {
		return locale.NewLocalizedError("error.api_key_already_exists")
	}

	return nil
}
