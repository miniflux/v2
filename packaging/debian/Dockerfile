FROM docker.io/golang:1.24-bookworm AS build

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update -q && \
    apt-get install -y -qq build-essential devscripts dh-make debhelper && \
    mkdir -p /build/debian

ADD . /src

CMD ["/src/packaging/debian/build.sh"]
