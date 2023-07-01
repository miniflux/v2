// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package timer // import "miniflux.app/timer"

import (
	"time"

	"miniflux.app/logger"
)

// ExecutionTime returns the elapsed time of a block of code.
func ExecutionTime(start time.Time, name string) {
	elapsed := time.Since(start)
	logger.Debug("%s took %s", name, elapsed)
}
