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

// AddWebAuthnCredential handles storage of webauthn credentials.
func (s *Storage) AddWebAuthnCredential(userID int64, handle []byte, credential *webauthn.Credential) error {
	query := `
		INSERT INTO webauthn_credentials
			(handle, cred_id, user_id, public_key, attestation_type, aaguid, sign_count, clone_warning, backup_eligible, backup_state)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
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
		credential.Flags.BackupEligible,
		credential.Flags.BackupState,
	)
	return err
}

func (s *Storage) WebAuthnCredentialByHandle(handle []byte) (int64, *model.WebAuthnCredential, error) {
	var credential model.WebAuthnCredential
	var userID int64
	var backupEligible sql.NullBool
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
			name,
			backup_eligible,
			backup_state
		FROM
			webauthn_credentials
		WHERE
			handle = $1
	`
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
			&credential.Name,
			&backupEligible,
			&credential.Credential.Flags.BackupState,
		)

	if err != nil {
		return 0, nil, err
	}

	if backupEligible.Valid {
		credential.Credential.Flags.BackupEligible = backupEligible.Bool
		credential.BackupEligibleKnown = true
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
			last_seen_on,
			backup_eligible,
			backup_state
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
	for rows.Next() {
		var cred model.WebAuthnCredential
		var backupEligible sql.NullBool
		err = rows.Scan(
			&cred.Handle,
			&cred.Credential.ID,
			&cred.Credential.PublicKey,
			&cred.Credential.AttestationType,
			&cred.Credential.Authenticator.AAGUID,
			&cred.Credential.Authenticator.SignCount,
			&cred.Credential.Authenticator.CloneWarning,
			&cred.Name,
			&cred.AddedOn,
			&cred.LastSeenOn,
			&backupEligible,
			&cred.Credential.Flags.BackupState,
		)
		if err != nil {
			return nil, err
		}

		if backupEligible.Valid {
			cred.Credential.Flags.BackupEligible = backupEligible.Bool
			cred.BackupEligibleKnown = true
		}

		creds = append(creds, cred)
	}
	return creds, nil
}

// WebAuthnSaveLogin writes back the per-assertion fields (sign count, clone warning, backup state, BE) the WebAuthn spec requires after every successful login.
func (s *Storage) WebAuthnSaveLogin(handle []byte, credential *webauthn.Credential) error {
	query := `
		UPDATE webauthn_credentials
		SET last_seen_on = NOW(),
			sign_count = $1,
			clone_warning = $2,
			backup_eligible = $3,
			backup_state = $4
		WHERE handle = $5
	`
	_, err := s.db.Exec(
		query,
		credential.Authenticator.SignCount,
		credential.Authenticator.CloneWarning,
		credential.Flags.BackupEligible,
		credential.Flags.BackupState,
		handle,
	)
	if err != nil {
		return fmt.Errorf(`store: unable to update webauthn credential after login: %v`, err)
	}
	return nil
}

func (s *Storage) WebAuthnUpdateName(userID int64, handle []byte, name string) (int64, error) {
	query := "UPDATE webauthn_credentials SET name=$1 WHERE handle=$2 AND user_id=$3"
	result, err := s.db.Exec(query, name, handle, userID)
	if err != nil {
		return 0, fmt.Errorf(`store: unable to update name for webauthn credential: %v`, err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf(`store: unable to update name for webauthn credential: %v`, err)
	}
	return rows, nil
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
