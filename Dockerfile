FROM golang:1.12-alpine3.10 as build
ENV GO111MODULE=on
WORKDIR /go/src/app
RUN apk add --no-cache --update build-base git
COPY . .
RUN make linux-__MINIFLUX_ARCH__ VERSION=__MINIFLUX_VERSION__

FROM __BASEIMAGE_ARCH__/alpine:3.10.0
EXPOSE 8080
ENV LISTEN_ADDR 0.0.0.0:8080
RUN apk --no-cache add ca-certificates tzdata
COPY --from=build /go/src/app/miniflux-linux-__MINIFLUX_ARCH__ /usr/bin/miniflux
USER nobody
CMD ["/usr/bin/miniflux"]
