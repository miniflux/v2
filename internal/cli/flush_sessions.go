// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "influxeed-engine/v2/internal/cli"

import (
	"fmt"

	"influxeed-engine/v2/internal/storage"
)

func flushSessions(store *storage.Storage) {
	fmt.Println("Flushing all sessions (disconnect users)")
	if err := store.FlushAllSessions(); err != nil {
		printErrorAndExit(err)
	}
}
