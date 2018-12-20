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

// HasMediaCache checks if the given URL has cache.
func (s *Storage) HasMediaCache(URL string) bool {
	defer timer.ExecutionTime(time.Now(), "[Storage:HasMediaCache]")
	var result int
	query := `SELECT count(*) as c FROM medias WHERE url_hash=$1 AND success='T'`
	s.db.QueryRow(query, media.URLHash(URL)).Scan(&result)
	return result != 0
}

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
	WHERE f.id=$1 and m.success='t'
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
	medias, err := s.getUncachedMedias(days)
	if err != nil {
		return err
	}
	for i, m := range medias {
		logger.Debug("[Storage:CacheMedias] caching medias (%d of %d) %s", i+1, len(medias), m.URL)
		err = s.cacheMedia(m)
		if err != nil {
			logger.Error("[Storage:CacheMedias] unable to cache media %s: %v", m.URL, err)
		}
	}
	return nil
}

func (s *Storage) cacheMedia(m *model.Media) error {
	fm, err := media.FindMedia(m.URL)
	if err != nil {
		return err
	}
	fm.ID = m.ID
	return s.UpdateMedia(fm)
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
		logger.Debug("[Storage:CreateEntryMedias] Caching media (%d of %d) for %s : <%s>", i+1, len(urls), entry.Title, u)
		m := &model.Media{URL: u, URLHash: media.URLHash(u)}
		err = s.MediaByHash(m)
		if err != nil {
			return err
		} else if m.ID == 0 {
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
	return nil
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
		UPDATE medias 
		SET mime_type='', content = E''::bytea, size=0, success='f'
		WHERE id in (
			SELECT m.id
			FROM feeds f
			INNER JOIN entries e on f.id=e.feed_id
			INNER JOIN entry_medias em ON e.id=em.entry_id
			INNER JOIN medias m ON m.id=em.media_id
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

	_ = s.CleanupMedias()

	return nil
}

func (s *Storage) getUncachedMedias(days int) (model.Medias, error) {
	query := `
	SELECT m.id, m.url, m.url_hash
	FROM feeds f
		INNER JOIN entries e on f.id=e.feed_id
		LEFT JOIN entry_medias em ON e.id=em.entry_id
		LEFT JOIN medias m ON em.media_id=m.id
	WHERE 
		f.cache_media='T' 
		AND e.starred='T' 
		AND m.success='f' 
		AND created_at > now () - '%d days'::interval LIMIT 5000
`
	query = fmt.Sprintf(query, days)

	medias := make(model.Medias, 0)
	rows, err := s.db.Query(query)
	defer rows.Close()
	if err == sql.ErrNoRows {
		return medias, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to fetch uncached medias: %v", err)
	}

	for rows.Next() {
		var media model.Media
		err := rows.Scan(&media.ID, &media.URL, &media.URLHash)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch medias row: %v", err)
		}
		medias = append(medias, &media)
	}
	return medias, nil
}
