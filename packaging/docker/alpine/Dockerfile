FROM docker.io/library/golang:bookworm AS build
RUN apt-get update && apt-get install -y --no-install-recommends build-essential git make && rm -rf /var/lib/apt/lists/*
ADD . /go/src/app
WORKDIR /go/src/app
# PIE build dynamically links libc — must compile on glibc (Debian) to match runtime.
RUN make miniflux

FROM docker.io/library/node:22-alpine AS defuddle
RUN apk add --no-cache git
RUN git clone --depth 1 https://github.com/kepano/defuddle /defuddle
WORKDIR /defuddle
RUN npm ci && npm run build

# Debian slim: Lightpanda is glibc-linked, Alpine musl lacks fcntl64 symbol —
# must use glibc-based image for native compatibility.
FROM docker.io/library/debian:bookworm-slim

LABEL org.opencontainers.image.title=Miniflux
LABEL org.opencontainers.image.description="Miniflux is a minimalist and opinionated feed reader"
LABEL org.opencontainers.image.licenses=Apache-2.0
LABEL org.opencontainers.image.source=https://github.com/naiba-forks/miniflux

EXPOSE 8080
ENV LISTEN_ADDR=0.0.0.0:8080
# TARGETARCH is provided by Docker buildx (amd64/arm64), map to Lightpanda naming (x86_64/aarch64)
ARG TARGETARCH
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata tini nodejs curl \
    && rm -rf /var/lib/apt/lists/* \
    && ARCH=$(echo "${TARGETARCH}" | sed 's/amd64/x86_64/;s/arm64/aarch64/') \
    && curl -fsSL -o /usr/bin/lightpanda \
       "https://github.com/lightpanda-io/browser/releases/download/nightly/lightpanda-${ARCH}-linux" \
    && chmod +x /usr/bin/lightpanda
COPY --from=build /go/src/app/miniflux /usr/bin/miniflux
COPY --from=defuddle /defuddle/dist /usr/share/miniflux/defuddle/dist
COPY --from=defuddle /defuddle/node_modules /usr/share/miniflux/defuddle/node_modules
COPY --from=defuddle /defuddle/package.json /usr/share/miniflux/defuddle/package.json
USER 65534
ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["/usr/bin/miniflux"]
