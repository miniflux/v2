// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage // import "miniflux.app/storage"

import (
	"fmt"
	"strings"
)

// Timezones returns all timezones supported by the database.
func (s *Storage) Timezones() (map[string]string, error) {
	timezones := make(map[string]string)
	rows, err := s.db.Query(`SELECT name FROM pg_timezone_names() ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch timezones: %v`, err)
	}
	defer rows.Close()

	for rows.Next() {
		var timezone string
		if err := rows.Scan(&timezone); err != nil {
			return nil, fmt.Errorf(`store: unable to fetch timezones row: %v`, err)
		}

		if !strings.HasPrefix(timezone, "posix") && !strings.HasPrefix(timezone, "SystemV") && timezone != "localtime" {
			timezones[timezone] = timezone
		}
	}

	return timezones, nil
}
