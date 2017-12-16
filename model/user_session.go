// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model

import "time"
import "fmt"

// UserSession represents a user session in the system.
type UserSession struct {
	ID        int64
	UserID    int64
	Token     string
	CreatedAt time.Time
	UserAgent string
	IP        string
}

func (s *UserSession) String() string {
	return fmt.Sprintf(`ID="%d", UserID="%d", IP="%s", Token="%s"`, s.ID, s.UserID, s.IP, s.Token)
}

// UserSessions represents a list of sessions.
type UserSessions []*UserSession
