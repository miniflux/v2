// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"errors"

	"miniflux.app/v2/internal/model"
)

// ValidateEnclosureUpdateRequest validates enclosure updates, ensuring media progression is not negative.
func ValidateEnclosureUpdateRequest(request *model.EnclosureUpdateRequest) error {
	if request.MediaProgression < 0 {
		return errors.New(`media progression must be a non-negative integer`)
	}

	return nil
}
