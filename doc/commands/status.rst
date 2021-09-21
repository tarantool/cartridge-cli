Checking instance status
========================

Run the ``status`` command to check the current status of one or more instances:

..  code-block:: bash

    cartridge status [INSTANCE_NAME...] [flags]

where ``[INSTANCE_NAME...]`` means that more than one instance can be specified.

If no ``INSTANCE_NAME`` is provided, all the instances from the
Cartridge instance configuration file are taken as arguments.
See the ``--cfg`` option below.

..  note::

    The instance(s) you are checking should be started with ``cartridge start -d``.

Flags
-----

..  container:: table

    ..  list-table::
        :widths: 20 80
        :header-rows: 0

        *   -   ``--name``
            -   Application name.
                By default, it is taken from the ``package`` field of the application's ``.rockspec``.
        *   -   ``--stateboard``
            -   Get the status of the application stateboard and the instances.
                Ignored if ``--stateboard-only`` is specified.
        *   -   ``--stateboard-only``
            -   Get only the application stateboard status.
                If specified, ``INSTANCE_NAME...``s are ignored.
        *   -   ``--run-dir``
            -   The directory where PID and socket files are stored.
                Defaults to ``./tmp/run``.
                ``run-dir`` is also a section of ``.cartridge.yml``.
                To learn more, see the `instance paths documentation <doc/instances_paths.rst>`__.
        *   -   ``--cfg``
            -   Path to the Cartridge instances configuration file.
                Defaults to ``./instances.yml``.
                ``cfg``is also a section of ``.cartridge.yml``.
                To learn more, see the `instance paths documentation <doc/instances_paths.rst>`__.

``status`` also supports `global flags <./global_flags.rst>`_.

..  note::

    Use the exact same ``run-dir`` as you did with ``cartridge start``.
    The PID files stored in that directory are used to stop running instances.
