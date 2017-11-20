// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main

//go:generate go run generate.go

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/miniflux/miniflux2/config"
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/reader/feed"
	"github.com/miniflux/miniflux2/scheduler"
	"github.com/miniflux/miniflux2/server"
	"github.com/miniflux/miniflux2/storage"
	"github.com/miniflux/miniflux2/version"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/ssh/terminal"
)

func run(cfg *config.Config, store *storage.Storage) {
	log.Println("Starting Miniflux...")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	feedHandler := feed.NewFeedHandler(store)
	server := server.NewServer(cfg, store, feedHandler)

	go func() {
		pool := scheduler.NewWorkerPool(feedHandler, cfg.GetInt("WORKER_POOL_SIZE", 5))
		scheduler.NewScheduler(store, pool, cfg.GetInt("POLLING_FREQUENCY", 30), cfg.GetInt("BATCH_SIZE", 10))
	}()

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
		cfg.Get("DATABASE_URL", "postgres://postgres:postgres@localhost/miniflux2?sslmode=disable"),
		cfg.GetInt("DATABASE_MAX_CONNS", 20),
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
		user := &model.User{IsAdmin: true}
		user.Username, user.Password = askCredentials()
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
