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

func (cp *configParser) postParsing() error {
	// Parse basePath and rootURL based on BASE_URL
	baseURL := cp.options.options["BASE_URL"].ParsedStringValue
	baseURL = strings.TrimSuffix(baseURL, "/")

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid BASE_URL: %v", err)
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "https" && scheme != "http" {
		return errors.New("BASE_URL scheme must be http or https")
	}

	cp.options.options["BASE_URL"].ParsedStringValue = baseURL
	cp.options.basePath = parsedURL.Path

	parsedURL.Path = ""
	cp.options.rootURL = parsedURL.String()

	// Parse YouTube embed domain based on YOUTUBE_EMBED_URL_OVERRIDE
	youTubeEmbedURLOverride := cp.options.options["YOUTUBE_EMBED_URL_OVERRIDE"].ParsedStringValue
	if youTubeEmbedURLOverride != "" {
		parsedYouTubeEmbedURL, err := url.Parse(youTubeEmbedURLOverride)
		if err != nil {
			return fmt.Errorf("invalid YOUTUBE_EMBED_URL_OVERRIDE: %v", err)
		}
		cp.options.youTubeEmbedDomain = parsedYouTubeEmbedURL.Hostname()
	}

	// Generate a media proxy private key if not set
	if len(cp.options.options["MEDIA_PROXY_PRIVATE_KEY"].ParsedBytesValue) == 0 {
		randomKey := make([]byte, 16)
		rand.Read(randomKey)
		cp.options.options["MEDIA_PROXY_PRIVATE_KEY"].ParsedBytesValue = randomKey
	}

	// Override LISTEN_ADDR with PORT if set (for compatibility reasons)
	if cp.options.Port() != "" {
		cp.options.options["LISTEN_ADDR"].ParsedStringList = []string{":" + cp.options.Port()}
		cp.options.options["LISTEN_ADDR"].RawValue = ":" + cp.options.Port()
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
		// Ignore unknown configuration keys to avoid parsing unrelated environment variables.
		return nil
	}

	// Validate the option if a validator is provided
	if field.Validator != nil {
		if err := field.Validator(value); err != nil {
			return fmt.Errorf("invalid value for key %s: %v", key, err)
		}
	}

	// Convert the raw value based on its type
	switch field.ValueType {
	case stringType:
		field.ParsedStringValue = parseStringValue(value, field.ParsedStringValue)
		field.RawValue = value
	case stringListType:
		field.ParsedStringList = parseStringListValue(value, field.ParsedStringList)
		field.RawValue = value
	case boolType:
		parsedValue, err := parseBoolValue(value, field.ParsedBoolValue)
		if err != nil {
			return fmt.Errorf("invalid boolean value for key %s: %v", key, err)
		}
		field.ParsedBoolValue = parsedValue
		field.RawValue = value
	case intType:
		field.ParsedIntValue = parseIntValue(value, field.ParsedIntValue)
		field.RawValue = value
	case int64Type:
		field.ParsedInt64Value = ParsedInt64Value(value, field.ParsedInt64Value)
		field.RawValue = value
	case secondType:
		field.ParsedDuration = parseDurationValue(value, time.Second, field.ParsedDuration)
		field.RawValue = value
	case minuteType:
		field.ParsedDuration = parseDurationValue(value, time.Minute, field.ParsedDuration)
		field.RawValue = value
	case hourType:
		field.ParsedDuration = parseDurationValue(value, time.Hour, field.ParsedDuration)
		field.RawValue = value
	case dayType:
		field.ParsedDuration = parseDurationValue(value, time.Hour*24, field.ParsedDuration)
		field.RawValue = value
	case urlType:
		parsedURL, err := parseURLValue(value, field.ParsedURLValue)
		if err != nil {
			return fmt.Errorf("invalid URL for key %s: %v", key, err)
		}
		field.ParsedURLValue = parsedURL
		field.RawValue = value
	case secretFileType:
		secretValue, err := readSecretFileValue(value)
		if err != nil {
			return fmt.Errorf("error reading secret file for key %s: %v", key, err)
		}
		if field.TargetKey != "" {
			if targetField, ok := cp.options.options[field.TargetKey]; ok {
				targetField.ParsedStringValue = secretValue
				targetField.RawValue = secretValue
			}
		}
		field.RawValue = value
	case bytesType:
		if value != "" {
			field.ParsedBytesValue = []byte(value)
			field.RawValue = value
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
