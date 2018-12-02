APP := miniflux
VERSION := $(shell git rev-parse --short HEAD)
BUILD_DATE := `date +%FT%T%z`
LD_FLAGS := "-s -w -X 'miniflux.app/version.Version=$(VERSION)' -X 'miniflux.app/version.BuildDate=$(BUILD_DATE)'"
PKG_LIST := $(shell go list ./... | grep -v /vendor/)
DB_URL := postgres://postgres:postgres@localhost/miniflux_test?sslmode=disable

.PHONY: generate
.PHONY: linux-amd64
.PHONY: linux-armv8
.PHONY: linux-armv7
.PHONY: linux-armv6
.PHONY: linux-armv5
.PHONY: darwin-amd64
.PHONY: freebsd-amd64
.PHONY: openbsd-amd64
.PHONY: windows-amd64
.PHONY: build run clean test lint integration-test clean-integration-test

generate:
	@ go generate

linux-amd64: generate
	@ GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags=$(LD_FLAGS) -o $(APP)-linux-amd64 main.go

linux-armv8: generate
	@ GOOS=linux GOARCH=arm64 go build -mod=vendor -ldflags=$(LD_FLAGS) -o $(APP)-linux-armv8 main.go

linux-armv7: generate
	@ GOOS=linux GOARCH=arm GOARM=7 go build -mod=vendor -ldflags=$(LD_FLAGS) -o $(APP)-linux-armv7 main.go

linux-armv6: generate
	@ GOOS=linux GOARCH=arm GOARM=6 go build -mod=vendor -ldflags=$(LD_FLAGS) -o $(APP)-linux-armv6 main.go

linux-armv5: generate
	@ GOOS=linux GOARCH=arm GOARM=5 go build -mod=vendor -ldflags=$(LD_FLAGS) -o $(APP)-linux-armv5 main.go

darwin-amd64: generate
	@ GOOS=darwin GOARCH=amd64 go build -mod=vendor -ldflags=$(LD_FLAGS) -o $(APP)-darwin-amd64 main.go

freebsd-amd64: generate
	@ GOOS=freebsd GOARCH=amd64 go build -mod=vendor -ldflags=$(LD_FLAGS) -o $(APP)-freebsd-amd64 main.go

openbsd-amd64: generate
	@ GOOS=openbsd GOARCH=amd64 go build -mod=vendor -ldflags=$(LD_FLAGS) -o $(APP)-openbsd-amd64 main.go

windows-amd64: generate
	@ GOOS=windows GOARCH=amd64 go build -mod=vendor -ldflags=$(LD_FLAGS) -o $(APP)-windows-amd64 main.go

build: linux-amd64 linux-armv8 linux-armv7 linux-armv6 linux-armv5 darwin-amd64 freebsd-amd64 openbsd-amd64 windows-amd64

run: generate
	@ go run main.go -debug

clean:
	@ rm -f $(APP)-*

test:
	go test -cover -race -count=1 ./...

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
	go test -v -tags=integration -count=1 miniflux.app/tests || cat /tmp/miniflux.log

clean-integration-test:
	@ kill -9 `cat /tmp/miniflux.pid`
	@ rm -f /tmp/miniflux.pid /tmp/miniflux.log
	@ rm miniflux-test
	@ psql -U postgres -c 'drop database if exists miniflux_test;'
