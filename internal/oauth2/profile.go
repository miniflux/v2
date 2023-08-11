// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package oauth2 // import "miniflux.app/v2/internal/oauth2"

import (
	"fmt"
)

// Profile is the OAuth2 user profile.
type Profile struct {
	Key      string
	ID       string
	Username string
}

func (p Profile) String() string {
	return fmt.Sprintf(`Key=%s ; ID=%s ; Username=%s`, p.Key, p.ID, p.Username)
}
