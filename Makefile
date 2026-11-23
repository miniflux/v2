APP             := miniflux
DOCKER_IMAGE    := miniflux/miniflux
VERSION         := $(shell git describe --tags --exact-match 2>/dev/null)
DOCKER_TAG      := $(or $(VERSION),dev)
LD_FLAGS        := "-s -w -X 'miniflux.app/v2/internal/version.Version=$(VERSION)'"
DB_URL          := postgres://postgres:postgres@localhost/miniflux_test?sslmode=disable
DOCKER_PLATFORM := amd64

export PGPASSWORD := postgres

.PHONY: \
	miniflux \
	miniflux-no-pie \
	linux-amd64 \
	linux-arm64 \
	linux-armv7 \
	linux-armv6 \
	linux-armv5 \
	linux-riscv64 \
	darwin-amd64 \
	darwin-arm64 \
	freebsd-amd64 \
	freebsd-arm64 \
	openbsd-amd64 \
	build \
	run \
	clean \
	add-string \
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
	@ go build -buildmode=pie -ldflags=$(LD_FLAGS) -o $(APP)

miniflux-no-pie:
	@ go build -ldflags=$(LD_FLAGS) -o $(APP)

linux-amd64:
	@ CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags=$(LD_FLAGS) -o $(APP)-$@
	@ sha256sum $(APP)-$@ > $(APP)-$@.sha256

linux-arm64:
	@ CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags=$(LD_FLAGS) -o $(APP)-$@
	@ sha256sum $(APP)-$@ > $(APP)-$@.sha256

linux-armv7:
	@ CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags=$(LD_FLAGS) -o $(APP)-$@
	@ sha256sum $(APP)-$@ > $(APP)-$@.sha256

linux-armv6:
	@ CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -ldflags=$(LD_FLAGS) -o $(APP)-$@
	@ sha256sum $(APP)-$@ > $(APP)-$@.sha256

linux-armv5:
	@ CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=5 go build -ldflags=$(LD_FLAGS) -o $(APP)-$@
	@ sha256sum $(APP)-$@ > $(APP)-$@.sha256

linux-riscv64:
	@ CGO_ENABLED=0 GOOS=linux GOARCH=riscv64 go build -ldflags=$(LD_FLAGS) -o $(APP)-$@
	@ sha256sum $(APP)-$@ > $(APP)-$@.sha256

darwin-amd64:
	@ CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags=$(LD_FLAGS) -o $(APP)-$@
	@ sha256sum $(APP)-$@ > $(APP)-$@.sha256

darwin-arm64:
	@ CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags=$(LD_FLAGS) -o $(APP)-$@
	@ sha256sum $(APP)-$@ > $(APP)-$@.sha256

freebsd-amd64:
	@ CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -ldflags=$(LD_FLAGS) -o $(APP)-$@
	@ sha256sum $(APP)-$@ > $(APP)-$@.sha256

freebsd-arm64:
	@ CGO_ENABLED=0 GOOS=freebsd GOARCH=arm64 go build -ldflags=$(LD_FLAGS) -o $(APP)-$@
	@ sha256sum $(APP)-$@ > $(APP)-$@.sha256

openbsd-amd64:
	@ CGO_ENABLED=0 GOOS=openbsd GOARCH=amd64 go build -ldflags=$(LD_FLAGS) -o $(APP)-$@
	@ sha256sum $(APP)-$@ > $(APP)-$@.sha256

build: linux-amd64 linux-arm64 linux-armv7 linux-armv6 linux-armv5 linux-riscv64 darwin-amd64 darwin-arm64 freebsd-amd64 freebsd-arm64 openbsd-amd64

run:
	@ LOG_DATE_TIME=1 LOG_LEVEL=debug RUN_MIGRATIONS=1 CREATE_ADMIN=1 ADMIN_USERNAME=admin ADMIN_PASSWORD=test123 go run .

clean:
	@ rm -f $(APP)-* $(APP) $(APP)*.rpm $(APP)*.deb

add-string:
	$(if $(KEY),,$(error KEY is required, e.g. make add-string KEY=page.example.title VAL="Example"))
	$(if $(VAL),,$(error VAL is required, e.g. make add-string KEY=page.example.title VAL="Example"))
	cd internal/locale/translations && \
	for file in *.json; do \
		jq --indent 4 --arg key "$(KEY)" --arg val "$(VAL)" \
		   '. + {($$key): $$val} | to_entries | sort_by(.key) | from_entries' "$$file" > tmp && \
		mv tmp "$$file"; \
	done

test:
	go test -cover -race -count=1 ./...

lint:
	go vet ./...
	gofmt -d -e .
	golangci-lint run

integration-test:
	psql -U postgres -c 'drop database if exists miniflux_test;'
	psql -U postgres -c 'create database miniflux_test;'

	DATABASE_URL=$(DB_URL) \
	ADMIN_USERNAME=admin \
	ADMIN_PASSWORD=test123 \
	CREATE_ADMIN=1 \
	RUN_MIGRATIONS=1 \
	LOG_LEVEL=debug \
	FETCHER_ALLOW_PRIVATE_NETWORKS=1 \
	INTEGRATION_ALLOW_PRIVATE_NETWORKS=1 \
	go run . >/tmp/miniflux.log 2>&1 & echo "$$!" > "/tmp/miniflux.pid"

	for attempt in $$(seq 1 120); do \
		curl -fs -o /dev/null http://127.0.0.1:8080/healthcheck && break; \
		if [ "$$attempt" = 120 ]; then echo "Miniflux did not start:"; cat /tmp/miniflux.log; exit 1; fi; \
		sleep 1; \
	done

	TEST_MINIFLUX_BASE_URL=http://127.0.0.1:8080 \
	TEST_MINIFLUX_ADMIN_USERNAME=admin \
	TEST_MINIFLUX_ADMIN_PASSWORD=test123 \
	go test -v -count=1 ./internal/api

clean-integration-test:
	@ go_run_pid=$$(cat /tmp/miniflux.pid 2>/dev/null); \
	  if [ -n "$$go_run_pid" ]; then \
	    pkill -P "$$go_run_pid"; \
	    for attempt in $$(seq 1 30); do kill -0 "$$go_run_pid" 2>/dev/null || break; sleep 1; done; \
	  fi
	@ rm -f /tmp/miniflux.pid /tmp/miniflux.log
	@ psql -U postgres -c 'drop database if exists miniflux_test;'

docker-image:
	docker build --pull -t $(DOCKER_IMAGE):$(DOCKER_TAG) -f packaging/docker/alpine/Dockerfile .

docker-image-distroless:
	docker build --pull -t $(DOCKER_IMAGE):$(DOCKER_TAG)-distroless -f packaging/docker/distroless/Dockerfile .

docker-images:
	$(if $(VERSION),,$(error VERSION is required to push images, e.g. make docker-images VERSION=2.3.0))
	docker buildx build \
		--platform linux/amd64,linux/arm64,linux/arm/v7,linux/arm/v6,linux/riscv64 \
		--file packaging/docker/alpine/Dockerfile \
		--tag $(DOCKER_IMAGE):$(VERSION) \
		--push .

rpm: clean
	@ docker build \
		-t miniflux-rpm-builder \
		-f packaging/rpm/Dockerfile \
		.
	@ docker run --rm \
		-v $(CURDIR):/root/rpmbuild/RPMS/x86_64 miniflux-rpm-builder \
		rpmbuild -bb --define "_miniflux_version $(VERSION)" /root/rpmbuild/SPECS/miniflux.spec

debian:
	@ docker buildx build --load \
		--platform linux/$(DOCKER_PLATFORM) \
		-t miniflux-deb-builder \
		-f packaging/debian/Dockerfile \
		.
	@ docker run --rm --platform linux/$(DOCKER_PLATFORM) \
		-v $(CURDIR):/pkg miniflux-deb-builder

debian-packages: clean
	$(MAKE) debian DOCKER_PLATFORM=amd64
	$(MAKE) debian DOCKER_PLATFORM=arm64
	$(MAKE) debian DOCKER_PLATFORM=arm/v7
	$(MAKE) debian DOCKER_PLATFORM=riscv64
