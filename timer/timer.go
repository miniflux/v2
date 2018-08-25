// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

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
