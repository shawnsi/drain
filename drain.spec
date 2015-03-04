Name:           drain
Version:        0.0.2
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
* Wed Sep 24 2014 Shawn Siefkas <shawn@siefk.as> - 0.0.2-1
- Fixing iptables path for sudo support

* Wed Sep 24 2014 Shawn Siefkas <shawn@siefk.as> - 0.0.1-1
- Initial Spec File
