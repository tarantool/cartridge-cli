.. _cartridge-cli.admin:

===============================================================================
Running admin functions
===============================================================================

``cartridge admin`` command is used to call admin function provided by application.

-------------------------------------------------------------------------------
Usage
-------------------------------------------------------------------------------

.. code-block:: bash

    cartridge admin [ADMIN_FUNC_NAME] [flags]

Command flags:

* ``--name`` - application name (required)
* ``--list`` - list available admin functions
* ``--help`` - help for admin function
* ``--instance`` - name of instance to connect to
* ``--run-dir`` - directory where instance's sockets are placed
  (defaults to ``/var/run/tarantool``)

-------------------------------------------------------------------------------
How does it work?
-------------------------------------------------------------------------------

Your application can provide *admin functions* that should be registered using
`admin extension <https://github.com/tarantool/cartridge-cli-extensions/blob/master/doc/admin.md>`_.
The default application contains the
`probe <https://github.com/tarantool/cartridge-cli-extensions/blob/master/doc/admin.md#example>`_
function that probes the instance by specified the URI.

.. NOTE::

    If your function calls ``print``, message is displayed on ``cartridge admin``
    call (since ``cartridge-cli-extensions``
    `1.1.0 <https://github.com/tarantool/cartridge-cli-extensions/releases/tag/1.1.0>`_).


.. NOTE::

    Your admin functions shouldn't accept arguments with names
    that conflict with ``cartridge admin`` options' names:

    * ``name``
    * ``list``
    * ``help``
    * ``instance``
    * ``run_dir``
    * ``debug``
    * ``quiet``
    * ``verbose``

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Connecting to instance
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

* If the ``--instance`` flag is specified, CLI checks if the
  ``<run-dir>/<name>.<instance>.control`` socket is *available* and if so,
  uses it to run an admin command.

* Otherwise, CLI checks all ``<run-dir>/<name>.*.control`` sockets and uses the
  first *available* socket to run an admin command.

What does *available* socket mean?
It means that it's possible to connect to the socket.
To make search for an available socket more clear, use ``--verbose`` flag.

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Example
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

This example shows the usage of the
`probe <https://github.com/tarantool/cartridge-cli-extensions/blob/master/doc/admin.md#example>`_
admin function.

*******************************************************************************
Get functions help
*******************************************************************************

Get a list of available admin functions:

.. code-block:: bash

    cartridge admin --name APPNAME --list

       • Available admin functions:

    probe  Probe instance


Get help for a specific function:

.. code-block:: bash

    cartridge admin --name APPNAME probe --help

       • Admin function "probe" usage:

    Probe instance

    Args:
      --uri string  Instance URI

*******************************************************************************
Call an admin function
*******************************************************************************

Call a function with an argument:

.. code-block:: bash

    cartridge admin --name APPNAME probe --uri localhost:3301

       • Probe "localhost:3301": OK
