Enter an instance
=================

``cartridge enter`` allows connecting to an instance started with ``cartridge start``.
The connection uses the instance's console socket placed in ``run-dir``.

..  code-block:: bash

    cartridge enter [INSTANCE_NAME] [flags]

Flags
-----

..  container:: table

    ..  list-table::
        :widths: 20 80
        :header-rows: 0

        *   -   ``--name``
            -   Application name.
        *   -   ``--run-dir``
            -   The directory where PID and socket files are stored.
                Defaults to ``./tmp/run``.
                ``run-dir`` is also a section of ``.cartridge.yml``.
                Learn more about
                :doc:`instance paths </book/cartridge/cartridge_cli/instance_paths>`.
