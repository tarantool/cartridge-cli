===============================================================================
Checking status of instance running locally in background
===============================================================================

To check the current instance status, use the ``status`` command:

.. code-block:: bash

    cartridge status [INSTANCE_NAME...] [flags]

where ``[INSTANCE_NAME...]`` means that several instances can be specified.

If no ``INSTANCE_NAME`` is provided, all the instances from the
Cartridge instances configuration file are taken as arguments (see the ``--cfg``
option below).

We also need an application name (``APP_NAME``) to pass it to the instances while
started.
By default, the ``APP_NAME`` is taken from the ``package`` field of application
rockspec placed in the current directory, but also it can be defined explicitly
via the ``--name`` option (see description below).

.. note::

   Instance(s) should be started via ``cartridge start -d``.

-------------------------------------------------------------------------------
Options
-------------------------------------------------------------------------------

* ``--name`` defines the application name.
  By default, it is taken from the ``package`` field of application rockspec.

* ``--stateboard`` get application stateboard status as well as instances.
  Ignored if ``--stateboard-only`` is specified.

* ``--stateboard-only`` get only application stateboard status.
  If specified, ``INSTANCE_NAME...`` are ignored.

* ``--run-dir`` is the directory where PID and socket files are stored.
  Defaults to ``./tmp/run``.
  ``.cartridge.yml`` section: ``run-dir``.
  See `instances paths doc <doc/instances_paths.rst>`_ for details.

* ``--cfg`` is the Cartridge instances configuration file.
  Defaults to ``./instances.yml``.
  ``.cartridge.yml`` section: ``cfg``.
  See `instances paths doc <doc/instances_paths.rst>`_ for details.

Command also supports `global flags <./global_flags.rst>`_.

.. note::

   ``run-dir`` should be exactly the same as used in the ``cartridge start``
   command. PID files stored there are used to check instances statuses.
