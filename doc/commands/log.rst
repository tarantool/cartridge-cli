Get instance logs
=================

To get the logs of an instance running in the background, use the ``log`` command:

..  code-block:: bash

    cartridge log [INSTANCE_NAME...] [flags]

which means that more than one instance name can be specified.

Options
-------

..  container:: table

    ..  list-table::
        :widths: 20 80
        :header-rows: 0

        *   -   ``-f, --follow``
            -   Output appended data as the log grows.
        *   -   ``-n, --lines int``
            -   Number of last lines to be displayed. Defaults to 15.
        *   -   ``--stateboard``
            -   Get both stateboard and instance logs.
                Ignored if ``--stateboard-only`` is specified.
        *   -   ``--stateboard-only``
            -   Get only stateboard logs.
                If specified, ``INSTANCE_NAME...`` is ignored.
        *   -   ``--log-dir``
            -   The directory that stores logs for instances that are running in the background.
                Defaults to ``./tmp/log``.
                ``log-dir`` is also a section of ``.cartridge.yml``.
                Learn more about
                :doc:`instance paths </book/cartridge/cartridge_cli/instance_paths>`.
        *   -   ``--run-dir``
            -   The directory where PID and socket files are stored.
                Defaults to ``./tmp/run``.
                ``run-dir`` is also a section of ``.cartridge.yml``.
                Learn more about
                :doc:`instance paths </book/cartridge/cartridge_cli/instance_paths>`.
        *   -   ``--cfg``
            -   Path to the Cartridge instances configuration file.
                Defaults to ``./instances.yml``.
                ``cfg``is also a section of ``.cartridge.yml``.
                Learn more about
                :doc:`instance paths </book/cartridge/cartridge_cli/instance_paths>`.

``log`` also supports :doc:`global flags </book/cartridge/cartridge_cli/commands/global_flags>`.

..  note::

    Use the exact same ``log-dir`` as you did with ``cartridge start``.
    The logs are retrieved from the files stored in that directory.
