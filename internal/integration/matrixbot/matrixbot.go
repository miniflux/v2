// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package matrixbot // import "miniflux.app/v2/internal/integration/matrixbot"

import (
	"fmt"
	"strings"

	"miniflux.app/v2/internal/model"
)

// PushEntry pushes entries to matrix chat using integration settings provided
func PushEntries(feed *model.Feed, entries model.Entries, matrixBaseURL, matrixUsername, matrixPassword, matrixRoomID string) error {
	client := NewClient(matrixBaseURL)
	discovery, err := client.DiscoverEndpoints()
	if err != nil {
		return err
	}

	loginResponse, err := client.Login(discovery.HomeServerInformation.BaseURL, matrixUsername, matrixPassword)
	if err != nil {
		return err
	}

	var textMessages []string
	var formattedTextMessages []string

	for _, entry := range entries {
		textMessages = append(textMessages, fmt.Sprintf(`[%s] %s - %s`, feed.Title, entry.Title, entry.URL))
		formattedTextMessages = append(formattedTextMessages, fmt.Sprintf(`<li><strong>%s</strong>: <a href=%q>%s</a></li>`, feed.Title, entry.URL, entry.Title))
	}

	_, err = client.SendFormattedTextMessage(
		discovery.HomeServerInformation.BaseURL,
		loginResponse.AccessToken,
		matrixRoomID,
		strings.Join(textMessages, "\n"),
		"<ul>"+strings.Join(formattedTextMessages, "\n")+"</ul>",
	)

	return err
}
