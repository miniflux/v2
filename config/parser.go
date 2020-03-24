// Copyright 2019 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package config // import "miniflux.app/config"

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	url_parser "net/url"
	"os"
	"strconv"
	"strings"

	"miniflux.app/logger"
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
		if len(line) > 0 && !strings.HasPrefix(line, "#") && strings.Index(line, "=") > 0 {
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
		case "LOG_DATE_TIME":
			p.opts.logDateTime = parseBool(value, defaultLogDateTime)
		case "DEBUG":
			p.opts.debug = parseBool(value, defaultDebug)
		case "BASE_URL":
			p.opts.baseURL, p.opts.rootURL, p.opts.basePath, p.opts.host, err = parseBaseURL(value)
			if err != nil {
				return err
			}
		case "CANONICAL_HOST_REDIRECT":
			p.opts.canonicalHostRedirect = parseBool(value, defaultCanonicalHostRedirect)
		case "PORT":
			port = value
		case "LISTEN_ADDR":
			p.opts.listenAddr = parseString(value, defaultListenAddr)
		case "DATABASE_URL":
			p.opts.databaseURL = parseString(value, defaultDatabaseURL)
		case "DATABASE_MAX_CONNS":
			p.opts.databaseMaxConns = parseInt(value, defaultDatabaseMaxConns)
		case "DATABASE_MIN_CONNS":
			p.opts.databaseMinConns = parseInt(value, defaultDatabaseMinConns)
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
		case "CERT_CACHE":
			p.opts.certCache = parseString(value, defaultCertCache)
		case "CLEANUP_FREQUENCY_HOURS":
			p.opts.cleanupFrequencyHours = parseInt(value, defaultCleanupFrequencyHours)
		case "CLEANUP_ARCHIVE_READ_DAYS":
			p.opts.cleanupArchiveReadDays = parseInt(value, defaultCleanupArchiveReadDays)
		case "CLEANUP_REMOVE_SESSIONS_DAYS":
			p.opts.cleanupRemoveSessionsDays = parseInt(value, defaultCleanupRemoveSessionsDays)
		case "CLEANUP_FREQUENCY":
			logger.Error("[Config] CLEANUP_FREQUENCY has been deprecated in favor of CLEANUP_FREQUENCY_HOURS.")

			if p.opts.cleanupFrequencyHours != defaultCleanupFrequencyHours {
				logger.Error("[Config] Ignoring CLEANUP_FREQUENCY as CLEANUP_FREQUENCY_HOURS is already specified.")
			} else {
				p.opts.cleanupFrequencyHours = parseInt(value, defaultCleanupFrequencyHours)
			}
		case "ARCHIVE_READ_DAYS":
			logger.Error("[Config] ARCHIVE_READ_DAYS has been deprecated in favor of CLEANUP_ARCHIVE_READ_DAYS.")

			if p.opts.cleanupArchiveReadDays != defaultCleanupArchiveReadDays {
				logger.Error("[Config] Ignoring ARCHIVE_READ_DAYS as CLEANUP_ARCHIVE_READ_DAYS is already specified.")
			} else {
				p.opts.cleanupArchiveReadDays = parseInt(value, defaultCleanupArchiveReadDays)
			}
		case "WORKER_POOL_SIZE":
			p.opts.workerPoolSize = parseInt(value, defaultWorkerPoolSize)
		case "POLLING_FREQUENCY":
			p.opts.pollingFrequency = parseInt(value, defaultPollingFrequency)
		case "BATCH_SIZE":
			p.opts.batchSize = parseInt(value, defaultBatchSize)
		case "PROXY_IMAGES":
			p.opts.proxyImages = parseString(value, defaultProxyImages)
		case "CREATE_ADMIN":
			p.opts.createAdmin = parseBool(value, defaultCreateAdmin)
		case "POCKET_CONSUMER_KEY":
			p.opts.pocketConsumerKey = parseString(value, defaultPocketConsumerKey)
		case "OAUTH2_USER_CREATION":
			p.opts.oauth2UserCreationAllowed = parseBool(value, defaultOAuth2UserCreation)
		case "OAUTH2_CLIENT_ID":
			p.opts.oauth2ClientID = parseString(value, defaultOAuth2ClientID)
		case "OAUTH2_CLIENT_SECRET":
			p.opts.oauth2ClientSecret = parseString(value, defaultOAuth2ClientSecret)
		case "OAUTH2_REDIRECT_URL":
			p.opts.oauth2RedirectURL = parseString(value, defaultOAuth2RedirectURL)
		case "OAUTH2_OIDC_DISCOVERY_ENDPOINT":
			p.opts.oauth2OidcDiscoveryEndpoint = parseString(value, defaultOAuth2OidcDiscoveryEndpoint)
		case "OAUTH2_PROVIDER":
			p.opts.oauth2Provider = parseString(value, defaultOAuth2Provider)
		case "HTTP_CLIENT_TIMEOUT":
			p.opts.httpClientTimeout = parseInt(value, defaultHTTPClientTimeout)
		case "HTTP_CLIENT_MAX_BODY_SIZE":
			p.opts.httpClientMaxBodySize = int64(parseInt(value, defaultHTTPClientMaxBodySize) * 1024 * 1024)
		case "AUTH_PROXY_HEADER":
			p.opts.authProxyHeader = parseString(value, defaultAuthProxyHeader)
		case "AUTH_PROXY_USER_CREATION":
			p.opts.authProxyUserCreation = parseBool(value, defaultAuthProxyUserCreation)
		}
	}

	if port != "" {
		p.opts.listenAddr = ":" + port
	}
	return nil
}

func parseBaseURL(value string) (string, string, string, string, error) {
	if value == "" {
		return defaultBaseURL, defaultRootURL, "", defaultHost, nil
	}

	if value[len(value)-1:] == "/" {
		value = value[:len(value)-1]
	}

	url, err := url_parser.Parse(value)
	if err != nil {
		return "", "", "", "", fmt.Errorf("Invalid BASE_URL: %v", err)
	}

	scheme := strings.ToLower(url.Scheme)
	if scheme != "https" && scheme != "http" {
		return "", "", "", "", errors.New("Invalid BASE_URL: scheme must be http or https")
	}

	host := strings.ToLower(url.Hostname() + url.Port())

	basePath := url.Path
	url.Path = ""
	return value, url.String(), basePath, host, nil
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
