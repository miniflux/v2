APP := miniflux
DOCKER_IMAGE := miniflux/miniflux
VERSION := $(shell git rev-parse --short HEAD)
BUILD_DATE := `date +%FT%T%z`
LD_FLAGS := "-s -w -X 'miniflux.app/version.Version=$(VERSION)' -X 'miniflux.app/version.BuildDate=$(BUILD_DATE)'"
PKG_LIST := $(shell go list ./... | grep -v /vendor/)
DB_URL := postgres://postgres:postgres@localhost/miniflux_test?sslmode=disable

export GO111MODULE=on

.PHONY: generate \
	miniflux \
	linux-amd64 \
	linux-armv8 \
	linux-armv7 \
	linux-armv6 \
	linux-armv5 \
	linux-x86 \
	darwin-amd64 \
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
	docker-images \
	docker-manifest

generate:
	@ go generate -mod=vendor

miniflux: generate
	@ go build -mod=vendor -ldflags=$(LD_FLAGS) -o $(APP) main.go

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

# NOTE: unsupported targets
netbsd-amd64: generate
	@ GOOS=netbsd GOARCH=amd64 go build -mod=vendor -ldflags=$(LD_FLAGS) -o $(APP)-netbsd-amd64 main.go

linux-x86: generate
	@ GOOS=linux GOARCH=386 go build -mod=vendor -ldflags=$(LD_FLAGS) -o $(APP)-linux-x86 main.go

freebsd-x86: generate
	@ GOOS=freebsd GOARCH=386 go build -mod=vendor -ldflags=$(LD_FLAGS) -o $(APP)-freebsd-x86 main.go

netbsd-x86: generate
	@ GOOS=netbsd GOARCH=386 go build -mod=vendor -ldflags=$(LD_FLAGS) -o $(APP)-netbsd-x86 main.go

openbsd-x86: generate
	@ GOOS=openbsd GOARCH=386 go build -mod=vendor -ldflags=$(LD_FLAGS) -o $(APP)-freebsd-x86 main.go

windows-x86: generate
	@ GOOS=windows GOARCH=386 go build -mod=vendor -ldflags=$(LD_FLAGS) -o $(APP)-windows-x86 main.go

run: generate
	@ go run -mod=vendor main.go -debug

clean:
	@ rm -f $(APP)-* $(APP)

test:
	go test -mod=vendor -cover -race -count=1 ./...

lint:
	@ golint -set_exit_status ${PKG_LIST}

integration-test:
	psql -U postgres -c 'drop database if exists miniflux_test;'
	psql -U postgres -c 'create database miniflux_test;'
	DATABASE_URL=$(DB_URL) go run -mod=vendor main.go -migrate
	DATABASE_URL=$(DB_URL) ADMIN_USERNAME=admin ADMIN_PASSWORD=test123 go run -mod=vendor main.go -create-admin
	go build -mod=vendor -o miniflux-test main.go
	DATABASE_URL=$(DB_URL) ./miniflux-test -debug >/tmp/miniflux.log 2>&1 & echo "$$!" > "/tmp/miniflux.pid"
	while ! echo exit | nc localhost 8080; do sleep 1; done >/dev/null
	go test -mod=vendor -v -tags=integration -count=1 miniflux.app/tests || cat /tmp/miniflux.log

clean-integration-test:
	@ kill -9 `cat /tmp/miniflux.pid`
	@ rm -f /tmp/miniflux.pid /tmp/miniflux.log
	@ rm miniflux-test
	@ psql -U postgres -c 'drop database if exists miniflux_test;'

docker-images:
	for arch in amd64 arm32v6 arm64v8; do \
	  case $${arch} in \
		amd64   ) miniflux_arch="amd64";; \
		arm32v6 ) miniflux_arch="armv6";; \
		arm64v8 ) miniflux_arch="armv8";; \
	  esac ;\
	  cp Dockerfile Dockerfile.$${arch} && \
	  sed -i"" -e "s|__BASEIMAGE_ARCH__|$${arch}|g" Dockerfile.$${arch} && \
	  sed -i"" -e "s|__MINIFLUX_VERSION__|$(VERSION)|g" Dockerfile.$${arch} && \
	  sed -i"" -e "s|__MINIFLUX_ARCH__|$${miniflux_arch}|g" Dockerfile.$${arch} && \
	  docker build --pull -f Dockerfile.$${arch} -t $(DOCKER_IMAGE):$${arch}-$(VERSION) . && \
	  docker tag $(DOCKER_IMAGE):$${arch}-${VERSION} $(DOCKER_IMAGE):$${arch}-latest && \
	  rm -f Dockerfile.$${arch}* ;\
	done

docker-manifest:
	for version in $(VERSION) latest; do \
		docker push $(DOCKER_IMAGE):amd64-$${version} && \
		docker push $(DOCKER_IMAGE):arm32v6-$${version} && \
		docker push $(DOCKER_IMAGE):arm64v8-$${version} && \
		docker manifest create --amend $(DOCKER_IMAGE):$${version} \
			$(DOCKER_IMAGE):amd64-$${version} \
			$(DOCKER_IMAGE):arm32v6-$${version} \
			$(DOCKER_IMAGE):arm64v8-$${version} && \
		docker manifest annotate $(DOCKER_IMAGE):$${version} \
			$(DOCKER_IMAGE):arm32v6-$${version} --os linux --arch arm --variant v6 && \
		docker manifest annotate $(DOCKER_IMAGE):$${version} \
			$(DOCKER_IMAGE):arm64v8-$${version} --os linux --arch arm64 --variant v8 && \
		docker manifest push --purge $(DOCKER_IMAGE):$${version} ;\
	done
