FROM docker.io/library/golang:alpine3.23 AS build
RUN apk add --no-cache build-base git make
ADD . /go/src/app
WORKDIR /go/src/app
RUN make miniflux

FROM docker.io/library/node:22-alpine AS defuddle
RUN git clone --depth 1 https://github.com/kepano/defuddle /defuddle
WORKDIR /defuddle
RUN npm ci && npm run build

FROM docker.io/library/alpine:3.23

LABEL org.opencontainers.image.title=Miniflux
LABEL org.opencontainers.image.description="Miniflux is a minimalist and opinionated feed reader"
LABEL org.opencontainers.image.licenses=Apache-2.0
LABEL org.opencontainers.image.source=https://github.com/naiba-forks/miniflux

EXPOSE 8080
ENV LISTEN_ADDR=0.0.0.0:8080
RUN apk --no-cache add ca-certificates tzdata tini nodejs \
    && ARCH=$(uname -m | sed 's/aarch64/aarch64/;s/x86_64/x86_64/') \
    && wget -q -O /usr/bin/lightpanda \
       "https://github.com/lightpanda-io/browser/releases/download/nightly/lightpanda-${ARCH}-linux" \
    && chmod +x /usr/bin/lightpanda
COPY --from=build /go/src/app/miniflux /usr/bin/miniflux
COPY --from=defuddle /defuddle/dist /usr/share/miniflux/defuddle/dist
COPY --from=defuddle /defuddle/node_modules /usr/share/miniflux/defuddle/node_modules
COPY --from=defuddle /defuddle/package.json /usr/share/miniflux/defuddle/package.json
USER 65534
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/usr/bin/miniflux"]
