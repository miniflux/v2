// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config // import "miniflux.app/v2/internal/config"

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/version"
)

const (
	defaultHTTPS                              = false
	defaultLogFile                            = "stderr"
	defaultLogDateTime                        = false
	defaultLogFormat                          = "text"
	defaultLogLevel                           = "info"
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
	defaultForceRefreshInterval               = 30
	defaultBatchSize                          = 100
	defaultPollingScheduler                   = "round_robin"
	defaultSchedulerEntryFrequencyMinInterval = 5
	defaultSchedulerEntryFrequencyMaxInterval = 24 * 60
	defaultSchedulerEntryFrequencyFactor      = 1
	defaultSchedulerRoundRobinMinInterval     = 60
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
	defaultProxyHTTPClientTimeout             = 120
	defaultProxyOption                        = "http-only"
	defaultProxyMediaTypes                    = "image"
	defaultProxyUrl                           = ""
	defaultFetchOdyseeWatchTime               = false
	defaultFetchYouTubeWatchTime              = false
	defaultYouTubeEmbedUrlOverride            = "https://www.youtube-nocookie.com/embed/"
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
	defaultHTTPServerTimeout                  = 300
	defaultAuthProxyHeader                    = ""
	defaultAuthProxyUserCreation              = false
	defaultMaintenanceMode                    = false
	defaultMaintenanceMessage                 = "Miniflux is currently under maintenance"
	defaultMetricsCollector                   = false
	defaultMetricsRefreshInterval             = 60
	defaultMetricsAllowedNetworks             = "127.0.0.1/8"
	defaultMetricsUsername                    = ""
	defaultMetricsPassword                    = ""
	defaultWatchdog                           = true
	defaultInvidiousInstance                  = "yewtu.be"
	defaultWebAuthn                           = false
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
	logFile                            string
	logDateTime                        bool
	logFormat                          string
	logLevel                           string
	hsts                               bool
	httpService                        bool
	schedulerService                   bool
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
	forceRefreshInterval               int
	batchSize                          int
	pollingScheduler                   string
	schedulerEntryFrequencyMinInterval int
	schedulerEntryFrequencyMaxInterval int
	schedulerEntryFrequencyFactor      int
	schedulerRoundRobinMinInterval     int
	pollingParsingErrorLimit           int
	workerPoolSize                     int
	createAdmin                        bool
	adminUsername                      string
	adminPassword                      string
	proxyHTTPClientTimeout             int
	proxyOption                        string
	proxyMediaTypes                    []string
	proxyUrl                           string
	fetchOdyseeWatchTime               bool
	fetchYouTubeWatchTime              bool
	youTubeEmbedUrlOverride            string
	oauth2UserCreationAllowed          bool
	oauth2ClientID                     string
	oauth2ClientSecret                 string
	oauth2RedirectURL                  string
	oidcDiscoveryEndpoint              string
	oauth2Provider                     string
	pocketConsumerKey                  string
	httpClientTimeout                  int
	httpClientMaxBodySize              int64
	httpClientProxy                    string
	httpClientUserAgent                string
	httpServerTimeout                  int
	authProxyHeader                    string
	authProxyUserCreation              bool
	maintenanceMode                    bool
	maintenanceMessage                 string
	metricsCollector                   bool
	metricsRefreshInterval             int
	metricsAllowedNetworks             []string
	metricsUsername                    string
	metricsPassword                    string
	watchdog                           bool
	invidiousInstance                  string
	proxyPrivateKey                    []byte
	webAuthn                           bool
}

// NewOptions returns Options with default values.
func NewOptions() *Options {
	return &Options{
		HTTPS:                              defaultHTTPS,
		logFile:                            defaultLogFile,
		logDateTime:                        defaultLogDateTime,
		logFormat:                          defaultLogFormat,
		logLevel:                           defaultLogLevel,
		hsts:                               defaultHSTS,
		httpService:                        defaultHTTPService,
		schedulerService:                   defaultSchedulerService,
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
		forceRefreshInterval:               defaultForceRefreshInterval,
		batchSize:                          defaultBatchSize,
		pollingScheduler:                   defaultPollingScheduler,
		schedulerEntryFrequencyMinInterval: defaultSchedulerEntryFrequencyMinInterval,
		schedulerEntryFrequencyMaxInterval: defaultSchedulerEntryFrequencyMaxInterval,
		schedulerEntryFrequencyFactor:      defaultSchedulerEntryFrequencyFactor,
		schedulerRoundRobinMinInterval:     defaultSchedulerRoundRobinMinInterval,
		pollingParsingErrorLimit:           defaultPollingParsingErrorLimit,
		workerPoolSize:                     defaultWorkerPoolSize,
		createAdmin:                        defaultCreateAdmin,
		proxyHTTPClientTimeout:             defaultProxyHTTPClientTimeout,
		proxyOption:                        defaultProxyOption,
		proxyMediaTypes:                    []string{defaultProxyMediaTypes},
		proxyUrl:                           defaultProxyUrl,
		fetchOdyseeWatchTime:               defaultFetchOdyseeWatchTime,
		fetchYouTubeWatchTime:              defaultFetchYouTubeWatchTime,
		youTubeEmbedUrlOverride:            defaultYouTubeEmbedUrlOverride,
		oauth2UserCreationAllowed:          defaultOAuth2UserCreation,
		oauth2ClientID:                     defaultOAuth2ClientID,
		oauth2ClientSecret:                 defaultOAuth2ClientSecret,
		oauth2RedirectURL:                  defaultOAuth2RedirectURL,
		oidcDiscoveryEndpoint:              defaultOAuth2OidcDiscoveryEndpoint,
		oauth2Provider:                     defaultOAuth2Provider,
		pocketConsumerKey:                  defaultPocketConsumerKey,
		httpClientTimeout:                  defaultHTTPClientTimeout,
		httpClientMaxBodySize:              defaultHTTPClientMaxBodySize * 1024 * 1024,
		httpClientProxy:                    defaultHTTPClientProxy,
		httpClientUserAgent:                defaultHTTPClientUserAgent,
		httpServerTimeout:                  defaultHTTPServerTimeout,
		authProxyHeader:                    defaultAuthProxyHeader,
		authProxyUserCreation:              defaultAuthProxyUserCreation,
		maintenanceMode:                    defaultMaintenanceMode,
		maintenanceMessage:                 defaultMaintenanceMessage,
		metricsCollector:                   defaultMetricsCollector,
		metricsRefreshInterval:             defaultMetricsRefreshInterval,
		metricsAllowedNetworks:             []string{defaultMetricsAllowedNetworks},
		metricsUsername:                    defaultMetricsUsername,
		metricsPassword:                    defaultMetricsPassword,
		watchdog:                           defaultWatchdog,
		invidiousInstance:                  defaultInvidiousInstance,
		proxyPrivateKey:                    crypto.GenerateRandomBytes(16),
		webAuthn:                           defaultWebAuthn,
	}
}

func (o *Options) LogFile() string {
	return o.logFile
}

// LogDateTime returns true if the date/time should be displayed in log messages.
func (o *Options) LogDateTime() bool {
	return o.logDateTime
}

// LogFormat returns the log format.
func (o *Options) LogFormat() string {
	return o.logFormat
}

// LogLevel returns the log level.
func (o *Options) LogLevel() string {
	return o.logLevel
}

// SetLogLevel sets the log level.
func (o *Options) SetLogLevel(level string) {
	o.logLevel = level
}

// HasMaintenanceMode returns true if maintenance mode is enabled.
func (o *Options) HasMaintenanceMode() bool {
	return o.maintenanceMode
}

// MaintenanceMessage returns maintenance message.
func (o *Options) MaintenanceMessage() string {
	return o.maintenanceMessage
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

// ForceRefreshInterval returns the force refresh interval
func (o *Options) ForceRefreshInterval() int {
	return o.forceRefreshInterval
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

// SchedulerEntryFrequencyFactor returns the factor for the entry frequency scheduler.
func (o *Options) SchedulerEntryFrequencyFactor() int {
	return o.schedulerEntryFrequencyFactor
}

func (o *Options) SchedulerRoundRobinMinInterval() int {
	return o.schedulerRoundRobinMinInterval
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

// OIDCDiscoveryEndpoint returns the OAuth2 OIDC discovery endpoint.
func (o *Options) OIDCDiscoveryEndpoint() string {
	return o.oidcDiscoveryEndpoint
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

// YouTubeEmbedUrlOverride returns YouTube URL which will be used for embeds
func (o *Options) YouTubeEmbedUrlOverride() string {
	return o.youTubeEmbedUrlOverride
}

// FetchOdyseeWatchTime returns true if the Odysee video duration
// should be fetched and used as a reading time.
func (o *Options) FetchOdyseeWatchTime() bool {
	return o.fetchOdyseeWatchTime
}

// ProxyOption returns "none" to never proxy, "http-only" to proxy non-HTTPS, "all" to always proxy.
func (o *Options) ProxyOption() string {
	return o.proxyOption
}

// ProxyMediaTypes returns a slice of media types to proxy.
func (o *Options) ProxyMediaTypes() []string {
	return o.proxyMediaTypes
}

// ProxyUrl returns a string of a URL to use to proxy image requests
func (o *Options) ProxyUrl() string {
	return o.proxyUrl
}

// ProxyHTTPClientTimeout returns the time limit in seconds before the proxy HTTP client cancel the request.
func (o *Options) ProxyHTTPClientTimeout() int {
	return o.proxyHTTPClientTimeout
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

// HTTPServerTimeout returns the time limit in seconds before the HTTP server cancel the request.
func (o *Options) HTTPServerTimeout() int {
	return o.httpServerTimeout
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

func (o *Options) MetricsUsername() string {
	return o.metricsUsername
}

func (o *Options) MetricsPassword() string {
	return o.metricsPassword
}

// HTTPClientUserAgent returns the global User-Agent header for miniflux.
func (o *Options) HTTPClientUserAgent() string {
	return o.httpClientUserAgent
}

// HasWatchdog returns true if the systemd watchdog is enabled.
func (o *Options) HasWatchdog() bool {
	return o.watchdog
}

// InvidiousInstance returns the invidious instance used by miniflux
func (o *Options) InvidiousInstance() string {
	return o.invidiousInstance
}

// ProxyPrivateKey returns the private key used by the media proxy
func (o *Options) ProxyPrivateKey() []byte {
	return o.proxyPrivateKey
}

// WebAuthn returns true if WebAuthn logins are supported
func (o *Options) WebAuthn() bool {
	return o.webAuthn
}

// SortedOptions returns options as a list of key value pairs, sorted by keys.
func (o *Options) SortedOptions(redactSecret bool) []*Option {
	var keyValues = map[string]interface{}{
		"ADMIN_PASSWORD":                         redactSecretValue(o.adminPassword, redactSecret),
		"ADMIN_USERNAME":                         o.adminUsername,
		"AUTH_PROXY_HEADER":                      o.authProxyHeader,
		"AUTH_PROXY_USER_CREATION":               o.authProxyUserCreation,
		"BASE_PATH":                              o.basePath,
		"BASE_URL":                               o.baseURL,
		"BATCH_SIZE":                             o.batchSize,
		"CERT_DOMAIN":                            o.certDomain,
		"CERT_FILE":                              o.certFile,
		"CLEANUP_ARCHIVE_BATCH_SIZE":             o.cleanupArchiveBatchSize,
		"CLEANUP_ARCHIVE_READ_DAYS":              o.cleanupArchiveReadDays,
		"CLEANUP_ARCHIVE_UNREAD_DAYS":            o.cleanupArchiveUnreadDays,
		"CLEANUP_FREQUENCY_HOURS":                o.cleanupFrequencyHours,
		"CLEANUP_REMOVE_SESSIONS_DAYS":           o.cleanupRemoveSessionsDays,
		"CREATE_ADMIN":                           o.createAdmin,
		"DATABASE_CONNECTION_LIFETIME":           o.databaseConnectionLifetime,
		"DATABASE_MAX_CONNS":                     o.databaseMaxConns,
		"DATABASE_MIN_CONNS":                     o.databaseMinConns,
		"DATABASE_URL":                           redactSecretValue(o.databaseURL, redactSecret),
		"DISABLE_HSTS":                           !o.hsts,
		"DISABLE_HTTP_SERVICE":                   !o.httpService,
		"DISABLE_SCHEDULER_SERVICE":              !o.schedulerService,
		"FETCH_YOUTUBE_WATCH_TIME":               o.fetchYouTubeWatchTime,
		"FETCH_ODYSEE_WATCH_TIME":                o.fetchOdyseeWatchTime,
		"HTTPS":                                  o.HTTPS,
		"HTTP_CLIENT_MAX_BODY_SIZE":              o.httpClientMaxBodySize,
		"HTTP_CLIENT_PROXY":                      o.httpClientProxy,
		"HTTP_CLIENT_TIMEOUT":                    o.httpClientTimeout,
		"HTTP_CLIENT_USER_AGENT":                 o.httpClientUserAgent,
		"HTTP_SERVER_TIMEOUT":                    o.httpServerTimeout,
		"HTTP_SERVICE":                           o.httpService,
		"INVIDIOUS_INSTANCE":                     o.invidiousInstance,
		"KEY_FILE":                               o.certKeyFile,
		"LISTEN_ADDR":                            o.listenAddr,
		"LOG_FILE":                               o.logFile,
		"LOG_DATE_TIME":                          o.logDateTime,
		"LOG_FORMAT":                             o.logFormat,
		"LOG_LEVEL":                              o.logLevel,
		"MAINTENANCE_MESSAGE":                    o.maintenanceMessage,
		"MAINTENANCE_MODE":                       o.maintenanceMode,
		"METRICS_ALLOWED_NETWORKS":               strings.Join(o.metricsAllowedNetworks, ","),
		"METRICS_COLLECTOR":                      o.metricsCollector,
		"METRICS_PASSWORD":                       redactSecretValue(o.metricsPassword, redactSecret),
		"METRICS_REFRESH_INTERVAL":               o.metricsRefreshInterval,
		"METRICS_USERNAME":                       o.metricsUsername,
		"OAUTH2_CLIENT_ID":                       o.oauth2ClientID,
		"OAUTH2_CLIENT_SECRET":                   redactSecretValue(o.oauth2ClientSecret, redactSecret),
		"OAUTH2_OIDC_DISCOVERY_ENDPOINT":         o.oidcDiscoveryEndpoint,
		"OAUTH2_PROVIDER":                        o.oauth2Provider,
		"OAUTH2_REDIRECT_URL":                    o.oauth2RedirectURL,
		"OAUTH2_USER_CREATION":                   o.oauth2UserCreationAllowed,
		"POCKET_CONSUMER_KEY":                    redactSecretValue(o.pocketConsumerKey, redactSecret),
		"POLLING_FREQUENCY":                      o.pollingFrequency,
		"FORCE_REFRESH_INTERVAL":                 o.forceRefreshInterval,
		"POLLING_PARSING_ERROR_LIMIT":            o.pollingParsingErrorLimit,
		"POLLING_SCHEDULER":                      o.pollingScheduler,
		"PROXY_HTTP_CLIENT_TIMEOUT":              o.proxyHTTPClientTimeout,
		"PROXY_MEDIA_TYPES":                      o.proxyMediaTypes,
		"PROXY_OPTION":                           o.proxyOption,
		"PROXY_PRIVATE_KEY":                      redactSecretValue(string(o.proxyPrivateKey), redactSecret),
		"PROXY_URL":                              o.proxyUrl,
		"ROOT_URL":                               o.rootURL,
		"RUN_MIGRATIONS":                         o.runMigrations,
		"SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL": o.schedulerEntryFrequencyMaxInterval,
		"SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL": o.schedulerEntryFrequencyMinInterval,
		"SCHEDULER_ENTRY_FREQUENCY_FACTOR":       o.schedulerEntryFrequencyFactor,
		"SCHEDULER_ROUND_ROBIN_MIN_INTERVAL":     o.schedulerRoundRobinMinInterval,
		"SCHEDULER_SERVICE":                      o.schedulerService,
		"SERVER_TIMING_HEADER":                   o.serverTimingHeader,
		"WATCHDOG":                               o.watchdog,
		"WORKER_POOL_SIZE":                       o.workerPoolSize,
		"YOUTUBE_EMBED_URL_OVERRIDE":             o.youTubeEmbedUrlOverride,
		"WEBAUTHN":                               o.webAuthn,
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

	for _, option := range o.SortedOptions(false) {
		fmt.Fprintf(&builder, "%s=%v\n", option.Key, option.Value)
	}

	return builder.String()
}

func redactSecretValue(value string, redactSecret bool) string {
	if redactSecret && value != "" {
		return "<secret>"
	}
	return value
}
