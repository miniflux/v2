// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "miniflux.app/cli"

import (
	"net/http"
	"time"

	"miniflux.app/config"
	"miniflux.app/logger"
)

func doHealthCheck(healthCheckEndpoint string) {
	if healthCheckEndpoint == "auto" {
		healthCheckEndpoint = "http://" + config.Opts.ListenAddr() + config.Opts.BasePath() + "/healthcheck"
	}

	logger.Debug(`Executing health check on %s`, healthCheckEndpoint)

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(healthCheckEndpoint)
	if err != nil {
		logger.Fatal(`Health check failure: %v`, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logger.Fatal(`Health check failed with status code %d`, resp.StatusCode)
	}

	logger.Debug(`Health check is OK`)
}
