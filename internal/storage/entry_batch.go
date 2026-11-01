// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"miniflux.app/v2/internal/model"

	"github.com/lib/pq"
)

// entryBatchSize caps how many entries go into one statement, which bounds
// the size of a single query message for feeds with thousands of entries.
const entryBatchSize = 500

// Tags are passed as one JSON array per entry because PostgreSQL can't
// unnest a ragged two-dimensional array. A JSON null keeps tags NULL, as
// passing a nil slice to pq.Array does.
const entryTagsFromJSON = `CASE WHEN e.tags = 'null' THEN NULL ELSE ARRAY(SELECT jsonb_array_elements_text(e.tags::jsonb)) END`

// entryColumns holds entry fields as parallel arrays for unnest().
type entryColumns struct {
	hashes, titles, urls, commentsURLs, dates, contents, authors []string
	tsvTitles, tsvContents, tags, languages                      []string
	readingTimes                                                 []int64
}

func newEntryColumns(entries model.Entries) (*entryColumns, error) {
	n := len(entries)
	c := &entryColumns{
		hashes:       make([]string, n),
		titles:       make([]string, n),
		urls:         make([]string, n),
		commentsURLs: make([]string, n),
		dates:        make([]string, n),
		contents:     make([]string, n),
		authors:      make([]string, n),
		tsvTitles:    make([]string, n),
		tsvContents:  make([]string, n),
		tags:         make([]string, n),
		languages:    make([]string, n),
		readingTimes: make([]int64, n),
	}

	for i, entry := range entries {
		tags, err := json.Marshal(entry.Tags)
		if err != nil {
			return nil, fmt.Errorf(`store: unable to encode tags of entry %q: %v`, entry.URL, err)
		}

		c.hashes[i] = entry.Hash
		c.titles[i] = entry.Title
		c.urls[i] = entry.URL
		c.commentsURLs[i] = entry.CommentsURL
		c.dates[i] = entry.Date.Format(time.RFC3339Nano)
		c.contents[i] = entry.Content
		c.authors[i] = entry.Author
		c.tsvTitles[i], c.tsvContents[i] = truncateTitleAndContentForTSVectorField(entry.Title, entry.Content)
		c.tags[i] = string(tags)
		c.languages[i] = entry.Language
		c.readingTimes[i] = int64(entry.ReadingTime)
	}

	return c, nil
}

// uniqueEntriesByHash drops entries whose hash appeared earlier in the list.
// A feed document can list the same item more than once, and the duplicates
// would violate the unique (feed_id, hash) constraint.
func uniqueEntriesByHash(entries model.Entries) model.Entries {
	seen := make(map[string]struct{}, len(entries))
	unique := make(model.Entries, 0, len(entries))
	for _, entry := range entries {
		if _, ok := seen[entry.Hash]; ok {
			continue
		}
		seen[entry.Hash] = struct{}{}
		unique = append(unique, entry)
	}
	return unique
}

func chunkEntries(entries model.Entries) []model.Entries {
	var chunks []model.Entries
	for start := 0; start < len(entries); start += entryBatchSize {
		chunks = append(chunks, entries[start:min(start+entryBatchSize, len(entries))])
	}
	return chunks
}

// createEntries inserts entries of one feed and their enclosures, and returns
// the entries that were created, in their original order. Entries must have
// unique hashes. Entries that are tombstoned or that already exist are
// skipped.
func (s *Storage) createEntries(tx *sql.Tx, userID, feedID int64, entries model.Entries) (model.Entries, error) {
	var created model.Entries
	for _, chunk := range chunkEntries(entries) {
		chunkCreated, err := s.createEntriesChunk(tx, userID, feedID, chunk)
		if err != nil {
			return nil, err
		}
		created = append(created, chunkCreated...)
	}

	if err := s.createEnclosures(tx, userID, created); err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Storage) createEntriesChunk(tx *sql.Tx, userID, feedID int64, entries model.Entries) (model.Entries, error) {
	c, err := newEntryColumns(entries)
	if err != nil {
		return nil, err
	}

	// The NOT EXISTS guard makes the tombstone check atomic with the insert, so a
	// concurrent archive committing between an earlier existence check and this statement
	// cannot bring a deleted entry back as unread. ON CONFLICT covers an entry inserted
	// by a concurrent refresh of the same feed.
	query := `
		INSERT INTO entries
			(
				title,
				hash,
				url,
				comments_url,
				published_at,
				content,
				author,
				user_id,
				feed_id,
				reading_time,
				changed_at,
				document_vectors,
				tags,
				language
			)
		SELECT
			e.title,
			e.hash,
			e.url,
			e.comments_url,
			e.published_at,
			e.content,
			e.author,
			$1,
			$2,
			e.reading_time,
			now(),
			setweight(to_tsvector(e.tsv_title), 'A') || setweight(to_tsvector(e.tsv_content), 'B'),
			` + entryTagsFromJSON + `,
			e.language
		FROM
			unnest(
				$3::text[], $4::text[], $5::text[], $6::text[], $7::timestamptz[], $8::text[],
				$9::text[], $10::int[], $11::text[], $12::text[], $13::text[], $14::text[]
			) WITH ORDINALITY AS e(
				title, hash, url, comments_url, published_at, content,
				author, reading_time, tsv_title, tsv_content, tags, language, position
			)
		WHERE NOT EXISTS (
			SELECT 1 FROM entry_tombstones t WHERE t.feed_id=$2 AND t.hash=e.hash
		)
		ORDER BY e.position
		ON CONFLICT (feed_id, hash) DO NOTHING
		RETURNING
			id, hash, status, created_at, changed_at
	`

	rows, err := tx.Query(
		query,
		userID,
		feedID,
		pq.Array(c.titles),
		pq.Array(c.hashes),
		pq.Array(c.urls),
		pq.Array(c.commentsURLs),
		pq.Array(c.dates),
		pq.Array(c.contents),
		pq.Array(c.authors),
		pq.Array(c.readingTimes),
		pq.Array(c.tsvTitles),
		pq.Array(c.tsvContents),
		pq.Array(c.tags),
		pq.Array(c.languages),
	)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to create entries (feed #%d): %v`, feedID, err)
	}
	defer rows.Close()

	byHash := make(map[string]*model.Entry, len(entries))
	for _, entry := range entries {
		byHash[entry.Hash] = entry
	}

	createdHashes := make(map[string]struct{}, len(entries))
	for rows.Next() {
		var (
			id                   int64
			hash, status         string
			createdAt, changedAt time.Time
		)
		if err := rows.Scan(&id, &hash, &status, &createdAt, &changedAt); err != nil {
			return nil, fmt.Errorf(`store: unable to fetch created entry (feed #%d): %v`, feedID, err)
		}
		entry := byHash[hash]
		entry.ID = id
		entry.UserID = userID
		entry.FeedID = feedID
		entry.Status = status
		entry.CreatedAt = createdAt
		entry.ChangedAt = changedAt
		createdHashes[hash] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(`store: unable to create entries (feed #%d): %v`, feedID, err)
	}

	created := make(model.Entries, 0, len(createdHashes))
	for _, entry := range entries {
		if _, ok := createdHashes[entry.Hash]; ok {
			created = append(created, entry)
		}
	}
	return created, nil
}

// updateEntries updates stored entries of one feed, matched by hash, and
// replaces their enclosures. Entries must have unique hashes and their ID set.
// The published date is left alone because some feeds have no dates and the
// parser falls back to the current time, which would reorder the history.
func (s *Storage) updateEntries(tx *sql.Tx, userID, feedID int64, entries model.Entries) error {
	for _, chunk := range chunkEntries(entries) {
		if err := s.updateEntriesChunk(tx, userID, feedID, chunk); err != nil {
			return err
		}
	}
	return s.replaceEnclosures(tx, userID, entries)
}

func (s *Storage) updateEntriesChunk(tx *sql.Tx, userID, feedID int64, entries model.Entries) error {
	c, err := newEntryColumns(entries)
	if err != nil {
		return err
	}

	query := `
		UPDATE
			entries
		SET
			title=e.title,
			url=e.url,
			comments_url=e.comments_url,
			content=e.content,
			author=e.author,
			reading_time=e.reading_time,
			document_vectors = setweight(to_tsvector(e.tsv_title), 'A') || setweight(to_tsvector(e.tsv_content), 'B'),
			tags=` + entryTagsFromJSON + `,
			language=e.language
		FROM
			unnest(
				$3::text[], $4::text[], $5::text[], $6::text[], $7::text[], $8::text[],
				$9::int[], $10::text[], $11::text[], $12::text[], $13::text[]
			) AS e(
				hash, title, url, comments_url, content, author,
				reading_time, tsv_title, tsv_content, tags, language
			)
		WHERE
			entries.user_id=$1 AND entries.feed_id=$2 AND entries.hash=e.hash
	`

	_, err = tx.Exec(
		query,
		userID,
		feedID,
		pq.Array(c.hashes),
		pq.Array(c.titles),
		pq.Array(c.urls),
		pq.Array(c.commentsURLs),
		pq.Array(c.contents),
		pq.Array(c.authors),
		pq.Array(c.readingTimes),
		pq.Array(c.tsvTitles),
		pq.Array(c.tsvContents),
		pq.Array(c.tags),
		pq.Array(c.languages),
	)
	if err != nil {
		return fmt.Errorf(`store: unable to update entries (feed #%d): %v`, feedID, err)
	}
	return nil
}

// entryIDsByHash returns the IDs of the stored entries of a feed with the given hashes.
func (s *Storage) entryIDsByHash(tx *sql.Tx, feedID int64, hashes []string) (map[string]int64, error) {
	ids := make(map[string]int64)
	if len(hashes) == 0 {
		return ids, nil
	}

	// Note: This query uses entries_feed_id_hash_key index (filtering on user_id is not necessary).
	rows, err := tx.Query(`SELECT hash, id FROM entries WHERE feed_id=$1 AND hash=ANY($2)`, feedID, pq.Array(hashes))
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch entry IDs (feed #%d): %v`, feedID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var hash string
		var id int64
		if err := rows.Scan(&hash, &id); err != nil {
			return nil, fmt.Errorf(`store: unable to fetch entry IDs (feed #%d): %v`, feedID, err)
		}
		ids[hash] = id
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(`store: unable to fetch entry IDs (feed #%d): %v`, feedID, err)
	}
	return ids, nil
}

// KnownEntryHashes returns which of the given hashes belong to entries that
// are stored or tombstoned for the feed. The crawler skips those, because
// they are either already stored or would be discarded.
func (s *Storage) KnownEntryHashes(feedID int64, hashes []string) (map[string]struct{}, error) {
	known := make(map[string]struct{})
	if feedID == 0 || len(hashes) == 0 {
		return known, nil
	}

	query := `
		SELECT hash FROM entries WHERE feed_id=$1 AND hash=ANY($2)
		UNION
		SELECT hash FROM entry_tombstones WHERE feed_id=$1 AND hash=ANY($2)
	`
	rows, err := s.db.Query(query, feedID, pq.Array(hashes))
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch known entry hashes (feed #%d): %v`, feedID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			return nil, fmt.Errorf(`store: unable to fetch known entry hashes (feed #%d): %v`, feedID, err)
		}
		known[hash] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(`store: unable to fetch known entry hashes (feed #%d): %v`, feedID, err)
	}
	return known, nil
}

// enclosureColumns holds enclosure fields as parallel arrays for unnest().
// Enclosures with an empty URL are left out.
type enclosureColumns struct {
	enclosures         []*model.Enclosure
	entryIDs           []int64
	urls, mimeTypes    []string
	sizes, progression []int64
}

func newEnclosureColumns(userID int64, entries model.Entries) *enclosureColumns {
	c := &enclosureColumns{}
	for _, entry := range entries {
		for _, enclosure := range entry.Enclosures {
			enclosure.EntryID = entry.ID
			enclosure.UserID = userID

			enclosureURL := strings.TrimSpace(enclosure.URL)
			if enclosureURL == "" {
				continue
			}

			c.enclosures = append(c.enclosures, enclosure)
			c.entryIDs = append(c.entryIDs, entry.ID)
			c.urls = append(c.urls, enclosureURL)
			c.mimeTypes = append(c.mimeTypes, enclosure.MimeType)
			c.sizes = append(c.sizes, enclosure.Size)
			c.progression = append(c.progression, enclosure.MediaProgression)
		}
	}
	return c
}

// createEnclosures inserts the enclosures of stored entries, skipping the
// ones that already exist.
func (s *Storage) createEnclosures(tx *sql.Tx, userID int64, entries model.Entries) error {
	c := newEnclosureColumns(userID, entries)
	if len(c.enclosures) == 0 {
		return nil
	}

	query := `
		INSERT INTO enclosures
			(url, size, mime_type, entry_id, user_id, media_progression)
		SELECT
			e.url, e.size, e.mime_type, e.entry_id, $1, e.media_progression
		FROM
			unnest($2::text[], $3::bigint[], $4::text[], $5::bigint[], $6::int[])
				WITH ORDINALITY AS e(url, size, mime_type, entry_id, media_progression, position)
		ORDER BY e.position
		ON CONFLICT (user_id, entry_id, sha256(url::bytea)) DO NOTHING
		RETURNING
			id, entry_id, url
	`
	rows, err := tx.Query(
		query,
		userID,
		pq.Array(c.urls),
		pq.Array(c.sizes),
		pq.Array(c.mimeTypes),
		pq.Array(c.entryIDs),
		pq.Array(c.progression),
	)
	if err != nil {
		return fmt.Errorf(`store: unable to create enclosures: %w`, err)
	}
	defer rows.Close()

	type enclosureKey struct {
		entryID int64
		url     string
	}
	ids := make(map[enclosureKey]int64, len(c.enclosures))
	for rows.Next() {
		var key enclosureKey
		var id int64
		if err := rows.Scan(&id, &key.entryID, &key.url); err != nil {
			return fmt.Errorf(`store: unable to fetch created enclosure: %w`, err)
		}
		ids[key] = id
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf(`store: unable to create enclosures: %w`, err)
	}

	// Only the first of several enclosures with the same URL is inserted.
	for i, enclosure := range c.enclosures {
		key := enclosureKey{c.entryIDs[i], c.urls[i]}
		if id, ok := ids[key]; ok {
			enclosure.ID = id
			delete(ids, key)
		}
	}

	return nil
}

// replaceEnclosures makes the enclosures of stored entries match the ones
// listed in the feed, adding new ones and deleting those no longer listed.
func (s *Storage) replaceEnclosures(tx *sql.Tx, userID int64, entries model.Entries) error {
	if len(entries) == 0 {
		return nil
	}

	if err := s.createEnclosures(tx, userID, entries); err != nil {
		return err
	}

	entryIDs := make([]int64, len(entries))
	for i, entry := range entries {
		entryIDs[i] = entry.ID
	}
	keep := newEnclosureColumns(userID, entries)

	query := `
		DELETE FROM
			enclosures
		WHERE
			user_id=$1 AND
			entry_id=ANY($2) AND
			NOT EXISTS (
				SELECT 1
				FROM unnest($3::bigint[], $4::text[]) AS k(entry_id, url)
				WHERE k.entry_id=enclosures.entry_id AND k.url=enclosures.url
			)
	`
	if _, err := tx.Exec(query, userID, pq.Array(entryIDs), pq.Array(keep.entryIDs), pq.Array(keep.urls)); err != nil {
		return fmt.Errorf(`store: unable to delete old enclosures: %v`, err)
	}

	return nil
}
