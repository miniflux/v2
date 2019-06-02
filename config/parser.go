// Copyright 2019 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package config // import "miniflux.app/config"

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func parse() (opts *Options, err error) {
	opts = &Options{}
	opts.baseURL, opts.rootURL, opts.basePath, err = parseBaseURL()
	if err != nil {
		return nil, err
	}

	opts.debug = getBooleanValue("DEBUG")
	opts.listenAddr = parseListenAddr()

	opts.databaseURL = getStringValue("DATABASE_URL", defaultDatabaseURL)
	opts.databaseMaxConns = getIntValue("DATABASE_MAX_CONNS", defaultDatabaseMaxConns)
	opts.databaseMinConns = getIntValue("DATABASE_MIN_CONNS", defaultDatabaseMinConns)
	opts.runMigrations = getBooleanValue("RUN_MIGRATIONS")

	opts.hsts = !getBooleanValue("DISABLE_HSTS")
	opts.HTTPS = getBooleanValue("HTTPS")

	opts.schedulerService = !getBooleanValue("DISABLE_SCHEDULER_SERVICE")
	opts.httpService = !getBooleanValue("DISABLE_HTTP_SERVICE")

	opts.certFile = getStringValue("CERT_FILE", defaultCertFile)
	opts.certKeyFile = getStringValue("KEY_FILE", defaultKeyFile)
	opts.certDomain = getStringValue("CERT_DOMAIN", defaultCertDomain)
	opts.certCache = getStringValue("CERT_CACHE", defaultCertCache)

	opts.cleanupFrequency = getIntValue("CLEANUP_FREQUENCY", defaultCleanupFrequency)
	opts.workerPoolSize = getIntValue("WORKER_POOL_SIZE", defaultWorkerPoolSize)
	opts.pollingFrequency = getIntValue("POLLING_FREQUENCY", defaultPollingFrequency)
	opts.batchSize = getIntValue("BATCH_SIZE", defaultBatchSize)
	opts.archiveReadDays = getIntValue("ARCHIVE_READ_DAYS", defaultArchiveReadDays)
	opts.proxyImages = getStringValue("PROXY_IMAGES", defaultProxyImages)
	opts.createAdmin = getBooleanValue("CREATE_ADMIN")
	opts.pocketConsumerKey = getStringValue("POCKET_CONSUMER_KEY", "")

	opts.oauth2UserCreationAllowed = getBooleanValue("OAUTH2_USER_CREATION")
	opts.oauth2ClientID = getStringValue("OAUTH2_CLIENT_ID", defaultOAuth2ClientID)
	opts.oauth2ClientSecret = getStringValue("OAUTH2_CLIENT_SECRET", defaultOAuth2ClientSecret)
	opts.oauth2RedirectURL = getStringValue("OAUTH2_REDIRECT_URL", defaultOAuth2RedirectURL)
	opts.oauth2Provider = getStringValue("OAUTH2_PROVIDER", defaultOAuth2Provider)

	opts.httpClientTimeout = getIntValue("HTTP_CLIENT_TIMEOUT", defaultHTTPClientTimeout)
	opts.httpClientMaxBodySize = int64(getIntValue("HTTP_CLIENT_MAX_BODY_SIZE", defaultHTTPClientMaxBodySize) * 1024 * 1024)

	return opts, nil
}

func parseBaseURL() (string, string, string, error) {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		return defaultBaseURL, defaultBaseURL, "", nil
	}

	if baseURL[len(baseURL)-1:] == "/" {
		baseURL = baseURL[:len(baseURL)-1]
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", "", "", fmt.Errorf("Invalid BASE_URL: %v", err)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "https" && scheme != "http" {
		return "", "", "", errors.New("Invalid BASE_URL: scheme must be http or https")
	}

	basePath := u.Path
	u.Path = ""
	return baseURL, u.String(), basePath, nil
}

func parseListenAddr() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}

	return getStringValue("LISTEN_ADDR", defaultListenAddr)
}

func getBooleanValue(key string) bool {
	value := strings.ToLower(os.Getenv(key))
	if value == "1" || value == "yes" || value == "true" || value == "on" {
		return true
	}
	return false
}

func getStringValue(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getIntValue(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return v
}
