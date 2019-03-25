// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

import (
	"fmt"
)

// WebpushSubscription represents a WebPush subscription for a given user.
type WebpushSubscription struct {
	ID           int64
	UserID       int64
	Subscription string
}

func (u *WebpushSubscription) String() string {
	return fmt.Sprintf(`ID="%d", UserID="%d", Subscription="%s"`, u.ID, u.UserID, u.Subscription)
}

// UserWebpushSubscriptions represents a list of sessions.
type UserWebpushSubscriptions []*WebpushSubscription
