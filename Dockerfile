# Stage 1

FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git make gcc musl-dev

WORKDIR /app

RUN git clone https://github.com/miniflux/v2.git .

RUN go mod download

RUN make linux-amd64

# Stage 2

#FROM alpine:3.20
FROM alpine:3.7

WORKDIR /app
EXPOSE 8080

RUN apk add  --no-cache ca-certificates tzdata

COPY --from=builder /app/miniflux-linux-amd64 /usr/local/bin/miniflux

RUN adduser -D -g  "" -u 10001 miniflux
USER miniflux

ENTRYPOINT ["/usr/local/bin/miniflux"]
