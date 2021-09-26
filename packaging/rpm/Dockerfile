FROM golang:1 AS build
ADD . /go/src/app
WORKDIR /go/src/app
RUN make miniflux

FROM centos:latest
RUN dnf install -y rpm-build
RUN mkdir -p /root/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
RUN echo "%_topdir /root/rpmbuild" >> .rpmmacros
COPY --from=build /go/src/app/miniflux /root/rpmbuild/SOURCES/miniflux
COPY --from=build /go/src/app/LICENSE /root/rpmbuild/SOURCES/
COPY --from=build /go/src/app/ChangeLog /root/rpmbuild/SOURCES/
COPY --from=build /go/src/app/miniflux.1 /root/rpmbuild/SOURCES/
COPY --from=build /go/src/app/packaging/systemd/miniflux.service /root/rpmbuild/SOURCES/
COPY --from=build /go/src/app/packaging/miniflux.conf /root/rpmbuild/SOURCES/
COPY --from=build /go/src/app/packaging/rpm/miniflux.spec /root/rpmbuild/SPECS/miniflux.spec
