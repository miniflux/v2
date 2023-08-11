// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package timer // import "miniflux.app/v2/internal/timer"

import (
	"time"

	"miniflux.app/v2/internal/logger"
)

// ExecutionTime returns the elapsed time of a block of code.
func ExecutionTime(start time.Time, name string) {
	elapsed := time.Since(start)
	logger.Debug("%s took %s", name, elapsed)
}
