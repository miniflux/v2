%undefine _disable_source_fetch

Name:    influxeed-engine
Version: %{_influxeed-engine_version}
Release: 1.0
Summary: Minimalist and opinionated feed reader
URL: https://influxeed-engine.app/
License: ASL 2.0
Source0: influxeed-engine
Source1: influxeed-engine.service
Source2: influxeed-engine.conf
Source3: influxeed-engine.1
Source4: LICENSE
BuildRoot: %{_topdir}/BUILD/%{name}-%{version}-%{release}
BuildArch: x86_64
Requires(pre): shadow-utils

%{?systemd_ordering}

AutoReqProv: no

%define __strip /bin/true
%define __os_install_post %{nil}

%description
%{summary}

%install
mkdir -p %{buildroot}%{_bindir}
install -p -m 755 %{SOURCE0} %{buildroot}%{_bindir}/influxeed-engine
install -D -m 644 %{SOURCE1} %{buildroot}%{_unitdir}/influxeed-engine.service
install -D -m 600 %{SOURCE2} %{buildroot}%{_sysconfdir}/influxeed-engine.conf
install -D -m 644 %{SOURCE3} %{buildroot}%{_mandir}/man1/influxeed-engine.1
install -D -m 644 %{SOURCE4} %{buildroot}%{_docdir}/influxeed-engine/LICENSE

%files
%defattr(755,root,root)
%{_bindir}/influxeed-engine
%{_docdir}/influxeed-engine
%defattr(644,root,root)
%{_unitdir}/influxeed-engine.service
%{_mandir}/man1/influxeed-engine.1*
%{_docdir}/influxeed-engine/*
%defattr(600,root,root)
%config(noreplace) %{_sysconfdir}/influxeed-engine.conf

%pre
getent group influxeed-engine >/dev/null || groupadd -r influxeed-engine
getent passwd influxeed-engine >/dev/null || \
    useradd -r -g influxeed-engine -d /dev/null -s /sbin/nologin \
    -c "influxeed-engine Daemon" influxeed-engine
exit 0

%post
%systemd_post influxeed-engine.service

%preun
%systemd_preun influxeed-engine.service

%postun
%systemd_postun_with_restart influxeed-engine.service
