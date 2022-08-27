FROM golang:latest AS build
ADD . /go/src/app
WORKDIR /go/src/app
RUN go build \
    -o miniflux \
    -ldflags="-s -w -X 'miniflux.app/version.Version=`git describe --tags --abbrev=0`' -X 'miniflux.app/version.Commit=`git rev-parse --short HEAD`' -X 'miniflux.app/version.BuildDate=`date +%FT%T%z`'" \
    main.go

FROM gcr.io/distroless/base

LABEL org.opencontainers.image.title=Miniflux
LABEL org.opencontainers.image.description="Miniflux is a minimalist and opinionated feed reader"
LABEL org.opencontainers.image.vendor="Frédéric Guillot"
LABEL org.opencontainers.image.licenses=Apache-2.0
LABEL org.opencontainers.image.url=https://miniflux.app
LABEL org.opencontainers.image.source=https://github.com/miniflux/v2
LABEL org.opencontainers.image.documentation=https://miniflux.app/docs/

EXPOSE 8080
ENV LISTEN_ADDR 0.0.0.0:8080
COPY --from=build /go/src/app/miniflux /usr/bin/miniflux
USER nonroot:nonroot
CMD ["/usr/bin/miniflux"]
