// Copyright 2019 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package config // import "miniflux.app/config"

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"miniflux.app/version"
)

const (
	defaultHTTPS                              = false
	defaultLogDateTime                        = false
	defaultHSTS                               = true
	defaultHTTPService                        = true
	defaultSchedulerService                   = true
	defaultDebug                              = false
	defaultTiming                             = false
	defaultBaseURL                            = "http://localhost"
	defaultRootURL                            = "http://localhost"
	defaultBasePath                           = ""
	defaultWorkerPoolSize                     = 5
	defaultPollingFrequency                   = 60
	defaultBatchSize                          = 100
	defaultPollingScheduler                   = "round_robin"
	defaultSchedulerEntryFrequencyMinInterval = 5
	defaultSchedulerEntryFrequencyMaxInterval = 24 * 60
	defaultPollingParsingErrorLimit           = 3
	defaultRunMigrations                      = false
	defaultDatabaseURL                        = "user=postgres password=postgres dbname=miniflux2 sslmode=disable"
	defaultDatabaseMaxConns                   = 20
	defaultDatabaseMinConns                   = 1
	defaultDatabaseConnectionLifetime         = 5
	defaultListenAddr                         = "127.0.0.1:8080"
	defaultCertFile                           = ""
	defaultKeyFile                            = ""
	defaultCertDomain                         = ""
	defaultCleanupFrequencyHours              = 24
	defaultCleanupArchiveReadDays             = 60
	defaultCleanupArchiveUnreadDays           = 180
	defaultCleanupArchiveBatchSize            = 10000
	defaultCleanupRemoveSessionsDays          = 30
	defaultProxyImages                        = "http-only"
	defaultFetchYouTubeWatchTime              = false
	defaultCreateAdmin                        = false
	defaultAdminUsername                      = ""
	defaultAdminPassword                      = ""
	defaultOAuth2UserCreation                 = false
	defaultOAuth2ClientID                     = ""
	defaultOAuth2ClientSecret                 = ""
	defaultOAuth2RedirectURL                  = ""
	defaultOAuth2OidcDiscoveryEndpoint        = ""
	defaultOAuth2Provider                     = ""
	defaultPocketConsumerKey                  = ""
	defaultHTTPClientTimeout                  = 20
	defaultHTTPClientMaxBodySize              = 15
	defaultHTTPClientProxy                    = ""
	defaultAuthProxyHeader                    = ""
	defaultAuthProxyUserCreation              = false
	defaultMaintenanceMode                    = false
	defaultMaintenanceMessage                 = "Miniflux is currently under maintenance"
	defaultMetricsCollector                   = false
	defaultMetricsRefreshInterval             = 60
	defaultMetricsAllowedNetworks             = "127.0.0.1/8"
	defaultWatchdog                           = true
)

var defaultHTTPClientUserAgent = "Mozilla/5.0 (compatible; Miniflux/" + version.Version + "; +https://miniflux.app)"

// Option contains a key to value map of a single option. It may be used to output debug strings.
type Option struct {
	Key   string
	Value interface{}
}

// Options contains configuration options.
type Options struct {
	HTTPS                              bool
	logDateTime                        bool
	hsts                               bool
	httpService                        bool
	schedulerService                   bool
	debug                              bool
	serverTimingHeader                 bool
	baseURL                            string
	rootURL                            string
	basePath                           string
	databaseURL                        string
	databaseMaxConns                   int
	databaseMinConns                   int
	databaseConnectionLifetime         int
	runMigrations                      bool
	listenAddr                         string
	certFile                           string
	certDomain                         string
	certKeyFile                        string
	cleanupFrequencyHours              int
	cleanupArchiveReadDays             int
	cleanupArchiveUnreadDays           int
	cleanupArchiveBatchSize            int
	cleanupRemoveSessionsDays          int
	pollingFrequency                   int
	batchSize                          int
	pollingScheduler                   string
	schedulerEntryFrequencyMinInterval int
	schedulerEntryFrequencyMaxInterval int
	pollingParsingErrorLimit           int
	workerPoolSize                     int
	createAdmin                        bool
	adminUsername                      string
	adminPassword                      string
	proxyImages                        string
	fetchYouTubeWatchTime              bool
	oauth2UserCreationAllowed          bool
	oauth2ClientID                     string
	oauth2ClientSecret                 string
	oauth2RedirectURL                  string
	oauth2OidcDiscoveryEndpoint        string
	oauth2Provider                     string
	pocketConsumerKey                  string
	httpClientTimeout                  int
	httpClientMaxBodySize              int64
	httpClientProxy                    string
	httpClientUserAgent                string
	authProxyHeader                    string
	authProxyUserCreation              bool
	maintenanceMode                    bool
	maintenanceMessage                 string
	metricsCollector                   bool
	metricsRefreshInterval             int
	metricsAllowedNetworks             []string
	watchdog                           bool
}

// NewOptions returns Options with default values.
func NewOptions() *Options {
	return &Options{
		HTTPS:                              defaultHTTPS,
		logDateTime:                        defaultLogDateTime,
		hsts:                               defaultHSTS,
		httpService:                        defaultHTTPService,
		schedulerService:                   defaultSchedulerService,
		debug:                              defaultDebug,
		serverTimingHeader:                 defaultTiming,
		baseURL:                            defaultBaseURL,
		rootURL:                            defaultRootURL,
		basePath:                           defaultBasePath,
		databaseURL:                        defaultDatabaseURL,
		databaseMaxConns:                   defaultDatabaseMaxConns,
		databaseMinConns:                   defaultDatabaseMinConns,
		databaseConnectionLifetime:         defaultDatabaseConnectionLifetime,
		runMigrations:                      defaultRunMigrations,
		listenAddr:                         defaultListenAddr,
		certFile:                           defaultCertFile,
		certDomain:                         defaultCertDomain,
		certKeyFile:                        defaultKeyFile,
		cleanupFrequencyHours:              defaultCleanupFrequencyHours,
		cleanupArchiveReadDays:             defaultCleanupArchiveReadDays,
		cleanupArchiveUnreadDays:           defaultCleanupArchiveUnreadDays,
		cleanupArchiveBatchSize:            defaultCleanupArchiveBatchSize,
		cleanupRemoveSessionsDays:          defaultCleanupRemoveSessionsDays,
		pollingFrequency:                   defaultPollingFrequency,
		batchSize:                          defaultBatchSize,
		pollingScheduler:                   defaultPollingScheduler,
		schedulerEntryFrequencyMinInterval: defaultSchedulerEntryFrequencyMinInterval,
		schedulerEntryFrequencyMaxInterval: defaultSchedulerEntryFrequencyMaxInterval,
		pollingParsingErrorLimit:           defaultPollingParsingErrorLimit,
		workerPoolSize:                     defaultWorkerPoolSize,
		createAdmin:                        defaultCreateAdmin,
		proxyImages:                        defaultProxyImages,
		fetchYouTubeWatchTime:              defaultFetchYouTubeWatchTime,
		oauth2UserCreationAllowed:          defaultOAuth2UserCreation,
		oauth2ClientID:                     defaultOAuth2ClientID,
		oauth2ClientSecret:                 defaultOAuth2ClientSecret,
		oauth2RedirectURL:                  defaultOAuth2RedirectURL,
		oauth2OidcDiscoveryEndpoint:        defaultOAuth2OidcDiscoveryEndpoint,
		oauth2Provider:                     defaultOAuth2Provider,
		pocketConsumerKey:                  defaultPocketConsumerKey,
		httpClientTimeout:                  defaultHTTPClientTimeout,
		httpClientMaxBodySize:              defaultHTTPClientMaxBodySize * 1024 * 1024,
		httpClientProxy:                    defaultHTTPClientProxy,
		httpClientUserAgent:                defaultHTTPClientUserAgent,
		authProxyHeader:                    defaultAuthProxyHeader,
		authProxyUserCreation:              defaultAuthProxyUserCreation,
		maintenanceMode:                    defaultMaintenanceMode,
		maintenanceMessage:                 defaultMaintenanceMessage,
		metricsCollector:                   defaultMetricsCollector,
		metricsRefreshInterval:             defaultMetricsRefreshInterval,
		metricsAllowedNetworks:             []string{defaultMetricsAllowedNetworks},
		watchdog:                           defaultWatchdog,
	}
}

// LogDateTime returns true if the date/time should be displayed in log messages.
func (o *Options) LogDateTime() bool {
	return o.logDateTime
}

// HasMaintenanceMode returns true if maintenance mode is enabled.
func (o *Options) HasMaintenanceMode() bool {
	return o.maintenanceMode
}

// MaintenanceMessage returns maintenance message.
func (o *Options) MaintenanceMessage() string {
	return o.maintenanceMessage
}

// HasDebugMode returns true if debug mode is enabled.
func (o *Options) HasDebugMode() bool {
	return o.debug
}

// HasServerTimingHeader returns true if server-timing headers enabled.
func (o *Options) HasServerTimingHeader() bool {
	return o.serverTimingHeader
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

// DatabaseConnectionLifetime returns the maximum amount of time a connection may be reused.
func (o *Options) DatabaseConnectionLifetime() time.Duration {
	return time.Duration(o.databaseConnectionLifetime) * time.Minute
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

// CleanupFrequencyHours returns the interval in hours for cleanup jobs.
func (o *Options) CleanupFrequencyHours() int {
	return o.cleanupFrequencyHours
}

// CleanupArchiveReadDays returns the number of days after which marking read items as removed.
func (o *Options) CleanupArchiveReadDays() int {
	return o.cleanupArchiveReadDays
}

// CleanupArchiveUnreadDays returns the number of days after which marking unread items as removed.
func (o *Options) CleanupArchiveUnreadDays() int {
	return o.cleanupArchiveUnreadDays
}

// CleanupArchiveBatchSize returns the number of entries to archive for each interval.
func (o *Options) CleanupArchiveBatchSize() int {
	return o.cleanupArchiveBatchSize
}

// CleanupRemoveSessionsDays returns the number of days after which to remove sessions.
func (o *Options) CleanupRemoveSessionsDays() int {
	return o.cleanupRemoveSessionsDays
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

// PollingScheduler returns the scheduler used for polling feeds.
func (o *Options) PollingScheduler() string {
	return o.pollingScheduler
}

// SchedulerEntryFrequencyMaxInterval returns the maximum interval in minutes for the entry frequency scheduler.
func (o *Options) SchedulerEntryFrequencyMaxInterval() int {
	return o.schedulerEntryFrequencyMaxInterval
}

// SchedulerEntryFrequencyMinInterval returns the minimum interval in minutes for the entry frequency scheduler.
func (o *Options) SchedulerEntryFrequencyMinInterval() int {
	return o.schedulerEntryFrequencyMinInterval
}

// PollingParsingErrorLimit returns the limit of errors when to stop polling.
func (o *Options) PollingParsingErrorLimit() int {
	return o.pollingParsingErrorLimit
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

// OAuth2OidcDiscoveryEndpoint returns the OAuth2 OIDC discovery endpoint.
func (o *Options) OAuth2OidcDiscoveryEndpoint() string {
	return o.oauth2OidcDiscoveryEndpoint
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

// AdminUsername returns the admin username if defined.
func (o *Options) AdminUsername() string {
	return o.adminUsername
}

// AdminPassword returns the admin password if defined.
func (o *Options) AdminPassword() string {
	return o.adminPassword
}

// FetchYouTubeWatchTime returns true if the YouTube video duration
// should be fetched and used as a reading time.
func (o *Options) FetchYouTubeWatchTime() bool {
	return o.fetchYouTubeWatchTime
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

// PocketConsumerKey returns the Pocket Consumer Key if configured.
func (o *Options) PocketConsumerKey(defaultValue string) string {
	if o.pocketConsumerKey != "" {
		return o.pocketConsumerKey
	}
	return defaultValue
}

// HTTPClientTimeout returns the time limit in seconds before the HTTP client cancel the request.
func (o *Options) HTTPClientTimeout() int {
	return o.httpClientTimeout
}

// HTTPClientMaxBodySize returns the number of bytes allowed for the HTTP client to transfer.
func (o *Options) HTTPClientMaxBodySize() int64 {
	return o.httpClientMaxBodySize
}

// HTTPClientProxy returns the proxy URL for HTTP client.
func (o *Options) HTTPClientProxy() string {
	return o.httpClientProxy
}

// HasHTTPClientProxyConfigured returns true if the HTTP proxy is configured.
func (o *Options) HasHTTPClientProxyConfigured() bool {
	return o.httpClientProxy != ""
}

// AuthProxyHeader returns an HTTP header name that contains username for
// authentication using auth proxy.
func (o *Options) AuthProxyHeader() string {
	return o.authProxyHeader
}

// IsAuthProxyUserCreationAllowed returns true if user creation is allowed for
// users authenticated using auth proxy.
func (o *Options) IsAuthProxyUserCreationAllowed() bool {
	return o.authProxyUserCreation
}

// HasMetricsCollector returns true if metrics collection is enabled.
func (o *Options) HasMetricsCollector() bool {
	return o.metricsCollector
}

// MetricsRefreshInterval returns the refresh interval in seconds.
func (o *Options) MetricsRefreshInterval() int {
	return o.metricsRefreshInterval
}

// MetricsAllowedNetworks returns the list of networks allowed to connect to the metrics endpoint.
func (o *Options) MetricsAllowedNetworks() []string {
	return o.metricsAllowedNetworks
}

// HTTPClientUserAgent returns the global User-Agent header for miniflux.
func (o *Options) HTTPClientUserAgent() string {
	return o.httpClientUserAgent
}

// HasWatchdog returns true if the systemd watchdog is enabled.
func (o *Options) HasWatchdog() bool {
	return o.watchdog
}

// SortedOptions returns options as a list of key value pairs, sorted by keys.
func (o *Options) SortedOptions() []*Option {
	var keyValues = map[string]interface{}{
		"ADMIN_PASSWORD":                         o.adminPassword,
		"ADMIN_USERNAME":                         o.adminUsername,
		"AUTH_PROXY_HEADER":                      o.authProxyHeader,
		"AUTH_PROXY_USER_CREATION":               o.authProxyUserCreation,
		"BASE_PATH":                              o.basePath,
		"BASE_URL":                               o.baseURL,
		"BATCH_SIZE":                             o.batchSize,
		"CERT_DOMAIN":                            o.certDomain,
		"CERT_FILE":                              o.certFile,
		"CLEANUP_ARCHIVE_READ_DAYS":              o.cleanupArchiveReadDays,
		"CLEANUP_ARCHIVE_UNREAD_DAYS":            o.cleanupArchiveUnreadDays,
		"CLEANUP_ARCHIVE_BATCH_SIZE":             o.cleanupArchiveBatchSize,
		"CLEANUP_FREQUENCY_HOURS":                o.cleanupFrequencyHours,
		"CLEANUP_REMOVE_SESSIONS_DAYS":           o.cleanupRemoveSessionsDays,
		"CREATE_ADMIN":                           o.createAdmin,
		"DATABASE_MAX_CONNS":                     o.databaseMaxConns,
		"DATABASE_MIN_CONNS":                     o.databaseMinConns,
		"DATABASE_CONNECTION_LIFETIME":           o.databaseConnectionLifetime,
		"DATABASE_URL":                           o.databaseURL,
		"DEBUG":                                  o.debug,
		"FETCH_YOUTUBE_WATCH_TIME":               o.fetchYouTubeWatchTime,
		"HSTS":                                   o.hsts,
		"HTTPS":                                  o.HTTPS,
		"HTTP_CLIENT_MAX_BODY_SIZE":              o.httpClientMaxBodySize,
		"HTTP_CLIENT_PROXY":                      o.httpClientProxy,
		"HTTP_CLIENT_TIMEOUT":                    o.httpClientTimeout,
		"HTTP_CLIENT_USER_AGENT":                 o.httpClientUserAgent,
		"HTTP_SERVICE":                           o.httpService,
		"KEY_FILE":                               o.certKeyFile,
		"LISTEN_ADDR":                            o.listenAddr,
		"LOG_DATE_TIME":                          o.logDateTime,
		"MAINTENANCE_MESSAGE":                    o.maintenanceMessage,
		"MAINTENANCE_MODE":                       o.maintenanceMode,
		"METRICS_ALLOWED_NETWORKS":               o.metricsAllowedNetworks,
		"METRICS_COLLECTOR":                      o.metricsCollector,
		"METRICS_REFRESH_INTERVAL":               o.metricsRefreshInterval,
		"OAUTH2_CLIENT_ID":                       o.oauth2ClientID,
		"OAUTH2_CLIENT_SECRET":                   o.oauth2ClientSecret,
		"OAUTH2_OIDC_DISCOVERY_ENDPOINT":         o.oauth2OidcDiscoveryEndpoint,
		"OAUTH2_PROVIDER":                        o.oauth2Provider,
		"OAUTH2_REDIRECT_URL":                    o.oauth2RedirectURL,
		"OAUTH2_USER_CREATION":                   o.oauth2UserCreationAllowed,
		"POCKET_CONSUMER_KEY":                    o.pocketConsumerKey,
		"POLLING_FREQUENCY":                      o.pollingFrequency,
		"POLLING_PARSING_ERROR_LIMIT":            o.pollingParsingErrorLimit,
		"POLLING_SCHEDULER":                      o.pollingScheduler,
		"PROXY_IMAGES":                           o.proxyImages,
		"ROOT_URL":                               o.rootURL,
		"RUN_MIGRATIONS":                         o.runMigrations,
		"SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL": o.schedulerEntryFrequencyMaxInterval,
		"SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL": o.schedulerEntryFrequencyMinInterval,
		"SCHEDULER_SERVICE":                      o.schedulerService,
		"SERVER_TIMING_HEADER":                   o.serverTimingHeader,
		"WORKER_POOL_SIZE":                       o.workerPoolSize,
		"WATCHDOG":                               o.watchdog,
	}

	keys := make([]string, 0, len(keyValues))
	for key := range keyValues {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var sortedOptions []*Option
	for _, key := range keys {
		sortedOptions = append(sortedOptions, &Option{Key: key, Value: keyValues[key]})
	}
	return sortedOptions
}

func (o *Options) String() string {
	var builder strings.Builder

	for _, option := range o.SortedOptions() {
		builder.WriteString(fmt.Sprintf("%s: %v\n", option.Key, option.Value))
	}

	return builder.String()
}
