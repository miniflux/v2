# Use this image to build the executable
FROM golang:1.16-alpine AS compiler

RUN apk add --no-cache git ca-certificates

WORKDIR $GOPATH/src/minify
COPY . .

RUN make install

# Final image containing the executable from the previous step
FROM alpine:3

COPY --from=compiler /bin/minify /bin/minify
