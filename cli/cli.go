// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package cli

import (
	"flag"
	"fmt"

	"github.com/miniflux/miniflux/logger"

	"github.com/miniflux/miniflux/config"
	"github.com/miniflux/miniflux/daemon"
	"github.com/miniflux/miniflux/storage"
	"github.com/miniflux/miniflux/version"
)

// Parse parses command line arguments.
func Parse() {
	flagInfo := flag.Bool("info", false, "Show application information")
	flagVersion := flag.Bool("version", false, "Show application version")
	flagMigrate := flag.Bool("migrate", false, "Migrate database schema")
	flagFlushSessions := flag.Bool("flush-sessions", false, "Flush all sessions (disconnect users)")
	flagCreateAdmin := flag.Bool("create-admin", false, "Create admin user")
	flagResetPassword := flag.Bool("reset-password", false, "Reset user password")
	flagDebugMode := flag.Bool("debug", false, "Enable debug mode (more verbose output)")
	flag.Parse()

	cfg := config.NewConfig()
	store := storage.NewStorage(
		cfg.DatabaseURL(),
		cfg.DatabaseMaxConnections(),
	)

	if *flagInfo {
		info()
		return
	}

	if *flagVersion {
		fmt.Println(version.Version)
		return
	}

	if *flagMigrate {
		store.Migrate()
		return
	}

	if *flagFlushSessions {
		flushSessions(store)
		return
	}

	if *flagCreateAdmin {
		createAdmin(store)
		return
	}

	if *flagResetPassword {
		resetPassword(store)
		return
	}

	if *flagDebugMode {
		logger.EnableDebug()
	}

	// start daemon
	daemon.Run(cfg, store)
}
