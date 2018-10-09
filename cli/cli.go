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

const (
	flagInfoHelp = "Show application information"
	flagVersionHelp = "Show application version"
	flagMigrateHelp = "Run SQL migrations"
	flagFlsuhSessionsHelp = "Flush all sessions (disconnect users)"
	flagCreateAdminHelp = "Create admin user"
	flagResetPasswordHelp = "Reset user password"
	flagResetFeedErrorsHelp = "Clear all feed errors for all users"
	flagDebugModeHelp = "Show debug logs"
)

// Parse parses command line arguments.
func Parse() {
	var (
		flagInfo bool
		flagVersion bool
		flagMigrate bool
		flagFlushSessions bool
		flagCreateAdmin bool
		flagResetPassword bool
		flagResetFeedErrors bool
		flagDebugMode bool
	)

	flag.BoolVar(&flagInfo, "info", false, flagInfoHelp)
	flag.BoolVar(&flagInfo, "i", false, flagInfoHelp)
	flag.BoolVar(&flagVersion, "version", false, flagVersionHelp)
	flag.BoolVar(&flagVersion, "v", false, flagVersionHelp)
	flag.BoolVar(&flagMigrate, "migrate", false, flagMigrateHelp)
	flag.BoolVar(&flagFlushSessions, "flush-sessions", false, flagFlsuhSessionsHelp)
	flag.BoolVar(&flagCreateAdmin, "create-admin", false, flagCreateAdminHelp)
	flag.BoolVar(&flagResetPassword, "reset-password", false, flagResetPasswordHelp)
	flag.BoolVar(&flagResetFeedErrors, "reset-feed-errors", false, flagResetFeedErrorsHelp)
	flag.BoolVar(&flagDebugMode,"debug", false, flagDebugModeHelp)
	flag.Parse()

	cfg := config.NewConfig()

	if flagDebugMode || cfg.HasDebugMode() {
		logger.EnableDebug()
	}

	db, err := database.NewConnectionPool(cfg.DatabaseURL(), cfg.DatabaseMinConns(), cfg.DatabaseMaxConns())
	if err != nil {
		logger.Fatal("Unable to connect to the database: %v", err)
	}
	defer db.Close()

	store := storage.NewStorage(db)

	if flagInfo {
		info()
		return
	}

	if flagVersion {
		fmt.Println(version.Version)
		return
	}

	if flagMigrate {
		database.Migrate(db)
		return
	}

	if flagResetFeedErrors {
		store.ResetFeedErrors()
		return
	}

	if flagFlushSessions {
		flushSessions(store)
		return
	}

	if flagCreateAdmin {
		createAdmin(store)
		return
	}

	if flagResetPassword {
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
