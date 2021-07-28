.. _cartridge-cli.failover:

===============================================================================
Configure Cartridge failover
===============================================================================

The ``cartridge failover`` command is used to configure Cartridge failover.

-------------------------------------------------------------------------------
Usage
-------------------------------------------------------------------------------

.. code-block:: bash

    cartridge failover [command] [flags] [args]

All ``failover`` sub-commands have these flags:

* ``--name`` - application name

-------------------------------------------------------------------------------
How it works
-------------------------------------------------------------------------------

Failover is configured using Cartridge Lua API.
All failover settings placed in ``failover.yml`` file (see ``--file`` flag).

To run the failover, the ``cartridge-cli`` connects to a random configured instance,
so you must have a topology configured (see
`cartridge replicasets <https://github.com/tarantool/cartridge-cli/blob/master/doc/replicasets.rst>`_ command)

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Failover parameters
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

* ``mode`` (required) - failover mode. Possible values are disabled, eventual and stateful.
* ``failover_timeout`` - timeout (in seconds), used by membership to mark suspect members as dead;
* ``fencing_enabled`` - abandon leadership when both the state provider quorum and at least one replica are lost (suitable in ``stateful`` mode only);
* ``fencing_timeout`` - time (in seconds) to actuate fencing after the check fails;
* ``fencing_pause`` - the period (in seconds) of performing the check;

Other parameters are mode-specific.

Read the `doc <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_dev/#failover-architecture>`_
to learn more about Cartridge failover.

"""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""
Eventual failover
"""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

If ``eventual`` mode is specified, there are no additional parameters.

Read the `doc <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_dev/#eventual-failover>`_
to learn more about ``eventual`` failover.

"""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""
Stateful failover
"""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""

``stateful`` failover requires these parameters:

* ``state_provider`` - external state provider type. Supported ``stateboard`` and ``etcd2`` providers.
* ``stateboard_params`` - configuration for stateboard:

  * ``uri`` (required) - stateboard instance URI;
  * ``password`` (required) - stateboard instance password;

* ``etcd2_params`` - configuration for etcd2:

  * ``prefix`` - prefix used for etcd keys: <prefix>/lock and <prefix>/leaders;
  * ``lock_delay`` - timeout (in seconds), determines lock's time-to-live (default value in Cartridge is 10);
  * ``endpoints`` - URIs that are used to discover and to access etcd cluster instances (default value in Cartridge is ['http://localhost:2379', 'http://localhost:4001']);
  * ``username``
  * ``password``

Read the `doc <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_dev/#stateful-failover>`_
to learn more about ``stateful`` failover.

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Configure failover described in a file
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    cartridge failover setup [flags]

Flags:

* ``--file`` - file where failover configuration is described
  (defaults to ``failover.yml``)

Example configuration:

.. code-block:: yaml

    mode: stateful
    state_provider: stateboard
    stateboard_params:
        uri: localhost:4401
        password: passwd
    failover_timeout: 15

For convenience, you can leave extra parameters. For example, suppose you want to configure a
``stateful stateboard`` failover instead of ``stateful etcd2`` failover. In this case, you can
leave the ``etcd2_params`` from the file and just add ``stateboard_params`` and change the
``state_provider``. Later, you wanted to switch the failover to eventual mode. You can also
not remove ``etcd2_params`` and ``stateboard_params`` from configuration file.

But, be careful: all parameters (``etcd2_params`` and ``stateboard_params`` when you specify
``eventual`` mode from example above) described in the configuration file will be applied anyway
on the Cartridge side.

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Configure failover with specified mode
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    cartridge replicasets set [mode] [flags]

Mode:

* ``stateful`` - stateful failover mode
* ``eventual`` - eventual failover mode
* ``disabled`` - disabled failover mode

Flags:

* ``--state-provider`` - failover state provider, can be ``stateboard`` or ``etcd2``. Used only for ``stateful`` mode
* ``--params`` - failover parameters, described in JSON-formatted string, for example ``"{'fencing_timeout': 10', 'fencing_enabled': true}"``
* ``--provider-params`` - failover provider parametrs, described in JSON-formatted string, for example ``"{'lock_delay': 14}"``

Unlike the ``setup`` command, you shouldn't pass unnecessary parameters. For example, you shouldn't
specify ``--state-provider`` flag when the mode is ``eventual``, otherwise you will get an error.

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Disable failover
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    cartridge failover disable [flags]


You can also disable failover with the ``set`` and ``setup`` commands
specifying ``disabled`` mode.

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
See current failover status
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    cartridge failover status [flags]
