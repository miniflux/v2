FROM golang:alpine AS alpha
RUN apk add --no-cache --update git
ADD . /go/src/app
WORKDIR /go/src/app
RUN go generate
RUN go build \
    -o miniflux \
    -ldflags="-s -w -X 'miniflux.app/version.Version=`git describe --tags --abbrev=0`' -X 'miniflux.app/version.Commit=`git rev-parse --short HEAD`' -X 'miniflux.app/version.BuildDate=`date +%FT%T%z`'" \
    main.go


FROM alpine

EXPOSE 1201

ENV LISTEN_ADDR 0.0.0.0:1201

RUN apk --no-cache add ca-certificates tzdata

COPY --from=alpha /go/src/app/miniflux /dist/miniflux

CMD ["/dist/miniflux"]
