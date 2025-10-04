// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"errors"

	"miniflux.app/v2/internal/model"
)

func ValidateEnclosureUpdateRequest(request *model.EnclosureUpdateRequest) error {
	if request.MediaProgression < 0 {
		return errors.New(`media progression must an positive integer`)
	}

	return nil
}
