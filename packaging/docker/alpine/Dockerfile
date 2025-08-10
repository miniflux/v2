FROM docker.io/library/golang:alpine3.22 AS build
RUN apk add --no-cache build-base git make
ADD . /go/src/app
WORKDIR /go/src/app
RUN make influxeed-engine

FROM docker.io/library/alpine:3.22

LABEL org.opencontainers.image.title=influxeed-engine
LABEL org.opencontainers.image.description="influxeed-engine is a minimalist and opinionated feed reader"
LABEL org.opencontainers.image.vendor="Frédéric Guillot"
LABEL org.opencontainers.image.licenses=Apache-2.0
LABEL org.opencontainers.image.url=https://influxeed-engine.app
LABEL org.opencontainers.image.source=https://github.com/influxeed-engine/v2
LABEL org.opencontainers.image.documentation=https://influxeed-engine.app/docs/

EXPOSE 8080
ENV LISTEN_ADDR=0.0.0.0:8080
RUN apk --no-cache add ca-certificates tzdata
COPY --from=build /go/src/app/influxeed-engine /usr/bin/influxeed-engine
USER 65534
CMD ["/usr/bin/influxeed-engine"]
