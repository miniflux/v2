// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model

import (
	"fmt"
	"time"

	"github.com/miniflux/miniflux/timezone"
)

// UserSession represents a user session in the system.
type UserSession struct {
	ID        int64
	UserID    int64
	Token     string
	CreatedAt time.Time
	UserAgent string
	IP        string
}

func (u *UserSession) String() string {
	return fmt.Sprintf(`ID="%d", UserID="%d", IP="%s", Token="%s"`, u.ID, u.UserID, u.IP, u.Token)
}

// UseTimezone converts creation date to the given timezone.
func (u *UserSession) UseTimezone(tz string) {
	u.CreatedAt = timezone.Convert(tz, u.CreatedAt)
}

// UserSessions represents a list of sessions.
type UserSessions []*UserSession

// UseTimezone converts creation date of all sessions to the given timezone.
func (u UserSessions) UseTimezone(tz string) {
	for _, session := range u {
		session.UseTimezone(tz)
	}
}
