Running admin functions
=======================

Use ``cartridge admin`` to call admin functions provided by the application.

..  code-block:: bash

    cartridge admin [ADMIN_FUNC_NAME] [flags]

Flags
-----

..  container:: table

    ..  list-table::
        :widths: 20 80
        :header-rows: 0

        *   -   ``--name``
            -   Application name (required)
        *   -   ``--list``
            -   List the available admin functions
        *   -   ``--help``
            -   Display help for an admin function
        *   -   ``--instance``
            -   Name of the instance to connect to
        *   -   ``--conn, -c``
            -   Address to connect to
        *   -   ``--run-dir``
            -   The directory to place the instance's sockets
                (defaults to ``/var/run/tarantool``)

``admin`` also supports :doc:`global flags </book/cartridge/cartridge_cli/global_flags>`.

Details
-------

Your application can provide *admin functions*. First, you have to register them using the
`admin extension <https://github.com/tarantool/cartridge-cli-extensions/blob/master/doc/admin.md>`_.
The example application contains a function named
`probe <https://github.com/tarantool/cartridge-cli-extensions/blob/master/doc/admin.md#example>`__,
which probes an instance at a specified URI.

..  note::

    If your function calls ``print``, the message is displayed on ``cartridge admin``
    call (since ``cartridge-cli-extensions``
    `1.1.0 <https://github.com/tarantool/cartridge-cli-extensions/releases/tag/1.1.0>`_).


..  note::

    Your admin functions shouldn't accept arguments with names
    that conflict with ``cartridge admin`` option names:

    *   ``name``
    *   ``list``
    *   ``help``
    *   ``instance``
    *   ``run_dir``
    *   ``debug``
    *   ``quiet``
    *   ``verbose``

Connecting to an instance
~~~~~~~~~~~~~~~~~~~~~~~~~

When the ``--conn`` flag is specified, CLI connects to the address provided.

When the ``--instance`` flag is specified, CLI checks if the socket
``<run-dir>/<name>.<instance>.control`` is *available* and if so,
uses it to run the admin command.
Otherwise, CLI checks all ``<run-dir>/<name>.*.control`` sockets and uses the
first *available* socket to run an admin command.

An *available* socket is one that can be connected to.
For more insight into the search for an available socket, use the ``--verbose`` flag.

Example
~~~~~~~

This example shows how to use the example admin function,
`probe <https://github.com/tarantool/cartridge-cli-extensions/blob/master/doc/admin.md#example>`__.

Get functions help
------------------

Get a list of available admin functions:

..  code-block:: bash

    cartridge admin --name APPNAME --list

       • Available admin functions:

    probe  Probe instance


Get help for a specific function:

..  code-block:: bash

    cartridge admin --name APPNAME probe --help

       • Admin function "probe" usage:

    Probe instance

    Args:
      --uri string  Instance URI

Call an admin function
----------------------

Call a function with an argument:

..  code-block:: bash

    cartridge admin --name APPNAME probe --uri localhost:3301

       • Probe "localhost:3301": OK
