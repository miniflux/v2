// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config // import "miniflux.app/v2/internal/config"

import (
	"maps"
	"net"
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
	parsedStringValue string
	parsedBoolValue   bool
	parsedIntValue    int
	parsedInt64Value  int64
	parsedDuration    time.Duration
	parsedStringList  []string
	parsedURLValue    *url.URL
	parsedBytesValue  []byte

	rawValue  string
	valueType configValueType
	secret    bool
	targetKey string

	validator func(string) error
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
				parsedStringValue: "",
				rawValue:          "",
				valueType:         stringType,
				secret:            true,
			},
			"ADMIN_PASSWORD_FILE": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         secretFileType,
				targetKey:         "ADMIN_PASSWORD",
			},
			"ADMIN_USERNAME": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         stringType,
			},
			"ADMIN_USERNAME_FILE": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         secretFileType,
				targetKey:         "ADMIN_USERNAME",
			},
			"AUTH_PROXY_HEADER": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         stringType,
			},
			"AUTH_PROXY_USER_CREATION": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"BASE_URL": {
				parsedStringValue: "http://localhost",
				rawValue:          "http://localhost",
				valueType:         stringType,
			},
			"BATCH_SIZE": {
				parsedIntValue: 100,
				rawValue:       "100",
				valueType:      intType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"CERT_DOMAIN": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         stringType,
			},
			"CERT_FILE": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         stringType,
			},
			"CLEANUP_ARCHIVE_BATCH_SIZE": {
				parsedIntValue: 10000,
				rawValue:       "10000",
				valueType:      intType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"CLEANUP_ARCHIVE_READ_DAYS": {
				parsedDuration: time.Hour * 24 * 60,
				rawValue:       "60",
				valueType:      dayType,
			},
			"CLEANUP_ARCHIVE_UNREAD_DAYS": {
				parsedDuration: time.Hour * 24 * 180,
				rawValue:       "180",
				valueType:      dayType,
			},
			"CLEANUP_FREQUENCY_HOURS": {
				parsedDuration: time.Hour * 24,
				rawValue:       "24",
				valueType:      hourType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"CLEANUP_REMOVE_SESSIONS_DAYS": {
				parsedDuration: time.Hour * 24 * 30,
				rawValue:       "30",
				valueType:      dayType,
			},
			"CREATE_ADMIN": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"DATABASE_CONNECTION_LIFETIME": {
				parsedDuration: time.Minute * 5,
				rawValue:       "5",
				valueType:      minuteType,
				validator: func(rawValue string) error {
					return validateGreaterThan(rawValue, 0)
				},
			},
			"DATABASE_MAX_CONNS": {
				parsedIntValue: 20,
				rawValue:       "20",
				valueType:      intType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"DATABASE_MIN_CONNS": {
				parsedIntValue: 1,
				rawValue:       "1",
				valueType:      intType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 0)
				},
			},
			"DATABASE_URL": {
				parsedStringValue: "user=postgres password=postgres dbname=miniflux2 sslmode=disable",
				rawValue:          "user=postgres password=postgres dbname=miniflux2 sslmode=disable",
				valueType:         stringType,
				secret:            true,
			},
			"DATABASE_URL_FILE": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         secretFileType,
				targetKey:         "DATABASE_URL",
			},
			"DISABLE_API": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"DISABLE_HSTS": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"DISABLE_HTTP_SERVICE": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"DISABLE_LOCAL_AUTH": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"DISABLE_SCHEDULER_SERVICE": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"FETCH_BILIBILI_WATCH_TIME": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"FETCH_NEBULA_WATCH_TIME": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"FETCH_ODYSEE_WATCH_TIME": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"FETCH_YOUTUBE_WATCH_TIME": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"FORCE_REFRESH_INTERVAL": {
				parsedDuration: 30 * time.Minute,
				rawValue:       "30",
				valueType:      minuteType,
				validator: func(rawValue string) error {
					return validateGreaterThan(rawValue, 0)
				},
			},
			"HTTP_CLIENT_MAX_BODY_SIZE": {
				parsedInt64Value: 15,
				rawValue:         "15",
				valueType:        int64Type,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"HTTP_CLIENT_PROXIES": {
				parsedStringList: []string{},
				rawValue:         "",
				valueType:        stringListType,
				secret:           true,
			},
			"HTTP_CLIENT_PROXY": {
				parsedURLValue: nil,
				rawValue:       "",
				valueType:      urlType,
				secret:         true,
			},
			"HTTP_CLIENT_TIMEOUT": {
				parsedDuration: 20 * time.Second,
				rawValue:       "20",
				valueType:      secondType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"HTTP_CLIENT_USER_AGENT": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         stringType,
			},
			"HTTP_SERVER_TIMEOUT": {
				parsedDuration: 300 * time.Second,
				rawValue:       "300",
				valueType:      secondType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"HTTPS": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"ICON_FETCH_ALLOW_PRIVATE_NETWORKS": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"INVIDIOUS_INSTANCE": {
				parsedStringValue: "yewtu.be",
				rawValue:          "yewtu.be",
				valueType:         stringType,
			},
			"KEY_FILE": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         stringType,
			},
			"LISTEN_ADDR": {
				parsedStringList: []string{"127.0.0.1:8080"},
				rawValue:         "127.0.0.1:8080",
				valueType:        stringListType,
			},
			"LOG_DATE_TIME": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"LOG_FILE": {
				parsedStringValue: "stderr",
				rawValue:          "stderr",
				valueType:         stringType,
			},
			"LOG_FORMAT": {
				parsedStringValue: "text",
				rawValue:          "text",
				valueType:         stringType,
				validator: func(rawValue string) error {
					return validateChoices(rawValue, []string{"text", "json"})
				},
			},
			"LOG_LEVEL": {
				parsedStringValue: "info",
				rawValue:          "info",
				valueType:         stringType,
				validator: func(rawValue string) error {
					return validateChoices(rawValue, []string{"debug", "info", "warning", "error"})
				},
			},
			"MAINTENANCE_MESSAGE": {
				parsedStringValue: "Miniflux is currently under maintenance",
				rawValue:          "Miniflux is currently under maintenance",
				valueType:         stringType,
			},
			"MAINTENANCE_MODE": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"MEDIA_PROXY_CUSTOM_URL": {
				rawValue:  "",
				valueType: urlType,
			},
			"MEDIA_PROXY_ALLOW_PRIVATE_NETWORKS": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"MEDIA_PROXY_HTTP_CLIENT_TIMEOUT": {
				parsedDuration: 120 * time.Second,
				rawValue:       "120",
				valueType:      secondType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"MEDIA_PROXY_MODE": {
				parsedStringValue: "http-only",
				rawValue:          "http-only",
				valueType:         stringType,
				validator: func(rawValue string) error {
					return validateChoices(rawValue, []string{"none", "http-only", "all"})
				},
			},
			"MEDIA_PROXY_PRIVATE_KEY": {
				valueType: bytesType,
				secret:    true,
			},
			"MEDIA_PROXY_RESOURCE_TYPES": {
				parsedStringList: []string{"image"},
				rawValue:         "image",
				valueType:        stringListType,
				validator: func(rawValue string) error {
					return validateListChoices(strings.Split(rawValue, ","), []string{"image", "video", "audio"})
				},
			},
			"METRICS_ALLOWED_NETWORKS": {
				parsedStringList: []string{"127.0.0.1/8"},
				rawValue:         "127.0.0.1/8",
				valueType:        stringListType,
			},
			"METRICS_COLLECTOR": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"METRICS_PASSWORD": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         stringType,
				secret:            true,
			},
			"METRICS_PASSWORD_FILE": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         secretFileType,
				targetKey:         "METRICS_PASSWORD",
			},
			"METRICS_REFRESH_INTERVAL": {
				parsedDuration: 60 * time.Second,
				rawValue:       "60",
				valueType:      secondType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"METRICS_USERNAME": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         stringType,
			},
			"METRICS_USERNAME_FILE": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         secretFileType,
				targetKey:         "METRICS_USERNAME",
			},
			"OAUTH2_CLIENT_ID": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         stringType,
				secret:            true,
			},
			"OAUTH2_CLIENT_ID_FILE": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         secretFileType,
				targetKey:         "OAUTH2_CLIENT_ID",
			},
			"OAUTH2_CLIENT_SECRET": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         stringType,
				secret:            true,
			},
			"OAUTH2_CLIENT_SECRET_FILE": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         secretFileType,
				targetKey:         "OAUTH2_CLIENT_SECRET",
			},
			"OAUTH2_OIDC_DISCOVERY_ENDPOINT": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         stringType,
			},
			"OAUTH2_OIDC_PROVIDER_NAME": {
				parsedStringValue: "OpenID Connect",
				rawValue:          "OpenID Connect",
				valueType:         stringType,
			},
			"OAUTH2_PROVIDER": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         stringType,
				validator: func(rawValue string) error {
					return validateChoices(rawValue, []string{"oidc", "google"})
				},
			},
			"OAUTH2_REDIRECT_URL": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         stringType,
			},
			"OAUTH2_USER_CREATION": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"POLLING_FREQUENCY": {
				parsedDuration: 60 * time.Minute,
				rawValue:       "60",
				valueType:      minuteType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"POLLING_LIMIT_PER_HOST": {
				parsedIntValue: 0,
				rawValue:       "0",
				valueType:      intType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 0)
				},
			},
			"POLLING_PARSING_ERROR_LIMIT": {
				parsedIntValue: 3,
				rawValue:       "3",
				valueType:      intType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 0)
				},
			},
			"POLLING_SCHEDULER": {
				parsedStringValue: "round_robin",
				rawValue:          "round_robin",
				valueType:         stringType,
				validator: func(rawValue string) error {
					return validateChoices(rawValue, []string{"round_robin", "entry_frequency"})
				},
			},
			"PORT": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         stringType,
				validator: func(rawValue string) error {
					return validateRange(rawValue, 1, 65535)
				},
			},
			"RUN_MIGRATIONS": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"SCHEDULER_ENTRY_FREQUENCY_FACTOR": {
				parsedIntValue: 1,
				rawValue:       "1",
				valueType:      intType,
			},
			"SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL": {
				parsedDuration: 24 * time.Hour,
				rawValue:       "1440",
				valueType:      minuteType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL": {
				parsedDuration: 5 * time.Minute,
				rawValue:       "5",
				valueType:      minuteType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"SCHEDULER_ROUND_ROBIN_MAX_INTERVAL": {
				parsedDuration: 1440 * time.Minute,
				rawValue:       "1440",
				valueType:      minuteType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"SCHEDULER_ROUND_ROBIN_MIN_INTERVAL": {
				parsedDuration: 60 * time.Minute,
				rawValue:       "60",
				valueType:      minuteType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"TRUSTED_REVERSE_PROXY_NETWORKS": {
				parsedStringList: []string{},
				rawValue:         "",
				valueType:        stringListType,
				validator: func(rawValue string) error {
					for ip := range strings.SplitSeq(rawValue, ",") {
						if _, _, err := net.ParseCIDR(ip); err != nil {
							return err
						}
					}

					return nil
				},
			},
			"WATCHDOG": {
				parsedBoolValue: true,
				rawValue:        "1",
				valueType:       boolType,
			},
			"WEBAUTHN": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"WORKER_POOL_SIZE": {
				parsedIntValue: 16,
				rawValue:       "16",
				valueType:      intType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"YOUTUBE_API_KEY": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         stringType,
				secret:            true,
			},
			"YOUTUBE_EMBED_URL_OVERRIDE": {
				parsedStringValue: "https://www.youtube-nocookie.com/embed/",
				rawValue:          "https://www.youtube-nocookie.com/embed/",
				valueType:         stringType,
			},
		},
	}
}

func (c *configOptions) AdminPassword() string {
	return c.options["ADMIN_PASSWORD"].parsedStringValue
}

func (c *configOptions) AdminUsername() string {
	return c.options["ADMIN_USERNAME"].parsedStringValue
}

func (c *configOptions) AuthProxyHeader() string {
	return c.options["AUTH_PROXY_HEADER"].parsedStringValue
}

func (c *configOptions) AuthProxyUserCreation() bool {
	return c.options["AUTH_PROXY_USER_CREATION"].parsedBoolValue
}

func (c *configOptions) BasePath() string {
	return c.basePath
}

func (c *configOptions) BaseURL() string {
	return c.options["BASE_URL"].parsedStringValue
}

func (c *configOptions) RootURL() string {
	return c.rootURL
}

func (c *configOptions) BatchSize() int {
	return c.options["BATCH_SIZE"].parsedIntValue
}

func (c *configOptions) CertDomain() string {
	return c.options["CERT_DOMAIN"].parsedStringValue
}

func (c *configOptions) CertFile() string {
	return c.options["CERT_FILE"].parsedStringValue
}

func (c *configOptions) CleanupArchiveBatchSize() int {
	return c.options["CLEANUP_ARCHIVE_BATCH_SIZE"].parsedIntValue
}

func (c *configOptions) CleanupArchiveReadInterval() time.Duration {
	return c.options["CLEANUP_ARCHIVE_READ_DAYS"].parsedDuration
}

func (c *configOptions) CleanupArchiveUnreadInterval() time.Duration {
	return c.options["CLEANUP_ARCHIVE_UNREAD_DAYS"].parsedDuration
}

func (c *configOptions) CleanupFrequency() time.Duration {
	return c.options["CLEANUP_FREQUENCY_HOURS"].parsedDuration
}

func (c *configOptions) CleanupRemoveSessionsInterval() time.Duration {
	return c.options["CLEANUP_REMOVE_SESSIONS_DAYS"].parsedDuration
}

func (c *configOptions) CreateAdmin() bool {
	return c.options["CREATE_ADMIN"].parsedBoolValue
}

func (c *configOptions) DatabaseConnectionLifetime() time.Duration {
	return c.options["DATABASE_CONNECTION_LIFETIME"].parsedDuration
}

func (c *configOptions) DatabaseMaxConns() int {
	return c.options["DATABASE_MAX_CONNS"].parsedIntValue
}

func (c *configOptions) DatabaseMinConns() int {
	return c.options["DATABASE_MIN_CONNS"].parsedIntValue
}

func (c *configOptions) DatabaseURL() string {
	return c.options["DATABASE_URL"].parsedStringValue
}

func (c *configOptions) DisableHSTS() bool {
	return c.options["DISABLE_HSTS"].parsedBoolValue
}

func (c *configOptions) DisableHTTPService() bool {
	return c.options["DISABLE_HTTP_SERVICE"].parsedBoolValue
}

func (c *configOptions) DisableLocalAuth() bool {
	return c.options["DISABLE_LOCAL_AUTH"].parsedBoolValue
}

func (c *configOptions) DisableSchedulerService() bool {
	return c.options["DISABLE_SCHEDULER_SERVICE"].parsedBoolValue
}

func (c *configOptions) FetchBilibiliWatchTime() bool {
	return c.options["FETCH_BILIBILI_WATCH_TIME"].parsedBoolValue
}

func (c *configOptions) FetchNebulaWatchTime() bool {
	return c.options["FETCH_NEBULA_WATCH_TIME"].parsedBoolValue
}

func (c *configOptions) FetchOdyseeWatchTime() bool {
	return c.options["FETCH_ODYSEE_WATCH_TIME"].parsedBoolValue
}

func (c *configOptions) FetchYouTubeWatchTime() bool {
	return c.options["FETCH_YOUTUBE_WATCH_TIME"].parsedBoolValue
}

func (c *configOptions) ForceRefreshInterval() time.Duration {
	return c.options["FORCE_REFRESH_INTERVAL"].parsedDuration
}

func (c *configOptions) HasHTTPClientProxiesConfigured() bool {
	return len(c.options["HTTP_CLIENT_PROXIES"].parsedStringList) > 0
}

func (c *configOptions) HasAPI() bool {
	return !c.options["DISABLE_API"].parsedBoolValue
}

func (c *configOptions) HasHTTPService() bool {
	return !c.options["DISABLE_HTTP_SERVICE"].parsedBoolValue
}

func (c *configOptions) HasHSTS() bool {
	return !c.options["DISABLE_HSTS"].parsedBoolValue
}

func (c *configOptions) HasHTTPClientProxyURLConfigured() bool {
	return c.options["HTTP_CLIENT_PROXY"].parsedURLValue != nil
}

func (c *configOptions) HasMaintenanceMode() bool {
	return c.options["MAINTENANCE_MODE"].parsedBoolValue
}

func (c *configOptions) HasMetricsCollector() bool {
	return c.options["METRICS_COLLECTOR"].parsedBoolValue
}

func (c *configOptions) HasSchedulerService() bool {
	return !c.options["DISABLE_SCHEDULER_SERVICE"].parsedBoolValue
}

func (c *configOptions) HasWatchdog() bool {
	return c.options["WATCHDOG"].parsedBoolValue
}

func (c *configOptions) HTTPClientMaxBodySize() int64 {
	return c.options["HTTP_CLIENT_MAX_BODY_SIZE"].parsedInt64Value * 1024 * 1024
}

func (c *configOptions) HTTPClientProxies() []string {
	return c.options["HTTP_CLIENT_PROXIES"].parsedStringList
}

func (c *configOptions) HTTPClientProxyURL() *url.URL {
	return c.options["HTTP_CLIENT_PROXY"].parsedURLValue
}

func (c *configOptions) HTTPClientTimeout() time.Duration {
	return c.options["HTTP_CLIENT_TIMEOUT"].parsedDuration
}

func (c *configOptions) HTTPClientUserAgent() string {
	if c.options["HTTP_CLIENT_USER_AGENT"].parsedStringValue != "" {
		return c.options["HTTP_CLIENT_USER_AGENT"].parsedStringValue
	}
	return defaultHTTPClientUserAgent
}

func (c *configOptions) HTTPServerTimeout() time.Duration {
	return c.options["HTTP_SERVER_TIMEOUT"].parsedDuration
}

func (c *configOptions) HTTPS() bool {
	return c.options["HTTPS"].parsedBoolValue
}

func (c *configOptions) IconFetchAllowPrivateNetworks() bool {
	return c.options["ICON_FETCH_ALLOW_PRIVATE_NETWORKS"].parsedBoolValue
}

func (c *configOptions) InvidiousInstance() string {
	return c.options["INVIDIOUS_INSTANCE"].parsedStringValue
}

func (c *configOptions) IsAuthProxyUserCreationAllowed() bool {
	return c.options["AUTH_PROXY_USER_CREATION"].parsedBoolValue
}

func (c *configOptions) IsDefaultDatabaseURL() bool {
	return c.options["DATABASE_URL"].rawValue == "user=postgres password=postgres dbname=miniflux2 sslmode=disable"
}

func (c *configOptions) IsOAuth2UserCreationAllowed() bool {
	return c.options["OAUTH2_USER_CREATION"].parsedBoolValue
}

func (c *configOptions) CertKeyFile() string {
	return c.options["KEY_FILE"].parsedStringValue
}

func (c *configOptions) ListenAddr() []string {
	return c.options["LISTEN_ADDR"].parsedStringList
}

func (c *configOptions) LogFile() string {
	return c.options["LOG_FILE"].parsedStringValue
}

func (c *configOptions) LogDateTime() bool {
	return c.options["LOG_DATE_TIME"].parsedBoolValue
}

func (c *configOptions) LogFormat() string {
	return c.options["LOG_FORMAT"].parsedStringValue
}

func (c *configOptions) LogLevel() string {
	return c.options["LOG_LEVEL"].parsedStringValue
}

func (c *configOptions) MaintenanceMessage() string {
	return c.options["MAINTENANCE_MESSAGE"].parsedStringValue
}

func (c *configOptions) MaintenanceMode() bool {
	return c.options["MAINTENANCE_MODE"].parsedBoolValue
}

func (c *configOptions) MediaCustomProxyURL() *url.URL {
	return c.options["MEDIA_PROXY_CUSTOM_URL"].parsedURLValue
}

func (c *configOptions) MediaProxyAllowPrivateNetworks() bool {
	return c.options["MEDIA_PROXY_ALLOW_PRIVATE_NETWORKS"].parsedBoolValue
}

func (c *configOptions) MediaProxyHTTPClientTimeout() time.Duration {
	return c.options["MEDIA_PROXY_HTTP_CLIENT_TIMEOUT"].parsedDuration
}

func (c *configOptions) MediaProxyMode() string {
	return c.options["MEDIA_PROXY_MODE"].parsedStringValue
}

func (c *configOptions) MediaProxyPrivateKey() []byte {
	return c.options["MEDIA_PROXY_PRIVATE_KEY"].parsedBytesValue
}

func (c *configOptions) MediaProxyResourceTypes() []string {
	return c.options["MEDIA_PROXY_RESOURCE_TYPES"].parsedStringList
}

func (c *configOptions) MetricsAllowedNetworks() []string {
	return c.options["METRICS_ALLOWED_NETWORKS"].parsedStringList
}

func (c *configOptions) MetricsCollector() bool {
	return c.options["METRICS_COLLECTOR"].parsedBoolValue
}

func (c *configOptions) MetricsPassword() string {
	return c.options["METRICS_PASSWORD"].parsedStringValue
}

func (c *configOptions) MetricsRefreshInterval() time.Duration {
	return c.options["METRICS_REFRESH_INTERVAL"].parsedDuration
}

func (c *configOptions) MetricsUsername() string {
	return c.options["METRICS_USERNAME"].parsedStringValue
}

func (c *configOptions) OAuth2ClientID() string {
	return c.options["OAUTH2_CLIENT_ID"].parsedStringValue
}

func (c *configOptions) OAuth2ClientSecret() string {
	return c.options["OAUTH2_CLIENT_SECRET"].parsedStringValue
}

func (c *configOptions) OAuth2OIDCDiscoveryEndpoint() string {
	return c.options["OAUTH2_OIDC_DISCOVERY_ENDPOINT"].parsedStringValue
}

func (c *configOptions) OAuth2OIDCProviderName() string {
	return c.options["OAUTH2_OIDC_PROVIDER_NAME"].parsedStringValue
}

func (c *configOptions) OAuth2Provider() string {
	return c.options["OAUTH2_PROVIDER"].parsedStringValue
}

func (c *configOptions) OAuth2RedirectURL() string {
	return c.options["OAUTH2_REDIRECT_URL"].parsedStringValue
}

func (c *configOptions) OAuth2UserCreation() bool {
	return c.options["OAUTH2_USER_CREATION"].parsedBoolValue
}

func (c *configOptions) PollingFrequency() time.Duration {
	return c.options["POLLING_FREQUENCY"].parsedDuration
}

func (c *configOptions) PollingLimitPerHost() int {
	return c.options["POLLING_LIMIT_PER_HOST"].parsedIntValue
}

func (c *configOptions) PollingParsingErrorLimit() int {
	return c.options["POLLING_PARSING_ERROR_LIMIT"].parsedIntValue
}

func (c *configOptions) PollingScheduler() string {
	return c.options["POLLING_SCHEDULER"].parsedStringValue
}

func (c *configOptions) Port() string {
	return c.options["PORT"].parsedStringValue
}

func (c *configOptions) RunMigrations() bool {
	return c.options["RUN_MIGRATIONS"].parsedBoolValue
}

func (c *configOptions) SetLogLevel(level string) {
	c.options["LOG_LEVEL"].parsedStringValue = level
	c.options["LOG_LEVEL"].rawValue = level
}

func (c *configOptions) SetHTTPSValue(value bool) {
	c.options["HTTPS"].parsedBoolValue = value
	if value {
		c.options["HTTPS"].rawValue = "1"
	} else {
		c.options["HTTPS"].rawValue = "0"
	}
}

func (c *configOptions) SchedulerEntryFrequencyFactor() int {
	return c.options["SCHEDULER_ENTRY_FREQUENCY_FACTOR"].parsedIntValue
}

func (c *configOptions) SchedulerEntryFrequencyMaxInterval() time.Duration {
	return c.options["SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL"].parsedDuration
}

func (c *configOptions) SchedulerEntryFrequencyMinInterval() time.Duration {
	return c.options["SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL"].parsedDuration
}

func (c *configOptions) SchedulerRoundRobinMaxInterval() time.Duration {
	return c.options["SCHEDULER_ROUND_ROBIN_MAX_INTERVAL"].parsedDuration
}

func (c *configOptions) SchedulerRoundRobinMinInterval() time.Duration {
	return c.options["SCHEDULER_ROUND_ROBIN_MIN_INTERVAL"].parsedDuration
}

func (c *configOptions) TrustedReverseProxyNetworks() []string {
	return c.options["TRUSTED_REVERSE_PROXY_NETWORKS"].parsedStringList
}

func (c *configOptions) Watchdog() bool {
	return c.options["WATCHDOG"].parsedBoolValue
}

func (c *configOptions) WebAuthn() bool {
	return c.options["WEBAUTHN"].parsedBoolValue
}

func (c *configOptions) WorkerPoolSize() int {
	return c.options["WORKER_POOL_SIZE"].parsedIntValue
}

func (c *configOptions) YouTubeAPIKey() string {
	return c.options["YOUTUBE_API_KEY"].parsedStringValue
}

func (c *configOptions) YouTubeEmbedUrlOverride() string {
	return c.options["YOUTUBE_EMBED_URL_OVERRIDE"].parsedStringValue
}

func (c *configOptions) YouTubeEmbedDomain() string {
	return c.youTubeEmbedDomain
}

func (c *configOptions) ConfigMap(redactSecret bool) []*optionPair {
	sortedKeys := slices.Sorted(maps.Keys(c.options))
	sortedOptions := make([]*optionPair, 0, len(sortedKeys))
	for _, key := range sortedKeys {
		value := c.options[key]
		displayValue := value.rawValue
		if displayValue != "" && redactSecret && value.secret {
			displayValue = "<redacted>"
		}
		sortedOptions = append(sortedOptions, &optionPair{Key: key, Value: displayValue})
	}
	return sortedOptions
}

func (c *configOptions) String() string {
	var builder strings.Builder

	for _, option := range c.ConfigMap(false) {
		builder.WriteString(option.Key)
		builder.WriteByte('=')
		builder.WriteString(option.Value)
		builder.WriteByte('\n')
	}

	return builder.String()
}
