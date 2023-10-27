package model // import "miniflux.app/v2/internal/model"

import (
	"database/sql/driver"
	jsonenc "encoding/json"
	"errors"
	"fmt"

	"github.com/go-webauthn/webauthn/webauthn"
)

// handle marshalling / unmarshalling session data
type WebAuthnSession struct {
	*webauthn.SessionData
}

func (s WebAuthnSession) Value() (driver.Value, error) {
	return jsonenc.Marshal(s)
}

func (s *WebAuthnSession) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return jsonenc.Unmarshal(b, &s)
}

func (s WebAuthnSession) String() string {
	if s.SessionData == nil {
		return "{}"
	}
	return fmt.Sprintf("{Challenge: %s, UserID: %x}", s.SessionData.Challenge, s.SessionData.UserID)
}
