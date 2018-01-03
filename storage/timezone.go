// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage

import (
	"fmt"
	"time"

	"github.com/miniflux/miniflux/timer"
)

// Timezones returns all timezones supported by the database.
func (s *Storage) Timezones() (map[string]string, error) {
	defer timer.ExecutionTime(time.Now(), "[Storage:Timezones]")

	timezones := make(map[string]string)
	query := `select name from pg_timezone_names() order by name asc`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch timezones: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var timezone string
		if err := rows.Scan(&timezone); err != nil {
			return nil, fmt.Errorf("unable to fetch timezones row: %v", err)
		}

		timezones[timezone] = timezone
	}

	return timezones, nil
}
