// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package cli // import "miniflux.app/cli"

import (
	"flag"
	"fmt"

	"miniflux.app/config"
	"miniflux.app/database"
	"miniflux.app/logger"
	"miniflux.app/storage"
	"miniflux.app/version"
)

const (
	flagInfoHelp            = "Show application information"
	flagVersionHelp         = "Show application version"
	flagMigrateHelp         = "Run SQL migrations"
	flagFlushSessionsHelp   = "Flush all sessions (disconnect users)"
	flagCreateAdminHelp     = "Create admin user"
	flagResetPasswordHelp   = "Reset user password"
	flagResetFeedErrorsHelp = "Clear all feed errors for all users"
	flagDebugModeHelp       = "Show debug logs"
	flagConfigFileHelp      = "Load configuration file"
	flagConfigDumpHelp      = "Print parsed configuration values"
)

// Parse parses command line arguments.
func Parse() {
	var (
		err                 error
		flagInfo            bool
		flagVersion         bool
		flagMigrate         bool
		flagFlushSessions   bool
		flagCreateAdmin     bool
		flagResetPassword   bool
		flagResetFeedErrors bool
		flagDebugMode       bool
		flagConfigFile      string
		flagConfigDump      bool
	)

	flag.BoolVar(&flagInfo, "info", false, flagInfoHelp)
	flag.BoolVar(&flagInfo, "i", false, flagInfoHelp)
	flag.BoolVar(&flagVersion, "version", false, flagVersionHelp)
	flag.BoolVar(&flagVersion, "v", false, flagVersionHelp)
	flag.BoolVar(&flagMigrate, "migrate", false, flagMigrateHelp)
	flag.BoolVar(&flagFlushSessions, "flush-sessions", false, flagFlushSessionsHelp)
	flag.BoolVar(&flagCreateAdmin, "create-admin", false, flagCreateAdminHelp)
	flag.BoolVar(&flagResetPassword, "reset-password", false, flagResetPasswordHelp)
	flag.BoolVar(&flagResetFeedErrors, "reset-feed-errors", false, flagResetFeedErrorsHelp)
	flag.BoolVar(&flagDebugMode, "debug", false, flagDebugModeHelp)
	flag.StringVar(&flagConfigFile, "config-file", "", flagConfigFileHelp)
	flag.StringVar(&flagConfigFile, "c", "", flagConfigFileHelp)
	flag.BoolVar(&flagConfigDump, "config-dump", false, flagConfigDumpHelp)
	flag.Parse()

	cfg := config.NewParser()

	if flagConfigFile != "" {
		config.Opts, err = cfg.ParseFile(flagConfigFile)
		if err != nil {
			logger.Fatal("%v", err)
		}
	}

	config.Opts, err = cfg.ParseEnvironmentVariables()
	if err != nil {
		logger.Fatal("%v", err)
	}

	if flagConfigDump {
		fmt.Print(config.Opts)
		return
	}

	if config.Opts.LogDateTime() {
		logger.EnableDateTime()
	}

	if flagDebugMode || config.Opts.HasDebugMode() {
		logger.EnableDebug()
	}

	if flagInfo {
		info()
		return
	}

	if flagVersion {
		fmt.Println(version.Version)
		return
	}

	if config.Opts.IsDefaultDatabaseURL() {
		logger.Info("The default value for DATABASE_URL is used")
	}

	db, err := database.NewConnectionPool(
		config.Opts.DatabaseURL(),
		config.Opts.DatabaseMinConns(),
		config.Opts.DatabaseMaxConns(),
	)
	if err != nil {
		logger.Fatal("Unable to connect to the database: %v", err)
	}
	defer db.Close()

	if flagMigrate {
		database.Migrate(db)
		return
	}

	store := storage.NewStorage(db)

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
	if config.Opts.RunMigrations() {
		database.Migrate(db)
	}

	if err := database.IsSchemaUpToDate(db); err != nil {
		logger.Fatal(`You must run the SQL migrations, %v`, err)
	}

	// Create admin user and start the deamon.
	if config.Opts.CreateAdmin() {
		createAdmin(store)
	}

	startDaemon(store)
}
