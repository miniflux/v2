FROM docker.io/library/golang:alpine3.22 AS build
RUN apk add --no-cache build-base git make
ADD . /go/src/app
WORKDIR /go/src/app
RUN make miniflux

FROM docker.io/library/alpine:3.22



EXPOSE 8080
ENV LISTEN_ADDR=0.0.0.0:8080
RUN apk --no-cache add ca-certificates tzdata
COPY --from=build /go/src/app/miniflux /usr/bin/miniflux
USER 65534
CMD ["/usr/bin/miniflux"]
