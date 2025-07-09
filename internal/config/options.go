// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config // import "miniflux.app/v2/internal/config"

import (
	"fmt"
	"maps"
	"net/url"
	"slices"
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
	defaultBaseURL                            = "http://localhost"
	defaultRootURL                            = "http://localhost"
	defaultBasePath                           = ""
	defaultWorkerPoolSize                     = 16
	defaultPollingFrequency                   = 60
	defaultForceRefreshInterval               = 30
	defaultBatchSize                          = 100
	defaultPollingScheduler                   = "round_robin"
	defaultSchedulerEntryFrequencyMinInterval = 5
	defaultSchedulerEntryFrequencyMaxInterval = 24 * 60
	defaultSchedulerEntryFrequencyFactor      = 1
	defaultSchedulerRoundRobinMinInterval     = 60
	defaultSchedulerRoundRobinMaxInterval     = 1440
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
	defaultMediaProxyHTTPClientTimeout        = 120
	defaultMediaProxyMode                     = "http-only"
	defaultMediaResourceTypes                 = "image"
	defaultMediaProxyURL                      = ""
	defaultFilterEntryMaxAgeDays              = 0
	defaultFetchBilibiliWatchTime             = false
	defaultFetchNebulaWatchTime               = false
	defaultFetchOdyseeWatchTime               = false
	defaultFetchYouTubeWatchTime              = false
	defaultYouTubeApiKey                      = ""
	defaultYouTubeEmbedUrlOverride            = "https://www.youtube-nocookie.com/embed/"
	defaultCreateAdmin                        = false
	defaultAdminUsername                      = ""
	defaultAdminPassword                      = ""
	defaultOAuth2UserCreation                 = false
	defaultOAuth2ClientID                     = ""
	defaultOAuth2ClientSecret                 = ""
	defaultOAuth2RedirectURL                  = ""
	defaultOAuth2OidcDiscoveryEndpoint        = ""
	defaultOauth2OidcProviderName             = "OpenID Connect"
	defaultOAuth2Provider                     = ""
	defaultDisableLocalAuth                   = false
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

// option contains a key to value map of a single option. It may be used to output debug strings.
type option struct {
	Key   string
	Value any
}

// options contains configuration options.
type options struct {
	HTTPS                              bool
	logFile                            string
	logDateTime                        bool
	logFormat                          string
	logLevel                           string
	hsts                               bool
	httpService                        bool
	schedulerService                   bool
	baseURL                            string
	rootURL                            string
	basePath                           string
	databaseURL                        string
	databaseMaxConns                   int
	databaseMinConns                   int
	databaseConnectionLifetime         int
	runMigrations                      bool
	listenAddr                         []string
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
	schedulerRoundRobinMaxInterval     int
	pollingParsingErrorLimit           int
	workerPoolSize                     int
	createAdmin                        bool
	adminUsername                      string
	adminPassword                      string
	mediaProxyHTTPClientTimeout        int
	mediaProxyMode                     string
	mediaProxyResourceTypes            []string
	mediaProxyCustomURL                string
	fetchBilibiliWatchTime             bool
	fetchNebulaWatchTime               bool
	fetchOdyseeWatchTime               bool
	fetchYouTubeWatchTime              bool
	filterEntryMaxAgeDays              int
	youTubeApiKey                      string
	youTubeEmbedUrlOverride            string
	youTubeEmbedDomain                 string
	oauth2UserCreationAllowed          bool
	oauth2ClientID                     string
	oauth2ClientSecret                 string
	oauth2RedirectURL                  string
	oidcDiscoveryEndpoint              string
	oidcProviderName                   string
	oauth2Provider                     string
	disableLocalAuth                   bool
	httpClientTimeout                  int
	httpClientMaxBodySize              int64
	httpClientProxyURL                 *url.URL
	httpClientProxies                  []string
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
	mediaProxyPrivateKey               []byte
	webAuthn                           bool
}

// NewOptions returns Options with default values.
func NewOptions() *options {
	return &options{
		HTTPS:                              defaultHTTPS,
		logFile:                            defaultLogFile,
		logDateTime:                        defaultLogDateTime,
		logFormat:                          defaultLogFormat,
		logLevel:                           defaultLogLevel,
		hsts:                               defaultHSTS,
		httpService:                        defaultHTTPService,
		schedulerService:                   defaultSchedulerService,
		baseURL:                            defaultBaseURL,
		rootURL:                            defaultRootURL,
		basePath:                           defaultBasePath,
		databaseURL:                        defaultDatabaseURL,
		databaseMaxConns:                   defaultDatabaseMaxConns,
		databaseMinConns:                   defaultDatabaseMinConns,
		databaseConnectionLifetime:         defaultDatabaseConnectionLifetime,
		runMigrations:                      defaultRunMigrations,
		listenAddr:                         []string{defaultListenAddr},
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
		schedulerRoundRobinMaxInterval:     defaultSchedulerRoundRobinMaxInterval,
		pollingParsingErrorLimit:           defaultPollingParsingErrorLimit,
		workerPoolSize:                     defaultWorkerPoolSize,
		createAdmin:                        defaultCreateAdmin,
		mediaProxyHTTPClientTimeout:        defaultMediaProxyHTTPClientTimeout,
		mediaProxyMode:                     defaultMediaProxyMode,
		mediaProxyResourceTypes:            []string{defaultMediaResourceTypes},
		mediaProxyCustomURL:                defaultMediaProxyURL,
		filterEntryMaxAgeDays:              defaultFilterEntryMaxAgeDays,
		fetchBilibiliWatchTime:             defaultFetchBilibiliWatchTime,
		fetchNebulaWatchTime:               defaultFetchNebulaWatchTime,
		fetchOdyseeWatchTime:               defaultFetchOdyseeWatchTime,
		fetchYouTubeWatchTime:              defaultFetchYouTubeWatchTime,
		youTubeApiKey:                      defaultYouTubeApiKey,
		youTubeEmbedUrlOverride:            defaultYouTubeEmbedUrlOverride,
		oauth2UserCreationAllowed:          defaultOAuth2UserCreation,
		oauth2ClientID:                     defaultOAuth2ClientID,
		oauth2ClientSecret:                 defaultOAuth2ClientSecret,
		oauth2RedirectURL:                  defaultOAuth2RedirectURL,
		oidcDiscoveryEndpoint:              defaultOAuth2OidcDiscoveryEndpoint,
		oidcProviderName:                   defaultOauth2OidcProviderName,
		oauth2Provider:                     defaultOAuth2Provider,
		disableLocalAuth:                   defaultDisableLocalAuth,
		httpClientTimeout:                  defaultHTTPClientTimeout,
		httpClientMaxBodySize:              defaultHTTPClientMaxBodySize * 1024 * 1024,
		httpClientProxyURL:                 nil,
		httpClientProxies:                  []string{},
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
		mediaProxyPrivateKey:               crypto.GenerateRandomBytes(16),
		webAuthn:                           defaultWebAuthn,
	}
}

func (o *options) LogFile() string {
	return o.logFile
}

// LogDateTime returns true if the date/time should be displayed in log messages.
func (o *options) LogDateTime() bool {
	return o.logDateTime
}

// LogFormat returns the log format.
func (o *options) LogFormat() string {
	return o.logFormat
}

// LogLevel returns the log level.
func (o *options) LogLevel() string {
	return o.logLevel
}

// SetLogLevel sets the log level.
func (o *options) SetLogLevel(level string) {
	o.logLevel = level
}

// HasMaintenanceMode returns true if maintenance mode is enabled.
func (o *options) HasMaintenanceMode() bool {
	return o.maintenanceMode
}

// MaintenanceMessage returns maintenance message.
func (o *options) MaintenanceMessage() string {
	return o.maintenanceMessage
}

// BaseURL returns the application base URL with path.
func (o *options) BaseURL() string {
	return o.baseURL
}

// RootURL returns the base URL without path.
func (o *options) RootURL() string {
	return o.rootURL
}

// BasePath returns the application base path according to the base URL.
func (o *options) BasePath() string {
	return o.basePath
}

// IsDefaultDatabaseURL returns true if the default database URL is used.
func (o *options) IsDefaultDatabaseURL() bool {
	return o.databaseURL == defaultDatabaseURL
}

// DatabaseURL returns the database URL.
func (o *options) DatabaseURL() string {
	return o.databaseURL
}

// DatabaseMaxConns returns the maximum number of database connections.
func (o *options) DatabaseMaxConns() int {
	return o.databaseMaxConns
}

// DatabaseMinConns returns the minimum number of database connections.
func (o *options) DatabaseMinConns() int {
	return o.databaseMinConns
}

// DatabaseConnectionLifetime returns the maximum amount of time a connection may be reused.
func (o *options) DatabaseConnectionLifetime() time.Duration {
	return time.Duration(o.databaseConnectionLifetime) * time.Minute
}

// ListenAddr returns the listen address for the HTTP server.
func (o *options) ListenAddr() []string {
	return o.listenAddr
}

// CertFile returns the SSL certificate filename if any.
func (o *options) CertFile() string {
	return o.certFile
}

// CertKeyFile returns the private key filename for custom SSL certificate.
func (o *options) CertKeyFile() string {
	return o.certKeyFile
}

// CertDomain returns the domain to use for Let's Encrypt certificate.
func (o *options) CertDomain() string {
	return o.certDomain
}

// CleanupFrequencyHours returns the interval in hours for cleanup jobs.
func (o *options) CleanupFrequencyHours() int {
	return o.cleanupFrequencyHours
}

// CleanupArchiveReadDays returns the number of days after which marking read items as removed.
func (o *options) CleanupArchiveReadDays() int {
	return o.cleanupArchiveReadDays
}

// CleanupArchiveUnreadDays returns the number of days after which marking unread items as removed.
func (o *options) CleanupArchiveUnreadDays() int {
	return o.cleanupArchiveUnreadDays
}

// CleanupArchiveBatchSize returns the number of entries to archive for each interval.
func (o *options) CleanupArchiveBatchSize() int {
	return o.cleanupArchiveBatchSize
}

// CleanupRemoveSessionsDays returns the number of days after which to remove sessions.
func (o *options) CleanupRemoveSessionsDays() int {
	return o.cleanupRemoveSessionsDays
}

// WorkerPoolSize returns the number of background worker.
func (o *options) WorkerPoolSize() int {
	return o.workerPoolSize
}

// PollingFrequency returns the interval to refresh feeds in the background.
func (o *options) PollingFrequency() int {
	return o.pollingFrequency
}

// ForceRefreshInterval returns the force refresh interval
func (o *options) ForceRefreshInterval() int {
	return o.forceRefreshInterval
}

// BatchSize returns the number of feeds to send for background processing.
func (o *options) BatchSize() int {
	return o.batchSize
}

// PollingScheduler returns the scheduler used for polling feeds.
func (o *options) PollingScheduler() string {
	return o.pollingScheduler
}

// SchedulerEntryFrequencyMaxInterval returns the maximum interval in minutes for the entry frequency scheduler.
func (o *options) SchedulerEntryFrequencyMaxInterval() int {
	return o.schedulerEntryFrequencyMaxInterval
}

// SchedulerEntryFrequencyMinInterval returns the minimum interval in minutes for the entry frequency scheduler.
func (o *options) SchedulerEntryFrequencyMinInterval() int {
	return o.schedulerEntryFrequencyMinInterval
}

// SchedulerEntryFrequencyFactor returns the factor for the entry frequency scheduler.
func (o *options) SchedulerEntryFrequencyFactor() int {
	return o.schedulerEntryFrequencyFactor
}

func (o *options) SchedulerRoundRobinMinInterval() int {
	return o.schedulerRoundRobinMinInterval
}

func (o *options) SchedulerRoundRobinMaxInterval() int {
	return o.schedulerRoundRobinMaxInterval
}

// PollingParsingErrorLimit returns the limit of errors when to stop polling.
func (o *options) PollingParsingErrorLimit() int {
	return o.pollingParsingErrorLimit
}

// IsOAuth2UserCreationAllowed returns true if user creation is allowed for OAuth2 users.
func (o *options) IsOAuth2UserCreationAllowed() bool {
	return o.oauth2UserCreationAllowed
}

// OAuth2ClientID returns the OAuth2 Client ID.
func (o *options) OAuth2ClientID() string {
	return o.oauth2ClientID
}

// OAuth2ClientSecret returns the OAuth2 client secret.
func (o *options) OAuth2ClientSecret() string {
	return o.oauth2ClientSecret
}

// OAuth2RedirectURL returns the OAuth2 redirect URL.
func (o *options) OAuth2RedirectURL() string {
	return o.oauth2RedirectURL
}

// OIDCDiscoveryEndpoint returns the OAuth2 OIDC discovery endpoint.
func (o *options) OIDCDiscoveryEndpoint() string {
	return o.oidcDiscoveryEndpoint
}

// OIDCProviderName returns the OAuth2 OIDC provider's display name
func (o *options) OIDCProviderName() string {
	return o.oidcProviderName
}

// OAuth2Provider returns the name of the OAuth2 provider configured.
func (o *options) OAuth2Provider() string {
	return o.oauth2Provider
}

// DisableLocalAUth returns true if the local user database should not be used to authenticate users
func (o *options) DisableLocalAuth() bool {
	return o.disableLocalAuth
}

// HasHSTS returns true if HTTP Strict Transport Security is enabled.
func (o *options) HasHSTS() bool {
	return o.hsts
}

// RunMigrations returns true if the environment variable RUN_MIGRATIONS is not empty.
func (o *options) RunMigrations() bool {
	return o.runMigrations
}

// CreateAdmin returns true if the environment variable CREATE_ADMIN is not empty.
func (o *options) CreateAdmin() bool {
	return o.createAdmin
}

// AdminUsername returns the admin username if defined.
func (o *options) AdminUsername() string {
	return o.adminUsername
}

// AdminPassword returns the admin password if defined.
func (o *options) AdminPassword() string {
	return o.adminPassword
}

// FetchYouTubeWatchTime returns true if the YouTube video duration
// should be fetched and used as a reading time.
func (o *options) FetchYouTubeWatchTime() bool {
	return o.fetchYouTubeWatchTime
}

// YouTubeApiKey returns the YouTube API key if defined.
func (o *options) YouTubeApiKey() string {
	return o.youTubeApiKey
}

// YouTubeEmbedUrlOverride returns the YouTube embed URL override if defined.
func (o *options) YouTubeEmbedUrlOverride() string {
	return o.youTubeEmbedUrlOverride
}

// YouTubeEmbedDomain returns the domain used for YouTube embeds.
func (o *options) YouTubeEmbedDomain() string {
	if o.youTubeEmbedDomain != "" {
		return o.youTubeEmbedDomain
	}
	return "www.youtube-nocookie.com"
}

// FetchNebulaWatchTime returns true if the Nebula video duration
// should be fetched and used as a reading time.
func (o *options) FetchNebulaWatchTime() bool {
	return o.fetchNebulaWatchTime
}

// FetchOdyseeWatchTime returns true if the Odysee video duration
// should be fetched and used as a reading time.
func (o *options) FetchOdyseeWatchTime() bool {
	return o.fetchOdyseeWatchTime
}

// FetchBilibiliWatchTime returns true if the Bilibili video duration
// should be fetched and used as a reading time.
func (o *options) FetchBilibiliWatchTime() bool {
	return o.fetchBilibiliWatchTime
}

// MediaProxyMode returns "none" to never proxy, "http-only" to proxy non-HTTPS, "all" to always proxy.
func (o *options) MediaProxyMode() string {
	return o.mediaProxyMode
}

// MediaProxyResourceTypes returns a slice of resource types to proxy.
func (o *options) MediaProxyResourceTypes() []string {
	return o.mediaProxyResourceTypes
}

// MediaCustomProxyURL returns the custom proxy URL for medias.
func (o *options) MediaCustomProxyURL() string {
	return o.mediaProxyCustomURL
}

// MediaProxyHTTPClientTimeout returns the time limit in seconds before the proxy HTTP client cancel the request.
func (o *options) MediaProxyHTTPClientTimeout() int {
	return o.mediaProxyHTTPClientTimeout
}

// MediaProxyPrivateKey returns the private key used by the media proxy.
func (o *options) MediaProxyPrivateKey() []byte {
	return o.mediaProxyPrivateKey
}

// HasHTTPService returns true if the HTTP service is enabled.
func (o *options) HasHTTPService() bool {
	return o.httpService
}

// HasSchedulerService returns true if the scheduler service is enabled.
func (o *options) HasSchedulerService() bool {
	return o.schedulerService
}

// HTTPClientTimeout returns the time limit in seconds before the HTTP client cancel the request.
func (o *options) HTTPClientTimeout() int {
	return o.httpClientTimeout
}

// HTTPClientMaxBodySize returns the number of bytes allowed for the HTTP client to transfer.
func (o *options) HTTPClientMaxBodySize() int64 {
	return o.httpClientMaxBodySize
}

// HTTPClientProxyURL returns the client HTTP proxy URL if configured.
func (o *options) HTTPClientProxyURL() *url.URL {
	return o.httpClientProxyURL
}

// HasHTTPClientProxyURLConfigured returns true if the client HTTP proxy URL if configured.
func (o *options) HasHTTPClientProxyURLConfigured() bool {
	return o.httpClientProxyURL != nil
}

// HTTPClientProxies returns the list of proxies.
func (o *options) HTTPClientProxies() []string {
	return o.httpClientProxies
}

// HTTPClientProxiesString returns true if the list of rotating proxies are configured.
func (o *options) HasHTTPClientProxiesConfigured() bool {
	return len(o.httpClientProxies) > 0
}

// HTTPServerTimeout returns the time limit in seconds before the HTTP server cancel the request.
func (o *options) HTTPServerTimeout() int {
	return o.httpServerTimeout
}

// AuthProxyHeader returns an HTTP header name that contains username for
// authentication using auth proxy.
func (o *options) AuthProxyHeader() string {
	return o.authProxyHeader
}

// IsAuthProxyUserCreationAllowed returns true if user creation is allowed for
// users authenticated using auth proxy.
func (o *options) IsAuthProxyUserCreationAllowed() bool {
	return o.authProxyUserCreation
}

// HasMetricsCollector returns true if metrics collection is enabled.
func (o *options) HasMetricsCollector() bool {
	return o.metricsCollector
}

// MetricsRefreshInterval returns the refresh interval in seconds.
func (o *options) MetricsRefreshInterval() int {
	return o.metricsRefreshInterval
}

// MetricsAllowedNetworks returns the list of networks allowed to connect to the metrics endpoint.
func (o *options) MetricsAllowedNetworks() []string {
	return o.metricsAllowedNetworks
}

func (o *options) MetricsUsername() string {
	return o.metricsUsername
}

func (o *options) MetricsPassword() string {
	return o.metricsPassword
}

// HTTPClientUserAgent returns the global User-Agent header for miniflux.
func (o *options) HTTPClientUserAgent() string {
	return o.httpClientUserAgent
}

// HasWatchdog returns true if the systemd watchdog is enabled.
func (o *options) HasWatchdog() bool {
	return o.watchdog
}

// InvidiousInstance returns the invidious instance used by miniflux
func (o *options) InvidiousInstance() string {
	return o.invidiousInstance
}

// WebAuthn returns true if WebAuthn logins are supported
func (o *options) WebAuthn() bool {
	return o.webAuthn
}

// FilterEntryMaxAgeDays returns the number of days after which entries should be retained.
func (o *options) FilterEntryMaxAgeDays() int {
	return o.filterEntryMaxAgeDays
}

// SortedOptions returns options as a list of key value pairs, sorted by keys.
func (o *options) SortedOptions(redactSecret bool) []*option {
	var clientProxyURLRedacted string
	if o.httpClientProxyURL != nil {
		if redactSecret {
			clientProxyURLRedacted = o.httpClientProxyURL.Redacted()
		} else {
			clientProxyURLRedacted = o.httpClientProxyURL.String()
		}
	}

	var clientProxyURLsRedacted string
	if len(o.httpClientProxies) > 0 {
		if redactSecret {
			var proxyURLs []string
			for range o.httpClientProxies {
				proxyURLs = append(proxyURLs, "<redacted>")
			}
			clientProxyURLsRedacted = strings.Join(proxyURLs, ",")
		} else {
			clientProxyURLsRedacted = strings.Join(o.httpClientProxies, ",")
		}
	}

	var mediaProxyPrivateKeyValue string
	if len(o.mediaProxyPrivateKey) > 0 {
		mediaProxyPrivateKeyValue = "<binary-data>"
	}

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
		"FILTER_ENTRY_MAX_AGE_DAYS":              o.filterEntryMaxAgeDays,
		"FETCH_YOUTUBE_WATCH_TIME":               o.fetchYouTubeWatchTime,
		"FETCH_NEBULA_WATCH_TIME":                o.fetchNebulaWatchTime,
		"FETCH_ODYSEE_WATCH_TIME":                o.fetchOdyseeWatchTime,
		"FETCH_BILIBILI_WATCH_TIME":              o.fetchBilibiliWatchTime,
		"HTTPS":                                  o.HTTPS,
		"HTTP_CLIENT_MAX_BODY_SIZE":              o.httpClientMaxBodySize,
		"HTTP_CLIENT_PROXIES":                    clientProxyURLsRedacted,
		"HTTP_CLIENT_PROXY":                      clientProxyURLRedacted,
		"HTTP_CLIENT_TIMEOUT":                    o.httpClientTimeout,
		"HTTP_CLIENT_USER_AGENT":                 o.httpClientUserAgent,
		"HTTP_SERVER_TIMEOUT":                    o.httpServerTimeout,
		"HTTP_SERVICE":                           o.httpService,
		"INVIDIOUS_INSTANCE":                     o.invidiousInstance,
		"KEY_FILE":                               o.certKeyFile,
		"LISTEN_ADDR":                            strings.Join(o.listenAddr, ","),
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
		"OAUTH2_OIDC_PROVIDER_NAME":              o.oidcProviderName,
		"OAUTH2_PROVIDER":                        o.oauth2Provider,
		"OAUTH2_REDIRECT_URL":                    o.oauth2RedirectURL,
		"OAUTH2_USER_CREATION":                   o.oauth2UserCreationAllowed,
		"DISABLE_LOCAL_AUTH":                     o.disableLocalAuth,
		"POLLING_FREQUENCY":                      o.pollingFrequency,
		"FORCE_REFRESH_INTERVAL":                 o.forceRefreshInterval,
		"POLLING_PARSING_ERROR_LIMIT":            o.pollingParsingErrorLimit,
		"POLLING_SCHEDULER":                      o.pollingScheduler,
		"MEDIA_PROXY_HTTP_CLIENT_TIMEOUT":        o.mediaProxyHTTPClientTimeout,
		"MEDIA_PROXY_RESOURCE_TYPES":             o.mediaProxyResourceTypes,
		"MEDIA_PROXY_MODE":                       o.mediaProxyMode,
		"MEDIA_PROXY_PRIVATE_KEY":                mediaProxyPrivateKeyValue,
		"MEDIA_PROXY_CUSTOM_URL":                 o.mediaProxyCustomURL,
		"ROOT_URL":                               o.rootURL,
		"RUN_MIGRATIONS":                         o.runMigrations,
		"SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL": o.schedulerEntryFrequencyMaxInterval,
		"SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL": o.schedulerEntryFrequencyMinInterval,
		"SCHEDULER_ENTRY_FREQUENCY_FACTOR":       o.schedulerEntryFrequencyFactor,
		"SCHEDULER_ROUND_ROBIN_MIN_INTERVAL":     o.schedulerRoundRobinMinInterval,
		"SCHEDULER_ROUND_ROBIN_MAX_INTERVAL":     o.schedulerRoundRobinMaxInterval,
		"SCHEDULER_SERVICE":                      o.schedulerService,
		"WATCHDOG":                               o.watchdog,
		"WORKER_POOL_SIZE":                       o.workerPoolSize,
		"YOUTUBE_API_KEY":                        redactSecretValue(o.youTubeApiKey, redactSecret),
		"YOUTUBE_EMBED_URL_OVERRIDE":             o.youTubeEmbedUrlOverride,
		"WEBAUTHN":                               o.webAuthn,
	}

	sortedKeys := slices.Sorted(maps.Keys(keyValues))
	var sortedOptions = make([]*option, 0, len(sortedKeys))
	for _, key := range sortedKeys {
		sortedOptions = append(sortedOptions, &option{Key: key, Value: keyValues[key]})
	}
	return sortedOptions
}

func (o *options) String() string {
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
