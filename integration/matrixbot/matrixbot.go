// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package matrixbot // import "miniflux.app/integration/matrixbot"

import (
	"fmt"

	"miniflux.app/logger"
	"miniflux.app/model"

	"github.com/matrix-org/gomatrix"
)

// PushEntry pushes entries to matrix chat using integration settings provided
func PushEntries(entries model.Entries, serverURL, botLogin, botPassword, chatID string) error {
	bot, err := gomatrix.NewClient(serverURL, "", "")
	if err != nil {
		return fmt.Errorf("matrixbot: bot creation failed: %w", err)
	}

	resp, err := bot.Login(&gomatrix.ReqLogin{
		Type:     "m.login.password",
		User:     botLogin,
		Password: botPassword,
	})

	if err != nil {
		logger.Debug("matrixbot: login failed: %w", err)
		return fmt.Errorf("matrixbot: login failed, please check your credentials or turn on debug mode")
	}

	bot.SetCredentials(resp.UserID, resp.AccessToken)
	defer func() {
		bot.Logout()
		bot.ClearCredentials()
	}()

	message := ""
	for _, entry := range entries {
		message = message + entry.Title + " " + entry.URL + "\n"
	}

	if _, err = bot.SendText(chatID, message); err != nil {
		logger.Debug("matrixbot: sending message failed: %w", err)
		return fmt.Errorf("matrixbot: sending message failed, turn on debug mode for more informations")
	}

	return nil
}
