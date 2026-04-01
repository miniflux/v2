// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"


// Get the VAPID keys from the database
func (s *Storage) GetVAPIDKeys() (string, string, error) {

	var privateKey string
	var publicKey string
	query := `SELECT private_key, public_key FROM vapid_key`

	err := s.db.QueryRow(query).Scan(&privateKey, &publicKey)

	return privateKey, publicKey, err
}
