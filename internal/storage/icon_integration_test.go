// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"database/sql"
	"testing"
	"time"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/model"
)

func TestStoreFeedIconWhenAnotherTransactionCreatesTheSameIcon(t *testing.T) {
	db := openTestDatabase(t)
	store := NewStorage(db)

	suffix := crypto.GenerateRandomStringHex(10)
	feedID := createIconTestFeed(t, db, "icon_test_user_"+suffix)

	icon := &model.Icon{Hash: "icon-test-" + suffix, MimeType: "image/png", Content: []byte(suffix)}
	t.Cleanup(func() {
		if _, err := db.Exec(`DELETE FROM icons WHERE hash=$1`, icon.Hash); err != nil {
			t.Errorf(`Unable to remove icon %q: %v`, icon.Hash, err)
		}
	})

	// Another transaction creates the same icon and keeps it uncommitted, as when
	// two feeds with the same new icon are refreshed at the same time.
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	var concurrentIconID int64
	err = tx.QueryRow(
		`INSERT INTO icons (hash, mime_type, content, external_id) VALUES ($1, $2, $3, $4) RETURNING id`,
		icon.Hash, icon.MimeType, icon.Content, crypto.GenerateRandomStringHex(20),
	).Scan(&concurrentIconID)
	if err != nil {
		t.Fatal(err)
	}

	result := make(chan error, 1)
	go func() { result <- store.StoreFeedIcon(feedID, icon) }()

	// StoreFeedIcon does not see the uncommitted icon, so it inserts it too and waits for the other transaction.
	waitForQueryBlockedOnLock(t, db, "INSERT INTO icons")
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	if err := <-result; err != nil {
		t.Fatalf(`Unable to store the icon created by the other transaction: %v`, err)
	}
	if icon.ID != concurrentIconID {
		t.Errorf(`Expected icon ID %d, got %d`, concurrentIconID, icon.ID)
	}

	var feedIconID int64
	if err := db.QueryRow(`SELECT icon_id FROM feed_icons WHERE feed_id=$1`, feedID).Scan(&feedIconID); err != nil {
		t.Fatal(err)
	}
	if feedIconID != concurrentIconID {
		t.Errorf(`Expected feed %d to use icon %d, got %d`, feedID, concurrentIconID, feedIconID)
	}
}

// createIconTestFeed creates a user with a feed, removed with the user when the test ends, and returns the feed ID.
func createIconTestFeed(t *testing.T, db *sql.DB, username string) int64 {
	t.Helper()

	var userID, categoryID, feedID int64
	if err := db.QueryRow(`INSERT INTO users (username) VALUES ($1) RETURNING id`, username).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(`DELETE FROM users WHERE id=$1`, userID); err != nil {
			t.Errorf(`Unable to remove user %d: %v`, userID, err)
		}
	})

	if err := db.QueryRow(`INSERT INTO categories (user_id, title) VALUES ($1, 'Icons') RETURNING id`, userID).Scan(&categoryID); err != nil {
		t.Fatal(err)
	}

	err := db.QueryRow(
		`INSERT INTO feeds (user_id, category_id, title, feed_url, site_url) VALUES ($1, $2, 'Icons', 'https://example.org/feed.xml', 'https://example.org/') RETURNING id`,
		userID, categoryID,
	).Scan(&feedID)
	if err != nil {
		t.Fatal(err)
	}
	return feedID
}

// waitForQueryBlockedOnLock waits until a query that contains the given text waits for a lock.
func waitForQueryBlockedOnLock(t *testing.T, db *sql.DB, queryText string) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var blocked bool
		err := db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM pg_stat_activity
				WHERE datname = current_database() AND wait_event_type = 'Lock' AND query LIKE '%' || $1 || '%'
			)`, queryText).Scan(&blocked)
		if err != nil {
			t.Fatal(err)
		}
		if blocked {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf(`No query containing %q waited for a lock`, queryText)
}
