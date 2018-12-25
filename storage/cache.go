package storage // import "miniflux.app/storage"

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/reader/media"
	"miniflux.app/timer"
)

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
		if !m.Cached {
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
		if err != nil {
			logger.Error("[Storage:CacheMedias] unable to cache media %s: %v", m.URL, err)
		}
	}
	return nil
}

// getUncachedMedias gets medias which should be but not yet cached
// set entryID=0 to get all uncached medias.
// caching task for an entry has two parts:
// 1. make sure the media is cached
// 2. has an entry_medias record refers to the media and use_cache='T'
func (s *Storage) getUncachedMedias(days int) (model.Medias, map[int64]string, error) {
	mediaEntries := make(map[int64]string, 0)
	// FIXME: use created_at to ignore failed medias could have problem
	// when caching medias which created long time ago but never requires cache
	query := `
	SELECT m.id, m.url, m.url_hash, m.cached, string_agg(cast(e.id as TEXT),',') as eids
    FROM feeds f
        INNER JOIN entries e ON f.id=e.feed_id
        INNER JOIN entry_medias em ON e.id=em.entry_id
        INNER JOIN medias m ON em.media_id=m.id
    WHERE 
        f.cache_media='T' 
        AND e.starred='T' 
		AND (em.use_cache='F' OR m.cached='F')
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
		err := rows.Scan(&media.ID, &media.URL, &media.URLHash, &media.Cached, &entryIDs)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to fetch uncached medias row: %v", err)
		}
		medias = append(medias, &media)
		mediaEntries[media.ID] = entryIDs

	}
	return medias, mediaEntries, nil
}

// HasEntryCache indicates if an entry has cache.
func (s *Storage) HasEntryCache(entryID int64) (bool, error) {
	var result int
	query := `
		SELECT count(*) as c 
		FROM entry_medias em
			INNER JOIN medias m on em.media_id=m.id
		WHERE em.entry_id=$1 AND em.use_cache='T' AND m.cached='T'`
	err := s.db.QueryRow(query, entryID).Scan(&result)
	return result > 0, err
}

// CacheEntryMedias caches media of an entry.
func (s *Storage) CacheEntryMedias(userID, EntryID int64) error {
	medias, err := s.getEntryMedias(userID, EntryID)
	if err != nil {
		return err
	}
	if len(medias) == 0 {
		return nil
	}
	var buf bytes.Buffer
	for _, m := range medias {
		if !m.Cached {
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
		buf.WriteString(fmt.Sprintf("('%v','%v','T'),", EntryID, m.ID))
	}
	vals := buf.String()[:buf.Len()-1]
	sql := fmt.Sprintf(`
		INSERT INTO entry_medias (entry_id, media_id, use_cache)
		VALUES %s
		ON CONFLICT (entry_id, media_id) DO UPDATE
			SET use_cache='T'
	`, vals)
	_, err = s.db.Exec(sql)
	return err
}

func (s *Storage) getEntryMedias(userID, EntryID int64) (model.Medias, error) {
	query := `
		SELECT m.id, m.url, m.url_hash, m.cached
		FROM feeds f
			INNER JOIN entries e on f.id=e.feed_id
			INNER JOIN entry_medias em on e.id=em.entry_id
			INNER JOIN medias m on m.id=em.media_id
		WHERE f.user_id=$1 AND e.id=$2
`
	medias := make(model.Medias, 0)
	rows, err := s.db.Query(query, userID, EntryID)
	defer rows.Close()
	if err == sql.ErrNoRows {
		return medias, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to fetch entry medias: %v", err)
	}

	for rows.Next() {
		var media model.Media
		err := rows.Scan(&media.ID, &media.URL, &media.URLHash, &media.Cached)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch entry medias row: %v", err)
		}
		medias = append(medias, &media)

	}
	return medias, nil
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
		SET mime_type='', content = E''::bytea, size=0, cached='f'
		WHERE id in (
			SELECT id
			FROM medias
			WHERE cached='t' and id NOT IN(
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

// ClearAllCaches clears caches of all user and all entries.
func (s *Storage) ClearAllCaches() error {
	query := `
		UPDATE medias SET mime_type='', content = E''::bytea, size=0, cached='f' WHERE cached='t';
		UPDATE entry_medias set use_cache='f' WHERE use_cache='t';
	`
	if _, err := s.db.Exec(query); err != nil {
		return fmt.Errorf("unable to clear all caches: %v", err)
	}

	return nil
}

// ToggleEntryCache toggles entry cache.
func (s *Storage) ToggleEntryCache(userID int64, entryID int64) error {
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:ToggleEntryCache] userID=%d, entryID=%d", userID, entryID))

	has, err := s.HasEntryCache(entryID)
	if err != nil {
		return fmt.Errorf("unable to toggle cache for entry #%d: %v", entryID, err)
	}

	if has {
		query := `
			UPDATE entry_medias SET use_cache='f' WHERE entry_id in (
				SELECT e.id
				FROM feeds f
					INNER JOIN entries e on f.id=e.feed_id
				WHERE f.user_id=$1 AND e.id=$2
			);
		`
		result, err := s.db.Exec(query, userID, entryID)
		if err != nil {
			return fmt.Errorf("unable to toggle cache for user #%d, entry #%d: %v", userID, entryID, err)
		}

		count, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("unable to toogle cache for user #%d, entry #%d: %v", userID, entryID, err)
		}
		if count == 0 {
			return errors.New("nothing has been updated")
		}
		return err
	}

	return s.CacheEntryMedias(userID, entryID)
}
