// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import "miniflux.app/v2/internal/model"

// ValidateSubscriptionDiscovery validates subscription discovery requests.
func ValidateSubscriptionDiscovery(request *model.SubscriptionDiscoveryRequest) *ValidationError {
	if !IsValidURL(request.URL) {
		return NewValidationError("error.invalid_site_url")
	}

	return nil
}
