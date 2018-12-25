package storage // import "miniflux.app/storage"

import (
	"fmt"
	"time"

	"miniflux.app/timer"
)

// MediaStatisticsAll returns media count and size of specified feed.
func (s *Storage) MediaStatisticsAll() (count int, cacheCount int, cacheSize int, err error) {
	defer timer.ExecutionTime(time.Now(), "[Storage:MediaStatisticsAll]")

	err = s.db.QueryRow(`SELECT count(id) count FROM medias`).Scan(&count)

	if err != nil || count == 0 {
		return
	}

	err = s.db.QueryRow(`
		SELECT count(id) count, coalesce(sum(size),0) size
		FROM medias
		WHERE success='t'
	`).Scan(&cacheCount, &cacheSize)

	return
}

// MediaStatisticsByFeed returns media count and size of specified feed.
func (s *Storage) MediaStatisticsByFeed(feedID int64) (count int, cacheCount int, cacheSize int, err error) {
	defer timer.ExecutionTime(time.Now(), "[Storage:MediaStatisticsByFeed]")

	cond := fmt.Sprintf(`f.id=%d`, feedID)
	return s.mediaStatisticsByCond(cond)
}

// MediaStatisticsByUser returns media count and size of specified feed.
func (s *Storage) MediaStatisticsByUser(userID int64) (count int, cacheCount int, cacheSize int, err error) {
	defer timer.ExecutionTime(time.Now(), "[Storage:MediaStatisticsByUser]")

	cond := fmt.Sprintf(`f.user_id=%d`, userID)
	return s.mediaStatisticsByCond(cond)
}

// MediaStatisticsByEntry returns media count and size of specified feed.
func (s *Storage) MediaStatisticsByEntry(entryID int64) (count int, cacheCount int, cacheSize int, err error) {
	defer timer.ExecutionTime(time.Now(), "[Storage:MediaStatisticsByEntry]")

	cond := fmt.Sprintf(`e.id=%d`, entryID)
	return s.mediaStatisticsByCond(cond)
}

// MediaStatisticsByUser returns media count and size of specified feed.
func (s *Storage) mediaStatisticsByCond(cond string) (count int, cacheCount int, cacheSize int, err error) {

	query := fmt.Sprintf(`
	SELECT count(m.id) count
	FROM feeds f
		INNER JOIN entries e on f.id=e.feed_id
		INNER JOIN entry_medias em on e.id=em.entry_id
		INNER JOIN medias m on em.media_id=m.id
	WHERE %s`, cond)
	err = s.db.QueryRow(query).Scan(&count)

	if err != nil || count == 0 {
		return
	}

	query = fmt.Sprintf(`
	SELECT count(m.id) count, coalesce(sum(m.size),0) size
	FROM feeds f
		INNER JOIN entries e on f.id=e.feed_id
		INNER JOIN entry_medias em on e.id=em.entry_id
		INNER JOIN medias m on em.media_id=m.id
	WHERE %s AND em.use_cache='t' AND m.success='t'`, cond)
	err = s.db.QueryRow(query).Scan(&cacheCount, &cacheSize)

	return
}
