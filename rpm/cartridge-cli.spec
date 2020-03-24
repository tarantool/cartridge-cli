Name: cartridge-cli
Version: 1.0.0
Release: 1%{?dist}
Summary: Cartridge command line interface
Group: Applications/Databases
License: BSD
URL: https://github.com/tarantool/cartridge-cli
Source0: https://github.com/tarantool/%{name}/archive/%{version}/%{name}-%{version}.tar.gz
BuildRequires: cmake >= 2.8
BuildRequires: tarantool-devel >= 1.7.5.0
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
%doc README.md
%{!?_licensedir:%global license %doc}
%license LICENSE

%changelog
* Tue Oct 22 2019 Konstantin Nazarov <mail@knazarov.com>
- Initial version of the RPM spec

* Tue Mar 24 2020 Elizaveta Dokshina <eldokshina@mail.ru>
- Install only cartridge binary
