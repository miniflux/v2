// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package storage // import "miniflux.app/storage"

import (
	"context"
	"database/sql"
	"regexp"
	"time"

	"github.com/jaytaylor/html2text"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/wangbin/jiebago"
)

// Storage handles all operations related to the database.
type Storage struct {
	db              *sql.DB
	keywordsCounter *prometheus.CounterVec
	seg             *jiebago.Segmenter
}

// NewStorage returns a new Storage.
func NewStorage(db *sql.DB) *Storage {
	var seg jiebago.Segmenter
	seg.LoadDictionary("dict.txt")
	return &Storage{db, nil, &seg}
}

func (s *Storage) SetKeyWordsCounter(counter *prometheus.CounterVec) {
	s.keywordsCounter = counter
}

// ReplacePunctuation from giving string
func ReplacePunctuation(s string) string {
	return regexp.MustCompile("[【】、；‘，。/！@#￥%……&*（）——《》？：“” ]").ReplaceAllString(s, "")
}

// LogKeywordForContent, if counter is not set, skip process
func (s *Storage) LogKeywordForContent(content string) {
	if s.keywordsCounter == nil {
		return
	}

	plainText := content

	plainText, _ = html2text.FromString(plainText, html2text.Options{
		TextOnly: true,
	})

	for keyword := range s.seg.Cut(plainText, true) {
		keyword = ReplacePunctuation(keyword)
		if len(keyword) > 0 {
			if c, err := s.keywordsCounter.GetMetricWithLabelValues(keyword); err == nil {
				c.Inc()
			}
		}
	}
}

// DatabaseVersion returns the version of the database which is in use.
func (s *Storage) DatabaseVersion() string {
	var dbVersion string
	err := s.db.QueryRow(`SELECT current_setting('server_version')`).Scan(&dbVersion)
	if err != nil {
		return err.Error()
	}

	return dbVersion
}

// Ping checks if the database connection works.
func (s *Storage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.db.PingContext(ctx)
}

// DBStats returns database statistics.
func (s *Storage) DBStats() sql.DBStats {
	return s.db.Stats()
}
