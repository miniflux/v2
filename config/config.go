// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package config

import (
	"os"
	"strconv"
)

const (
	defaultBaseURL                 = "http://localhost"
	defaultDatabaseURL             = "postgres://postgres:postgres@localhost/miniflux2?sslmode=disable"
	defaultWorkerPoolSize          = 5
	defaultPollingFrequency        = 60
	defaultBatchSize               = 10
	defaultDatabaseMaxConns        = 20
	defaultListenAddr              = "127.0.0.1:8080"
	defaultCertFile                = ""
	defaultKeyFile                 = ""
	defaultCertDomain              = ""
	defaultCertCache               = "/tmp/cert_cache"
	defaultSessionCleanupFrequency = 24
)

// Config manages configuration parameters.
type Config struct {
	IsHTTPS bool
}

func (c *Config) get(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func (c *Config) getInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	v, _ := strconv.Atoi(value)
	return v
}

// HasDebugMode returns true if debug mode is enabled.
func (c *Config) HasDebugMode() bool {
	return c.get("DEBUG", "") != ""
}

// BaseURL returns the application base URL.
func (c *Config) BaseURL() string {
	return c.get("BASE_URL", defaultBaseURL)
}

// DatabaseURL returns the database URL.
func (c *Config) DatabaseURL() string {
	return c.get("DATABASE_URL", defaultDatabaseURL)
}

// DatabaseMaxConnections returns the number of maximum database connections.
func (c *Config) DatabaseMaxConnections() int {
	return c.getInt("DATABASE_MAX_CONNS", defaultDatabaseMaxConns)
}

// ListenAddr returns the listen address for the HTTP server.
func (c *Config) ListenAddr() string {
	return c.get("LISTEN_ADDR", defaultListenAddr)
}

// CertFile returns the SSL certificate filename if any.
func (c *Config) CertFile() string {
	return c.get("CERT_FILE", defaultCertFile)
}

// KeyFile returns the private key filename for custom SSL certificate.
func (c *Config) KeyFile() string {
	return c.get("KEY_FILE", defaultKeyFile)
}

// CertDomain returns the domain to use for Let's Encrypt certificate.
func (c *Config) CertDomain() string {
	return c.get("CERT_DOMAIN", defaultCertDomain)
}

// CertCache returns the directory to use for Let's Encrypt session cache.
func (c *Config) CertCache() string {
	return c.get("CERT_CACHE", defaultCertCache)
}

// SessionCleanupFrequency returns the interval for session cleanup.
func (c *Config) SessionCleanupFrequency() int {
	return c.getInt("SESSION_CLEANUP_FREQUENCY", defaultSessionCleanupFrequency)
}

// WorkerPoolSize returns the number of background worker.
func (c *Config) WorkerPoolSize() int {
	return c.getInt("WORKER_POOL_SIZE", defaultWorkerPoolSize)
}

// PollingFrequency returns the interval to refresh feeds in the background.
func (c *Config) PollingFrequency() int {
	return c.getInt("POLLING_FREQUENCY", defaultPollingFrequency)
}

// BatchSize returns the number of feeds to send for background processing.
func (c *Config) BatchSize() int {
	return c.getInt("BATCH_SIZE", defaultBatchSize)
}

// IsOAuth2UserCreationAllowed returns true if user creation is allowed for OAuth2 users.
func (c *Config) IsOAuth2UserCreationAllowed() bool {
	return c.getInt("OAUTH2_USER_CREATION", 0) == 1
}

// OAuth2ClientID returns the OAuth2 Client ID.
func (c *Config) OAuth2ClientID() string {
	return c.get("OAUTH2_CLIENT_ID", "")
}

// OAuth2ClientSecret returns the OAuth2 client secret.
func (c *Config) OAuth2ClientSecret() string {
	return c.get("OAUTH2_CLIENT_SECRET", "")
}

// OAuth2RedirectURL returns the OAuth2 redirect URL.
func (c *Config) OAuth2RedirectURL() string {
	return c.get("OAUTH2_REDIRECT_URL", "")
}

// OAuth2Provider returns the name of the OAuth2 provider configured.
func (c *Config) OAuth2Provider() string {
	return c.get("OAUTH2_PROVIDER", "")
}

// NewConfig returns a new Config.
func NewConfig() *Config {
	return &Config{IsHTTPS: os.Getenv("HTTPS") != ""}
}
