// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package config

import (
	"os"
	"strconv"
)

// Default config parameters values
const (
	DefaultBaseURL                 = "http://localhost"
	DefaultDatabaseURL             = "postgres://postgres:postgres@localhost/miniflux2?sslmode=disable"
	DefaultWorkerPoolSize          = 5
	DefaultPollingFrequency        = 60
	DefaultBatchSize               = 10
	DefaultDatabaseMaxConns        = 20
	DefaultListenAddr              = "127.0.0.1:8080"
	DefaultCertFile                = ""
	DefaultKeyFile                 = ""
	DefaultCertDomain              = ""
	DefaultCertCache               = "/tmp/cert_cache"
	DefaultSessionCleanupFrequency = 24
)

// Config manages configuration parameters.
type Config struct {
	IsHTTPS bool
}

// Get returns a config parameter value.
func (c *Config) Get(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

// GetInt returns a config parameter as integer.
func (c *Config) GetInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	v, _ := strconv.Atoi(value)
	return v
}

// BaseURL returns the application base URL.
func (c *Config) BaseURL() string {
	return c.Get("BASE_URL", DefaultBaseURL)
}

// DatabaseURL returns the database URL.
func (c *Config) DatabaseURL() string {
	return c.Get("DATABASE_URL", DefaultDatabaseURL)
}

// DatabaseMaxConnections returns the number of maximum database connections.
func (c *Config) DatabaseMaxConnections() int {
	return c.GetInt("DATABASE_MAX_CONNS", DefaultDatabaseMaxConns)
}

// ListenAddr returns the listen address for the HTTP server.
func (c *Config) ListenAddr() string {
	return c.Get("LISTEN_ADDR", DefaultListenAddr)
}

// CertFile returns the SSL certificate filename if any.
func (c *Config) CertFile() string {
	return c.Get("CERT_FILE", DefaultCertFile)
}

// KeyFile returns the private key filename for custom SSL certificate.
func (c *Config) KeyFile() string {
	return c.Get("KEY_FILE", DefaultKeyFile)
}

// CertDomain returns the domain to use for Let's Encrypt certificate.
func (c *Config) CertDomain() string {
	return c.Get("CERT_DOMAIN", DefaultCertDomain)
}

// CertCache returns the directory to use for Let's Encrypt session cache.
func (c *Config) CertCache() string {
	return c.Get("CERT_CACHE", DefaultCertCache)
}

// SessionCleanupFrequency returns the interval for session cleanup.
func (c *Config) SessionCleanupFrequency() int {
	return c.GetInt("SESSION_CLEANUP_FREQUENCY", DefaultSessionCleanupFrequency)
}

// WorkerPoolSize returns the number of background worker.
func (c *Config) WorkerPoolSize() int {
	return c.GetInt("WORKER_POOL_SIZE", DefaultWorkerPoolSize)
}

// PollingFrequency returns the interval to refresh feeds in the background.
func (c *Config) PollingFrequency() int {
	return c.GetInt("POLLING_FREQUENCY", DefaultPollingFrequency)
}

// BatchSize returns the number of feeds to send for background processing.
func (c *Config) BatchSize() int {
	return c.GetInt("BATCH_SIZE", DefaultBatchSize)
}

// NewConfig returns a new Config.
func NewConfig() *Config {
	return &Config{IsHTTPS: os.Getenv("HTTPS") != ""}
}
