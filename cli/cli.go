// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package cli // import "miniflux.app/cli"

import (
	"flag"
	"fmt"

	"miniflux.app/config"
	"miniflux.app/daemon"
	"miniflux.app/database"
	"miniflux.app/logger"
	"miniflux.app/storage"
	"miniflux.app/version"
)

// Parse parses command line arguments.
func Parse() {
	flagInfo := flag.Bool("info", false, "Show application information")
	flagVersion := flag.Bool("version", false, "Show application version")
	flagMigrate := flag.Bool("migrate", false, "Migrate database schema")
	flagFlushSessions := flag.Bool("flush-sessions", false, "Flush all sessions (disconnect users)")
	flagCreateAdmin := flag.Bool("create-admin", false, "Create admin user")
	flagResetPassword := flag.Bool("reset-password", false, "Reset user password")
	flagResetFeedErrors := flag.Bool("reset-feed-errors", false, "Clear all feed errors for all users")
	flagDebugMode := flag.Bool("debug", false, "Enable debug mode (more verbose output)")
	flag.Parse()

	cfg := config.NewConfig()

	if *flagDebugMode || cfg.HasDebugMode() {
		logger.EnableDebug()
	}

	db, err := database.NewConnectionPool(cfg.DatabaseURL(), cfg.DatabaseMinConns(), cfg.DatabaseMaxConns())
	if err != nil {
		logger.Fatal("Unable to connect to the database: %v", err)
	}
	defer db.Close()

	store := storage.NewStorage(db)

	if *flagInfo {
		info()
		return
	}

	if *flagVersion {
		fmt.Println(version.Version)
		return
	}

	if *flagMigrate {
		database.Migrate(db)
		return
	}

	if *flagResetFeedErrors {
		store.ResetFeedErrors()
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

	// Run migrations and start the deamon.
	if cfg.RunMigrations() {
		database.Migrate(db)
	}

	// Create admin user and start the deamon.
	if cfg.CreateAdmin() {
		createAdmin(store)
	}

	daemon.Run(cfg, store)
}
