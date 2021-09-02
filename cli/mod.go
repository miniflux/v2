package cli

import (
	"flag"
	"fmt"

	"miniflux.app/config"
	"miniflux.app/database"
	"miniflux.app/locale"
	"miniflux.app/logger"
	"miniflux.app/storage"
	"miniflux.app/ui/static"
	"miniflux.app/constant"
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
	flagHealthCheckHelp     = `Perform a health check on the given endpoint (the value "auto" try to guess the health check endpoint).`
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
	)

	flag.BoolVar(&flagConfigDump, "config-dump", false, flagConfigDumpHelp)
	flag.StringVar(&flagConfigFile, "config-file", "", flagConfigFileHelp)
	flag.BoolVar(&flagCreateAdmin, "create-admin", false, flagCreateAdminHelp)
	flag.BoolVar(&flagDebugMode, "debug", false, flagDebugModeHelp)
	flag.BoolVar(&flagFlushSessions, "flush-sessions", false, flagFlushSessionsHelp)
	flag.StringVar(&flagHealthCheck, "healthcheck", "", flagHealthCheckHelp)
	flag.BoolVar(&flagInfo, "info", false, flagInfoHelp)
	flag.BoolVar(&flagMigrate, "migrate", false, flagMigrateHelp)
	flag.BoolVar(&flagResetFeedErrors, "reset-feed-errors", false, flagResetFeedErrorsHelp)
	flag.BoolVar(&flagResetPassword, "reset-password", false, flagResetPasswordHelp)
	flag.BoolVar(&flagVersion, "version", false, flagVersionHelp)
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

	if flagHealthCheck != "" {
		doHealthCheck(flagHealthCheck)
		return
	}

	if flagInfo {
		info()
		return
	}

	if flagVersion {
		fmt.Println(constant.Version)
		return
	}

	if config.Opts.IsDefaultDatabaseURL() {
		logger.Info("The default value for DATABASE_URL is used")
	}

	logger.Debug("Loading translations...")
	if err := locale.LoadCatalogMessages(); err != nil {
		logger.Fatal("Unable to load translations: %v", err)
	}

	logger.Debug("Loading static assets...")
	if err := static.CalculateBinaryFileChecksums(); err != nil {
		logger.Fatal("Unable to calculate binary files checksum: %v", err)
	}

	if err := static.GenerateStylesheetsBundles(); err != nil {
		logger.Fatal("Unable to generate Stylesheet bundles: %v", err)
	}

	if err := static.GenerateJavascriptBundles(); err != nil {
		logger.Fatal("Unable to generate Javascript bundles: %v", err)
	}

	db, err := database.NewConnectionPool(
		config.Opts.DatabaseURL(),
		config.Opts.DatabaseMinConns(),
		config.Opts.DatabaseMaxConns(),
		config.Opts.DatabaseConnectionLifetime(),
	)
	if err != nil {
		logger.Fatal("Unable to initialize database connection pool: %v", err)
	}
	defer db.Close()

	store := storage.NewStorage(db)

	if err := store.Ping(); err != nil {
		logger.Fatal("Unable to connect to the database: %v", err)
	}

	if flagMigrate {
		if err := database.Migrate(db); err != nil {
			logger.Fatal(`%v`, err)
		}
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

	// Run migrations and start the daemon.
	if config.Opts.RunMigrations() {
		if err := database.Migrate(db); err != nil {
			logger.Fatal(`%v`, err)
		}
	}

	if err := database.IsSchemaUpToDate(db); err != nil {
		logger.Fatal(`You must run the SQL migrations, %v`, err)
	}

	// Create admin user and start the daemon.
	if config.Opts.CreateAdmin() {
		createAdmin(store)
	}

	startDaemon(store)
}
