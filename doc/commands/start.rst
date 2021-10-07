Starting application instances locally
======================================

You can start application instances for local development from the application directory:

..  code-block:: bash

    cartridge start [INSTANCE_NAME...] [flags]

where ``[INSTANCE_NAME...]`` means that more than one instance can be specified.

If no ``INSTANCE_NAME`` is provided, all the instances from the
Cartridge instance configuration file are taken as arguments.
See the ``--cfg`` option below.

During instance startup, the application name (``APP_NAME``) is passed to the instance.
By default, this variable is taken from the ``package`` field of the application's ``.rockspec``.
However, it can also be defined explicitly via the ``--name`` option (see description below).

Flags
-----

..  container:: table

    ..  list-table::
        :widths: 20 80
        :header-rows: 0

        *   -   ``--name``
            -   Application name.
                By default, it is taken from the ``package`` field of the application's ``.rockspec``.
        *   -   ``--timeout``
            -   Time to wait for the instance(s) to start in the background.
                Can be specified in seconds or in the duration form (``72h3m0.5s``).
                Can't be negative.
                A ``0`` timeout means that Tarantool will wait forever for the instance(s) to start.
                The default timeout is 60 seconds (``1m0s``).
        *   -   ``-d, --daemonize``
            -   Start the instance(s) in the background.
                With this option, Tarantool also waits until the application's init script
                finishes evaluating.
                This is useful if ``init.lua`` requires time-consuming startup
                from a snapshot, and Tarantool has to wait for the startup to complete.
                Another use case would be if your application's init script
                generates errors, so Tarantool can handle them.
        *   -   ``--stateboard``
            -   Start the application stateboard and the instances.
                Ignored if ``--stateboard-only`` is specified.
        *   -   ``--stateboard-only``
            -   Start only the application stateboard.
                If specified, ``INSTANCE_NAME...`` is ignored.
        *   -   ``--script``
            -   Application entry point.
                The default value is ``init.lua`` in the project root directory.
                ``script`` is also a section in ``.cartridge.yml``.
                Learn more about
                :doc:`instance paths </book/cartridge/cartridge_cli/instance_paths>`.
        *   -   ``--run-dir``
            -   The directory where PID and socket files are stored.
                Defaults to ``./tmp/run``.
                ``run-dir`` is also a section of ``.cartridge.yml``.
                Learn more about
                :doc:`instance paths </book/cartridge/cartridge_cli/instance_paths>`.
        *   -   ``--data-dir``
            -   The directory containing the working directories of instances.
                Defaults to ``./tmp/data``.
                ``data-dir`` is also a section of ``.cartridge.yml``.
                Learn more about
                :doc:`instance paths </book/cartridge/cartridge_cli/instance_paths>`.
        *   -   ``--log-dir``
            -   The directory that stores logs for instances that are running in the background.
                Defaults to ``./tmp/log``.
                ``log-dir`` is also a section of ``.cartridge.yml``.
                Learn more about
                :doc:`instance paths </book/cartridge/cartridge_cli/instance_paths>`.
        *   -   ``--cfg``
            -   Path to the Cartridge instances configuration file.
                Defaults to ``./instances.yml``.
                ``cfg``is also a section of ``.cartridge.yml``.
                Learn more about
                :doc:`instance paths </book/cartridge/cartridge_cli/instance_paths>`.

``start`` also supports :doc:`global flags </book/cartridge/cartridge_cli/commands/global_flags>`.

Details
-------

The ``cartridge start`` command starts a Tarantool instance with enforced
**environment variables**:

..  code-block:: bash

    TARANTOOL_APP_NAME="<name>"
    TARANTOOL_INSTANCE_NAME="<instance-name>"
    TARANTOOL_CFG="<cfg>"
    TARANTOOL_PID_FILE="<run-dir>/<app-name>.<instance-name>.pid"
    TARANTOOL_CONSOLE_SOCK="<run-dir>/<app-name>.<instance-name>.control"
    TARANTOOL_WORKDIR="<data-dir>/<app-name>.<instance-name>.control"

If the instance is started in the background, a notify socket path is passed additionally:

..  code-block:: bash

    NOTIFY_SOCKET="<data-dir>/<app-name>.<instance-name>.notify"

``cartridge.cfg()`` uses  ``TARANTOOL_APP_NAME`` and ``TARANTOOL_INSTANCE_NAME``
to read the instance's configuration from the file provided in ``TARANTOOL_CFG``.

test
