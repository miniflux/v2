// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package cli

import (
	"fmt"
	"runtime"

	"github.com/miniflux/miniflux/version"
)

func info() {
	fmt.Println("Version:", version.Version)
	fmt.Println("Build Date:", version.BuildDate)
	fmt.Println("Go Version:", runtime.Version())
}
