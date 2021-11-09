Packaging an application into RPM or DEB
========================================

``cartridge pack rpm|deb`` creates an RPM or DEB package.

Flags
-----

Use the following flags to control the local packaging of an RPM or DEB distribution.
For flags that are applicable for packaging any distribution type,
check the :doc:`packaging overview </book/cartridge/cartridge_cli/commands/pack>`.

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``--deps``
            -   Defines the dependencies of the package.
        *   -   ``--deps-file``
            -   Path to the file that contains package dependencies.
                Defaults to ``package-deps.txt`` in the application directory.
        *   -   ``--preinst``
            -   Path to the pre-install script for RPM and DEB packages.
        *   -   ``--postinst``
            -   Path to the post-install script for RPM and DEB packages.
        *   -   ``--unit-template``
            -   Path to the template for the ``systemd`` unit file.
        *   -   ``--instantiated-unit-template``
            -   Path to the template for the ``systemd`` instantiated unit file.
        *   -   ``--stateboard-unit-template``
            -   Path to the template for the stateboard ``systemd`` unit file.
        *   -   ``--unit-params-file``
            -   Path to the file that contains unit parameters for ``systemd`` unit files.
                Defaults to ``systemd-unit-params.yml`` in the application root directory.


Package contents
----------------

The resulting artifact name is ``<app-name>-<version>[-<suffix>].{rpm,deb}``.

The package name is ``<app-name>`` no matter what the artifact name is.

If you're using an open-source version of Tarantool, the package has a ``tarantool``
dependency (version >= ``<major>.<minor>`` and < ``<major+1>``, where
``<major>.<minor>`` is the version of Tarantool used for packaging the application).

The package contents are as follows:

*   Contents of the application directory.
    They will be placed at ``/usr/share/tarantool/<app-name>``.
    In case of Tarantool Enterprise, this directory also contains the
    ``tarantool`` and ``tarantoolctl`` binaries.

*   Unit files that allow running the application as a ``systemd`` service.
    They will be unpacked as ``/etc/systemd/system/<app-name>.service`` and
    ``/etc/systemd/system/<app-name>@.service``.

*   Application stateboard unit file. When unpacked, it is placed at
    ``/etc/systemd/system/<app-name>-stateboard.service``.
    This file will be packed only if the application contains
    ``stateboard.init.lua`` in its root directory.

*   The file ``/usr/lib/tmpfiles.d/<app-name>.conf``, which allows the instance to restart
    after server reboot.

Upon package installation, the following directories are created:

*   ``/etc/tarantool/conf.d/`` stores instance configuration.
*   ``/var/lib/tarantool/`` stores instance snapshots.
*   ``/var/run/tarantool/`` stores PID files and console sockets.

Dependencies
------------

The ``--deps`` and ``--deps-file`` flags require similar formats of dependency information.
However, ``--deps`` does not allow you to specify major and minor versions:

..  code-block:: bash

    # You can't do that:
    cartridge pack rpm --deps dependency_06>=4,<5 appname

    # Instead, do this:
    cartridge pack rpm --deps dependency_06>=4,dependency_06<5 appname

    # Or this:
    cartridge pack rpm --deps dependency_06>=4 --deps dependency_06<5 appname

``--deps-file`` lets you specify dependencies in a file (``package-deps.txt`` by default).
The file is located in the application root directory.
If you created your application from template, ``package-deps.txt`` is already there.

Example dependencies file
~~~~~~~~~~~~~~~~~~~~~~~~~

..  code-block:: bash

    dependency_01 >= 2.5
    dependency_01 <
    dependency_02 >= 1, < 5
    dependency_03==2
    dependency_04<5,>=1.5.3

Each line must describe a single dependency.
For each dependency, you can specify the major or minor version,
as well as the highest and lowest compatible versions.


..  _cartridge-cli-preinst_postinst:

Pre-install and post-install scripts
------------------------------------

You can add Bash scripts that will run before and after
the installation of your RPM/DEB package.
This might be useful, for example, if you want to set up symlinks.
Place these files in your application root directory.

``preinst.sh`` is the default name of the pre-install script.
``postinst.sh`` is the default name of the post-install script.

To specify different names, use ``cartridge pack`` with the
``--preinst`` and ``--postinst`` flags correspondingly.

Provide absolute paths to executables in the pre- and post-install scripts,
or use ``/bin/sh -c ''`` instead.

Example pre-/post-install script
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

..  code-block:: bash

    /bin/sh -c 'touch file-path'
    /bin/sh -c 'mkdir dir-path'
    # or
    /bin/mkdir dir-path


Customizing systemd unit files
------------------------------

Use the flags ``--unit-template``, ``--instantiated-unit-template``, and
``--stateboard-unit-template`` to customize standard unit files.

One reason to customize standard unit files
is if you want to deploy your RPM/DEB package on a platform
different from the one where you've built it.
In this case, ``ExecStartPre`` may contain an incorrect path to ``mkdir``.
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

Passing parameters to unit files
--------------------------------

You can pass certain parameters to your application's unit files
using a special file.
By default, it is ``systemd-unit-params.yml``, located in the project directory.
To use a different file, specify its name with the ``--unit-params-file`` flag.

For example, the ``fd-limit`` option lets you limit the number of file descriptors
determined by the ``LimitNOFILE`` parameter in the ``systemd`` unit file and
instantiated unit file.
Another example would be ``stateboard-fd-limit``, which lets you
set the file descriptor limit in the stateboard ``systemd`` unit file.

You can also pass parameters via environment variables with the systemd unit file.
To do so, specify the instance and stateboard arguments in the unit parameters file.
The parameter will convert to ``Environment=TARANTOOL_<PARAM>: <value>`` in the unit file.
Note that these variables have higher priority than the variables
in the instance configuration file (``--cfg``).

..  // these are YAML options, put them in a separate file?

Supported parameters
~~~~~~~~~~~~~~~~~~~~

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``fd-limit``
            -   ``LimitNOFILE`` for an application instance
        *   -   ``stateboard-fd-limit``
            -   ``LimitNOFILE`` for a stateboard instance
        *   -   ``instance-env``
            -   :doc:`cartridge.argparse </book/cartridge/cartridge_api/modules/cartridge.argparse>`
                environment variables (like ``net-msg-max``) for an application instance
        *   -   ``stateboard-env``
            -   :doc:`cartridge.argparse </book/cartridge/cartridge_api/modules/cartridge.argparse>`
                environment variables (like ``net-msg-max``) for a stateboard instance

Example
~~~~~~~

``systemd-unit-params.yml``:

..  code-block:: yaml

    fd-limit: 1024
    stateboard-fd-limit: 2048
    instance-env:
        app-name: 'my-app'
        net_msg_max: 1024
        pid_file: '/some/special/dir/my-app.%i.pid'
        my-param: 'something'
        # or
        # TARANTOOL_MY_PARAM: 'something'
    stateboard-env:
        app-name: 'my-app-stateboard'
        pid_file: '/some/special/dir/my-app-stateboard.pid'

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
instances, you might put the following ``myapp.yml`` file
in the ``/etc/tarantool/conf.d`` directory:

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
:ref:`configuring Cartridge application instances <cartridge-config-basic>`.

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

test

