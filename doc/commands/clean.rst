Cleaning up instance files
==========================

Locally running instances create a number of files,
such as the log file, the workdir, the console socket, the PID file, and the notify socket.
To remove all of these files for one or more instances, use the ``clean`` command:

..  code-block:: bash

    cartridge clean [INSTANCE_NAME...] [flags]

where ``[INSTANCE_NAME...]`` means that more than one instance name can be specified.

If no ``INSTANCE_NAME`` is provided, all the instances from the
Cartridge instance configuration file are taken as arguments.
See the ``--cfg`` option below.

Flags
-----

..  container:: table

    ..  list-table::
        :widths: 20 80
        :header-rows: 0

        *   -   ``--stateboard``
            -   Remove the stateboard files as well as the files of other instances.
                Ignored if ``--stateboard-only`` is specified.
        *   -   ``--stateboard-only``
            -   Remove only the application stateboard files.
                If this flag is specified, ``INSTANCE_NAME...`` is ignored.
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

``clean`` also supports :doc:`global flags </book/cartridge/cartridge_cli/commands/global_flags>`.
