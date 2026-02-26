// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package validator // import "miniflux.app/v2/internal/validator"

import (
	"testing"

	"miniflux.app/v2/internal/model"
)

func TestValidateEnclosureUpdateRequest(t *testing.T) {
	request := &model.EnclosureUpdateRequest{MediaProgression: -1}
	if err := ValidateEnclosureUpdateRequest(request); err == nil {
		t.Error("A negative media progression should generate an error")
	}

	request.MediaProgression = 0
	if err := ValidateEnclosureUpdateRequest(request); err != nil {
		t.Fatalf("Zero media progression should be accepted: %v", err)
	}

	request.MediaProgression = 42
	if err := ValidateEnclosureUpdateRequest(request); err != nil {
		t.Fatalf("Positive media progression should be accepted: %v", err)
	}
}
