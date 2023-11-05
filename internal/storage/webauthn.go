// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/go-webauthn/webauthn/webauthn"
	"miniflux.app/v2/internal/model"
)

// handle storage of webauthn credentials
func (s *Storage) AddWebAuthnCredential(userID int64, handle []byte, credential *webauthn.Credential) error {
	query := `
		INSERT INTO webauthn_credentials
			(handle, cred_id, user_id, public_key, attestation_type, aaguid, sign_count, clone_warning) 
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := s.db.Exec(
		query,
		handle,
		credential.ID,
		userID,
		credential.PublicKey,
		credential.AttestationType,
		credential.Authenticator.AAGUID,
		credential.Authenticator.SignCount,
		credential.Authenticator.CloneWarning,
	)
	return err
}

func (s *Storage) WebAuthnCredentialByHandle(handle []byte) (int64, *model.WebAuthnCredential, error) {
	var credential model.WebAuthnCredential
	var userID int64
	query := `
		SELECT
			user_id,
			cred_id,
			public_key,
			attestation_type,
			aaguid,
			sign_count,
			clone_warning,
			added_on,
			last_seen_on,
			name
		FROM
			webauthn_credentials
		WHERE
			handle = $1
	`
	var nullName sql.NullString
	err := s.db.
		QueryRow(query, handle).
		Scan(
			&userID,
			&credential.Credential.ID,
			&credential.Credential.PublicKey,
			&credential.Credential.AttestationType,
			&credential.Credential.Authenticator.AAGUID,
			&credential.Credential.Authenticator.SignCount,
			&credential.Credential.Authenticator.CloneWarning,
			&credential.AddedOn,
			&credential.LastSeenOn,
			&nullName,
		)

	if err != nil {
		return 0, nil, err
	}

	if nullName.Valid {
		credential.Name = nullName.String
	} else {
		credential.Name = ""
	}
	credential.Handle = handle
	return userID, &credential, err
}

func (s *Storage) WebAuthnCredentialsByUserID(userID int64) ([]model.WebAuthnCredential, error) {
	query := `
		SELECT
			handle,
			cred_id,
			public_key,
			attestation_type,
			aaguid,
			sign_count,
			clone_warning,
			name,
			added_on,
			last_seen_on
		FROM
			webauthn_credentials
		WHERE
			user_id = $1
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []model.WebAuthnCredential
	var nullName sql.NullString
	for rows.Next() {
		var cred model.WebAuthnCredential
		err = rows.Scan(
			&cred.Handle,
			&cred.Credential.ID,
			&cred.Credential.PublicKey,
			&cred.Credential.AttestationType,
			&cred.Credential.Authenticator.AAGUID,
			&cred.Credential.Authenticator.SignCount,
			&cred.Credential.Authenticator.CloneWarning,
			&nullName,
			&cred.AddedOn,
			&cred.LastSeenOn,
		)
		if err != nil {
			return nil, err
		}

		if nullName.Valid {
			cred.Name = nullName.String
		} else {
			cred.Name = ""
		}

		creds = append(creds, cred)
	}
	return creds, nil
}

func (s *Storage) WebAuthnSaveLogin(handle []byte) error {
	query := "UPDATE webauthn_credentials SET last_seen_on=NOW() WHERE handle=$1"
	_, err := s.db.Exec(query, handle)
	if err != nil {
		return fmt.Errorf(`store: unable to update last seen date for webauthn credential: %v`, err)
	}
	return nil
}

func (s *Storage) WebAuthnUpdateName(handle []byte, name string) error {
	query := "UPDATE webauthn_credentials SET name=$1 WHERE handle=$2"
	_, err := s.db.Exec(query, name, handle)
	if err != nil {
		return fmt.Errorf(`store: unable to update name for webauthn credential: %v`, err)
	}
	return nil
}

func (s *Storage) CountWebAuthnCredentialsByUserID(userID int64) int {
	var count int
	query := "SELECT COUNT(*) FROM webauthn_credentials WHERE user_id = $1"
	err := s.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		slog.Error("store: unable to count webauthn certs for user",
			slog.Int64("user_id", userID),
			slog.Any("error", err),
		)
		return 0
	}
	return count
}

func (s *Storage) DeleteCredentialByHandle(userID int64, handle []byte) error {
	query := "DELETE FROM webauthn_credentials WHERE user_id = $1 AND handle = $2"
	_, err := s.db.Exec(query, userID, handle)
	return err
}

func (s *Storage) DeleteAllWebAuthnCredentialsByUserID(userID int64) error {
	query := "DELETE FROM webauthn_credentials WHERE user_id = $1"
	_, err := s.db.Exec(query, userID)
	return err
}
