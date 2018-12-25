package storage // import "miniflux.app/storage"

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"miniflux.app/model"
	"miniflux.app/reader/media"
	"miniflux.app/timer"
)

// MediaByURL returns an Media by the url.
// Notice the media fetched could be an unsucessfully cached one.
// Remember to check Media.Cached.
func (s *Storage) MediaByURL(URL string) (*model.Media, error) {
	defer timer.ExecutionTime(time.Now(), "[Storage:MediaByURL]")

	m := &model.Media{URLHash: media.URLHash(URL)}
	err := s.MediaByHash(m)
	return m, err
}

// MediaByHash returns an Media by the url hash (checksum).
// Notice the media fetched could be an unsucessfully cached one.
// Remember to check Media.Cached.
func (s *Storage) MediaByHash(media *model.Media) error {
	defer timer.ExecutionTime(time.Now(), "[Storage:MediaByHash]")

	err := s.db.QueryRow(
		`SELECT id, url, mime_type, content, cached FROM medias WHERE url_hash=$1`,
		media.URLHash,
	).Scan(
		&media.ID,
		&media.URL,
		&media.MimeType,
		&media.Content,
		&media.Cached,
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
// Remember to check Media.Cached.
func (s *Storage) UserMediaByURL(URL string, userID int64) (*model.Media, error) {
	defer timer.ExecutionTime(time.Now(), "[Storage:UserMediaByURL]")

	m := &model.Media{URLHash: media.URLHash(URL)}
	err := s.UserMediaByHash(m, userID)
	return m, err
}

// UserMediaByHash returns an Media by the url hash (checksum).
// Notice the media fetched could be an unsucessfully cached one.
// Remember to check Media.Cached.
func (s *Storage) UserMediaByHash(media *model.Media, userID int64) error {
	defer timer.ExecutionTime(time.Now(), "[Storage:UserMediaByHash]")

	err := s.db.QueryRow(`
	SELECT m.id, m.url, m.mime_type, m.content, m.cached 
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
		&media.Cached,
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
	(url, url_hash, mime_type, content, size, cached)
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
		media.Cached,
	).Scan(&media.ID)

	if err != nil {
		return fmt.Errorf("Unable to create media: %v", err)
	}

	return nil
}

// CreateEntriesMedia creates media for a slice of entries at a time
func (s *Storage) CreateEntriesMedia(entries model.Entries) error {
	var err error
	cap := len(entries) * 15
	medias := make(map[string]string, cap)
	type IDSet map[int64]*int8
	// one url could be in multiple entries, and even appears many times in one entry
	// use IDSET to make sure one url could have multiple entries, but not duplicated
	urlEntries := make(map[string]IDSet, cap)
	mediaIDs := make(map[string]int64, cap)
	for _, entry := range entries {
		urls, err := media.ParseDocument(entry)
		if err != nil || len(urls) == 0 {
			continue
		}
		for _, u := range urls {
			hash := media.URLHash(u)
			medias[hash] = fmt.Sprintf("('%v','%v'),", strings.Replace(u, "'", "''", -1), hash)
			if _, ok := urlEntries[hash]; !ok {
				urlEntries[hash] = make(IDSet, 0)
			}
			urlEntries[hash][entry.ID] = nil
		}
	}

	if len(medias) == 0 {
		return nil
	}
	// insert medias records
	var buf bytes.Buffer
	for _, em := range medias {
		buf.WriteString(em)
	}
	vals := buf.String()[:buf.Len()-1]
	sql := fmt.Sprintf(`
		INSERT INTO medias (url, url_hash)
		VALUES %s
		ON CONFLICT (url_hash) DO UPDATE
			SET created_at=current_timestamp
		RETURNING id, url_hash
	`, vals)
	rows, err := s.db.Query(sql)
	defer rows.Close()
	if err != nil {
		return err
	}
	for rows.Next() {
		var m model.Media
		err = rows.Scan(&m.ID, &m.URLHash)
		if err != nil {
			return err
		}
		mediaIDs[m.URLHash] = m.ID
	}

	// insert entry_medias records
	buf.Reset()
	for hash, idSet := range urlEntries {
		for id := range idSet {
			buf.WriteString(fmt.Sprintf("(%v,%v),", id, mediaIDs[hash]))
		}
	}
	vals = buf.String()[:buf.Len()-1]
	sql = fmt.Sprintf(`INSERT INTO entry_medias (entry_id, media_id) VALUES %s`, vals)
	_, err = s.db.Exec(sql)
	return err
}

// UpdateMedia updates a media cache.
func (s *Storage) UpdateMedia(media *model.Media) error {
	defer timer.ExecutionTime(time.Now(), "[Storage:UpdateMedia]")
	query := `
	UPDATE medias
	SET mime_type=$2, content=$3, size=$4, cached=$5
	WHERE id = $1
`
	_, err := s.db.Exec(
		query,
		media.ID,
		normalizeMimeType(media.MimeType),
		media.Content,
		media.Size,
		media.Cached,
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
		WHERE m.cached='t' AND e.user_id=$1
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

// UpdateEntriesMedia updates media records for given entries
func (s *Storage) UpdateEntriesMedia(entries model.Entries) error {
	defer timer.ExecutionTime(time.Now(), "[Storage:UpdateEntryMedias]")

	var buf bytes.Buffer
	for _, entry := range entries {
		buf.WriteString(fmt.Sprintf("%d,", entry.ID))
	}
	vals := buf.String()[:buf.Len()-1]
	_, err := s.db.Exec(`DELETE FROM entry_medias WHERE entry_id in ($1)`, vals)
	if err != nil {
		return err
	}
	return s.CreateEntriesMedia(entries)
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
