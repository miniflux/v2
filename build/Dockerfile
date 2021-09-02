FROM golang:alpine AS build
RUN apk add --no-cache --update git
ADD . /go/src/app
WORKDIR /go/src/app
RUN go generate
RUN go build \
    -o miniflux \
    -ldflags="-s -w -X 'miniflux.app/version.Version=`git describe --tags --abbrev=0`' -X 'miniflux.app/version.Commit=`git rev-parse --short HEAD`' -X 'miniflux.app/version.BuildDate=`date +%FT%T%z`'" \
    main.go

FROM alpine:latest

EXPOSE 1201
ENV LISTEN_ADDR 0.0.0.0:1201
RUN apk --no-cache add ca-certificates tzdata
COPY --from=build /go/src/app/miniflux /usr/bin/miniflux
USER nobody
CMD ["/usr/bin/miniflux"]
