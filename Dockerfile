ARG BASE_IMAGE_ARCH="amd64"
ARG ALPINE_LINUX_VERSION="3.12"

FROM golang:1-alpine${ALPINE_LINUX_VERSION} as build
ARG APP_VERSION
ARG APP_ARCH="amd64"
WORKDIR /go/src/app
RUN apk add --no-cache --update build-base git
COPY . .
RUN make linux-${APP_ARCH} VERSION=${APP_VERSION}
RUN cp /go/src/app/miniflux-linux-${APP_ARCH} /go/src/app/miniflux

FROM ${BASE_IMAGE_ARCH}/alpine:${ALPINE_LINUX_VERSION}
EXPOSE 8080
ENV LISTEN_ADDR 0.0.0.0:8080
RUN apk --no-cache add ca-certificates tzdata
COPY --from=build /go/src/app/miniflux /usr/bin/miniflux
USER nobody
CMD ["/usr/bin/miniflux"]
