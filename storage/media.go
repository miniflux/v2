package storage // import "miniflux.app/storage"

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"miniflux.app/logger"

	"miniflux.app/model"
	"miniflux.app/reader/media"
	"miniflux.app/timer"
)

// MediaByURL returns an Media by the url.
// Notice the media fetched could be an unsucessfully cached one.
// Remember to check Media.Success.
func (s *Storage) MediaByURL(URL string) (*model.Media, error) {
	defer timer.ExecutionTime(time.Now(), "[Storage:MediaByURL]")

	m := &model.Media{URLHash: media.URLHash(URL)}
	err := s.MediaByHash(m)
	return m, err
}

// MediaByHash returns an Media by the url hash (checksum).
// Notice the media fetched could be an unsucessfully cached one.
// Remember to check Media.Success.
func (s *Storage) MediaByHash(media *model.Media) error {
	defer timer.ExecutionTime(time.Now(), "[Storage:MediaByHash]")

	err := s.db.QueryRow(
		`SELECT id, url, mime_type, content, success FROM medias WHERE url_hash=$1`,
		media.URLHash,
	).Scan(
		&media.ID,
		&media.URL,
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

// UserMediaByURL returns an Media by the url.
// Notice the media fetched could be an unsucessfully cached one.
// Remember to check Media.Success.
func (s *Storage) UserMediaByURL(URL string, userID int64) (*model.Media, error) {
	defer timer.ExecutionTime(time.Now(), "[Storage:UserMediaByURL]")

	m := &model.Media{URLHash: media.URLHash(URL)}
	err := s.UserMediaByHash(m, userID)
	return m, err
}

// UserMediaByHash returns an Media by the url hash (checksum).
// Notice the media fetched could be an unsucessfully cached one.
// Remember to check Media.Success.
func (s *Storage) UserMediaByHash(media *model.Media, userID int64) error {
	defer timer.ExecutionTime(time.Now(), "[Storage:UserMediaByHash]")

	err := s.db.QueryRow(`
	SELECT m.id, m.url, m.mime_type, m.content, m.success 
	FROM medias m
		INNER JOIN entry_medias em ON m.id=em.media_id
		INNER JOIN entries e ON e.id=em.entry_id
		INNER JOIN feeds f on f.id=e.feed_id
	WHERE m.url_hash=$1 AND em.use_cache='t' AND f.user_id=$2
`,
		media.URLHash,
		userID,
	).Scan(
		&media.ID,
		&media.URL,
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
	(url, url_hash, mime_type, content, size, success)
	VALUES
	($1, $2, $3, $4, $5, $6)
	RETURNING id
`
	err := s.db.QueryRow(
		query,
		media.URL,
		media.URLHash,
		normalizeMimeType(media.MimeType),
		media.Content,
		media.Size,
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
	SET mime_type=$2, content=$3, size=$4, success=$5
	WHERE id = $1
`
	_, err := s.db.Exec(
		query,
		media.ID,
		normalizeMimeType(media.MimeType),
		media.Content,
		media.Size,
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
func (s *Storage) MediaStatistics(feedID int64) (count int, cacheCount int, cacheSize int, err error) {
	defer timer.ExecutionTime(time.Now(), "[Storage:MediaStatistics]")

	query := `
	SELECT count(m.id) count
	FROM feeds f
		INNER JOIN entries e on f.id=e.feed_id
		INNER JOIN entry_medias em on e.id=em.entry_id
		INNER JOIN medias m on em.media_id=m.id
	WHERE f.id=$1
`
	err = s.db.QueryRow(
		query,
		feedID,
	).Scan(&count)

	if err != nil || count == 0 {
		return
	}

	query = `
	SELECT count(m.id) count, coalesce(sum(m.size),0) size
	FROM feeds f
		INNER JOIN entries e on f.id=e.feed_id
		INNER JOIN entry_medias em on e.id=em.entry_id
		INNER JOIN medias m on em.media_id=m.id
	WHERE f.id=$1 AND em.use_cache='t' AND m.success='t'
`
	err = s.db.QueryRow(
		query,
		feedID,
	).Scan(&cacheCount, &cacheSize)

	return
}

// CacheMedias caches recently created medias of starred entries.
// the days limit is to avoid always trying to cache failed medias
func (s *Storage) CacheMedias(days int) error {
	medias, mEntries, err := s.getUncachedMedias(days)
	if err != nil {
		return err
	}
	for i, m := range medias {
		logger.Debug("[Storage:CacheMedias] caching medias (%d of %d) %s", i+1, len(medias), m.URL)
		entries, _ := mEntries[m.ID]
		err = s.cacheMedia(m, entries)
		if err != nil {
			logger.Error("[Storage:CacheMedias] unable to cache media %s: %v", m.URL, err)
		}
	}
	return nil
}

func (s *Storage) cacheMedia(m *model.Media, entries string) error {
	if !m.Success {
		fm, err := media.FindMedia(m.URL)
		if err != nil {
			return err
		}
		fm.ID = m.ID
		err = s.UpdateMedia(fm)
		if err != nil {
			return err
		}
	}
	sql := fmt.Sprintf(`UPDATE entry_medias set use_cache='t' WHERE media_id=%d AND entry_id in (%s)`, m.ID, entries)
	_, err := s.db.Exec(sql)
	return err
}

// UpdateEntryMedias updates media records for given entry
func (s *Storage) UpdateEntryMedias(entry *model.Entry) error {
	defer timer.ExecutionTime(time.Now(), "[Storage:UpdateEntryMedias]")
	_, err := s.db.Exec(`DELETE FROM entry_medias WHERE entry_id=$1`, entry.ID)
	if err != nil {
		return err
	}
	return s.CreateEntryMedias(entry)
}

// CreateEntryMedias create media records for given entry, but not cache them
func (s *Storage) CreateEntryMedias(entry *model.Entry) error {
	var err error
	defer func() {
		if err != nil {
			logger.Error("[Storage:CreateEntryMedias] unable to create media records for entry id %d: %v", entry.ID, err)
		}
	}()
	urls, err := media.ParseDocument(entry)
	if err != nil || len(urls) == 0 {
		return err
	}
	entryMedias := make(map[string]int8, 0)
	for i, u := range urls {
		logger.Debug("[Storage:CreateEntryMedias] Create media (%d of %d) for %s : <%s>", i+1, len(urls), entry.Title, u)
		m := &model.Media{URL: u, URLHash: media.URLHash(u)}
		err = s.MediaByHash(m)
		if err != nil {
			return err
		}
		if m.ID == 0 {
			if err = s.CreateMedia(m); err != nil {
				return err
			}
		}
		// medias in an article could be duplicate, use map to remove them
		entryMedias[fmt.Sprintf("(%v,%v),", entry.ID, m.ID)] = 0
	}
	if len(entryMedias) == 0 {
		return nil
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
	return err
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

// RemoveFeedCaches removes all caches of a feed.
func (s *Storage) RemoveFeedCaches(userID, feedID int64) error {
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:RemoveFeedCaches] userID=%d, feedID=%d", userID, feedID))

	result, err := s.db.Exec(`
		UPDATE entry_medias 
		SET use_cache ='f'
		WHERE entry_id in (
			SELECT em.entry_id
            FROM feeds f
                INNER JOIN entries e on f.id=e.feed_id
                INNER JOIN entry_medias em ON e.id=em.entry_id
			WHERE f.id=$1 AND f.user_id=$2
		)
	`, feedID, userID)
	if err != nil {
		return fmt.Errorf("unable to remove cache for feed #%d: %v", feedID, err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("unable to remove cache for feed #%d: %v", feedID, err)
	}

	if count == 0 {
		return errors.New("no cache has been removed")
	}

	return nil
}

// CleanupCaches removes caches that has no entry refer to.
func (s *Storage) CleanupCaches() error {
	defer timer.ExecutionTime(time.Now(), "[Storage:CleanupCaches]")

	result, err := s.db.Exec(`
		UPDATE medias 
		SET mime_type='', content = E''::bytea, size=0, success='f'
		WHERE id in (
			SELECT id
			FROM medias
			WHERE success='t' and id NOT IN(
				SELECT media_id from entry_medias WHERE use_cache='t'
			)
		)
	`)
	if err != nil {
		return fmt.Errorf("unable to remove CleanupCaches: %v", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("unable to remove CleanupCaches: %v", err)
	}

	if count == 0 {
		return errors.New("no cache has been removed")
	}
	return nil
}

func (s *Storage) getUncachedMedias(days int) (model.Medias, map[int64]string, error) {
	mediaEntries := make(map[int64]string, 0)
	query := `
	SELECT m.id, m.url, m.url_hash, m.success, string_agg(cast(e.id as TEXT),',') as eids
    FROM feeds f
        INNER JOIN entries e ON f.id=e.feed_id
        INNER JOIN entry_medias em ON e.id=em.entry_id
        INNER JOIN medias m ON em.media_id=m.id
    WHERE 
        f.cache_media='T' 
        AND e.starred='T' 
	    AND em.use_cache='F'
		AND created_at > now()-'%d days'::interval
	GROUP BY m.id
    LIMIT 5000
`
	query = fmt.Sprintf(query, days)

	medias := make(model.Medias, 0)
	rows, err := s.db.Query(query)
	defer rows.Close()
	if err == sql.ErrNoRows {
		return medias, mediaEntries, nil
	} else if err != nil {
		return nil, nil, fmt.Errorf("unable to fetch uncached medias: %v", err)
	}

	for rows.Next() {
		var media model.Media
		var entryIDs string
		err := rows.Scan(&media.ID, &media.URL, &media.URLHash, &media.Success, &entryIDs)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to fetch uncached medias row: %v", err)
		}
		medias = append(medias, &media)
		mediaEntries[media.ID] = entryIDs

	}
	return medias, mediaEntries, nil
}
