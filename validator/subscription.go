// Copyright 2021 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package validator // import "miniflux.app/validator"

import "miniflux.app/model"

// ValidateSubscriptionDiscovery validates subscription discovery requests.
func ValidateSubscriptionDiscovery(request *model.SubscriptionDiscoveryRequest) *ValidationError {
	if !isValidURL(request.URL) {
		return NewValidationError("error.invalid_site_url")
	}

	return nil
}
