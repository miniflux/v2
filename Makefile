APP          := miniflux
DOCKER_IMAGE := miniflux/miniflux
VERSION      := $(shell git describe --tags --abbrev=0)
COMMIT       := $(shell git rev-parse --short HEAD)
BUILD_DATE   := `date +%FT%T%z`
LD_FLAGS     := "-s -w -X 'miniflux.app/version.Version=$(VERSION)' -X 'miniflux.app/version.Commit=$(COMMIT)' -X 'miniflux.app/version.BuildDate=$(BUILD_DATE)'"
PKG_LIST     := $(shell go list ./... | grep -v /vendor/)
DB_URL       := postgres://postgres:postgres@localhost/miniflux_test?sslmode=disable
DEB_IMG_ARCH := amd64

export PGPASSWORD := postgres

.PHONY: \
	miniflux \
	linux-amd64 \
	linux-arm64 \
	linux-armv7 \
	linux-armv6 \
	linux-armv5 \
	linux-x86 \
	darwin-amd64 \
	darwin-arm64 \
	freebsd-amd64 \
	freebsd-x86 \
	openbsd-amd64 \
	openbsd-x86 \
	netbsd-x86 \
	netbsd-amd64 \
	windows-amd64 \
	windows-x86 \
	build \
	run \
	clean \
	test \
	lint \
	integration-test \
	clean-integration-test \
	docker-image \
	docker-image-distroless \
	docker-images \
	rpm \
	debian \
	debian-packages

miniflux:
	@ go build -buildmode=pie -ldflags=$(LD_FLAGS) -o $(APP) main.go

linux-amd64:
	@ GOOS=linux GOARCH=amd64 go build -ldflags=$(LD_FLAGS) -o $(APP)-linux-amd64 main.go

linux-arm64:
	@ GOOS=linux GOARCH=arm64 go build -ldflags=$(LD_FLAGS) -o $(APP)-linux-arm64 main.go

linux-armv7:
	@ GOOS=linux GOARCH=arm GOARM=7 go build -ldflags=$(LD_FLAGS) -o $(APP)-linux-armv7 main.go

linux-armv6:
	@ GOOS=linux GOARCH=arm GOARM=6 go build -ldflags=$(LD_FLAGS) -o $(APP)-linux-armv6 main.go

linux-armv5:
	@ GOOS=linux GOARCH=arm GOARM=5 go build -ldflags=$(LD_FLAGS) -o $(APP)-linux-armv5 main.go

darwin-amd64:
	@ GOOS=darwin GOARCH=amd64 go build -ldflags=$(LD_FLAGS) -o $(APP)-darwin-amd64 main.go

darwin-arm64:
	@ GOOS=darwin GOARCH=arm64 go build -ldflags=$(LD_FLAGS) -o $(APP)-darwin-arm64 main.go

freebsd-amd64:
	@ GOOS=freebsd GOARCH=amd64 go build -ldflags=$(LD_FLAGS) -o $(APP)-freebsd-amd64 main.go

openbsd-amd64:
	@ GOOS=openbsd GOARCH=amd64 go build -ldflags=$(LD_FLAGS) -o $(APP)-openbsd-amd64 main.go

windows-amd64:
	@ GOOS=windows GOARCH=amd64 go build -ldflags=$(LD_FLAGS) -o $(APP)-windows-amd64 main.go

build: linux-amd64 linux-arm64 linux-armv7 linux-armv6 linux-armv5 darwin-amd64 darwin-arm64 freebsd-amd64 openbsd-amd64 windows-amd64

# NOTE: unsupported targets
netbsd-amd64:
	@ GOOS=netbsd GOARCH=amd64 go build -ldflags=$(LD_FLAGS) -o $(APP)-netbsd-amd64 main.go

linux-x86:
	@ GOOS=linux GOARCH=386 go build -ldflags=$(LD_FLAGS) -o $(APP)-linux-x86 main.go

freebsd-x86:
	@ GOOS=freebsd GOARCH=386 go build -ldflags=$(LD_FLAGS) -o $(APP)-freebsd-x86 main.go

netbsd-x86:
	@ GOOS=netbsd GOARCH=386 go build -ldflags=$(LD_FLAGS) -o $(APP)-netbsd-x86 main.go

openbsd-x86:
	@ GOOS=openbsd GOARCH=386 go build -ldflags=$(LD_FLAGS) -o $(APP)-freebsd-x86 main.go

windows-x86:
	@ GOOS=windows GOARCH=386 go build -ldflags=$(LD_FLAGS) -o $(APP)-windows-x86 main.go

run:
	@ LOG_DATE_TIME=1 go run main.go -debug

clean:
	@ rm -f $(APP)-* $(APP) $(APP)*.rpm $(APP)*.deb

test:
	go test -cover -race -count=1 ./...

lint:
	golint -set_exit_status ${PKG_LIST}

integration-test:
	psql -U postgres -c 'drop database if exists miniflux_test;'
	psql -U postgres -c 'create database miniflux_test;'
	go build -o miniflux-test main.go

	DATABASE_URL=$(DB_URL) \
	ADMIN_USERNAME=admin \
	ADMIN_PASSWORD=test123 \
	CREATE_ADMIN=1 \
	RUN_MIGRATIONS=1 \
	DEBUG=1 \
	./miniflux-test >/tmp/miniflux.log 2>&1 & echo "$$!" > "/tmp/miniflux.pid"
	
	while ! nc -z localhost 8080; do sleep 1; done
	go test -v -tags=integration -count=1 miniflux.app/tests

clean-integration-test:
	@ kill -9 `cat /tmp/miniflux.pid`
	@ rm -f /tmp/miniflux.pid /tmp/miniflux.log
	@ rm miniflux-test
	@ psql -U postgres -c 'drop database if exists miniflux_test;'

docker-image:
	docker build -t $(DOCKER_IMAGE):$(VERSION) -f packaging/docker/alpine/Dockerfile .

docker-image-distroless:
	docker build -t $(DOCKER_IMAGE):$(VERSION) -f packaging/docker/distroless/Dockerfile .

docker-images:
	docker buildx build \
		--platform linux/amd64,linux/arm64,linux/arm/v7,linux/arm/v6 \
		--file packaging/docker/alpine/Dockerfile \
		--tag $(DOCKER_IMAGE):$(VERSION) \
		--push .

rpm: clean
	@ docker build \
		-t miniflux-rpm-builder \
		-f packaging/rpm/Dockerfile \
		.
	@ docker run --rm \
		-v ${PWD}:/root/rpmbuild/RPMS/x86_64 miniflux-rpm-builder \
		rpmbuild -bb --define "_miniflux_version $(VERSION)" /root/rpmbuild/SPECS/miniflux.spec

debian:
	@ docker build \
		--build-arg BASE_IMAGE_ARCH=$(DEB_IMG_ARCH) \
		-t $(DEB_IMG_ARCH)/miniflux-deb-builder \
		-f packaging/debian/Dockerfile \
		.
	@ docker run --rm \
		-v ${PWD}:/pkg $(DEB_IMG_ARCH)/miniflux-deb-builder

debian-packages: clean
	$(MAKE) debian DEB_IMG_ARCH=amd64
	$(MAKE) debian DEB_IMG_ARCH=arm64v8
	$(MAKE) debian DEB_IMG_ARCH=arm32v7
