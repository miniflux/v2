APP = miniflux
VERSION = $(shell git rev-parse --short HEAD)
BUILD_DATE = `date +%FT%T%z`

.PHONY: build-linux build-darwin build run clean test

build-linux:
	@ go generate
	@ GOOS=linux GOARCH=amd64 go build -ldflags="-X 'miniflux/version.Version=$(VERSION)' -X 'miniflux/version.BuildDate=$(BUILD_DATE)'" -o $(APP)-linux-amd64 main.go

build-darwin:
	@ go generate
	@ GOOS=darwin GOARCH=amd64 go build -ldflags="-X 'miniflux/version.Version=$(VERSION)' -X 'miniflux/version.BuildDate=$(BUILD_DATE)'" -o $(APP)-darwin-amd64 main.go

build: build-linux build-darwin

run:
	@ go generate
	@ go run main.go

clean:
	@ rm -f $(APP)-*

test:
	go test -cover -race ./...
