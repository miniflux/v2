// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"testing"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/http/server/servertest"
	"miniflux.app/v2/internal/model"
)

const (
	adminUsername = "admin"
	adminPassword = "test123"

	// regularUserPassword is the password of the users created by newRegularUser.
	regularUserPassword = "regular_test_user_password"
)

func TestMain(m *testing.M) {
	servertest.Main(m, createAdminUser)
}

// createAdminUser creates the admin user that the tests share, unless an earlier run left it in the database.
func createAdminUser(server *servertest.Server) error {
	if server.Store.UserExists(adminUsername) {
		return nil
	}

	_, err := server.Store.CreateUser(&model.UserCreationRequest{
		Username: adminUsername,
		Password: adminPassword,
		IsAdmin:  true,
	})
	return err
}

func randomUsername() string {
	return fmt.Sprintf("regular_test_user_%10d", rand.Int())
}

func newAdminClient(t *testing.T) *miniflux.Client {
	t.Helper()

	return miniflux.NewClient(servertest.Require(t).URL, adminUsername, adminPassword)
}

// deleteUserOnCleanup removes the user when the test and its subtests finish,
// and fails the test if the user cannot be removed.
func deleteUserOnCleanup(t *testing.T, userID int64) {
	t.Helper()

	client := newAdminClient(t)
	t.Cleanup(func() {
		if err := client.DeleteUser(userID); err != nil {
			t.Errorf(`Unable to delete user %d: %v`, userID, err)
		}
	})
}

// newRegularUser creates a regular user that is removed when the test ends,
// and returns a client authenticated as that user.
func newRegularUser(t *testing.T) (*miniflux.Client, *miniflux.User) {
	t.Helper()

	user, err := newAdminClient(t).CreateUser(randomUsername(), regularUserPassword, false)
	if err != nil {
		t.Fatal(err)
	}
	deleteUserOnCleanup(t, user.ID)

	return miniflux.NewClient(servertest.Require(t).URL, user.Username, regularUserPassword), user
}

// requestStatus sends an API request authenticated as the regular user and returns the status code.
// It covers routes that the client does not use, such as aliases and alternative methods.
func requestStatus(t *testing.T, user *miniflux.User, method, path string) int {
	t.Helper()

	request, err := http.NewRequest(method, servertest.Require(t).URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.SetBasicAuth(user.Username, regularUserPassword)

	status, _ := servertest.Send(t, request)
	return status
}

// listFeedEntries returns the entries of the feed that match the filter.
func listFeedEntries(t *testing.T, client *miniflux.Client, feedID int64, filter *miniflux.Filter) *miniflux.EntryResultSet {
	t.Helper()

	result, err := client.FeedEntries(feedID, filter)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
