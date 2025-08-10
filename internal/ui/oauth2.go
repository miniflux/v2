// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "influxeed-engine/v2/internal/ui"

import (
	"context"

	"influxeed-engine/v2/internal/config"
	"influxeed-engine/v2/internal/oauth2"
)

func getOAuth2Manager(ctx context.Context) *oauth2.Manager {
	return oauth2.NewManager(
		ctx,
		config.Opts.OAuth2ClientID(),
		config.Opts.OAuth2ClientSecret(),
		config.Opts.OAuth2RedirectURL(),
		config.Opts.OIDCDiscoveryEndpoint(),
	)
}
