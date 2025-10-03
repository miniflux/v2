// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"miniflux.app/v2/internal/botauth"
	"miniflux.app/v2/internal/crypto"
)

func (s *Storage) CreateWebAuthBothKeys() error {
	privateKey, publicKey, err := crypto.GenerateEd25519Keys()
	if err != nil {
		return err
	}

	query := `INSERT INTO web_bot_auth (private_key, public_key) VALUES ($1, $2)`
	if _, err := s.db.Exec(query, privateKey, publicKey); err != nil {
		return err
	}

	return nil
}

func (s *Storage) WebAuthBothKeys() (keyPairs botauth.KeyPairs, err error) {
	query := `SELECT private_key, public_key FROM web_bot_auth ORDER BY created_at DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var privateKey, publicKey []byte
		if err := rows.Scan(&privateKey, &publicKey); err != nil {
			return nil, err
		}
		keyPair, err := botauth.NewKeyPair(privateKey, publicKey)
		if err != nil {
			return nil, err
		}
		keyPairs = append(keyPairs, keyPair)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return keyPairs, nil
}

func (s *Storage) HasWebAuthBothKeys() (bool, error) {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM web_bot_auth`).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}
