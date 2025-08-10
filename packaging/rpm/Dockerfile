FROM golang:1 AS build
ENV CGO_ENABLED=0
ADD . /go/src/app
WORKDIR /go/src/app
RUN make influxeed-engine

FROM rockylinux:9
RUN dnf install --setopt=install_weak_deps=False -y rpm-build systemd-rpm-macros
RUN mkdir -p /root/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
RUN echo "%_topdir /root/rpmbuild" >> .rpmmacros
COPY --from=build /go/src/app/influxeed-engine /root/rpmbuild/SOURCES/influxeed-engine
COPY --from=build /go/src/app/LICENSE /root/rpmbuild/SOURCES/
COPY --from=build /go/src/app/influxeed-engine.1 /root/rpmbuild/SOURCES/
COPY --from=build /go/src/app/packaging/systemd/influxeed-engine.service /root/rpmbuild/SOURCES/
COPY --from=build /go/src/app/packaging/influxeed-engine.conf /root/rpmbuild/SOURCES/
COPY --from=build /go/src/app/packaging/rpm/influxeed-engine.spec /root/rpmbuild/SPECS/influxeed-engine.spec
