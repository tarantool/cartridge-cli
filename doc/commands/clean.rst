===============================================================================
Cleanup application instances running locally files
===============================================================================

To remove instance(s) running locally files (log, workdir, console socket, PID-file and notify socket),
use the ``clean`` command:

.. code-block:: bash

    cartridge clean [INSTANCE_NAME...] [flags]

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

* ``--stateboard`` removes stateboard files as well as other instances.
  Ignored if ``--stateboard-only`` is specified.

* ``--stateboard-only`` removes only the application stateboard files.
  If specified, ``INSTANCE_NAME...`` are ignored.

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
