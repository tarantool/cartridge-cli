.. _cartridge-cli.connect:

===============================================================================
Connect to running instances
===============================================================================

-------------------------------------------------------------------------------
Enter instance started via ``cartridge start``
-------------------------------------------------------------------------------

.. code-block:: bash

    cartridge enter INSTANCE_NAME [flags]

Flags:

* ``--name`` - application name
* ``--run-dir`` - directory where PID and socket files are stored
  (defaults to ./tmp/run or "run-dir" in .cartridge.yml)

Connects to instance via it's console socket placed in ``run-dir``.

-------------------------------------------------------------------------------
Connect to instance by specified address
-------------------------------------------------------------------------------

.. code-block:: bash

    cartridge connect URI [flags]

Instance address or path to UNIX socket can be specified.
Username and password can be passed as a part of URI or by flags (has greater priority):

* ``-u, --username``
* ``-p, --password``
