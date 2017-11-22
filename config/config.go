// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package config

import (
	"os"
	"strconv"
)

// Default config parameters values
const (
	DefaultBaseURL          = "http://localhost"
	DefaultDatabaseURL      = "postgres://postgres:postgres@localhost/miniflux2?sslmode=disable"
	DefaultWorkerPoolSize   = 5
	DefaultPollingFrequency = 60
	DefaultBatchSize        = 10
	DefaultDatabaseMaxConns = 20
	DefaultListenAddr       = "127.0.0.1:8080"
	DefaultCertFile         = ""
	DefaultKeyFile          = ""
	DefaultCertDomain       = ""
	DefaultCertCache        = "/tmp/cert_cache"
)

// Config manages configuration parameters.
type Config struct{}

// Get returns a config parameter value.
func (c *Config) Get(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

// GetInt returns a config parameter as integer.
func (c *Config) GetInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	v, _ := strconv.Atoi(value)
	return v
}

// NewConfig returns a new Config.
func NewConfig() *Config {
	return &Config{}
}
