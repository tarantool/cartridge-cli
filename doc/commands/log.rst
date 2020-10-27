===============================================================================
Get logs of instance running locally in background
===============================================================================

To get logs of the instance running in background, use the ``log`` command:

.. code-block:: bash

    cartridge log [INSTANCE_NAME...] [flags]

-------------------------------------------------------------------------------
Options
-------------------------------------------------------------------------------

* ``-f, --follow`` outputs appended data as the log grows.

* ``-n, --lines int`` is the number of lines to output (from the end).
  Defaults to 15.

* ``--stateboard`` get stateboard logs as well as instances.
  Ignored if ``--stateboard-only`` is specified.

* ``--stateboard-only`` get only stateboard logs.
  If specified, ``INSTANCE_NAME...`` are ignored.

* ``--log-dir`` is the directory to store instances logs
  when running in background.
  Defaults to ``./tmp/log``.
  ``.cartridge.yml`` section: ``log-dir``.
  See `instances paths doc <doc/instances_paths.rst>`_ for details.

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

   ``log-dir`` should be exactly the same as used in the ``cartridge start``
   command. Log files stored there are used to get instances logs.
