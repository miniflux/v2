ARG BASE_IMAGE_ARCH="amd64"

FROM ${BASE_IMAGE_ARCH}/golang:1.22-bookworm AS build

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update -q && \
    apt-get install -y -qq build-essential devscripts dh-make debhelper && \
    mkdir -p /build/debian

ADD . /src

CMD ["/src/packaging/debian/build.sh"]
