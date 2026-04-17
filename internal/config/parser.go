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
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type configParser struct {
	options *configOptions
}

func NewConfigParser() *configParser {
	return &configParser{
		options: NewConfigOptions(),
	}
}

func (cp *configParser) ParseEnvironmentVariables() (*configOptions, error) {
	if err := cp.parseLines(os.Environ()); err != nil {
		return nil, err
	}

	return cp.options, nil
}

func (cp *configParser) ParseFile(filename string) (*configOptions, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	if err := cp.parseLines(parseFileContent(fp)); err != nil {
		return nil, err
	}

	return cp.options, nil
}

// Validate checks for invalid or incomplete option combinations.
func (c *configOptions) Validate() error {
	if c.OAuth2Provider() == "oidc" && c.OAuth2OIDCDiscoveryEndpoint() == "" {
		return errors.New("OAUTH2_OIDC_DISCOVERY_ENDPOINT must be configured when using the OIDC provider")
	}

	if c.DisableLocalAuth() {
		if c.OAuth2Provider() == "" && c.AuthProxyHeader() == "" {
			return errors.New("DISABLE_LOCAL_AUTH is enabled but neither OAUTH2_PROVIDER nor AUTH_PROXY_HEADER is set. Please enable at least one authentication source")
		}
	}

	if c.AuthProxyHeader() != "" && len(c.TrustedReverseProxyNetworks()) == 0 {
		return errors.New("TRUSTED_REVERSE_PROXY_NETWORKS must be configured when AUTH_PROXY_HEADER is used")
	}

	if (c.CertFile() != "") != (c.CertKeyFile() != "") {
		return errors.New("CERT_FILE and KEY_FILE must both be provided")
	}

	if c.CertDomain() != "" && c.CertFile() != "" {
		return errors.New("CERT_DOMAIN and CERT_FILE/KEY_FILE are mutually exclusive")
	}

	if (c.MetricsUsername() != "") != (c.MetricsPassword() != "") {
		return errors.New("METRICS_USERNAME and METRICS_PASSWORD must both be provided")
	}

	if c.DatabaseMinConns() > c.DatabaseMaxConns() {
		return errors.New("DATABASE_MIN_CONNS must be less than or equal to DATABASE_MAX_CONNS")
	}

	if c.SchedulerRoundRobinMinInterval() > c.SchedulerRoundRobinMaxInterval() {
		return errors.New("SCHEDULER_ROUND_ROBIN_MIN_INTERVAL must be less than or equal to SCHEDULER_ROUND_ROBIN_MAX_INTERVAL")
	}

	if c.SchedulerEntryFrequencyMinInterval() > c.SchedulerEntryFrequencyMaxInterval() {
		return errors.New("SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL must be less than or equal to SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL")
	}

	return nil
}

func (cp *configParser) postParsing() error {
	// Parse basePath and rootURL based on BASE_URL
	baseURL := cp.options.options["BASE_URL"].parsedStringValue
	baseURL = strings.TrimSuffix(baseURL, "/")

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid BASE_URL: %v", err)
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "https" && scheme != "http" {
		return errors.New("BASE_URL scheme must be http or https")
	}

	cp.options.options["BASE_URL"].parsedStringValue = baseURL
	cp.options.basePath = parsedURL.Path

	parsedURL.Path = ""
	cp.options.rootURL = parsedURL.String()

	// Parse YouTube embed domain based on YOUTUBE_EMBED_URL_OVERRIDE
	youTubeEmbedURLOverride := cp.options.options["YOUTUBE_EMBED_URL_OVERRIDE"].parsedStringValue
	if youTubeEmbedURLOverride != "" {
		parsedYouTubeEmbedURL, err := url.Parse(youTubeEmbedURLOverride)
		if err != nil {
			return fmt.Errorf("invalid YOUTUBE_EMBED_URL_OVERRIDE: %v", err)
		}
		cp.options.youTubeEmbedDomain = parsedYouTubeEmbedURL.Hostname()
	}

	// Generate a media proxy private key if not set
	if len(cp.options.options["MEDIA_PROXY_PRIVATE_KEY"].parsedBytesValue) == 0 {
		randomKey := make([]byte, 16)
		rand.Read(randomKey)
		cp.options.options["MEDIA_PROXY_PRIVATE_KEY"].parsedBytesValue = randomKey
	}

	// Override LISTEN_ADDR with PORT if set (for compatibility reasons)
	if cp.options.Port() != "" {
		cp.options.options["LISTEN_ADDR"].parsedStringList = []string{":" + cp.options.Port()}
		cp.options.options["LISTEN_ADDR"].rawValue = ":" + cp.options.Port()
	}

	return nil
}

func (cp *configParser) parseLines(lines []string) error {
	for lineNum, line := range lines {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("unable to parse configuration, invalid format on line %d", lineNum)
		}

		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if err := cp.parseLine(key, value); err != nil {
			return err
		}
	}

	if err := cp.postParsing(); err != nil {
		return err
	}

	return nil
}

func (cp *configParser) parseLine(key, value string) error {
	field, exists := cp.options.options[key]
	if !exists {
		if key == "FILTER_ENTRY_MAX_AGE_DAYS" {
			slog.Warn("Configuration option FILTER_ENTRY_MAX_AGE_DAYS is deprecated; use user filter rule max-age:<duration> instead")
		}
		// Ignore unknown configuration keys to avoid parsing unrelated environment variables.
		return nil
	}

	// Validate the option if a validator is provided
	if field.validator != nil {
		if err := field.validator(value); err != nil {
			return fmt.Errorf("invalid value for key %s: %v", key, err)
		}
	}

	// Convert the raw value based on its type
	switch field.valueType {
	case stringType:
		field.parsedStringValue = parseStringValue(value, field.parsedStringValue)
		field.rawValue = value
	case stringListType:
		field.parsedStringList = parseStringListValue(value, field.parsedStringList)
		field.rawValue = value
	case boolType:
		parsedValue, err := parseBoolValue(value, field.parsedBoolValue)
		if err != nil {
			return fmt.Errorf("invalid boolean value for key %s: %v", key, err)
		}
		field.parsedBoolValue = parsedValue
		field.rawValue = value
	case intType:
		field.parsedIntValue = parseIntValue(value, field.parsedIntValue)
		field.rawValue = value
	case int64Type:
		field.parsedInt64Value = ParsedInt64Value(value, field.parsedInt64Value)
		field.rawValue = value
	case secondType:
		field.parsedDuration = parseDurationValue(value, time.Second, field.parsedDuration)
		field.rawValue = value
	case minuteType:
		field.parsedDuration = parseDurationValue(value, time.Minute, field.parsedDuration)
		field.rawValue = value
	case hourType:
		field.parsedDuration = parseDurationValue(value, time.Hour, field.parsedDuration)
		field.rawValue = value
	case dayType:
		field.parsedDuration = parseDurationValue(value, time.Hour*24, field.parsedDuration)
		field.rawValue = value
	case urlType:
		parsedURL, err := parseURLValue(value, field.parsedURLValue)
		if err != nil {
			return fmt.Errorf("invalid URL for key %s: %v", key, err)
		}
		field.parsedURLValue = parsedURL
		field.rawValue = value
	case secretFileType:
		secretValue, err := readSecretFileValue(value)
		if err != nil {
			return fmt.Errorf("error reading secret file for key %s: %v", key, err)
		}
		if field.targetKey != "" {
			if targetField, ok := cp.options.options[field.targetKey]; ok {
				targetField.parsedStringValue = secretValue
				targetField.rawValue = secretValue
			}
		}
		field.rawValue = value
	case bytesType:
		if value != "" {
			field.parsedBytesValue = []byte(value)
			field.rawValue = value
		}
	}

	return nil
}

func parseStringValue(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func parseBoolValue(value string, fallback bool) (bool, error) {
	if value == "" {
		return fallback, nil
	}

	value = strings.ToLower(value)
	if value == "1" || value == "yes" || value == "true" || value == "on" {
		return true, nil
	}
	if value == "0" || value == "no" || value == "false" || value == "off" {
		return false, nil
	}

	return false, fmt.Errorf("invalid boolean value: %q", value)
}

func parseIntValue(value string, fallback int) int {
	if value == "" {
		return fallback
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return v
}

func ParsedInt64Value(value string, fallback int64) int64 {
	if value == "" {
		return fallback
	}

	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}

	return v
}

func parseStringListValue(value string, fallback []string) []string {
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

func parseDurationValue(value string, unit time.Duration, fallback time.Duration) time.Duration {
	if value == "" {
		return fallback
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return time.Duration(v) * unit
}

func parseURLValue(value string, fallback *url.URL) (*url.URL, error) {
	if value == "" {
		return fallback, nil
	}

	parsedURL, err := url.Parse(value)
	if err != nil {
		return fallback, err
	}

	return parsedURL, nil
}

func readSecretFileValue(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	value := string(bytes.TrimSpace(data))
	if value == "" {
		return "", errors.New("secret file is empty")
	}

	return value, nil
}

func parseFileContent(r io.Reader) (lines []string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") && strings.Index(line, "=") > 0 {
			lines = append(lines, line)
		}
	}
	return lines
}
