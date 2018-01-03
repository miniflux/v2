// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package cli

import (
	"fmt"
	"os"

	"github.com/miniflux/miniflux/storage"
)

func flushSessions(store *storage.Storage) {
	fmt.Println("Flushing all sessions (disconnect users)")
	if err := store.FlushAllSessions(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
