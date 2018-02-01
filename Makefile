APP := miniflux
VERSION=$(shell git rev-parse --short HEAD)
BUILD_DATE=`date +%FT%T%z`
PKG_LIST := $(shell go list ./... | grep -v /vendor/)
DB_URL := postgres://postgres:postgres@localhost/miniflux_test?sslmode=disable

.PHONY: linux linux-arm darwin build run clean test lint integration-test clean-integration-test

linux:
	@ go generate
	@ GOOS=linux GOARCH=amd64 go build -ldflags="-X 'github.com/miniflux/miniflux/version.Version=$(VERSION)' -X 'github.com/miniflux/miniflux/version.BuildDate=$(BUILD_DATE)'" -o $(APP)-linux-amd64 main.go

linux-arm:
	@ go generate
	@ GOOS=linux GOARCH=arm64 go build -ldflags="-X 'github.com/miniflux/miniflux/version.Version=$(VERSION)' -X 'github.com/miniflux/miniflux/version.BuildDate=$(BUILD_DATE)'" -o $(APP)-linux-armv8 main.go
	@ GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-X 'github.com/miniflux/miniflux/version.Version=$(VERSION)' -X 'github.com/miniflux/miniflux/version.BuildDate=$(BUILD_DATE)'" -o $(APP)-linux-armv7 main.go
	@ GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="-X 'github.com/miniflux/miniflux/version.Version=$(VERSION)' -X 'github.com/miniflux/miniflux/version.BuildDate=$(BUILD_DATE)'" -o $(APP)-linux-armv6 main.go
	@ GOOS=linux GOARCH=arm GOARM=5 go build -ldflags="-X 'github.com/miniflux/miniflux/version.Version=$(VERSION)' -X 'github.com/miniflux/miniflux/version.BuildDate=$(BUILD_DATE)'" -o $(APP)-linux-armv5 main.go

darwin:
	@ go generate
	@ GOOS=darwin GOARCH=amd64 go build -ldflags="-X 'github.com/miniflux/miniflux/version.Version=$(VERSION)' -X 'github.com/miniflux/miniflux/version.BuildDate=$(BUILD_DATE)'" -o $(APP)-darwin-amd64 main.go

build: linux linux-arm darwin

run:
	@ go generate
	@ go run main.go -debug

clean:
	@ rm -f $(APP)-*

test:
	go test -cover -race ./...

lint:
	@ golint -set_exit_status ${PKG_LIST}

integration-test:
	psql -U postgres -c 'drop database if exists miniflux_test;'
	psql -U postgres -c 'create database miniflux_test;'
	DATABASE_URL=$(DB_URL) go run main.go -migrate
	DATABASE_URL=$(DB_URL) ADMIN_USERNAME=admin ADMIN_PASSWORD=test123 go run main.go -create-admin
	go build -o miniflux-test main.go
	DATABASE_URL=$(DB_URL) ./miniflux-test -debug >/tmp/miniflux.log 2>&1 & echo "$$!" > "/tmp/miniflux.pid"
	while ! echo exit | nc localhost 8080; do sleep 1; done >/dev/null
	go test -v -tags=integration || cat /tmp/miniflux.log

clean-integration-test:
	@ kill -9 `cat /tmp/miniflux.pid`
	@ rm -f /tmp/miniflux.pid /tmp/miniflux.log
	@ rm miniflux-test
	@ psql -U postgres -c 'drop database if exists miniflux_test;'
