// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage // import "miniflux.app/v2/internal/storage"

import (
	"context"
	"database/sql"

	"golang.org/x/crypto/acme/autocert"
)

// Making sure that we're adhering to the autocert.Cache interface.
var _ autocert.Cache = (*CertificateCache)(nil)

// CertificateCache provides a SQL backend to the autocert cache.
type CertificateCache struct {
	storage *Storage
}

// NewCertificateCache creates an cache instance that can be used with autocert.Cache.
// It returns any errors that could happen while connecting to SQL.
func NewCertificateCache(storage *Storage) *CertificateCache {
	return &CertificateCache{
		storage: storage,
	}
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (c *CertificateCache) Get(ctx context.Context, key string) ([]byte, error) {
	query := `SELECT data::bytea FROM acme_cache WHERE key = $1`
	var data []byte
	err := c.storage.db.QueryRowContext(ctx, query, key).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, autocert.ErrCacheMiss
	}

	return data, err
}

// Put stores the data in the cache under the specified key.
func (c *CertificateCache) Put(ctx context.Context, key string, data []byte) error {
	query := `INSERT INTO acme_cache (key, data, updated_at) VALUES($1, $2::bytea, now())
	          ON CONFLICT (key) DO UPDATE SET data = $2::bytea, updated_at = now()`
	_, err := c.storage.db.ExecContext(ctx, query, key, data)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (c *CertificateCache) Delete(ctx context.Context, key string) error {
	query := `DELETE FROM acme_cache WHERE key = $1`
	_, err := c.storage.db.ExecContext(ctx, query, key)
	if err != nil {
		return err
	}
	return nil
}
