package storage // import "miniflux.app/storage"

import (
	"github.com/duo-labs/webauthn/webauthn"
	"miniflux.app/logger"
)

// public_key bytea not null,
// attestation_type varchar(255) not null,
// aaguid bytea,
// sign_count bigint,
// clone_warning bool

// handle storage of webauthn credentials
func (s *Storage) AddWebAuthnCredential(userID int64, handle []byte, credential *webauthn.Credential) error {
	query := `
		insert into webauthn_credentials 
			(handle, cred_id, user_id, public_key, attestation_type, aaguid, sign_count, clone_warning) 
			values ($1, $2, $3, $4, $5, $6, $7, $8)
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

func (s *Storage) WebAuthnCredentialByHandle(handle []byte) (int64, *webauthn.Credential, error) {
	var credential webauthn.Credential
	var userID int64
	query := "select user_id, cred_id, public_key, attestation_type, aaguid, sign_count, clone_warning from webauthn_credentials where handle = $1"
	err := s.db.
		QueryRow(query, handle).
		Scan(
			&userID,
			&credential.ID,
			&credential.PublicKey,
			&credential.AttestationType,
			&credential.Authenticator.AAGUID,
			&credential.Authenticator.SignCount,
			&credential.Authenticator.CloneWarning,
		)
	return userID, &credential, err
}

func (s *Storage) WebAuthnCredentialsByUserID(userID int64) ([]webauthn.Credential, error) {
	query := "select cred_id, public_key, attestation_type, aaguid, sign_count, clone_warning from webauthn_credentials where user_id = $1"
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []webauthn.Credential
	for rows.Next() {
		var cred webauthn.Credential
		err = rows.Scan(
			&cred.ID,
			&cred.PublicKey,
			&cred.AttestationType,
			&cred.Authenticator.AAGUID,
			&cred.Authenticator.SignCount,
			&cred.Authenticator.CloneWarning,
		)
		if err != nil {
			return nil, err
		}
		creds = append(creds, cred)
	}
	return creds, nil
}

func (s *Storage) CountWebAuthnCredentialsByUserID(userID int64) int {
	var count int
	query := "select count(*) from webauthn_credentials where user_id = $1"
	err := s.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		logger.Error(`store: unable to count webauthn certs for user #%d: %v`, userID, err)
		return 0
	}
	return count
}

func (s *Storage) DeleteAllWebAuthnCredentialsByUserID(userID int64) error {
	query := "delete from webauthn_credentials where user_id = $1"
	_, err := s.db.Exec(query, userID)
	return err
}
