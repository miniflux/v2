package storage // import "miniflux.app/storage"

import (
	"database/sql"
	"fmt"
	"time"

	"miniflux.app/logger"

	"miniflux.app/model"
	"miniflux.app/reader/media"
	"miniflux.app/timer"
)

// HasMedia checks if the given entry has an media cache.
func (s *Storage) HasMedia(entryID int64) bool {
	var result int
	query := `SELECT count(*) as c FROM entry_medias WHERE entry_id=$1`
	s.db.QueryRow(query, entryID).Scan(&result)
	return result != 0
}

// MediaByID returns an media by the ID.
func (s *Storage) MediaByID(mediaID int64) (*model.Media, error) {
	defer timer.ExecutionTime(time.Now(), "[Storage:MediaByID]")

	var media model.Media
	query := `SELECT id, url_hash, mime_type, content FROM medias WHERE id=$1`
	err := s.db.QueryRow(query, mediaID).Scan(&media.ID, &media.URLHash, &media.MimeType, &media.Content)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch media: %v", err)
	}

	return &media, nil
}

// MediasByEntryID returns a entry media.
func (s *Storage) MediasByEntryID(userID, entryID int64) (map[string]*model.Media, error) {
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:MediasByEntryID] userID=%d, entryID=%d", userID, entryID))
	query := `
		SELECT
		medias.id, medias.url_hash, medias.mime_type, medias.content
		FROM medias
		LEFT JOIN entry_medias ON entry_medias.media_id=medias.id
		LEFT JOIN entries ON entries.id=entry_medias.entry_id
		WHERE entries.user_id=$1 AND entries.id=$2
	`
	rows, err := s.db.Query(query, userID, entryID)
	if err != nil {
		return nil, fmt.Errorf("unable to get medias: %v", err)
	}
	defer rows.Close()

	medias := make(map[string]*model.Media, 0)
	for rows.Next() {
		var media model.Media
		err := rows.Scan(
			&media.ID,
			&media.URLHash,
			&media.MimeType,
			&media.Content,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch medias row: %v", err)
		}
		medias[media.URLHash] = &media
	}

	return medias, nil
}

// MediaByHash returns an Media by the hash (checksum).
func (s *Storage) MediaByHash(media *model.Media) error {
	defer timer.ExecutionTime(time.Now(), "[Storage:MediaByHash]")

	err := s.db.QueryRow(`SELECT id FROM medias WHERE url_hash=$1`, media.URLHash).Scan(&media.ID)
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return fmt.Errorf("Unable to fetch media by hash: %v", err)
	}

	return nil
}

// CreateMedia creates a new media cache.
func (s *Storage) CreateMedia(media *model.Media) error {
	defer timer.ExecutionTime(time.Now(), "[Storage:CreateMedia]")
	query := `
	INSERT INTO medias
	(url_hash, mime_type, content)
	VALUES
	($1, $2, $3)
	RETURNING id
`
	err := s.db.QueryRow(
		query,
		media.URLHash,
		normalizeMimeType(media.MimeType),
		media.Content,
	).Scan(&media.ID)

	if err != nil {
		return fmt.Errorf("Unable to create media: %v", err)
	}

	return nil
}

// CreateEntryMeidas creates an media and associate the media to the given entry.
func (s *Storage) CreateEntryMeidas(entryID int64, medias model.Medias) error {
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:CreateEntryMeidas] entryID=%d", entryID))

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for _, media := range medias {
		err := s.MediaByHash(media)
		if err != nil {
			tx.Rollback()
			return err
		}

		if media.ID == 0 {
			err := s.CreateMedia(media)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		_, err = s.db.Exec(`INSERT INTO entry_medias (entry_id, media_id) VALUES ($1, $2)`, entryID, media.ID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("unable to create entry media: %v", err)
		}
	}
	return tx.Commit()
}

// Medias returns all media caches tht belongs to a user.
func (s *Storage) Medias(userID int64) (model.Medias, error) {
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:Medias] userID=%d", userID))
	query := `
		SELECT
		medias.id, medias.url_hash, medias.mime_type, medias.content
		FROM medias
		LEFT JOIN entry_medias ON entry_medias.media_id=medias.id
		LEFT JOIN entries ON entries.id=entry_medias.entry_id
		WHERE entries.user_id=$1
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch medias: %v", err)
	}
	defer rows.Close()

	var medias model.Medias
	for rows.Next() {
		var media model.Media
		err := rows.Scan(&media.ID, &media.URLHash, &media.MimeType, &media.Content)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch medias row: %v", err)
		}
		medias = append(medias, &media)
	}

	return medias, nil
}

// CacheEntries caches medias for starred entries.
func (s *Storage) CacheEntries() error {
	entries, err := s.getUncachedEntries()
	if err != nil {
		return err
	}
	for _, entry := range entries {
		s.CacheEntry(entry)
	}
	return nil
}

// CacheEntry caches medias for given entry.
func (s *Storage) CacheEntry(entry *model.Entry) {
	var err error
	defer func() {
		if err != nil {
			logger.Error("[Storage:CacheEntry] unable to cache medias for entry id %d: %v", entry.ID, err)
			return
		}
	}()
	urls, err := media.ParseDocument(entry)
	if err != nil || len(urls) == 0 {
		// TODO: save status, don't try caching again
		return
	}
	entryMedias := make(map[string]int8, 0)
	for _, u := range urls {
		m := &model.Media{URLHash: media.URLHash(u)}
		err = s.MediaByHash(m)
		if err != nil {
			return
		}
		if m.ID == 0 {
			if m, err = media.FindMedia(u); err != nil {
				// TODO: don't fail all, but continue for the failed medias next time
				return
			}
			if err = s.CreateMedia(m); err != nil {
				return
			}
		}
		// medias in an article could be duplicate, use map to remove them
		entryMedias[fmt.Sprintf("(%v,%v),", entry.ID, m.ID)] = 0
	}
	if len(entryMedias) == 0 {
		return
	}
	rows := ""
	for em := range entryMedias {
		rows += em
	}
	rows = rows[:len(rows)-1]
	sql := fmt.Sprintf(`INSERT INTO entry_medias (entry_id, media_id) VALUES %s`, rows)
	_, err = s.db.Exec(sql)
}

// CleanupMedias deletes from the database medias those don't belong to any entries.
func (s *Storage) CleanupMedias() error {
	query := `
		DELETE FROM medias
		WHERE id IN (
			SELECT id 
			FROM medias m
			LEFT JOIN entry_medias em on m.id=em.media_id
			WHERE em.entry_id IS NULL
		)
	`
	if _, err := s.db.Exec(query); err != nil {
		return fmt.Errorf("unable to cleanup medias: %v", err)
	}

	return nil
}

func (s *Storage) getUncachedEntries() (model.Entries, error) {
	query := `
	SELECT e.id, e.user_id, e.url, e.content
	FROM entries e
	LEFT JOIN entry_medias em ON e.id=em.entry_id
	LEFT JOIN medias m ON em.media_id=m.id
	WHERE e.starred='T' AND m.id IS NULL
`
	if _, err := s.db.Exec(query); err != nil {
		return nil, fmt.Errorf("unable to archive read entries: %v", err)
	}

	entries := make(model.Entries, 0)

	rows, err := s.db.Query(query)
	defer rows.Close()
	if err == sql.ErrNoRows {
		return entries, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to fetch uncached entries: %v", err)
	}

	for rows.Next() {
		var entry model.Entry
		err := rows.Scan(&entry.ID, &entry.UserID, &entry.URL, &entry.Content)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch entries row: %v", err)
		}
		entries = append(entries, &entry)
	}
	return entries, nil
}
