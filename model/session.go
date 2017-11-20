// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model

import "time"
import "fmt"

type Session struct {
	ID        int64
	UserID    int64
	Token     string
	CreatedAt time.Time
	UserAgent string
	IP        string
}

func (s *Session) String() string {
	return fmt.Sprintf("ID=%d, UserID=%d, IP=%s", s.ID, s.UserID, s.IP)
}

type Sessions []*Session
