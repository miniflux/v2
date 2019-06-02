// Copyright 2019 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package config // import "miniflux.app/config"

const (
	defaultBaseURL            = "http://localhost"
	defaultWorkerPoolSize     = 5
	defaultPollingFrequency   = 60
	defaultBatchSize          = 10
	defaultDatabaseURL        = "user=postgres password=postgres dbname=miniflux2 sslmode=disable"
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

// Options contains configuration options.
type Options struct {
	HTTPS                     bool
	hsts                      bool
	httpService               bool
	schedulerService          bool
	debug                     bool
	baseURL                   string
	rootURL                   string
	basePath                  string
	databaseURL               string
	databaseMaxConns          int
	databaseMinConns          int
	runMigrations             bool
	listenAddr                string
	certFile                  string
	certDomain                string
	certCache                 string
	certKeyFile               string
	cleanupFrequency          int
	archiveReadDays           int
	pollingFrequency          int
	batchSize                 int
	workerPoolSize            int
	createAdmin               bool
	proxyImages               string
	oauth2UserCreationAllowed bool
	oauth2ClientID            string
	oauth2ClientSecret        string
	oauth2RedirectURL         string
	oauth2Provider            string
	pocketConsumerKey         string
}

// HasDebugMode returns true if debug mode is enabled.
func (o *Options) HasDebugMode() bool {
	return o.debug
}

// BaseURL returns the application base URL with path.
func (o *Options) BaseURL() string {
	return o.baseURL
}

// RootURL returns the base URL without path.
func (o *Options) RootURL() string {
	return o.rootURL
}

// BasePath returns the application base path according to the base URL.
func (o *Options) BasePath() string {
	return o.basePath
}

// IsDefaultDatabaseURL returns true if the default database URL is used.
func (o *Options) IsDefaultDatabaseURL() bool {
	return o.databaseURL == defaultDatabaseURL
}

// DatabaseURL returns the database URL.
func (o *Options) DatabaseURL() string {
	return o.databaseURL
}

// DatabaseMaxConns returns the maximum number of database connections.
func (o *Options) DatabaseMaxConns() int {
	return o.databaseMaxConns
}

// DatabaseMinConns returns the minimum number of database connections.
func (o *Options) DatabaseMinConns() int {
	return o.databaseMinConns
}

// ListenAddr returns the listen address for the HTTP server.
func (o *Options) ListenAddr() string {
	return o.listenAddr
}

// CertFile returns the SSL certificate filename if any.
func (o *Options) CertFile() string {
	return o.certFile
}

// CertKeyFile returns the private key filename for custom SSL certificate.
func (o *Options) CertKeyFile() string {
	return o.certKeyFile
}

// CertDomain returns the domain to use for Let's Encrypt certificate.
func (o *Options) CertDomain() string {
	return o.certDomain
}

// CertCache returns the directory to use for Let's Encrypt session cache.
func (o *Options) CertCache() string {
	return o.certCache
}

// CleanupFrequency returns the interval for cleanup jobs.
func (o *Options) CleanupFrequency() int {
	return o.cleanupFrequency
}

// WorkerPoolSize returns the number of background worker.
func (o *Options) WorkerPoolSize() int {
	return o.workerPoolSize
}

// PollingFrequency returns the interval to refresh feeds in the background.
func (o *Options) PollingFrequency() int {
	return o.pollingFrequency
}

// BatchSize returns the number of feeds to send for background processing.
func (o *Options) BatchSize() int {
	return o.batchSize
}

// IsOAuth2UserCreationAllowed returns true if user creation is allowed for OAuth2 users.
func (o *Options) IsOAuth2UserCreationAllowed() bool {
	return o.oauth2UserCreationAllowed
}

// OAuth2ClientID returns the OAuth2 Client ID.
func (o *Options) OAuth2ClientID() string {
	return o.oauth2ClientID
}

// OAuth2ClientSecret returns the OAuth2 client secret.
func (o *Options) OAuth2ClientSecret() string {
	return o.oauth2ClientSecret
}

// OAuth2RedirectURL returns the OAuth2 redirect URL.
func (o *Options) OAuth2RedirectURL() string {
	return o.oauth2RedirectURL
}

// OAuth2Provider returns the name of the OAuth2 provider configured.
func (o *Options) OAuth2Provider() string {
	return o.oauth2Provider
}

// HasHSTS returns true if HTTP Strict Transport Security is enabled.
func (o *Options) HasHSTS() bool {
	return o.hsts
}

// RunMigrations returns true if the environment variable RUN_MIGRATIONS is not empty.
func (o *Options) RunMigrations() bool {
	return o.runMigrations
}

// CreateAdmin returns true if the environment variable CREATE_ADMIN is not empty.
func (o *Options) CreateAdmin() bool {
	return o.createAdmin
}

// ProxyImages returns "none" to never proxy, "http-only" to proxy non-HTTPS, "all" to always proxy.
func (o *Options) ProxyImages() string {
	return o.proxyImages
}

// HasHTTPService returns true if the HTTP service is enabled.
func (o *Options) HasHTTPService() bool {
	return o.httpService
}

// HasSchedulerService returns true if the scheduler service is enabled.
func (o *Options) HasSchedulerService() bool {
	return o.schedulerService
}

// ArchiveReadDays returns the number of days after which marking read items as removed.
func (o *Options) ArchiveReadDays() int {
	return o.archiveReadDays
}

// PocketConsumerKey returns the Pocket Consumer Key if configured.
func (o *Options) PocketConsumerKey(defaultValue string) string {
	if o.pocketConsumerKey != "" {
		return o.pocketConsumerKey
	}
	return defaultValue
}
