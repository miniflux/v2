// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "miniflux.app/v2/internal/cli"

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/database"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/ui/static"
	"miniflux.app/v2/internal/version"
)

const (
	flagInfoHelp            = "Show build information"
	flagVersionHelp         = "Show application version"
	flagMigrateHelp         = "Run SQL migrations"
	flagFlushSessionsHelp   = "Flush all sessions (disconnect users)"
	flagCreateAdminHelp     = "Create admin user"
	flagResetPasswordHelp   = "Reset user password"
	flagResetFeedErrorsHelp = "Clear all feed errors for all users"
	flagDebugModeHelp       = "Show debug logs"
	flagConfigFileHelp      = "Load configuration file"
	flagConfigDumpHelp      = "Print parsed configuration values"
	flagHealthCheckHelp     = `Perform a health check on the given endpoint (the value "auto" try to guess the health check endpoint).`
	flagRefreshFeedsHelp    = "Refresh a batch of feeds and exit"
	flagRunCleanupTasksHelp = "Run cleanup tasks (delete old sessions and archives old entries)"
	flagExportUserFeedsHelp = "Export user feeds (provide the username as argument)"
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
		flagHealthCheck     string
		flagRefreshFeeds    bool
		flagRunCleanupTasks bool
		flagExportUserFeeds string
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
	flag.StringVar(&flagHealthCheck, "healthcheck", "", flagHealthCheckHelp)
	flag.BoolVar(&flagRefreshFeeds, "refresh-feeds", false, flagRefreshFeedsHelp)
	flag.BoolVar(&flagRunCleanupTasks, "run-cleanup-tasks", false, flagRunCleanupTasksHelp)
	flag.StringVar(&flagExportUserFeeds, "export-user-feeds", "", flagExportUserFeedsHelp)
	flag.Parse()

	cfg := config.NewParser()

	if flagConfigFile != "" {
		config.Opts, err = cfg.ParseFile(flagConfigFile)
		if err != nil {
			printErrorAndExit(err)
		}
	}

	config.Opts, err = cfg.ParseEnvironmentVariables()
	if err != nil {
		printErrorAndExit(err)
	}

	if flagConfigDump {
		fmt.Print(config.Opts)
		return
	}

	if flagDebugMode {
		config.Opts.SetLogLevel("debug")
	}

	logFile := config.Opts.LogFile()
	var logFileHandler io.Writer
	switch logFile {
	case "stdout":
		logFileHandler = os.Stdout
	case "stderr":
		logFileHandler = os.Stderr
	default:
		logFileHandler, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			printErrorAndExit(fmt.Errorf("unable to open log file: %v", err))
		}
		defer logFileHandler.(*os.File).Close()
	}

	if err := InitializeDefaultLogger(config.Opts.LogLevel(), logFileHandler, config.Opts.LogFormat(), config.Opts.LogDateTime()); err != nil {
		printErrorAndExit(err)
	}

	if flagHealthCheck != "" {
		doHealthCheck(flagHealthCheck)
		return
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
		slog.Info("The default value for DATABASE_URL is used")
	}

	if err := locale.LoadCatalogMessages(); err != nil {
		printErrorAndExit(fmt.Errorf("unable to load translations: %v", err))
	}

	if err := static.CalculateBinaryFileChecksums(); err != nil {
		printErrorAndExit(fmt.Errorf("unable to calculate binary file checksums: %v", err))
	}

	if err := static.GenerateStylesheetsBundles(); err != nil {
		printErrorAndExit(fmt.Errorf("unable to generate stylesheets bundles: %v", err))
	}

	if err := static.GenerateJavascriptBundles(); err != nil {
		printErrorAndExit(fmt.Errorf("unable to generate javascript bundles: %v", err))
	}

	db, err := database.NewConnectionPool(
		config.Opts.DatabaseURL(),
		config.Opts.DatabaseMinConns(),
		config.Opts.DatabaseMaxConns(),
		config.Opts.DatabaseConnectionLifetime(),
	)
	if err != nil {
		printErrorAndExit(fmt.Errorf("unable to connect to database: %v", err))
	}
	defer db.Close()

	store := storage.NewStorage(db)

	if err := store.Ping(); err != nil {
		printErrorAndExit(err)
	}

	if flagMigrate {
		if err := database.Migrate(db); err != nil {
			printErrorAndExit(err)
		}
		return
	}

	if flagResetFeedErrors {
		store.ResetFeedErrors()
		return
	}

	if flagExportUserFeeds != "" {
		exportUserFeeds(store, flagExportUserFeeds)
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

	// Run migrations and start the daemon.
	if config.Opts.RunMigrations() {
		if err := database.Migrate(db); err != nil {
			printErrorAndExit(err)
		}
	}

	if err := database.IsSchemaUpToDate(db); err != nil {
		printErrorAndExit(err)
	}

	// Create admin user and start the daemon.
	if config.Opts.CreateAdmin() {
		createAdmin(store)
	}

	if flagRefreshFeeds {
		refreshFeeds(store)
		return
	}

	if flagRunCleanupTasks {
		runCleanupTasks(store)
		return
	}

	startDaemon(store)
}

func printErrorAndExit(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}
