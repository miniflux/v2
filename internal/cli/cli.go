// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "miniflux.app/v2/internal/cli"

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/database"
	"miniflux.app/v2/internal/proxyrotator"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/ui/static"
	"miniflux.app/v2/internal/version"
)

const (
	flagInfoHelp             = "Show build information"
	flagVersionHelp          = "Show application version"
	flagMigrateHelp          = "Run SQL migrations"
	flagFlushSessionsHelp    = "Flush all sessions (disconnect users)"
	flagCreateAdminHelp      = "Create an admin user from an interactive terminal"
	flagResetPasswordHelp    = "Reset user password"
	flagResetFeedErrorsHelp  = "Clear all feed errors for all users"
	flagDebugModeHelp        = "Show debug logs"
	flagConfigFileHelp       = "Load configuration file"
	flagConfigDumpHelp       = "Print parsed configuration values"
	flagHealthCheckHelp      = `Perform a health check on the given endpoint (the value "auto" try to guess the health check endpoint).`
	flagRefreshFeedsHelp     = "Refresh a batch of feeds and exit"
	flagRunCleanupTasksHelp  = "Run cleanup tasks (delete old sessions and archives old entries)"
	flagExportUserFeedsHelp  = "Export user feeds (provide the username as argument)"
	flagResetNextCheckAtHelp = "Reset the next check time for all feeds"
)

// Parse parses command line arguments.
func Parse() {
	var (
		err                      error
		flagInfo                 bool
		flagVersion              bool
		flagMigrate              bool
		flagFlushSessions        bool
		flagCreateAdmin          bool
		flagResetPassword        bool
		flagResetFeedErrors      bool
		flagResetFeedNextCheckAt bool
		flagDebugMode            bool
		flagConfigFile           string
		flagConfigDump           bool
		flagHealthCheck          string
		flagRefreshFeeds         bool
		flagRunCleanupTasks      bool
		flagExportUserFeeds      string
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
	flag.BoolVar(&flagResetFeedNextCheckAt, "reset-feed-next-check-at", false, flagResetNextCheckAtHelp)
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

	if oauth2Provider := config.Opts.OAuth2Provider(); oauth2Provider != "" {
		if oauth2Provider != "oidc" && oauth2Provider != "google" {
			printErrorAndExit(fmt.Errorf(`unsupported OAuth2 provider: %q (Possible values are "google" or "oidc")`, oauth2Provider))
		}
	}

	if config.Opts.DisableLocalAuth() {
		switch {
		case config.Opts.OAuth2Provider() == "" && config.Opts.AuthProxyHeader() == "":
			printErrorAndExit(errors.New("DISABLE_LOCAL_AUTH is enabled but neither OAUTH2_PROVIDER nor AUTH_PROXY_HEADER is not set. Please enable at least one authentication source"))
		case config.Opts.OAuth2Provider() != "" && !config.Opts.IsOAuth2UserCreationAllowed():
			printErrorAndExit(errors.New("DISABLE_LOCAL_AUTH is enabled and an OAUTH2_PROVIDER is configured, but OAUTH2_USER_CREATION is not enabled"))
		case config.Opts.AuthProxyHeader() != "" && !config.Opts.IsAuthProxyUserCreationAllowed():
			printErrorAndExit(errors.New("DISABLE_LOCAL_AUTH is enabled and an AUTH_PROXY_HEADER is configured, but AUTH_PROXY_USER_CREATION is not enabled"))
		}
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
		if err := store.ResetFeedErrors(); err != nil {
			printErrorAndExit(err)
		}
		return
	}

	if flagResetFeedNextCheckAt {
		if err := store.ResetNextCheckAt(); err != nil {
			printErrorAndExit(err)
		}
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
		createAdminUserFromInteractiveTerminal(store)
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

	if config.Opts.CreateAdmin() {
		createAdminUserFromEnvironmentVariables(store)
	}

	if config.Opts.HasHTTPClientProxiesConfigured() {
		slog.Info("Initializing proxy rotation", slog.Int("proxies_count", len(config.Opts.HTTPClientProxies())))
		proxyrotator.ProxyRotatorInstance, err = proxyrotator.NewProxyRotator(config.Opts.HTTPClientProxies())
		if err != nil {
			printErrorAndExit(fmt.Errorf("unable to initialize proxy rotator: %v", err))
		}
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
