// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package config

import (
	"os"
	"strconv"
)

const (
	DefaultBaseURL = "http://localhost"
)

type Config struct {
}

func (c *Config) Get(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func (c *Config) GetInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	v, _ := strconv.Atoi(value)
	return v
}

func NewConfig() *Config {
	return &Config{}
}
