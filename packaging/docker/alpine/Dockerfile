FROM docker.io/library/golang:alpine3.23 AS build
RUN apk add --no-cache build-base git make
ADD . /go/src/app
WORKDIR /go/src/app
RUN make miniflux

FROM docker.io/library/alpine:3.23

LABEL org.opencontainers.image.title=Miniflux
LABEL org.opencontainers.image.description="Miniflux is a minimalist and opinionated feed reader"
LABEL org.opencontainers.image.licenses=Apache-2.0
LABEL org.opencontainers.image.source=https://github.com/naiba-forks/miniflux

EXPOSE 8080
ENV LISTEN_ADDR=0.0.0.0:8080
# Chromium 128+ crashpad requires writable XDG directories for its crash
# database.  Running as nobody (UID 65534) means $HOME is not writable,
# causing SIGTRAP on aarch64.  Point XDG dirs to /tmp so crashpad can
# initialise without crashing.
ENV XDG_CONFIG_HOME=/tmp/.chromium
ENV XDG_CACHE_HOME=/tmp/.chromium
ARG PINCHTAB_VERSION=0.7.8
RUN apk --no-cache add ca-certificates tzdata chromium tini \
    && ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/') \
    && wget -q -O /usr/bin/pinchtab \
       "https://github.com/pinchtab/pinchtab/releases/download/v${PINCHTAB_VERSION}/pinchtab-linux-${ARCH}" \
    && chmod +x /usr/bin/pinchtab
COPY --from=build /go/src/app/miniflux /usr/bin/miniflux
USER 65534
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/usr/bin/miniflux"]
