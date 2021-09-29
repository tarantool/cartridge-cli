Packing an application into RPM and DEB
=======================================

``cartridge pack rpm|deb`` creates an RPM or DEB package.

Package contents
----------------

The resulting artifact name is ``<app-name>-<version>[-<suffix>].{rpm,deb}``.

The package name is ``<app-name>`` no matter what the artifact name is.

If you're using an open-source version of Tarantool, the package has a ``tarantool``
dependency (version >= ``<major>.<minor>`` and < ``<major+1>``, where
``<major>.<minor>`` is the version of Tarantool used for packaging the application).

The package contents are as follows:

*   The contents of the distribution directory, placed in ``/usr/share/tarantool/<app-name>``.
    For Tarantool Enterprise, this directory also contains ``tarantool`` and
    ``tarantoolctl`` binaries.

*   Unit files that allow running the application as a ``systemd`` service:
    ``/etc/systemd/system/<app-name>.service`` and
    ``/etc/systemd/system/<app-name>@.service``.

*   Application stateboard unit file:
    ``/etc/systemd/system/<app-name>-stateboard.service``.
    This file will be packed only if the application contains
    ``stateboard.init.lua`` in its root directory.

*   The file ``/usr/lib/tmpfiles.d/<app-name>.conf``, which allows the instance to restart
    after server reboot.

Upon package installation, the following directories are created:

*   ``/etc/tarantool/conf.d/`` stores instance configuration.
*   ``/var/lib/tarantool/`` stores instance snapshots.
*   ``/var/run/tarantool/`` stores PID files and console sockets.

Customizing systemd unit files
------------------------------

Use the flags ``--unit-template``, ``--instantiated-unit-template``, and
``--stateboard-unit-template`` to customize standard unit files.

One reason to customize standard unit files
is if you want to deploy your RPM/DEB package on a platform
different from the one where you've built it.
In this case, ``ExecStartPre`` may contain an incorrect path to `mkdir`.
As a hotfix, we suggest editing the unit files.

The unit files can contain `text templates <https://golang.org/pkg/text/template/>`__.

Example
~~~~~~~
This is an instantiated unit file.

..  code-block:: kconfig

    [Unit]
    Description=Tarantool Cartridge app {{ .Name }}@%i
    After=network.target

    [Service]
    Type=simple
    ExecStartPre=/bin/sh -c 'mkdir -p {{ .InstanceWorkDir }}'
    ExecStart={{ .Tarantool }} {{ .AppEntrypointPath }}
    Restart=on-failure
    RestartSec=2
    User=tarantool
    Group=tarantool

    Environment=TARANTOOL_APP_NAME={{ .Name }}
    Environment=TARANTOOL_WORKDIR={{ .InstanceWorkDir }}
    Environment=TARANTOOL_CFG={{ .ConfPath }}
    Environment=TARANTOOL_PID_FILE={{ .InstancePidFile }}
    Environment=TARANTOOL_CONSOLE_SOCK={{ .InstanceConsoleSock }}
    Environment=TARANTOOL_INSTANCE_NAME=%i

    LimitCORE=infinity
    # Disable OOM killer
    OOMScoreAdjust=-1000
    # Increase fd limit for Vinyl
    LimitNOFILE=65535

    # Systemd waits until all xlogs are recovered
    TimeoutStartSec=86400s
    # Give a reasonable amount of time to close xlogs
    TimeoutStopSec=10s

    [Install]
    WantedBy=multi-user.target
    Alias={{ .Name }}.%i

Supported variables
~~~~~~~~~~~~~~~~~~~

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``Name``
            -   Application name.
        *   -   ``StateboardName``
            -   Application stateboard name (``<app-name>-stateboard``).
        *   -   ``DefaultWorkDir``
            -   Default instance working directory
                (``/var/lib/tarantool/<app-name>.default``).
        *   -   ``InstanceWorkDir``
            -   Application instance working directory
                (``/var/lib/tarantool/<app-name>.<instance-name>``).
        *   -   ``StateboardWorkDir``
            -   Stateboard working directory
                (``/var/lib/tarantool/<app-name>-stateboard``).
        *   -   ``DefaultPidFile``
            -   Default instance PID file (``/var/run/tarantool/<app-name>.default.pid``).
        *   -   ``InstancePidFile``
            -   Application instance PID file
                (``/var/run/tarantool/<app-name>.<instance-name>.pid``).
        *   -   ``StateboardPidFile``
            -   Stateboard PID file (``/var/run/tarantool/<app-name>-stateboard.pid``).
        *   -   ``DefaultConsoleSock``
            -   Default instance console socket
                (``/var/run/tarantool/<app-name>.default.control``).
        *   -   ``InstanceConsoleSock``
            -   Application instance console socket
                (``/var/run/tarantool/<app-name>.<instance-name>.control``).
        *   -   ``StateboardConsoleSock``
            -   Stateboard console socket (``/var/run/tarantool/<app-name>-stateboard.control``).
        *   -   ``ConfPath``
            -   Path to the application instances config (``/etc/tarantool/conf.d``).
        *   -   ``AppEntrypointPath``
            -   Path to the application entrypoint
                (``/usr/share/tarantool/<app-name>/init.lua``).
        *   -   ``StateboardEntrypointPath``
            -   Path to the stateboard entrypoint
                (``/usr/share/tarantool/<app-name>/stateboard.init.lua``).

Installation
------------

If you are using open-source Tarantool, your application package has
Tarantool as a dependency.
In this case, before installing your RPM/DEB package, you have to enable the Tarantool repo
to allow your package manager to install this dependency correctly:

..  code-block:: bash

    curl -L https://tarantool.io/installer.sh | VER=${TARANTOOL_VERSION} bash

After this, you can install the application package.

Starting application instances
------------------------------

After you've installed the package, configure the instances you want to start.

For example, if your application name is ``myapp`` and you want to start two
instances, put the file ``myapp.yml`` into the ``/etc/tarantool/conf.d`` directory:

..  code-block:: yaml

    myapp:
      cluster_cookie: secret-cookie

    myapp.instance-1:
      http_port: 8081
      advertise_uri: localhost:3301

    myapp.instance-2:
      http_port: 8082
      advertise_uri: localhost:3302

Learn more about
:ref:`configuring Cartridge application instances </book/cartridge/cartridge_dev/#configuring-instances>`.

Now start the instances you've configured:

..  code-block:: bash

    systemctl start myapp@instance-1
    systemctl start myapp@instance-2

If you use stateful failover, start the application stateboard as well.
Make sure that your application has ``stateboard.init.lua`` in its root directory.

Add the ``myapp-stateboard`` section to ``/etc/tarantool/conf.d/myapp.yml``:

..  code-block:: yaml

    myapp-stateboard:
      listen: localhost:3310
      password: passwd

Then start the stateboard service:

..  code-block:: bash

    systemctl start myapp-stateboard
