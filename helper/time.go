// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package helper

import (
	"log"
	"time"
)

// ExecutionTime returns the elapsed time of a block of code.
func ExecutionTime(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
