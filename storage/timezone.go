// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage

import (
	"fmt"
	"github.com/miniflux/miniflux2/helper"
	"time"
)

func (s *Storage) GetTimezones() (map[string]string, error) {
	defer helper.ExecutionTime(time.Now(), "[Storage:GetTimezones]")

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
