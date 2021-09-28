..  _cartridge-cli_enter:

Connect to running instances
============================

Enter an instance
-----------------

``cartridge enter`` allows connecting to an instance started with ``cartridge start``.
The connection uses the instance's console socket placed in ``run-dir``.

..  code-block:: bash

    cartridge enter INSTANCE_NAME [flags]

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

..  _cartridge-cli_connect:

Connect to an instance at a specific address
--------------------------------------------

.. code-block:: bash

    cartridge connect URI [flags]

Specify the instance's address or path to its UNIX socket.
Username and password can be passed as part of the URI
or via the following flags (has greater priority):

* ``-u, --username``
* ``-p, --password``
