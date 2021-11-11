Stopping instances
==================

To stop one or more instances that are running locally in the background, run:

..  code-block:: bash

    cartridge stop [INSTANCE_NAME...] [flags]

where ``[INSTANCE_NAME...]`` means that more than one instance can be specified.

If no ``INSTANCE_NAME`` is provided, all the instances from the
Cartridge instance configuration file are taken as arguments.
See the ``--cfg`` option below.

Flags
-----

..  container:: table

    ..  list-table::
        :widths: 20 80
        :header-rows: 0

        *   -   ``--name``
            -   Application name.
                By default, it is taken from the ``package`` field
                of the application's ``.rockspec``.
        *   -   ``-f, --force``
            -   Force stop the instance(s) with a SIGKILL.
                By default, the instances receive a SIGTERM.
        *   -   ``--stateboard``
            -   Stop the application
                :ref:`stateboard <cartridge-stateful_failover>`
                and the instances.
                Ignored if ``--stateboard-only`` is specified.
        *   -   ``--stateboard-only``
            -   Stop only the application stateboard.
                If specified, ``INSTANCE_NAME...`` is ignored.
        *   -   ``--run-dir``
            -   The directory where PID and socket files are stored.
                Defaults to ``./tmp/run``.
                ``run-dir`` is also a section of ``.cartridge.yml``.
                Learn more about
                :doc:`instance paths </book/cartridge/cartridge_cli/instance-paths>`.
        *   -   ``--cfg``
            -   Path to the Cartridge instances configuration file.
                Defaults to ``./instances.yml``.
                ``cfg``is also a section of ``.cartridge.yml``.
                Learn more about
                :doc:`instance paths </book/cartridge/cartridge_cli/instance-paths>`.

..  note::

    Use the exact same ``run-dir`` as you did with ``cartridge start``.
    The PID files stored in that directory are used to stop running instances.

