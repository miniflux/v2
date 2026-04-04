// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package oauth2 // import "miniflux.app/v2/internal/oauth2"

import (
	"fmt"
)

// UserProfile represents a user's profile retrieved from an OAuth2 provider.
type UserProfile struct {
	Key      string
	ID       string
	Username string
}

// String returns a formatted string representation of the user profile.
func (p UserProfile) String() string {
	return fmt.Sprintf(`Key=%s ; ID=%s ; Username=%s`, p.Key, p.ID, p.Username)
}
