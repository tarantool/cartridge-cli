===============================================================================
Starting application instances locally
===============================================================================

For local development, you can start application instances locally from
application directory.

.. code-block:: bash

    cartridge start [INSTANCE_NAME...] [flags]

where ``[INSTANCE_NAME...]`` means that several instances can be specified.

If no ``INSTANCE_NAME`` is provided, all the instances from the
Cartridge instances configuration file are taken as arguments (see the ``--cfg``
option below).

We also need an application name (``APP_NAME``) to pass it to the instances while
started.
By default, the ``APP_NAME`` is taken from the ``package`` field of application
rockspec placed in the current directory, but also it can be defined explicitly
via the ``--name`` option (see description below).

-------------------------------------------------------------------------------
Options
-------------------------------------------------------------------------------

* ``--name`` defines the application name.
  By default, it is taken from the ``package`` field of application rockspec.

* ``--timeout`` Time to wait for instance(s) start in background.
  Can be specified in seconds or in the duration form (``72h3m0.5s``).
  Timeout can't be negative.
  Timeout ``0`` means no timeout (wait for instance(s) start forever).
  The default timeout is 60 seconds (``1m0s``).

* ``-d, --daemonize`` starts the instance in background.
  With this option, Tarantool also waits until the application's main script is
  finished.
  For example, it is useful if the ``init.lua`` requires time-consuming startup
  from snapshot, and Tarantool waits for the startup to complete.
  This is also useful if the application's main script generates errors, and
  Tarantool can handle them.

* ``--stateboard`` starts the application stateboard as well as instances.
  Ignored if ``--stateboard-only`` is specified.

* ``--stateboard-only`` starts only the application stateboard.
  If specified, ``INSTANCE_NAME...`` are ignored.

* ``--script`` is the application's entry point.
  The default value is ``init.lua`` in the project root.
  ``.cartridge.yml`` section: ``script``.
  See `instances paths doc <doc/instances_paths.rst>`_ for details.

* ``--run-dir`` is the directory where PID and socket files are stored.
  Defaults to ``./tmp/run``.
  ``.cartridge.yml`` section: ``run-dir``.
  See `instances paths doc <doc/instances_paths.rst>`_ for details.

* ``--data-dir`` is the directory where instances' working directories are placed.
  Defaults to ``./tmp/data``.
  ``.cartridge.yml`` section: ``data-dir``.
  See `instances paths doc <doc/instances_paths.rst>`_ for details.

* ``--log-dir`` is the directory to store instances logs
  when running in background.
  Defaults to ``./tmp/log``.
  ``.cartridge.yml`` section: ``log-dir``.
  See `instances paths doc <doc/instances_paths.rst>`_ for details.

* ``--cfg`` is the Cartridge instances configuration file.
  Defaults to ``./instances.yml``.
  ``.cartridge.yml`` section: ``cfg``.
  See `instances paths doc <doc/instances_paths.rst>`_ for details.

Command also supports `global flags <./global_flags.rst>`_.

-------------------------------------------------------------------------------
Details
-------------------------------------------------------------------------------

The ``cartridge start`` command starts a Tarantool instance with enforced
**environment variables**:

.. code-block:: bash

    TARANTOOL_APP_NAME="<name>"
    TARANTOOL_INSTANCE_NAME="<instance-name>"
    TARANTOOL_CFG="<cfg>"
    TARANTOOL_PID_FILE="<run-dir>/<app-name>.<instance-name>.pid"
    TARANTOOL_CONSOLE_SOCK="<run-dir>/<app-name>.<instance-name>.control"
    TARANTOOL_WORKDIR="<data-dir>/<app-name>.<instance-name>.control"

When started in background, a notify socket path is passed additionally:

.. code-block:: bash

    NOTIFY_SOCKET="<data-dir>/<app-name>.<instance-name>.notify"

``cartridge.cfg()`` uses  ``TARANTOOL_APP_NAME`` and ``TARANTOOL_INSTANCE_NAME``
to read the instance's configuration from the file provided in ``TARANTOOL_CFG``.
