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
# Install runtime dependencies: ca-certificates, tzdata, and chromium for pinchtab JS rendering.
# Pinchtab binary must be provided externally (e.g. COPY from a builder stage or download).
RUN apk --no-cache add ca-certificates tzdata chromium dumb-init
COPY --from=build /go/src/app/miniflux /usr/bin/miniflux
USER 65534
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/usr/bin/miniflux"]
