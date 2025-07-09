// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config // import "miniflux.app/v2/internal/config"

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// parser handles configuration parsing.
type parser struct {
	opts *options
}

// NewParser returns a new Parser.
func NewParser() *parser {
	return &parser{
		opts: NewOptions(),
	}
}

// ParseEnvironmentVariables loads configuration values from environment variables.
func (p *parser) ParseEnvironmentVariables() (*options, error) {
	err := p.parseLines(os.Environ())
	if err != nil {
		return nil, err
	}
	return p.opts, nil
}

// ParseFile loads configuration values from a local file.
func (p *parser) ParseFile(filename string) (*options, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	err = p.parseLines(p.parseFileContent(fp))
	if err != nil {
		return nil, err
	}
	return p.opts, nil
}

func (p *parser) parseFileContent(r io.Reader) (lines []string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") && strings.Index(line, "=") > 0 {
			lines = append(lines, line)
		}
	}
	return lines
}

func (p *parser) parseLines(lines []string) (err error) {
	var port string

	for lineNum, line := range lines {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("config: unable to parse configuration, invalid format on line %d", lineNum)
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)

		switch key {
		case "LOG_FILE":
			p.opts.logFile = parseString(value, defaultLogFile)
		case "LOG_DATE_TIME":
			p.opts.logDateTime = parseBool(value, defaultLogDateTime)
		case "LOG_LEVEL":
			parsedValue := parseString(value, defaultLogLevel)
			if parsedValue == "debug" || parsedValue == "info" || parsedValue == "warning" || parsedValue == "error" {
				p.opts.logLevel = parsedValue
			}
		case "LOG_FORMAT":
			parsedValue := parseString(value, defaultLogFormat)
			if parsedValue == "json" || parsedValue == "text" {
				p.opts.logFormat = parsedValue
			}
		case "BASE_URL":
			p.opts.baseURL, p.opts.rootURL, p.opts.basePath, err = parseBaseURL(value)
			if err != nil {
				return err
			}
		case "PORT":
			port = value
		case "LISTEN_ADDR":
			p.opts.listenAddr = parseStringList(value, []string{defaultListenAddr})
		case "DATABASE_URL":
			p.opts.databaseURL = parseString(value, defaultDatabaseURL)
		case "DATABASE_URL_FILE":
			p.opts.databaseURL = readSecretFile(value, defaultDatabaseURL)
		case "DATABASE_MAX_CONNS":
			p.opts.databaseMaxConns = parseInt(value, defaultDatabaseMaxConns)
		case "DATABASE_MIN_CONNS":
			p.opts.databaseMinConns = parseInt(value, defaultDatabaseMinConns)
		case "DATABASE_CONNECTION_LIFETIME":
			p.opts.databaseConnectionLifetime = parseInt(value, defaultDatabaseConnectionLifetime)
		case "FILTER_ENTRY_MAX_AGE_DAYS":
			p.opts.filterEntryMaxAgeDays = parseInt(value, defaultFilterEntryMaxAgeDays)
		case "RUN_MIGRATIONS":
			p.opts.runMigrations = parseBool(value, defaultRunMigrations)
		case "DISABLE_HSTS":
			p.opts.hsts = !parseBool(value, defaultHSTS)
		case "HTTPS":
			p.opts.HTTPS = parseBool(value, defaultHTTPS)
		case "DISABLE_SCHEDULER_SERVICE":
			p.opts.schedulerService = !parseBool(value, defaultSchedulerService)
		case "DISABLE_HTTP_SERVICE":
			p.opts.httpService = !parseBool(value, defaultHTTPService)
		case "CERT_FILE":
			p.opts.certFile = parseString(value, defaultCertFile)
		case "KEY_FILE":
			p.opts.certKeyFile = parseString(value, defaultKeyFile)
		case "CERT_DOMAIN":
			p.opts.certDomain = parseString(value, defaultCertDomain)
		case "CLEANUP_FREQUENCY_HOURS":
			p.opts.cleanupFrequencyHours = parseInt(value, defaultCleanupFrequencyHours)
		case "CLEANUP_ARCHIVE_READ_DAYS":
			p.opts.cleanupArchiveReadDays = parseInt(value, defaultCleanupArchiveReadDays)
		case "CLEANUP_ARCHIVE_UNREAD_DAYS":
			p.opts.cleanupArchiveUnreadDays = parseInt(value, defaultCleanupArchiveUnreadDays)
		case "CLEANUP_ARCHIVE_BATCH_SIZE":
			p.opts.cleanupArchiveBatchSize = parseInt(value, defaultCleanupArchiveBatchSize)
		case "CLEANUP_REMOVE_SESSIONS_DAYS":
			p.opts.cleanupRemoveSessionsDays = parseInt(value, defaultCleanupRemoveSessionsDays)
		case "WORKER_POOL_SIZE":
			p.opts.workerPoolSize = parseInt(value, defaultWorkerPoolSize)
		case "POLLING_FREQUENCY":
			p.opts.pollingFrequency = parseInt(value, defaultPollingFrequency)
		case "FORCE_REFRESH_INTERVAL":
			p.opts.forceRefreshInterval = parseInt(value, defaultForceRefreshInterval)
		case "BATCH_SIZE":
			p.opts.batchSize = parseInt(value, defaultBatchSize)
		case "POLLING_SCHEDULER":
			p.opts.pollingScheduler = strings.ToLower(parseString(value, defaultPollingScheduler))
		case "SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL":
			p.opts.schedulerEntryFrequencyMaxInterval = parseInt(value, defaultSchedulerEntryFrequencyMaxInterval)
		case "SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL":
			p.opts.schedulerEntryFrequencyMinInterval = parseInt(value, defaultSchedulerEntryFrequencyMinInterval)
		case "SCHEDULER_ENTRY_FREQUENCY_FACTOR":
			p.opts.schedulerEntryFrequencyFactor = parseInt(value, defaultSchedulerEntryFrequencyFactor)
		case "SCHEDULER_ROUND_ROBIN_MIN_INTERVAL":
			p.opts.schedulerRoundRobinMinInterval = parseInt(value, defaultSchedulerRoundRobinMinInterval)
		case "SCHEDULER_ROUND_ROBIN_MAX_INTERVAL":
			p.opts.schedulerRoundRobinMaxInterval = parseInt(value, defaultSchedulerRoundRobinMaxInterval)
		case "POLLING_PARSING_ERROR_LIMIT":
			p.opts.pollingParsingErrorLimit = parseInt(value, defaultPollingParsingErrorLimit)
		case "MEDIA_PROXY_HTTP_CLIENT_TIMEOUT":
			p.opts.mediaProxyHTTPClientTimeout = parseInt(value, defaultMediaProxyHTTPClientTimeout)
		case "MEDIA_PROXY_MODE":
			p.opts.mediaProxyMode = parseString(value, defaultMediaProxyMode)
		case "MEDIA_PROXY_RESOURCE_TYPES":
			p.opts.mediaProxyResourceTypes = parseStringList(value, []string{defaultMediaResourceTypes})
		case "MEDIA_PROXY_PRIVATE_KEY":
			randomKey := make([]byte, 16)
			if _, err := rand.Read(randomKey); err != nil {
				return fmt.Errorf("config: unable to generate random key: %w", err)
			}
			p.opts.mediaProxyPrivateKey = parseBytes(value, randomKey)
		case "MEDIA_PROXY_CUSTOM_URL":
			p.opts.mediaProxyCustomURL = parseString(value, defaultMediaProxyURL)
		case "CREATE_ADMIN":
			p.opts.createAdmin = parseBool(value, defaultCreateAdmin)
		case "ADMIN_USERNAME":
			p.opts.adminUsername = parseString(value, defaultAdminUsername)
		case "ADMIN_USERNAME_FILE":
			p.opts.adminUsername = readSecretFile(value, defaultAdminUsername)
		case "ADMIN_PASSWORD":
			p.opts.adminPassword = parseString(value, defaultAdminPassword)
		case "ADMIN_PASSWORD_FILE":
			p.opts.adminPassword = readSecretFile(value, defaultAdminPassword)
		case "OAUTH2_USER_CREATION":
			p.opts.oauth2UserCreationAllowed = parseBool(value, defaultOAuth2UserCreation)
		case "OAUTH2_CLIENT_ID":
			p.opts.oauth2ClientID = parseString(value, defaultOAuth2ClientID)
		case "OAUTH2_CLIENT_ID_FILE":
			p.opts.oauth2ClientID = readSecretFile(value, defaultOAuth2ClientID)
		case "OAUTH2_CLIENT_SECRET":
			p.opts.oauth2ClientSecret = parseString(value, defaultOAuth2ClientSecret)
		case "OAUTH2_CLIENT_SECRET_FILE":
			p.opts.oauth2ClientSecret = readSecretFile(value, defaultOAuth2ClientSecret)
		case "OAUTH2_REDIRECT_URL":
			p.opts.oauth2RedirectURL = parseString(value, defaultOAuth2RedirectURL)
		case "OAUTH2_OIDC_DISCOVERY_ENDPOINT":
			p.opts.oidcDiscoveryEndpoint = parseString(value, defaultOAuth2OidcDiscoveryEndpoint)
		case "OAUTH2_OIDC_PROVIDER_NAME":
			p.opts.oidcProviderName = parseString(value, defaultOauth2OidcProviderName)
		case "OAUTH2_PROVIDER":
			p.opts.oauth2Provider = parseString(value, defaultOAuth2Provider)
		case "DISABLE_LOCAL_AUTH":
			p.opts.disableLocalAuth = parseBool(value, defaultDisableLocalAuth)
		case "HTTP_CLIENT_TIMEOUT":
			p.opts.httpClientTimeout = parseInt(value, defaultHTTPClientTimeout)
		case "HTTP_CLIENT_MAX_BODY_SIZE":
			p.opts.httpClientMaxBodySize = int64(parseInt(value, defaultHTTPClientMaxBodySize) * 1024 * 1024)
		case "HTTP_CLIENT_PROXY":
			p.opts.httpClientProxyURL, err = url.Parse(parseString(value, defaultHTTPClientProxy))
			if err != nil {
				return fmt.Errorf("config: invalid HTTP_CLIENT_PROXY value: %w", err)
			}
		case "HTTP_CLIENT_PROXIES":
			p.opts.httpClientProxies = parseStringList(value, []string{})
		case "HTTP_CLIENT_USER_AGENT":
			p.opts.httpClientUserAgent = parseString(value, defaultHTTPClientUserAgent)
		case "HTTP_SERVER_TIMEOUT":
			p.opts.httpServerTimeout = parseInt(value, defaultHTTPServerTimeout)
		case "AUTH_PROXY_HEADER":
			p.opts.authProxyHeader = parseString(value, defaultAuthProxyHeader)
		case "AUTH_PROXY_USER_CREATION":
			p.opts.authProxyUserCreation = parseBool(value, defaultAuthProxyUserCreation)
		case "MAINTENANCE_MODE":
			p.opts.maintenanceMode = parseBool(value, defaultMaintenanceMode)
		case "MAINTENANCE_MESSAGE":
			p.opts.maintenanceMessage = parseString(value, defaultMaintenanceMessage)
		case "METRICS_COLLECTOR":
			p.opts.metricsCollector = parseBool(value, defaultMetricsCollector)
		case "METRICS_REFRESH_INTERVAL":
			p.opts.metricsRefreshInterval = parseInt(value, defaultMetricsRefreshInterval)
		case "METRICS_ALLOWED_NETWORKS":
			p.opts.metricsAllowedNetworks = parseStringList(value, []string{defaultMetricsAllowedNetworks})
		case "METRICS_USERNAME":
			p.opts.metricsUsername = parseString(value, defaultMetricsUsername)
		case "METRICS_USERNAME_FILE":
			p.opts.metricsUsername = readSecretFile(value, defaultMetricsUsername)
		case "METRICS_PASSWORD":
			p.opts.metricsPassword = parseString(value, defaultMetricsPassword)
		case "METRICS_PASSWORD_FILE":
			p.opts.metricsPassword = readSecretFile(value, defaultMetricsPassword)
		case "FETCH_BILIBILI_WATCH_TIME":
			p.opts.fetchBilibiliWatchTime = parseBool(value, defaultFetchBilibiliWatchTime)
		case "FETCH_NEBULA_WATCH_TIME":
			p.opts.fetchNebulaWatchTime = parseBool(value, defaultFetchNebulaWatchTime)
		case "FETCH_ODYSEE_WATCH_TIME":
			p.opts.fetchOdyseeWatchTime = parseBool(value, defaultFetchOdyseeWatchTime)
		case "FETCH_YOUTUBE_WATCH_TIME":
			p.opts.fetchYouTubeWatchTime = parseBool(value, defaultFetchYouTubeWatchTime)
		case "YOUTUBE_API_KEY":
			p.opts.youTubeApiKey = parseString(value, defaultYouTubeApiKey)
		case "YOUTUBE_EMBED_URL_OVERRIDE":
			p.opts.youTubeEmbedUrlOverride = parseString(value, defaultYouTubeEmbedUrlOverride)
		case "WATCHDOG":
			p.opts.watchdog = parseBool(value, defaultWatchdog)
		case "INVIDIOUS_INSTANCE":
			p.opts.invidiousInstance = parseString(value, defaultInvidiousInstance)
		case "WEBAUTHN":
			p.opts.webAuthn = parseBool(value, defaultWebAuthn)
		}
	}

	if port != "" {
		p.opts.listenAddr = []string{":" + port}
	}

	youtubeEmbedURL, err := url.Parse(p.opts.youTubeEmbedUrlOverride)
	if err != nil {
		return fmt.Errorf("config: invalid YOUTUBE_EMBED_URL_OVERRIDE value: %w", err)
	}
	p.opts.youTubeEmbedDomain = youtubeEmbedURL.Hostname()

	return nil
}

func parseBaseURL(value string) (string, string, string, error) {
	if value == "" {
		return defaultBaseURL, defaultRootURL, "", nil
	}

	value = strings.TrimSuffix(value, "/")

	parsedURL, err := url.Parse(value)
	if err != nil {
		return "", "", "", fmt.Errorf("config: invalid BASE_URL: %w", err)
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "https" && scheme != "http" {
		return "", "", "", errors.New("config: invalid BASE_URL: scheme must be http or https")
	}

	basePath := parsedURL.Path
	parsedURL.Path = ""
	return value, parsedURL.String(), basePath, nil
}

func parseBool(value string, fallback bool) bool {
	if value == "" {
		return fallback
	}

	value = strings.ToLower(value)
	if value == "1" || value == "yes" || value == "true" || value == "on" {
		return true
	}

	return false
}

func parseInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return v
}

func parseString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func parseStringList(value string, fallback []string) []string {
	if value == "" {
		return fallback
	}

	var strList []string
	present := make(map[string]bool)

	for item := range strings.SplitSeq(value, ",") {
		if itemValue := strings.TrimSpace(item); itemValue != "" {
			if !present[itemValue] {
				present[itemValue] = true
				strList = append(strList, itemValue)
			}
		}
	}

	return strList
}

func parseBytes(value string, fallback []byte) []byte {
	if value == "" {
		return fallback
	}

	return []byte(value)
}

func readSecretFile(filename, fallback string) string {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fallback
	}

	value := string(bytes.TrimSpace(data))
	if value == "" {
		return fallback
	}

	return value
}
