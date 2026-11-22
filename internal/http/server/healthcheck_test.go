// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"miniflux.app/v2/internal/storage"
)

const sensitiveDatabaseErrorMessage = "dial tcp 10.0.0.5:5432: password authentication failed for user miniflux"

type failingPingDriver struct{}

func (failingPingDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New(sensitiveDatabaseErrorMessage)
}

func init() {
	sql.Register("failing-ping", failingPingDriver{})
}

func TestReadinessProbeDoesNotLeakDatabaseError(t *testing.T) {
	db, err := sql.Open("failing-ping", "")
	if err != nil {
		t.Fatalf("unable to open database: %v", err)
	}
	defer db.Close()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readiness", nil)
	newReadinessProbe(storage.NewStorage(db))(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}

	body := recorder.Body.String()
	if strings.Contains(body, sensitiveDatabaseErrorMessage) || strings.Contains(body, "10.0.0.5") {
		t.Fatalf("response body leaks the database error: %q", body)
	}

	if strings.TrimSpace(body) != "Database Connection Error" {
		t.Fatalf("unexpected response body: %q", body)
	}
}
