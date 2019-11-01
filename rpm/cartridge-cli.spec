Name: cartridge-cli
Version: 1.0.0
Release: 1%{?dist}
Summary: Cartridge command line interface
Group: Applications/Databases
License: BSD
URL: https://github.com/tarantool/cartridge-cli
Source0: https://github.com/tarantool/%{name}/archive/%{version}/%{name}-%{version}.tar.gz
BuildRequires: cmake >= 2.8
BuildRequires: /usr/bin/prove
Requires: tarantool >= 1.7.5.0

%description
This package provides a command line interface for cartridge

%global debug_package %{nil}
%prep
%setup -q -n %{name}-%{version}

%build
%cmake . -DCMAKE_BUILD_TYPE=RelWithDebInfo -DVERSION=%{version}
make %{?_smp_mflags}

%install
%make_install

%files
%{_bindir}/cartridge
%{_datarootdir}/tarantool/*
%doc README.md
%{!?_licensedir:%global license %doc}
%license LICENSE

%changelog
* Thu Feb 18 2016 Roman Tsisyk <roman@tarantool.org> 1.0.0-1
- Initial version of the RPM spec
