Name:           drain
Version:        0.0.7
Release:        1%{?dist}
Summary:        TCP Drain

License:        MIT
Source0:        drain-%{version}.tar.gz

BuildRequires:  git
BuildRequires:  golang
BuildArch:      x86_64

%description
TCP Drain

%prep
%setup -q -c -n drain-%{version}
mkdir -p $RPM_BUILD_ROOT/usr/src/github.com/shawnsi/drain
cp -pr * $RPM_BUILD_ROOT/usr/src/github.com/shawnsi/drain

%build
GOPATH=$RPM_BUILD_ROOT/usr go get github.com/docopt/docopt-go

%install
GOPATH=$RPM_BUILD_ROOT/usr go install github.com/shawnsi/drain
rm -fr $RPM_BUILD_ROOT/usr/{pkg,src}

%files
/usr/bin/drain

%changelog
* Wed Jun 17 2015 Shawn Siefkas <shawn@siefk.as> - 0.0.7-1
- Reject with TCP RST

* Tue Jun 16 2015 Shawn Siefkas <shawn@siefk.as> - 0.0.6-1
- Reject all connections after reject timeout met

* Tue Jun 16 2015 Shawn Siefkas <shawn@siefk.as> - 0.0.5-1
- Adding kill to start

* Tue Jun 16 2015 Shawn Siefkas <shawn@siefk.as> - 0.0.4-1
- Adding timeout to monitor

* Wed Mar 04 2015 Shawn Siefkas <shawn@siefk.as> - 0.0.3-1
- Adding debug logging

* Wed Sep 24 2014 Shawn Siefkas <shawn@siefk.as> - 0.0.2-1
- Fixing iptables path for sudo support

* Wed Sep 24 2014 Shawn Siefkas <shawn@siefk.as> - 0.0.1-1
- Initial Spec File
