// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package config // import "miniflux.app/config"

import (
	"net/url"
	"os"
	"strconv"
	"strings"

	"miniflux.app/logger"
)

const (
	defaultBaseURL            = "http://localhost"
	defaultDatabaseURL        = "user=postgres password=postgres dbname=miniflux2 sslmode=disable"
	defaultWorkerPoolSize     = 5
	defaultPollingFrequency   = 60
	defaultBatchSize          = 10
	defaultDatabaseMaxConns   = 20
	defaultDatabaseMinConns   = 1
	defaultArchiveReadDays    = 60
	defaultListenAddr         = "127.0.0.1:8080"
	defaultCertFile           = ""
	defaultKeyFile            = ""
	defaultCertDomain         = ""
	defaultCertCache          = "/tmp/cert_cache"
	defaultCleanupFrequency   = 24
	defaultProxyImages        = "http-only"
	defaultOAuth2ClientID     = ""
	defaultOAuth2ClientSecret = ""
	defaultOAuth2RedirectURL  = ""
	defaultOAuth2Provider     = ""
)

// Config manages configuration parameters.
type Config struct {
	IsHTTPS  bool
	baseURL  string
	rootURL  string
	basePath string
}

func (c *Config) parseBaseURL() {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		return
	}

	if baseURL[len(baseURL)-1:] == "/" {
		baseURL = baseURL[:len(baseURL)-1]
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		logger.Error("Invalid BASE_URL: %v", err)
		return
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "https" && scheme != "http" {
		logger.Error("Invalid BASE_URL: scheme must be http or https")
		return
	}

	c.baseURL = baseURL
	c.basePath = u.Path

	u.Path = ""
	c.rootURL = u.String()
}

// HasDebugMode returns true if debug mode is enabled.
func (c *Config) HasDebugMode() bool {
	return getBooleanValue("DEBUG")
}

// BaseURL returns the application base URL with path.
func (c *Config) BaseURL() string {
	return c.baseURL
}

// RootURL returns the base URL without path.
func (c *Config) RootURL() string {
	return c.rootURL
}

// BasePath returns the application base path according to the base URL.
func (c *Config) BasePath() string {
	return c.basePath
}

// DatabaseURL returns the database URL.
func (c *Config) DatabaseURL() string {
	value, exists := os.LookupEnv("DATABASE_URL")
	if !exists {
		logger.Info("The environment variable DATABASE_URL is not configured (the default value is used instead)")
	}

	if value == "" {
		value = defaultDatabaseURL
	}

	return value
}

// DatabaseMaxConns returns the maximum number of database connections.
func (c *Config) DatabaseMaxConns() int {
	return getIntValue("DATABASE_MAX_CONNS", defaultDatabaseMaxConns)
}

// DatabaseMinConns returns the minimum number of database connections.
func (c *Config) DatabaseMinConns() int {
	return getIntValue("DATABASE_MIN_CONNS", defaultDatabaseMinConns)
}

// ListenAddr returns the listen address for the HTTP server.
func (c *Config) ListenAddr() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}

	return getStringValue("LISTEN_ADDR", defaultListenAddr)
}

// CertFile returns the SSL certificate filename if any.
func (c *Config) CertFile() string {
	return getStringValue("CERT_FILE", defaultCertFile)
}

// KeyFile returns the private key filename for custom SSL certificate.
func (c *Config) KeyFile() string {
	return getStringValue("KEY_FILE", defaultKeyFile)
}

// CertDomain returns the domain to use for Let's Encrypt certificate.
func (c *Config) CertDomain() string {
	return getStringValue("CERT_DOMAIN", defaultCertDomain)
}

// CertCache returns the directory to use for Let's Encrypt session cache.
func (c *Config) CertCache() string {
	return getStringValue("CERT_CACHE", defaultCertCache)
}

// CleanupFrequency returns the interval for cleanup jobs.
func (c *Config) CleanupFrequency() int {
	return getIntValue("CLEANUP_FREQUENCY", defaultCleanupFrequency)
}

// WorkerPoolSize returns the number of background worker.
func (c *Config) WorkerPoolSize() int {
	return getIntValue("WORKER_POOL_SIZE", defaultWorkerPoolSize)
}

// PollingFrequency returns the interval to refresh feeds in the background.
func (c *Config) PollingFrequency() int {
	return getIntValue("POLLING_FREQUENCY", defaultPollingFrequency)
}

// BatchSize returns the number of feeds to send for background processing.
func (c *Config) BatchSize() int {
	return getIntValue("BATCH_SIZE", defaultBatchSize)
}

// IsOAuth2UserCreationAllowed returns true if user creation is allowed for OAuth2 users.
func (c *Config) IsOAuth2UserCreationAllowed() bool {
	return getBooleanValue("OAUTH2_USER_CREATION")
}

// OAuth2ClientID returns the OAuth2 Client ID.
func (c *Config) OAuth2ClientID() string {
	return getStringValue("OAUTH2_CLIENT_ID", defaultOAuth2ClientID)
}

// OAuth2ClientSecret returns the OAuth2 client secret.
func (c *Config) OAuth2ClientSecret() string {
	return getStringValue("OAUTH2_CLIENT_SECRET", defaultOAuth2ClientSecret)
}

// OAuth2RedirectURL returns the OAuth2 redirect URL.
func (c *Config) OAuth2RedirectURL() string {
	return getStringValue("OAUTH2_REDIRECT_URL", defaultOAuth2RedirectURL)
}

// OAuth2Provider returns the name of the OAuth2 provider configured.
func (c *Config) OAuth2Provider() string {
	return getStringValue("OAUTH2_PROVIDER", defaultOAuth2Provider)
}

// HasHSTS returns true if HTTP Strict Transport Security is enabled.
func (c *Config) HasHSTS() bool {
	return !getBooleanValue("DISABLE_HSTS")
}

// RunMigrations returns true if the environment variable RUN_MIGRATIONS is not empty.
func (c *Config) RunMigrations() bool {
	return getBooleanValue("RUN_MIGRATIONS")
}

// CreateAdmin returns true if the environment variable CREATE_ADMIN is not empty.
func (c *Config) CreateAdmin() bool {
	return getBooleanValue("CREATE_ADMIN")
}

// PocketConsumerKey returns the Pocket Consumer Key if defined as environment variable.
func (c *Config) PocketConsumerKey(defaultValue string) string {
	return getStringValue("POCKET_CONSUMER_KEY", defaultValue)
}

// ProxyImages returns "none" to never proxy, "http-only" to proxy non-HTTPS, "all" to always proxy.
func (c *Config) ProxyImages() string {
	return getStringValue("PROXY_IMAGES", defaultProxyImages)
}

// HasHTTPService returns true if the HTTP service is enabled.
func (c *Config) HasHTTPService() bool {
	return !getBooleanValue("DISABLE_HTTP_SERVICE")
}

// HasSchedulerService returns true if the scheduler service is enabled.
func (c *Config) HasSchedulerService() bool {
	return !getBooleanValue("DISABLE_SCHEDULER_SERVICE")
}

// ArchiveReadDays returns the number of days after which marking read items as removed.
func (c *Config) ArchiveReadDays() int {
	return getIntValue("ARCHIVE_READ_DAYS", defaultArchiveReadDays)
}

// NewConfig returns a new Config.
func NewConfig() *Config {
	cfg := &Config{
		baseURL: defaultBaseURL,
		rootURL: defaultBaseURL,
		IsHTTPS: getBooleanValue("HTTPS"),
	}

	cfg.parseBaseURL()
	return cfg
}

func getBooleanValue(key string) bool {
	value := strings.ToLower(os.Getenv(key))
	if value == "1" || value == "yes" || value == "true" || value == "on" {
		return true
	}
	return false
}

func getStringValue(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getIntValue(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return v
}
