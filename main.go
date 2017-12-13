// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main

//go:generate go run generate.go
//go:generate gofmt -s -w sql/sql.go
//go:generate gofmt -s -w server/static/css.go
//go:generate gofmt -s -w server/static/bin.go
//go:generate gofmt -s -w server/static/js.go
//go:generate gofmt -s -w server/template/views.go
//go:generate gofmt -s -w server/template/common.go
//go:generate gofmt -s -w locale/translations.go

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/miniflux/miniflux/config"
	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/reader/feed"
	"github.com/miniflux/miniflux/scheduler"
	"github.com/miniflux/miniflux/server"
	"github.com/miniflux/miniflux/storage"
	"github.com/miniflux/miniflux/version"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/ssh/terminal"
)

func run(cfg *config.Config, store *storage.Storage) {
	log.Println("Starting Miniflux...")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	feedHandler := feed.NewFeedHandler(store)
	pool := scheduler.NewWorkerPool(feedHandler, cfg.GetInt("WORKER_POOL_SIZE", config.DefaultWorkerPoolSize))
	server := server.NewServer(cfg, store, pool, feedHandler)

	scheduler.NewScheduler(
		store,
		pool,
		cfg.GetInt("POLLING_FREQUENCY", config.DefaultPollingFrequency),
		cfg.GetInt("BATCH_SIZE", config.DefaultBatchSize),
	)

	<-stop
	log.Println("Shutting down the server...")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	server.Shutdown(ctx)
	store.Close()
	log.Println("Server gracefully stopped")
}

func askCredentials() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, _ := reader.ReadString('\n')

	fmt.Print("Enter Password: ")
	bytePassword, _ := terminal.ReadPassword(0)

	fmt.Printf("\n")
	return strings.TrimSpace(username), strings.TrimSpace(string(bytePassword))
}

func main() {
	flagInfo := flag.Bool("info", false, "Show application information")
	flagVersion := flag.Bool("version", false, "Show application version")
	flagMigrate := flag.Bool("migrate", false, "Migrate database schema")
	flagFlushSessions := flag.Bool("flush-sessions", false, "Flush all sessions (disconnect users)")
	flagCreateAdmin := flag.Bool("create-admin", false, "Create admin user")
	flag.Parse()

	cfg := config.NewConfig()
	store := storage.NewStorage(
		cfg.Get("DATABASE_URL", config.DefaultDatabaseURL),
		cfg.GetInt("DATABASE_MAX_CONNS", config.DefaultDatabaseMaxConns),
	)

	if *flagInfo {
		fmt.Println("Version:", version.Version)
		fmt.Println("Build Date:", version.BuildDate)
		fmt.Println("Go Version:", runtime.Version())
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
		fmt.Println("Flushing all sessions (disconnect users)")
		if err := store.FlushAllSessions(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	if *flagCreateAdmin {
		user := &model.User{
			Username: os.Getenv("ADMIN_USERNAME"),
			Password: os.Getenv("ADMIN_PASSWORD"),
			IsAdmin:  true,
		}

		if user.Username == "" || user.Password == "" {
			user.Username, user.Password = askCredentials()
		}

		if err := user.ValidateUserCreation(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := store.CreateUser(user); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		return
	}

	run(cfg, store)
}
