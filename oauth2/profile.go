// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package oauth2 // import "miniflux.app/oauth2"

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
	return fmt.Sprintf(`ID=%s ; Username=%s`, p.ID, p.Username)
}
