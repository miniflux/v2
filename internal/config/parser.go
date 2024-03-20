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

// Parser handles configuration parsing.
type Parser struct {
	opts *Options
}

// NewParser returns a new Parser.
func NewParser() *Parser {
	return &Parser{
		opts: NewOptions(),
	}
}

// ParseEnvironmentVariables loads configuration values from environment variables.
func (p *Parser) ParseEnvironmentVariables() (*Options, error) {
	err := p.parseLines(os.Environ())
	if err != nil {
		return nil, err
	}
	return p.opts, nil
}

// ParseFile loads configuration values from a local file.
func (p *Parser) ParseFile(filename string) (*Options, error) {
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

func (p *Parser) parseFileContent(r io.Reader) (lines []string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") && strings.Index(line, "=") > 0 {
			lines = append(lines, line)
		}
	}
	return lines
}

func (p *Parser) parseLines(lines []string) (err error) {
	var port string

	for _, line := range lines {
		fields := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])

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
		case "DEBUG":
			parsedValue := parseBool(value, defaultDebug)
			if parsedValue {
				p.opts.logLevel = "debug"
			}
		case "SERVER_TIMING_HEADER":
			p.opts.serverTimingHeader = parseBool(value, defaultTiming)
		case "BASE_URL":
			p.opts.baseURL, p.opts.rootURL, p.opts.basePath, err = parseBaseURL(value)
			if err != nil {
				return err
			}
		case "PORT":
			port = value
		case "LISTEN_ADDR":
			p.opts.listenAddr = parseString(value, defaultListenAddr)
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
		case "POLLING_PARSING_ERROR_LIMIT":
			p.opts.pollingParsingErrorLimit = parseInt(value, defaultPollingParsingErrorLimit)
		// kept for compatibility purpose
		case "PROXY_IMAGES":
			p.opts.proxyOption = parseString(value, defaultProxyOption)
		case "PROXY_HTTP_CLIENT_TIMEOUT":
			p.opts.proxyHTTPClientTimeout = parseInt(value, defaultProxyHTTPClientTimeout)
		case "PROXY_OPTION":
			p.opts.proxyOption = parseString(value, defaultProxyOption)
		case "PROXY_MEDIA_TYPES":
			p.opts.proxyMediaTypes = parseStringList(value, []string{defaultProxyMediaTypes})
		// kept for compatibility purpose
		case "PROXY_IMAGE_URL":
			p.opts.proxyUrl = parseString(value, defaultProxyUrl)
		case "PROXY_URL":
			p.opts.proxyUrl = parseString(value, defaultProxyUrl)
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
		case "POCKET_CONSUMER_KEY":
			p.opts.pocketConsumerKey = parseString(value, defaultPocketConsumerKey)
		case "POCKET_CONSUMER_KEY_FILE":
			p.opts.pocketConsumerKey = readSecretFile(value, defaultPocketConsumerKey)
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
		case "OAUTH2_PROVIDER":
			p.opts.oauth2Provider = parseString(value, defaultOAuth2Provider)
		case "HTTP_CLIENT_TIMEOUT":
			p.opts.httpClientTimeout = parseInt(value, defaultHTTPClientTimeout)
		case "HTTP_CLIENT_MAX_BODY_SIZE":
			p.opts.httpClientMaxBodySize = int64(parseInt(value, defaultHTTPClientMaxBodySize) * 1024 * 1024)
		case "HTTP_CLIENT_PROXY":
			p.opts.httpClientProxy = parseString(value, defaultHTTPClientProxy)
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
		case "FETCH_ODYSEE_WATCH_TIME":
			p.opts.fetchOdyseeWatchTime = parseBool(value, defaultFetchOdyseeWatchTime)
		case "FETCH_YOUTUBE_WATCH_TIME":
			p.opts.fetchYouTubeWatchTime = parseBool(value, defaultFetchYouTubeWatchTime)
		case "YOUTUBE_EMBED_URL_OVERRIDE":
			p.opts.youTubeEmbedUrlOverride = parseString(value, defaultYouTubeEmbedUrlOverride)
		case "WATCHDOG":
			p.opts.watchdog = parseBool(value, defaultWatchdog)
		case "INVIDIOUS_INSTANCE":
			p.opts.invidiousInstance = parseString(value, defaultInvidiousInstance)
		case "PROXY_PRIVATE_KEY":
			randomKey := make([]byte, 16)
			rand.Read(randomKey)
			p.opts.proxyPrivateKey = parseBytes(value, randomKey)
		case "WEBAUTHN":
			p.opts.webAuthn = parseBool(value, defaultWebAuthn)
		}
	}

	if port != "" {
		p.opts.listenAddr = ":" + port
	}
	return nil
}

func parseBaseURL(value string) (string, string, string, error) {
	if value == "" {
		return defaultBaseURL, defaultRootURL, "", nil
	}

	if value[len(value)-1:] == "/" {
		value = value[:len(value)-1]
	}

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
	strMap := make(map[string]bool)

	items := strings.Split(value, ",")
	for _, item := range items {
		itemValue := strings.TrimSpace(item)

		if _, found := strMap[itemValue]; !found {
			strMap[itemValue] = true
			strList = append(strList, itemValue)
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
