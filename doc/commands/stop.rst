Stopping instances running locally in background
================================================

To stop one or more running instances, say:

.. code-block:: bash

    cartridge stop [INSTANCE_NAME...] [flags]

where ``[INSTANCE_NAME...]`` means that several instances can be specified.

If no ``INSTANCE_NAME`` is provided, all the instances from the
Cartridge instances configuration file are taken as arguments (see the ``--cfg``
option below).

We also need an application name (``APP_NAME``) to pass it to the instances while
started.
By default, the ``APP_NAME`` is taken from the ``package`` field of application
rockspec placed in the current directory, but also it can be defined explicitly
via the ``--name`` option (see description below).

Options
-------

* ``--name`` defines the application name.
  By default, it is taken from the ``package`` field of application rockspec.

* ``-f, --force`` indicates if instance(s) stop should be forced (sends SIGKILL).

* ``--stateboard`` stops application stateboard as well as instances.
  Ignored if ``--stateboard-only`` is specified.

* ``--stateboard-only`` stops only the application stateboard.
  If specified, ``INSTANCE_NAME...`` are ignored.

* ``--run-dir`` is the directory where PID and socket files are stored.
  Defaults to ``./tmp/run``.
  ``.cartridge.yml`` section: ``run-dir``.
  See `instances paths doc <doc/instances_paths.rst>`_ for details.

* ``--cfg`` is the Cartridge instances configuration file.
  Defaults to ``./instances.yml``.
  ``.cartridge.yml`` section: ``cfg``.
  See `instances paths doc <doc/instances_paths.rst>`_ for details.

.. note::

   ``run-dir`` should be exactly the same as used in the ``cartridge start``
   command. PID files stored there are used to stop the running instances.
