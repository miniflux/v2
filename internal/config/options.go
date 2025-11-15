// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config // import "miniflux.app/v2/internal/config"

import (
	"fmt"
	"maps"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"time"
)

type optionPair struct {
	Key   string
	Value string
}

type configValueType int

const (
	stringListType configValueType = iota
	boolType
	intType
	int64Type
	urlType
	secondType
	minuteType
	hourType
	dayType
	bytesType
)

type configValue struct {
	value any
	str   string

	ParsedStringValue string
	ParsedIntValue    int
	ParsedInt64Value  int64
	ParsedDuration    time.Duration
	ParsedStringList  []string
	ParsedURLValue    *url.URL
	ParsedBytesValue  []byte

	RawValue  string
	ValueType configValueType
	Secret    bool
	TargetKey string

	Validator func(string) error
}

type configOptions struct {
	rootURL            string
	basePath           string
	youTubeEmbedDomain string
	options            map[string]*configValue
}

func anyToStr(a any) string {
	if a == nil {
		return ""
	}
	if ret, ok := a.(string); ok {
		return ret
	}
	panic(fmt.Sprintf("expected string, got %q", reflect.TypeOf(a)))
}
func anyToBool(a any) bool {
	if a == nil {
		return false
	}
	if ret, ok := a.(bool); ok {
		return ret
	}
	panic(fmt.Sprintf("expected bool, got %q", reflect.TypeOf(a)))
}

// NewConfigOptions creates a new instance of ConfigOptions with default values.
func NewConfigOptions() *configOptions {
	return &configOptions{
		rootURL:            "http://localhost",
		basePath:           "",
		youTubeEmbedDomain: "www.youtube-nocookie.com",
		options: map[string]*configValue{
			"ADMIN_PASSWORD": {
				value:  "",
				Secret: true,
			},
			"ADMIN_PASSWORD_FILE": {
				value:     "",
				TargetKey: "ADMIN_PASSWORD",
			},
			"ADMIN_USERNAME": {
				value: "",
			},
			"ADMIN_USERNAME_FILE": {
				value:     "",
				TargetKey: "ADMIN_USERNAME",
			},
			"AUTH_PROXY_HEADER": {
				value: "",
			},
			"AUTH_PROXY_USER_CREATION": {
				value: false,
			},
			"BASE_URL": {
				value: "http://localhost",
			},
			"BATCH_SIZE": {
				ParsedIntValue: 100,
				RawValue:       "100",
				ValueType:      intType,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"CERT_DOMAIN": {
				value: "",
			},
			"CERT_FILE": {
				value: "",
			},
			"CLEANUP_ARCHIVE_BATCH_SIZE": {
				ParsedIntValue: 10000,
				RawValue:       "10000",
				ValueType:      intType,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"CLEANUP_ARCHIVE_READ_DAYS": {
				ParsedDuration: time.Hour * 24 * 60,
				RawValue:       "60",
				ValueType:      dayType,
			},
			"CLEANUP_ARCHIVE_UNREAD_DAYS": {
				ParsedDuration: time.Hour * 24 * 180,
				RawValue:       "180",
				ValueType:      dayType,
			},
			"CLEANUP_FREQUENCY_HOURS": {
				ParsedDuration: time.Hour * 24,
				RawValue:       "24",
				ValueType:      hourType,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"CLEANUP_REMOVE_SESSIONS_DAYS": {
				ParsedDuration: time.Hour * 24 * 30,
				RawValue:       "30",
				ValueType:      dayType,
			},
			"CREATE_ADMIN": {
				value: false,
			},
			"DATABASE_CONNECTION_LIFETIME": {
				ParsedDuration: time.Minute * 5,
				RawValue:       "5",
				ValueType:      minuteType,
				Validator: func(rawValue string) error {
					return validateGreaterThan(rawValue, 0)
				},
			},
			"DATABASE_MAX_CONNS": {
				ParsedIntValue: 20,
				RawValue:       "20",
				ValueType:      intType,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"DATABASE_MIN_CONNS": {
				ParsedIntValue: 1,
				RawValue:       "1",
				ValueType:      intType,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 0)
				},
			},
			"DATABASE_URL": {
				value:  "user=postgres password=postgres dbname=miniflux2 sslmode=disable",
				Secret: true,
			},
			"DATABASE_URL_FILE": {
				value:     "",
				TargetKey: "DATABASE_URL",
			},
			"DISABLE_HSTS": {
				value: false,
			},
			"DISABLE_HTTP_SERVICE": {
				value: false,
			},
			"DISABLE_LOCAL_AUTH": {
				value: false,
			},
			"DISABLE_SCHEDULER_SERVICE": {
				value: false,
			},
			"FETCH_BILIBILI_WATCH_TIME": {
				value: false,
			},
			"FETCH_NEBULA_WATCH_TIME": {
				value: false,
			},
			"FETCH_ODYSEE_WATCH_TIME": {
				value: false,
			},
			"FETCH_YOUTUBE_WATCH_TIME": {
				value: false,
			},
			"FILTER_ENTRY_MAX_AGE_DAYS": {
				ParsedIntValue: 0,
				RawValue:       "0",
				ValueType:      intType,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 0)
				},
			},
			"FORCE_REFRESH_INTERVAL": {
				ParsedDuration: 30 * time.Minute,
				RawValue:       "30",
				ValueType:      minuteType,
				Validator: func(rawValue string) error {
					return validateGreaterThan(rawValue, 0)
				},
			},
			"HTTP_CLIENT_MAX_BODY_SIZE": {
				ParsedInt64Value: 15,
				RawValue:         "15",
				ValueType:        int64Type,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"HTTP_CLIENT_PROXIES": {
				ParsedStringList: []string{},
				RawValue:         "",
				ValueType:        stringListType,
				Secret:           true,
			},
			"HTTP_CLIENT_PROXY": {
				ParsedURLValue: nil,
				RawValue:       "",
				ValueType:      urlType,
				Secret:         true,
			},
			"HTTP_CLIENT_TIMEOUT": {
				ParsedDuration: 20 * time.Second,
				RawValue:       "20",
				ValueType:      secondType,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"HTTP_CLIENT_USER_AGENT": {
				value: "",
			},
			"HTTP_SERVER_TIMEOUT": {
				ParsedDuration: 300 * time.Second,
				RawValue:       "300",
				ValueType:      secondType,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"HTTPS": {
				value: false,
			},
			"INVIDIOUS_INSTANCE": {
				value: "yewtu.be",
			},
			"KEY_FILE": {
				value: "",
			},
			"LISTEN_ADDR": {
				ParsedStringList: []string{"127.0.0.1:8080"},
				RawValue:         "127.0.0.1:8080",
				ValueType:        stringListType,
			},
			"LOG_DATE_TIME": {
				value: false,
			},
			"LOG_FILE": {
				value: "stderr",
			},
			"LOG_FORMAT": {
				value: "text",
				Validator: func(rawValue string) error {
					return validateChoices(rawValue, []string{"text", "json"})
				},
			},
			"LOG_LEVEL": {
				value: "info",
				Validator: func(rawValue string) error {
					return validateChoices(rawValue, []string{"debug", "info", "warning", "error"})
				},
			},
			"MAINTENANCE_MESSAGE": {
				value: "Miniflux is currently under maintenance",
			},
			"MAINTENANCE_MODE": {
				value: false,
			},
			"MEDIA_PROXY_CUSTOM_URL": {
				RawValue:  "",
				ValueType: urlType,
			},
			"MEDIA_PROXY_HTTP_CLIENT_TIMEOUT": {
				ParsedDuration: 120 * time.Second,
				RawValue:       "120",
				ValueType:      secondType,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"MEDIA_PROXY_MODE": {
				value: "http-only",
				Validator: func(rawValue string) error {
					return validateChoices(rawValue, []string{"none", "http-only", "all"})
				},
			},
			"MEDIA_PROXY_PRIVATE_KEY": {
				ValueType: bytesType,
				Secret:    true,
			},
			"MEDIA_PROXY_RESOURCE_TYPES": {
				ParsedStringList: []string{"image"},
				RawValue:         "image",
				ValueType:        stringListType,
				Validator: func(rawValue string) error {
					return validateListChoices(strings.Split(rawValue, ","), []string{"image", "video", "audio"})
				},
			},
			"METRICS_ALLOWED_NETWORKS": {
				ParsedStringList: []string{"127.0.0.1/8"},
				RawValue:         "127.0.0.1/8",
				ValueType:        stringListType,
			},
			"METRICS_COLLECTOR": {
				value: false,
			},
			"METRICS_PASSWORD": {
				value:  "",
				Secret: true,
			},
			"METRICS_PASSWORD_FILE": {
				value:     "",
				TargetKey: "METRICS_PASSWORD",
			},
			"METRICS_REFRESH_INTERVAL": {
				ParsedDuration: 60 * time.Second,
				RawValue:       "60",
				ValueType:      secondType,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"METRICS_USERNAME": {
				value: "",
			},
			"METRICS_USERNAME_FILE": {
				value:     "",
				TargetKey: "METRICS_USERNAME",
			},
			"OAUTH2_CLIENT_ID": {
				value:  "",
				Secret: true,
			},
			"OAUTH2_CLIENT_ID_FILE": {
				value:     "",
				TargetKey: "OAUTH2_CLIENT_ID",
			},
			"OAUTH2_CLIENT_SECRET": {
				value:  "",
				Secret: true,
			},
			"OAUTH2_CLIENT_SECRET_FILE": {
				value:     "",
				TargetKey: "OAUTH2_CLIENT_SECRET",
			},
			"OAUTH2_OIDC_DISCOVERY_ENDPOINT": {
				value: "",
			},
			"OAUTH2_OIDC_PROVIDER_NAME": {
				value: "OpenID Connect",
			},
			"OAUTH2_PROVIDER": {
				value: "",
				Validator: func(rawValue string) error {
					return validateChoices(rawValue, []string{"oidc", "google"})
				},
			},
			"OAUTH2_REDIRECT_URL": {
				value: "",
			},
			"OAUTH2_USER_CREATION": {
				value: false,
			},
			"POLLING_FREQUENCY": {
				ParsedDuration: 60 * time.Minute,
				RawValue:       "60",
				ValueType:      minuteType,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"POLLING_LIMIT_PER_HOST": {
				ParsedIntValue: 0,
				RawValue:       "0",
				ValueType:      intType,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 0)
				},
			},
			"POLLING_PARSING_ERROR_LIMIT": {
				ParsedIntValue: 3,
				RawValue:       "3",
				ValueType:      intType,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 0)
				},
			},
			"POLLING_SCHEDULER": {
				value: "round_robin",
				Validator: func(rawValue string) error {
					return validateChoices(rawValue, []string{"round_robin", "entry_frequency"})
				},
			},
			"PORT": {
				value: "",
				Validator: func(rawValue string) error {
					return validateRange(rawValue, 1, 65535)
				},
			},
			"RUN_MIGRATIONS": {
				value: false,
			},
			"SCHEDULER_ENTRY_FREQUENCY_FACTOR": {
				ParsedIntValue: 1,
				RawValue:       "1",
				ValueType:      intType,
			},
			"SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL": {
				ParsedDuration: 24 * time.Hour,
				RawValue:       "1440",
				ValueType:      minuteType,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL": {
				ParsedDuration: 5 * time.Minute,
				RawValue:       "5",
				ValueType:      minuteType,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"SCHEDULER_ROUND_ROBIN_MAX_INTERVAL": {
				ParsedDuration: 1440 * time.Minute,
				RawValue:       "1440",
				ValueType:      minuteType,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"SCHEDULER_ROUND_ROBIN_MIN_INTERVAL": {
				ParsedDuration: 60 * time.Minute,
				RawValue:       "60",
				ValueType:      minuteType,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"WATCHDOG": {
				value: true,
			},
			"WEBAUTHN": {
				value: false,
			},
			"WORKER_POOL_SIZE": {
				ParsedIntValue: 16,
				RawValue:       "16",
				ValueType:      intType,
				Validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"YOUTUBE_API_KEY": {
				value:  "",
				Secret: true,
			},
			"YOUTUBE_EMBED_URL_OVERRIDE": {
				value: "https://www.youtube-nocookie.com/embed/",
			},
		},
	}
}

func (c *configOptions) AdminPassword() string {
	return anyToStr(c.options["ADMIN_PASSWORD"].value)
}

func (c *configOptions) AdminUsername() string {
	return anyToStr(c.options["ADMIN_USERNAME"].value)
}

func (c *configOptions) AuthProxyHeader() string {
	return anyToStr(c.options["AUTH_PROXY_HEADER"].value)
}

func (c *configOptions) AuthProxyUserCreation() bool {
	return anyToBool(c.options["AUTH_PROXY_USER_CREATION"].value)
}

func (c *configOptions) BasePath() string {
	return c.basePath
}

func (c *configOptions) BaseURL() string {
	return anyToStr(c.options["BASE_URL"].value)
}

func (c *configOptions) RootURL() string {
	return c.rootURL
}

func (c *configOptions) BatchSize() int {
	return c.options["BATCH_SIZE"].ParsedIntValue
}

func (c *configOptions) CertDomain() string {
	return anyToStr(c.options["CERT_DOMAIN"].value)
}

func (c *configOptions) CertFile() string {
	return anyToStr(c.options["CERT_FILE"].value)
}

func (c *configOptions) CleanupArchiveBatchSize() int {
	return c.options["CLEANUP_ARCHIVE_BATCH_SIZE"].ParsedIntValue
}

func (c *configOptions) CleanupArchiveReadInterval() time.Duration {
	return c.options["CLEANUP_ARCHIVE_READ_DAYS"].ParsedDuration
}

func (c *configOptions) CleanupArchiveUnreadInterval() time.Duration {
	return c.options["CLEANUP_ARCHIVE_UNREAD_DAYS"].ParsedDuration
}

func (c *configOptions) CleanupFrequency() time.Duration {
	return c.options["CLEANUP_FREQUENCY_HOURS"].ParsedDuration
}

func (c *configOptions) CleanupRemoveSessionsInterval() time.Duration {
	return c.options["CLEANUP_REMOVE_SESSIONS_DAYS"].ParsedDuration
}

func (c *configOptions) CreateAdmin() bool {
	return anyToBool(c.options["CREATE_ADMIN"].value)
}

func (c *configOptions) DatabaseConnectionLifetime() time.Duration {
	return c.options["DATABASE_CONNECTION_LIFETIME"].ParsedDuration
}

func (c *configOptions) DatabaseMaxConns() int {
	return c.options["DATABASE_MAX_CONNS"].ParsedIntValue
}

func (c *configOptions) DatabaseMinConns() int {
	return c.options["DATABASE_MIN_CONNS"].ParsedIntValue
}

func (c *configOptions) DatabaseURL() string {
	return anyToStr(c.options["DATABASE_URL"].value)
}

func (c *configOptions) DisableHSTS() bool {
	return anyToBool(c.options["DISABLE_HSTS"].value)
}

func (c *configOptions) DisableHTTPService() bool {
	return anyToBool(c.options["DISABLE_HTTP_SERVICE"].value)
}

func (c *configOptions) DisableLocalAuth() bool {
	return anyToBool(c.options["DISABLE_LOCAL_AUTH"].value)
}

func (c *configOptions) DisableSchedulerService() bool {
	return anyToBool(c.options["DISABLE_SCHEDULER_SERVICE"].value)
}

func (c *configOptions) FetchBilibiliWatchTime() bool {
	return anyToBool(c.options["FETCH_BILIBILI_WATCH_TIME"].value)
}

func (c *configOptions) FetchNebulaWatchTime() bool {
	return anyToBool(c.options["FETCH_NEBULA_WATCH_TIME"].value)
}

func (c *configOptions) FetchOdyseeWatchTime() bool {
	return anyToBool(c.options["FETCH_ODYSEE_WATCH_TIME"].value)
}

func (c *configOptions) FetchYouTubeWatchTime() bool {
	return anyToBool(c.options["FETCH_YOUTUBE_WATCH_TIME"].value)
}

func (c *configOptions) FilterEntryMaxAgeDays() int {
	return c.options["FILTER_ENTRY_MAX_AGE_DAYS"].ParsedIntValue
}

func (c *configOptions) ForceRefreshInterval() time.Duration {
	return c.options["FORCE_REFRESH_INTERVAL"].ParsedDuration
}

func (c *configOptions) HasHTTPClientProxiesConfigured() bool {
	return len(c.options["HTTP_CLIENT_PROXIES"].ParsedStringList) > 0
}

func (c *configOptions) HasHTTPService() bool {
	return !anyToBool(c.options["DISABLE_HTTP_SERVICE"].value)
}

func (c *configOptions) HasHSTS() bool {
	return !anyToBool(c.options["DISABLE_HSTS"].value)
}

func (c *configOptions) HasHTTPClientProxyURLConfigured() bool {
	return c.options["HTTP_CLIENT_PROXY"].ParsedURLValue != nil
}

func (c *configOptions) HasMaintenanceMode() bool {
	return anyToBool(c.options["MAINTENANCE_MODE"].value)
}

func (c *configOptions) HasMetricsCollector() bool {
	return anyToBool(c.options["METRICS_COLLECTOR"].value)
}

func (c *configOptions) HasSchedulerService() bool {
	return !anyToBool(c.options["DISABLE_SCHEDULER_SERVICE"].value)
}

func (c *configOptions) HasWatchdog() bool {
	return anyToBool(c.options["WATCHDOG"].value)
}

func (c *configOptions) HTTPClientMaxBodySize() int64 {
	return c.options["HTTP_CLIENT_MAX_BODY_SIZE"].ParsedInt64Value * 1024 * 1024
}

func (c *configOptions) HTTPClientProxies() []string {
	return c.options["HTTP_CLIENT_PROXIES"].ParsedStringList
}

func (c *configOptions) HTTPClientProxyURL() *url.URL {
	return c.options["HTTP_CLIENT_PROXY"].ParsedURLValue
}

func (c *configOptions) HTTPClientTimeout() time.Duration {
	return c.options["HTTP_CLIENT_TIMEOUT"].ParsedDuration
}

func (c *configOptions) HTTPClientUserAgent() string {
	s := anyToStr(c.options["HTTP_CLIENT_USER_AGENT"].value)
	if s != "" {
		return s
	}
	return defaultHTTPClientUserAgent
}

func (c *configOptions) HTTPServerTimeout() time.Duration {
	return c.options["HTTP_SERVER_TIMEOUT"].ParsedDuration
}

func (c *configOptions) HTTPS() bool {
	return anyToBool(c.options["HTTPS"].value)
}

func (c *configOptions) InvidiousInstance() string {
	return anyToStr(c.options["INVIDIOUS_INSTANCE"].value)
}

func (c *configOptions) IsAuthProxyUserCreationAllowed() bool {
	return anyToBool(c.options["AUTH_PROXY_USER_CREATION"].value)
}

func (c *configOptions) IsDefaultDatabaseURL() bool {
	if c.options["DATABASE_URL"].str != "" {
		return c.options["DATABASE_URL"].str == "user=postgres password=postgres dbname=miniflux2 sslmode=disable"
	}
	return anyToStr(c.options["DATABASE_URL"].value) == "user=postgres password=postgres dbname=miniflux2 sslmode=disable"
}

func (c *configOptions) IsOAuth2UserCreationAllowed() bool {
	return anyToBool(c.options["OAUTH2_USER_CREATION"].value)
}

func (c *configOptions) CertKeyFile() string {
	return anyToStr(c.options["KEY_FILE"].value)
}

func (c *configOptions) ListenAddr() []string {
	return c.options["LISTEN_ADDR"].ParsedStringList
}

func (c *configOptions) LogFile() string {
	return anyToStr(c.options["LOG_FILE"].value)
}

func (c *configOptions) LogDateTime() bool {
	return anyToBool(c.options["LOG_DATE_TIME"].value)
}

func (c *configOptions) LogFormat() string {
	return anyToStr(c.options["LOG_FORMAT"].value)
}

func (c *configOptions) LogLevel() string {
	return anyToStr(c.options["LOG_LEVEL"].value)
}

func (c *configOptions) MaintenanceMessage() string {
	return anyToStr(c.options["MAINTENANCE_MESSAGE"].value)
}

func (c *configOptions) MaintenanceMode() bool {
	return anyToBool(c.options["MAINTENANCE_MODE"].value)
}

func (c *configOptions) MediaCustomProxyURL() *url.URL {
	return c.options["MEDIA_PROXY_CUSTOM_URL"].ParsedURLValue
}

func (c *configOptions) MediaProxyHTTPClientTimeout() time.Duration {
	return c.options["MEDIA_PROXY_HTTP_CLIENT_TIMEOUT"].ParsedDuration
}

func (c *configOptions) MediaProxyMode() string {
	return anyToStr(c.options["MEDIA_PROXY_MODE"].value)
}

func (c *configOptions) MediaProxyPrivateKey() []byte {
	return c.options["MEDIA_PROXY_PRIVATE_KEY"].ParsedBytesValue
}

func (c *configOptions) MediaProxyResourceTypes() []string {
	return c.options["MEDIA_PROXY_RESOURCE_TYPES"].ParsedStringList
}

func (c *configOptions) MetricsAllowedNetworks() []string {
	return c.options["METRICS_ALLOWED_NETWORKS"].ParsedStringList
}

func (c *configOptions) MetricsCollector() bool {
	return anyToBool(c.options["METRICS_COLLECTOR"].value)
}

func (c *configOptions) MetricsPassword() string {
	return anyToStr(c.options["METRICS_PASSWORD"].value)
}

func (c *configOptions) MetricsRefreshInterval() time.Duration {
	return c.options["METRICS_REFRESH_INTERVAL"].ParsedDuration
}

func (c *configOptions) MetricsUsername() string {
	return anyToStr(c.options["METRICS_USERNAME"].value)
}

func (c *configOptions) OAuth2ClientID() string {
	return anyToStr(c.options["OAUTH2_CLIENT_ID"].value)
}

func (c *configOptions) OAuth2ClientSecret() string {
	return anyToStr(c.options["OAUTH2_CLIENT_SECRET"].value)
}

func (c *configOptions) OAuth2OIDCDiscoveryEndpoint() string {
	return anyToStr(c.options["OAUTH2_OIDC_DISCOVERY_ENDPOINT"].value)
}

func (c *configOptions) OAuth2OIDCProviderName() string {
	return anyToStr(c.options["OAUTH2_OIDC_PROVIDER_NAME"].value)
}

func (c *configOptions) OAuth2Provider() string {
	return anyToStr(c.options["OAUTH2_PROVIDER"].value)
}

func (c *configOptions) OAuth2RedirectURL() string {
	return anyToStr(c.options["OAUTH2_REDIRECT_URL"].value)
}

func (c *configOptions) OAuth2UserCreation() bool {
	return anyToBool(c.options["OAUTH2_USER_CREATION"].value)
}

func (c *configOptions) PollingFrequency() time.Duration {
	return c.options["POLLING_FREQUENCY"].ParsedDuration
}

func (c *configOptions) PollingLimitPerHost() int {
	return c.options["POLLING_LIMIT_PER_HOST"].ParsedIntValue
}

func (c *configOptions) PollingParsingErrorLimit() int {
	return c.options["POLLING_PARSING_ERROR_LIMIT"].ParsedIntValue
}

func (c *configOptions) PollingScheduler() string {
	return anyToStr(c.options["POLLING_SCHEDULER"].value)
}

func (c *configOptions) Port() string {
	return anyToStr(c.options["PORT"].value)
}

func (c *configOptions) RunMigrations() bool {
	return anyToBool(c.options["RUN_MIGRATIONS"].value)
}

func (c *configOptions) SetLogLevel(level string) {
	c.options["LOG_LEVEL"].value = level
	c.options["LOG_LEVEL"].str = level
}

func (c *configOptions) SetHTTPSValue(value bool) {
	c.options["HTTPS"].value = value
}

func (c *configOptions) SchedulerEntryFrequencyFactor() int {
	return c.options["SCHEDULER_ENTRY_FREQUENCY_FACTOR"].ParsedIntValue
}

func (c *configOptions) SchedulerEntryFrequencyMaxInterval() time.Duration {
	return c.options["SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL"].ParsedDuration
}

func (c *configOptions) SchedulerEntryFrequencyMinInterval() time.Duration {
	return c.options["SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL"].ParsedDuration
}

func (c *configOptions) SchedulerRoundRobinMaxInterval() time.Duration {
	return c.options["SCHEDULER_ROUND_ROBIN_MAX_INTERVAL"].ParsedDuration
}

func (c *configOptions) SchedulerRoundRobinMinInterval() time.Duration {
	return c.options["SCHEDULER_ROUND_ROBIN_MIN_INTERVAL"].ParsedDuration
}

func (c *configOptions) Watchdog() bool {
	return anyToBool(c.options["WATCHDOG"].value)
}

func (c *configOptions) WebAuthn() bool {
	return anyToBool(c.options["WEBAUTHN"].value)
}

func (c *configOptions) WorkerPoolSize() int {
	return c.options["WORKER_POOL_SIZE"].ParsedIntValue
}

func (c *configOptions) YouTubeAPIKey() string {
	return anyToStr(c.options["YOUTUBE_API_KEY"].value)
}

func (c *configOptions) YouTubeEmbedUrlOverride() string {
	return anyToStr(c.options["YOUTUBE_EMBED_URL_OVERRIDE"].value)
}

func (c *configOptions) YouTubeEmbedDomain() string {
	return c.youTubeEmbedDomain
}

func (c *configOptions) ConfigMap(redactSecret bool) []*optionPair {
	sortedKeys := slices.Sorted(maps.Keys(c.options))
	sortedOptions := make([]*optionPair, 0, len(sortedKeys))
	for _, key := range sortedKeys {
		value := c.options[key]
		displayValue := value.str
		if displayValue == "" {
			displayValue = value.RawValue
		}
		if redactSecret && value.Secret && displayValue != "" {
			displayValue = "<redacted>"
		}
		sortedOptions = append(sortedOptions, &optionPair{Key: key, Value: displayValue})
	}
	return sortedOptions
}

func (c *configOptions) String() string {
	var builder strings.Builder

	for _, option := range c.ConfigMap(false) {
		fmt.Fprintf(&builder, "%s=%v\n", option.Key, option.Value)
	}

	return builder.String()
}
