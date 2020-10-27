===============================================================================
Packing an application into RPM and DEB
===============================================================================

``cartridge pack rpm|deb`` creates an RPM or DEB package.

-------------------------------------------------------------------------------
Package contents
-------------------------------------------------------------------------------

The result artifact name is ``<app-name>-<version>[-<suffix>].{rpm,deb}``.

The package name is ``<app-name>`` no matter what the artifact name is.

If you use an opensource version of Tarantool, the package has a ``tarantool``
dependency (version >= ``<major>.<minor>`` and < ``<major+1>``, where
``<major>.<minor>`` is the version of Tarantool used for packing the application).

The package contents is as follows:

* the contents of the distribution directory, placed in the
  ``/usr/share/tarantool/<app-name>`` directory
  (for Tarantool Enterprise, this directory also contains ``tarantool`` and
  ``tarantoolctl`` binaries);

* unit files for running the application as a ``systemd`` service:
  ``/etc/systemd/system/<app-name>.service`` and
  ``/etc/systemd/system/<app-name>@.service``;

* application stateboard unit file:
  ``/etc/systemd/system/<app-name>-stateboard.service``
  (will be packed only if the application contains ``stateboard.init.lua`` in its root);

* the file ``/usr/lib/tmpfiles.d/<app-name>.conf`` that allows the instance to restart
  after server restart.

The following directories are created:

* ``/etc/tarantool/conf.d/`` — directory for instances configuration;
* ``/var/lib/tarantool/`` — directory to store instances snapshots;
* ``/var/run/tarantool/`` — directory to store PID-files and console sockets.

-------------------------------------------------------------------------------
Customizing systemd unit-files
-------------------------------------------------------------------------------

Use the options ``--unit-template``, ``--instantiated-unit-template`` and
``--stateboard-unit-template`` to customize standard unit files.

You may need it first of all for DEB packages, if your build platform
is different from the deployment platform. In this case, ``ExecStartPre`` may
contain an incorrect path to `mkdir`. As a hotfix, we suggest customizing the
unit files.

Templates can contain `text templates <Templates_>`_.

.. _Templates: https://golang.org/pkg/text/template/

Example of an instantiated unit file:

.. code-block:: kconfig

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

Supported variables:

* ``Name`` — the application name;
* ``StateboardName`` — the application stateboard name (``<app-name>-stateboard``);

* ``DefaultWorkDir`` — default instance working directory (``/var/lib/tarantool/<app-name>.default``);
* ``InstanceWorkDir`` — application instance working directory (``/var/lib/tarantool/<app-name>.<instance-name>``);
* ``StateboardWorkDir`` — stateboard working directory (``/var/lib/tarantool/<app-name>-stateboard``);

* ``DefaultPidFile`` — default instance pid file (``/var/run/tarantool/<app-name>.default.pid``);
* ``InstancePidFile`` — application instance pid file (``/var/run/tarantool/<app-name>.<instance-name>.pid``);
* ``StateboardPidFile`` — stateboard pid file (``/var/run/tarantool/<app-name>-stateboard.pid``);

* ``DefaultConsoleSock`` — default instance console socket (``/var/run/tarantool/<app-name>.default.control``);
* ``InstanceConsoleSock`` — application instance console socket (``/var/run/tarantool/<app-name>.<instance-name>.control``);
* ``StateboardConsoleSock`` — stateboard console socket (``/var/run/tarantool/<app-name>-stateboard.control``);

* ``ConfPath`` — path to the application instances config (``/etc/tarantool/conf.d``);

* ``AppEntrypointPath`` — path to the application entrypoint (``/usr/share/tarantool/<app-name>/init.lua``);
* ``StateboardEntrypointPath`` — path to the stateboard entrypoint (``/usr/share/tarantool/<app-name>/stateboard.init.lua``);

-------------------------------------------------------------------------------
Installation
-------------------------------------------------------------------------------

If you are using opensource Tarantool, then your application package has
Tarantool dependency.
In this case, before package installation you need to enable the Tarantool repo
to allow your package manager install this dependency correctly:

* for both RPM and DEB:

  .. code-block:: bash

      curl -L https://tarantool.io/installer.sh | VER=${TARANTOOL_VERSION} bash

Now, you can simply install an application package.

-------------------------------------------------------------------------------
Starting application instances
-------------------------------------------------------------------------------

After package installation you need to specify configuration for instances to start.

For example, if your application is named ``myapp`` and you want to start two
instances, put the ``myapp.yml`` file into the ``/etc/tarantool/conf.d`` directory.

.. code-block:: yaml

    myapp:
      cluster_cookie: secret-cookie

    myapp.instance-1:
      http_port: 8081
      advertise_uri: localhost:3301

    myapp.instance-2:
      http_port: 8082
      advertise_uri: localhost:3302

For more details about instances configuration see the
`documentation <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_dev/#configuring-instances>`_.

Now, start the configured instances:

.. code-block:: bash

    systemctl start myapp@instance-1
    systemctl start myapp@instance-2

If you use stateful failover, you need to start application stateboard.

(Remember that your application should contain ``stateboard.init.lua`` in its
root.)

Add the ``myapp-stateboard`` section to ``/etc/tarantool/conf.d/myapp.yml``:

.. code-block:: yaml

    myapp-stateboard:
      listen: localhost:3310
      password: passwd

Then, start the stateboard service:

.. code-block:: bash

    systemctl start myapp-stateboard
