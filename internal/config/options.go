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
)

type optionPair struct {
	Key   string
	Value string
}

type configValueType int

const (
	stringType configValueType = iota
	stringListType
	boolType
	intType
	int64Type
	urlType
	secondType
	minuteType
	hourType
	dayType
	secretFileType
	bytesType
)

type configValue struct {
	ParsedStringValue string
	ParsedBoolValue   bool
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

// NewConfigOptions creates a new instance of ConfigOptions with default values.
func NewConfigOptions() *configOptions {
	return &configOptions{
		rootURL:            "http://localhost",
		basePath:           "",
		youTubeEmbedDomain: "www.youtube-nocookie.com",
		options: map[string]*configValue{
			"ADMIN_PASSWORD": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         stringType,
				Secret:            true,
			},
			"ADMIN_PASSWORD_FILE": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         secretFileType,
				TargetKey:         "ADMIN_PASSWORD",
			},
			"ADMIN_USERNAME": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         stringType,
			},
			"ADMIN_USERNAME_FILE": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         secretFileType,
				TargetKey:         "ADMIN_USERNAME",
			},
			"AUTH_PROXY_HEADER": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         stringType,
			},
			"AUTH_PROXY_USER_CREATION": {
				ParsedBoolValue: false,
				RawValue:        "0",
				ValueType:       boolType,
			},
			"BASE_URL": {
				ParsedStringValue: "http://localhost",
				RawValue:          "http://localhost",
				ValueType:         stringType,
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
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         stringType,
			},
			"CERT_FILE": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         stringType,
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
				ParsedBoolValue: false,
				RawValue:        "0",
				ValueType:       boolType,
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
				ParsedStringValue: "user=postgres password=postgres dbname=miniflux2 sslmode=disable",
				RawValue:          "user=postgres password=postgres dbname=miniflux2 sslmode=disable",
				ValueType:         stringType,
				Secret:            true,
			},
			"DATABASE_URL_FILE": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         secretFileType,
				TargetKey:         "DATABASE_URL",
			},
			"DISABLE_HSTS": {
				ParsedBoolValue: false,
				RawValue:        "0",
				ValueType:       boolType,
			},
			"DISABLE_HTTP_SERVICE": {
				ParsedBoolValue: false,
				RawValue:        "0",
				ValueType:       boolType,
			},
			"DISABLE_LOCAL_AUTH": {
				ParsedBoolValue: false,
				RawValue:        "0",
				ValueType:       boolType,
			},
			"DISABLE_SCHEDULER_SERVICE": {
				ParsedBoolValue: false,
				RawValue:        "0",
				ValueType:       boolType,
			},
			"FETCH_BILIBILI_WATCH_TIME": {
				ParsedBoolValue: false,
				RawValue:        "0",
				ValueType:       boolType,
			},
			"FETCH_NEBULA_WATCH_TIME": {
				ParsedBoolValue: false,
				RawValue:        "0",
				ValueType:       boolType,
			},
			"FETCH_ODYSEE_WATCH_TIME": {
				ParsedBoolValue: false,
				RawValue:        "0",
				ValueType:       boolType,
			},
			"FETCH_YOUTUBE_WATCH_TIME": {
				ParsedBoolValue: false,
				RawValue:        "0",
				ValueType:       boolType,
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
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         stringType,
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
				ParsedBoolValue: false,
				RawValue:        "0",
				ValueType:       boolType,
			},
			"INVIDIOUS_INSTANCE": {
				ParsedStringValue: "yewtu.be",
				RawValue:          "yewtu.be",
				ValueType:         stringType,
			},
			"KEY_FILE": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         stringType,
			},
			"LISTEN_ADDR": {
				ParsedStringList: []string{"127.0.0.1:8080"},
				RawValue:         "127.0.0.1:8080",
				ValueType:        stringListType,
			},
			"LOG_DATE_TIME": {
				ParsedBoolValue: false,
				RawValue:        "0",
				ValueType:       boolType,
			},
			"LOG_FILE": {
				ParsedStringValue: "stderr",
				RawValue:          "stderr",
				ValueType:         stringType,
			},
			"LOG_FORMAT": {
				ParsedStringValue: "text",
				RawValue:          "text",
				ValueType:         stringType,
				Validator: func(rawValue string) error {
					return validateChoices(rawValue, []string{"text", "json"})
				},
			},
			"LOG_LEVEL": {
				ParsedStringValue: "info",
				RawValue:          "info",
				ValueType:         stringType,
				Validator: func(rawValue string) error {
					return validateChoices(rawValue, []string{"debug", "info", "warning", "error"})
				},
			},
			"MAINTENANCE_MESSAGE": {
				ParsedStringValue: "Miniflux is currently under maintenance",
				RawValue:          "Miniflux is currently under maintenance",
				ValueType:         stringType,
			},
			"MAINTENANCE_MODE": {
				ParsedBoolValue: false,
				RawValue:        "0",
				ValueType:       boolType,
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
				ParsedStringValue: "http-only",
				RawValue:          "http-only",
				ValueType:         stringType,
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
				ParsedBoolValue: false,
				RawValue:        "0",
				ValueType:       boolType,
			},
			"METRICS_PASSWORD": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         stringType,
				Secret:            true,
			},
			"METRICS_PASSWORD_FILE": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         secretFileType,
				TargetKey:         "METRICS_PASSWORD",
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
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         stringType,
			},
			"METRICS_USERNAME_FILE": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         secretFileType,
				TargetKey:         "METRICS_USERNAME",
			},
			"OAUTH2_CLIENT_ID": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         stringType,
				Secret:            true,
			},
			"OAUTH2_CLIENT_ID_FILE": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         secretFileType,
				TargetKey:         "OAUTH2_CLIENT_ID",
			},
			"OAUTH2_CLIENT_SECRET": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         stringType,
				Secret:            true,
			},
			"OAUTH2_CLIENT_SECRET_FILE": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         secretFileType,
				TargetKey:         "OAUTH2_CLIENT_SECRET",
			},
			"OAUTH2_OIDC_DISCOVERY_ENDPOINT": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         stringType,
			},
			"OAUTH2_OIDC_PROVIDER_NAME": {
				ParsedStringValue: "OpenID Connect",
				RawValue:          "OpenID Connect",
				ValueType:         stringType,
			},
			"OAUTH2_PROVIDER": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         stringType,
				Validator: func(rawValue string) error {
					return validateChoices(rawValue, []string{"oidc", "google"})
				},
			},
			"OAUTH2_REDIRECT_URL": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         stringType,
			},
			"OAUTH2_USER_CREATION": {
				ParsedBoolValue: false,
				RawValue:        "0",
				ValueType:       boolType,
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
				ParsedStringValue: "round_robin",
				RawValue:          "round_robin",
				ValueType:         stringType,
				Validator: func(rawValue string) error {
					return validateChoices(rawValue, []string{"round_robin", "entry_frequency"})
				},
			},
			"PORT": {
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         stringType,
				Validator: func(rawValue string) error {
					return validateRange(rawValue, 1, 65535)
				},
			},
			"RUN_MIGRATIONS": {
				ParsedBoolValue: false,
				RawValue:        "0",
				ValueType:       boolType,
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
				ParsedBoolValue: true,
				RawValue:        "1",
				ValueType:       boolType,
			},
			"WEBAUTHN": {
				ParsedBoolValue: false,
				RawValue:        "0",
				ValueType:       boolType,
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
				ParsedStringValue: "",
				RawValue:          "",
				ValueType:         stringType,
				Secret:            true,
			},
			"YOUTUBE_EMBED_URL_OVERRIDE": {
				ParsedStringValue: "https://www.youtube-nocookie.com/embed/",
				RawValue:          "https://www.youtube-nocookie.com/embed/",
				ValueType:         stringType,
			},
		},
	}
}

func (c *configOptions) AdminPassword() string {
	return c.options["ADMIN_PASSWORD"].ParsedStringValue
}

func (c *configOptions) AdminUsername() string {
	return c.options["ADMIN_USERNAME"].ParsedStringValue
}

func (c *configOptions) AuthProxyHeader() string {
	return c.options["AUTH_PROXY_HEADER"].ParsedStringValue
}

func (c *configOptions) AuthProxyUserCreation() bool {
	return c.options["AUTH_PROXY_USER_CREATION"].ParsedBoolValue
}

func (c *configOptions) BasePath() string {
	return c.basePath
}

func (c *configOptions) BaseURL() string {
	return c.options["BASE_URL"].ParsedStringValue
}

func (c *configOptions) RootURL() string {
	return c.rootURL
}

func (c *configOptions) BatchSize() int {
	return c.options["BATCH_SIZE"].ParsedIntValue
}

func (c *configOptions) CertDomain() string {
	return c.options["CERT_DOMAIN"].ParsedStringValue
}

func (c *configOptions) CertFile() string {
	return c.options["CERT_FILE"].ParsedStringValue
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
	return c.options["CREATE_ADMIN"].ParsedBoolValue
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
	return c.options["DATABASE_URL"].ParsedStringValue
}

func (c *configOptions) DisableHSTS() bool {
	return c.options["DISABLE_HSTS"].ParsedBoolValue
}

func (c *configOptions) DisableHTTPService() bool {
	return c.options["DISABLE_HTTP_SERVICE"].ParsedBoolValue
}

func (c *configOptions) DisableLocalAuth() bool {
	return c.options["DISABLE_LOCAL_AUTH"].ParsedBoolValue
}

func (c *configOptions) DisableSchedulerService() bool {
	return c.options["DISABLE_SCHEDULER_SERVICE"].ParsedBoolValue
}

func (c *configOptions) FetchBilibiliWatchTime() bool {
	return c.options["FETCH_BILIBILI_WATCH_TIME"].ParsedBoolValue
}

func (c *configOptions) FetchNebulaWatchTime() bool {
	return c.options["FETCH_NEBULA_WATCH_TIME"].ParsedBoolValue
}

func (c *configOptions) FetchOdyseeWatchTime() bool {
	return c.options["FETCH_ODYSEE_WATCH_TIME"].ParsedBoolValue
}

func (c *configOptions) FetchYouTubeWatchTime() bool {
	return c.options["FETCH_YOUTUBE_WATCH_TIME"].ParsedBoolValue
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
	return !c.options["DISABLE_HTTP_SERVICE"].ParsedBoolValue
}

func (c *configOptions) HasHSTS() bool {
	return !c.options["DISABLE_HSTS"].ParsedBoolValue
}

func (c *configOptions) HasHTTPClientProxyURLConfigured() bool {
	return c.options["HTTP_CLIENT_PROXY"].ParsedURLValue != nil
}

func (c *configOptions) HasMaintenanceMode() bool {
	return c.options["MAINTENANCE_MODE"].ParsedBoolValue
}

func (c *configOptions) HasMetricsCollector() bool {
	return c.options["METRICS_COLLECTOR"].ParsedBoolValue
}

func (c *configOptions) HasSchedulerService() bool {
	return !c.options["DISABLE_SCHEDULER_SERVICE"].ParsedBoolValue
}

func (c *configOptions) HasWatchdog() bool {
	return c.options["WATCHDOG"].ParsedBoolValue
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
	if c.options["HTTP_CLIENT_USER_AGENT"].ParsedStringValue != "" {
		return c.options["HTTP_CLIENT_USER_AGENT"].ParsedStringValue
	}
	return defaultHTTPClientUserAgent
}

func (c *configOptions) HTTPServerTimeout() time.Duration {
	return c.options["HTTP_SERVER_TIMEOUT"].ParsedDuration
}

func (c *configOptions) HTTPS() bool {
	return c.options["HTTPS"].ParsedBoolValue
}

func (c *configOptions) InvidiousInstance() string {
	return c.options["INVIDIOUS_INSTANCE"].ParsedStringValue
}

func (c *configOptions) IsAuthProxyUserCreationAllowed() bool {
	return c.options["AUTH_PROXY_USER_CREATION"].ParsedBoolValue
}

func (c *configOptions) IsDefaultDatabaseURL() bool {
	return c.options["DATABASE_URL"].RawValue == "user=postgres password=postgres dbname=miniflux2 sslmode=disable"
}

func (c *configOptions) IsOAuth2UserCreationAllowed() bool {
	return c.options["OAUTH2_USER_CREATION"].ParsedBoolValue
}

func (c *configOptions) CertKeyFile() string {
	return c.options["KEY_FILE"].ParsedStringValue
}

func (c *configOptions) ListenAddr() []string {
	return c.options["LISTEN_ADDR"].ParsedStringList
}

func (c *configOptions) LogFile() string {
	return c.options["LOG_FILE"].ParsedStringValue
}

func (c *configOptions) LogDateTime() bool {
	return c.options["LOG_DATE_TIME"].ParsedBoolValue
}

func (c *configOptions) LogFormat() string {
	return c.options["LOG_FORMAT"].ParsedStringValue
}

func (c *configOptions) LogLevel() string {
	return c.options["LOG_LEVEL"].ParsedStringValue
}

func (c *configOptions) MaintenanceMessage() string {
	return c.options["MAINTENANCE_MESSAGE"].ParsedStringValue
}

func (c *configOptions) MaintenanceMode() bool {
	return c.options["MAINTENANCE_MODE"].ParsedBoolValue
}

func (c *configOptions) MediaCustomProxyURL() *url.URL {
	return c.options["MEDIA_PROXY_CUSTOM_URL"].ParsedURLValue
}

func (c *configOptions) MediaProxyHTTPClientTimeout() time.Duration {
	return c.options["MEDIA_PROXY_HTTP_CLIENT_TIMEOUT"].ParsedDuration
}

func (c *configOptions) MediaProxyMode() string {
	return c.options["MEDIA_PROXY_MODE"].ParsedStringValue
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
	return c.options["METRICS_COLLECTOR"].ParsedBoolValue
}

func (c *configOptions) MetricsPassword() string {
	return c.options["METRICS_PASSWORD"].ParsedStringValue
}

func (c *configOptions) MetricsRefreshInterval() time.Duration {
	return c.options["METRICS_REFRESH_INTERVAL"].ParsedDuration
}

func (c *configOptions) MetricsUsername() string {
	return c.options["METRICS_USERNAME"].ParsedStringValue
}

func (c *configOptions) OAuth2ClientID() string {
	return c.options["OAUTH2_CLIENT_ID"].ParsedStringValue
}

func (c *configOptions) OAuth2ClientSecret() string {
	return c.options["OAUTH2_CLIENT_SECRET"].ParsedStringValue
}

func (c *configOptions) OAuth2OIDCDiscoveryEndpoint() string {
	return c.options["OAUTH2_OIDC_DISCOVERY_ENDPOINT"].ParsedStringValue
}

func (c *configOptions) OAuth2OIDCProviderName() string {
	return c.options["OAUTH2_OIDC_PROVIDER_NAME"].ParsedStringValue
}

func (c *configOptions) OAuth2Provider() string {
	return c.options["OAUTH2_PROVIDER"].ParsedStringValue
}

func (c *configOptions) OAuth2RedirectURL() string {
	return c.options["OAUTH2_REDIRECT_URL"].ParsedStringValue
}

func (c *configOptions) OAuth2UserCreation() bool {
	return c.options["OAUTH2_USER_CREATION"].ParsedBoolValue
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
	return c.options["POLLING_SCHEDULER"].ParsedStringValue
}

func (c *configOptions) Port() string {
	return c.options["PORT"].ParsedStringValue
}

func (c *configOptions) RunMigrations() bool {
	return c.options["RUN_MIGRATIONS"].ParsedBoolValue
}

func (c *configOptions) SetLogLevel(level string) {
	c.options["LOG_LEVEL"].ParsedStringValue = level
	c.options["LOG_LEVEL"].RawValue = level
}

func (c *configOptions) SetHTTPSValue(value bool) {
	c.options["HTTPS"].ParsedBoolValue = value
	if value {
		c.options["HTTPS"].RawValue = "1"
	} else {
		c.options["HTTPS"].RawValue = "0"
	}
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
	return c.options["WATCHDOG"].ParsedBoolValue
}

func (c *configOptions) WebAuthn() bool {
	return c.options["WEBAUTHN"].ParsedBoolValue
}

func (c *configOptions) WorkerPoolSize() int {
	return c.options["WORKER_POOL_SIZE"].ParsedIntValue
}

func (c *configOptions) YouTubeAPIKey() string {
	return c.options["YOUTUBE_API_KEY"].ParsedStringValue
}

func (c *configOptions) YouTubeEmbedUrlOverride() string {
	return c.options["YOUTUBE_EMBED_URL_OVERRIDE"].ParsedStringValue
}

func (c *configOptions) YouTubeEmbedDomain() string {
	return c.youTubeEmbedDomain
}

func (c *configOptions) ConfigMap(redactSecret bool) []*optionPair {
	sortedKeys := slices.Sorted(maps.Keys(c.options))
	sortedOptions := make([]*optionPair, 0, len(sortedKeys))
	for _, key := range sortedKeys {
		value := c.options[key]
		displayValue := value.RawValue
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
