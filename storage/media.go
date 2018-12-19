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

// MediasByEntryID returns medias of an entry.
func (s *Storage) MediasByEntryID(userID, entryID int64) (map[string]*model.Media, error) {
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:MediasByEntryID] userID=%d, entryID=%d", userID, entryID))
	query := `
		SELECT
		m.id, m.url_hash, m.mime_type, m.content
		FROM medias m
		LEFT JOIN entry_medias em ON em.media_id=m.id
		LEFT JOIN entries e ON e.id=em.entry_id
		WHERE e.user_id=$1 AND e.id=$2 AND m.success='t'
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

// MediaByHash returns an Media by the url hash (checksum).
// Notice the media fetched could be an unsucessfully cached one.
// Remember to check Media.Success.
func (s *Storage) MediaByHash(media *model.Media) error {
	defer timer.ExecutionTime(time.Now(), "[Storage:MediaByHash]")

	err := s.db.QueryRow(
		`SELECT id, mime_type, content, success FROM medias WHERE url_hash=$1`,
		media.URLHash,
	).Scan(
		&media.ID,
		&media.MimeType,
		&media.Content,
		&media.Success,
	)
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
	(url, url_hash, mime_type, content, success)
	VALUES
	($1, $2, $3, $4, $5)
	RETURNING id
`
	err := s.db.QueryRow(
		query,
		media.URL,
		media.URLHash,
		normalizeMimeType(media.MimeType),
		media.Content,
		media.Success,
	).Scan(&media.ID)

	if err != nil {
		return fmt.Errorf("Unable to create media: %v", err)
	}

	return nil
}

// UpdateMedia updates a media cache.
func (s *Storage) UpdateMedia(media *model.Media) error {
	defer timer.ExecutionTime(time.Now(), "[Storage:UpdateMedia]")
	query := `
	UPDATE medias
	SET mime_type=$2, content=$3, success=$4
	WHERE id = $1
`
	_, err := s.db.Exec(
		query,
		media.ID,
		normalizeMimeType(media.MimeType),
		media.Content,
		media.Success,
	)

	if err != nil {
		return fmt.Errorf("Unable to update media: %v", err)
	}

	return nil
}

// Medias returns all media caches tht belongs to a user.
func (s *Storage) Medias(userID int64) (model.Medias, error) {
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:Medias] userID=%d", userID))
	query := `
		SELECT
		m.id, m.url_hash, m.mime_type, m.content
		FROM medias m
		LEFT JOIN entry_medias em ON em.media_id=medias.id
		LEFT JOIN entries e ON e.id=em.entry_id
		WHERE m.success='t' AND e.user_id=$1
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

// MediaStatistics returns media count and size of specified feed.
func (s *Storage) MediaStatistics(feedID int64) (count int, size int, err error) {
	defer timer.ExecutionTime(time.Now(), "[Storage:MediaStatistics]")
	query := `
	SELECT count(m.*) count, coalesce(sum(length(m.content)),0) size
	FROM feeds f
	INNER JOIN entries e on f.id=e.feed_id
	INNER JOIN entry_medias em on e.id=em.entry_id
	INNER JOIN medias m on em.media_id=m.id
	WHERE f.id=$1
`
	err = s.db.QueryRow(
		query,
		feedID,
	).Scan(&count, &size)
	return
}

// CacheEntries caches medias for starred entries.
func (s *Storage) CacheEntries() error {
	entries, err := s.getUncachedEntries()
	if err != nil {
		return err
	}
	for i, entry := range entries {
		logger.Debug("Caching entry (%d of %d) %s", i+1, len(entries), entry.Title)
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
		err = s.cacheNoMediaEntry(entry.ID)
		if err != nil {
			logger.Error("[Storage:CacheEntry] unable to add placeholder cache record for entry id %d: %v", entry.ID, err)
		}
		return
	}
	entryMedias := make(map[string]int8, 0)
	for i, u := range urls {
		logger.Debug("Caching media (%d of %d) for %s : <%s>", i+1, len(urls), entry.Title, u)
		m := &model.Media{URL: u, URLHash: media.URLHash(u)}
		err = s.MediaByHash(m)
		if err != nil {
			return
		} else if m.ID == 0 {
			m, err = media.FindMedia(u)
			if err != nil {
				// make a placeholder media in database
				// retry caching next time
				m = &model.Media{
					URL:      u,
					URLHash:  media.URLHash(u),
					MimeType: "",
					Content:  []byte{},
					Success:  false,
				}
			}
			if err = s.CreateMedia(m); err != nil {
				return
			}
		} else if !m.Success {
			fm, err := media.FindMedia(u)
			if err == nil {
				fm.ID = m.ID
				if err = s.UpdateMedia(fm); err != nil {
					return
				}
			}
		}

		// medias in an article could be duplicate, use map to remove them
		entryMedias[fmt.Sprintf("(%v,%v),", entry.ID, m.ID)] = 0
	}
	if len(entryMedias) == 0 {
		return
	}
	// 'entry_medias' records must insert together here
	// to make sure the records are inserted all or none.
	//
	// Otherwise, if the task is stop in the middle,
	// 'entry_medias' has part of all media records
	// 'getUncachedEntries' won't find and process the entry again
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

// RetryCache retry caching for recently failed meidas
func (s *Storage) RetryCache(days int) error {
	ms, err := s.getFailedMedias(days)
	if err != nil {
		return err
	}
	for _, m := range ms {
		if m.URL == "" {
			continue
		}
		fm, err := media.FindMedia(m.URL)
		if err != nil {
			logger.Error("[Storage:RetryCache] unable to download medias for media id %d: %v", m.ID, err)
			continue
		}
		m.MimeType = fm.MimeType
		m.Content = fm.Content
		m.Success = true
		err = s.UpdateMedia(m)
		if err != nil {
			logger.Error("[Storage:RetryCache] unable to update medias for media id %d: %v", m.ID, err)
		}
	}
	return nil
}

func (s *Storage) getFailedMedias(days int) (model.Medias, error) {
	query := fmt.Sprintf(`
		SELECT id, url
		FROM medias 
		WHERE success='f' AND created_at > now () - '%d days'::interval LIMIT 5000
	`, days)
	if _, err := s.db.Exec(query); err != nil {
		return nil, fmt.Errorf("unable to archive get failed medias: %v", err)
	}

	medias := make(model.Medias, 0)

	rows, err := s.db.Query(query)
	defer rows.Close()
	if err == sql.ErrNoRows {
		return medias, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to fetch uncached medias: %v", err)
	}

	for rows.Next() {
		var m model.Media
		err := rows.Scan(&m.ID, &m.URL)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch medias row: %v", err)
		}
		medias = append(medias, &m)
	}
	return medias, nil
}

func (s *Storage) getUncachedEntries() (model.Entries, error) {
	query := `
	SELECT e.id, e.user_id, e.url, e.title, e.content
	FROM feeds f
		INNER JOIN entries e on f.id=e.feed_id
		LEFT JOIN entry_medias em ON e.id=em.entry_id
		LEFT JOIN medias m ON em.media_id=m.id
	WHERE f.cache_media='T' AND e.starred='T' AND m.id IS NULL
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
		err := rows.Scan(&entry.ID, &entry.UserID, &entry.URL, &entry.Title, &entry.Content)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch entries row: %v", err)
		}
		entries = append(entries, &entry)
	}
	return entries, nil
}

// cacheNoMediaEntry add a fake media (empty url) record for entry without media
// so that getUncachedEntries don't get and parse it again
func (s *Storage) cacheNoMediaEntry(entryID int64) error {
	m := &model.Media{
		URL:      "",
		URLHash:  media.URLHash(""),
		MimeType: "",
		Content:  []byte{},
		Success:  false,
	}
	err := s.MediaByHash(m)
	if err != nil {
		return err
	}
	if m.ID == 0 {
		if err = s.CreateMedia(m); err != nil {
			return err
		}
	}
	_, err = s.db.Exec(`INSERT INTO entry_medias (entry_id, media_id) VALUES ($1,$2)`, entryID, m.ID)
	return err
}
