Configuring Cartridge failover
==============================

The ``cartridge failover`` command lets you configure Cartridge failover.

..  code-block:: bash

    cartridge failover [subcommand] [flags] [args]

Flags
-----

..  container:: table

    ..  list-table::
        :widths: 20 80
        :header-rows: 0

        *   -   ``--name``
            -   Application name.
        *   -   ``--file``
            -   Path to the file containing failover settings.
                Defaults to ``failover.yml``.

``failover`` also supports :doc:`global flags </book/cartridge/cartridge_cli/global-flags>`.


Details
-------

Failover is configured through the Cartridge Lua API.

To run the failover, ``cartridge-cli`` connects to a random configured instance,
so you must have a topology configured.
To learn more, see the
:doc:`cartridge replicasets </book/cartridge/cartridge_cli/commands/replicasets>` command.
You might also want to check out the documentation on
:ref:`Cartridge failover architecture <cartridge-failover>`.

You can manage failover in the following ways:

*   :ref:`Set a specific failover mode <cartridge-cli_failover-set>`
    with ``cartridge failover set``, passing the parameters via special flags.
*   Specify parameters through a :ref:`configuration file <cartridge-cli_failover-setup>`
    and make it the default file with ``cartridge failover setup``.
*   :ref:`Check failover status <cartridge-cli_failover-status>` with ``status``.
*   :ref:`Disable failover <cartridge-cli_failover-disable>` with ``disable``.


Subcommands
-----------

..  contents::
    :depth: 1
    :local:

..  _cartridge-cli_failover-set:

set
~~~

..  code-block:: bash

    cartridge failover set [mode] [flags]

This command lets you set a failover mode. Learn more about
:ref:`Cartridge failover modes <cartridge-failover>`.

Modes
^^^^^

* ``stateful``
* ``eventual``
* ``disabled``

Flags
^^^^^

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``--state-provider``
            -   Failover state provider. Can be ``stateboard`` or ``etcd2``.
                Used only in the ``stateful`` mode.
        *   -   ``--params``
            -   Failover parameters. Described in a JSON-formatted string like
                ``"{'fencing_timeout': 10', 'fencing_enabled': true}"``.
        *   -   ``--provider-params``
            -   Failover provider parameters. Described in a JSON-formatted string like
                ``"{'lock_delay': 14}"``.

To learn more about the parameters, check the corresponding
:ref:`section of this document <cartridge-cli_failover-parameters>`.

Unlike in the case with ``setup``, don't pass unnecessary parameters.
For example, don't specify the ``--state-provider`` flag
when the mode is ``eventual``, otherwise you will get an error.

..  _cartridge-cli_failover-setup:

setup
~~~~~

..  code-block:: bash

    cartridge failover setup --file [configuration file]

The failover configuration file defaults to ``failover.yml``.
See the :ref:`full description of parameters <cartridge-cli_failover-parameters>`
to include in the failover configuration.

Example
^^^^^^^

..  code-block:: yaml

    mode: stateful
    state_provider: stateboard
    stateboard_params:
        uri: localhost:4401
        password: passwd
    failover_timeout: 15

You can leave extra parameters in the file, which may be convenient.
Suppose you have ``stateful etcd2`` failover configured
and want to change it to ``stateful stateboard``.
You don't have to delete ``etcd2_params`` from the file, but you can just
add ``stateboard_params`` and change the ``state_provider``.
Then you might want to switch the failover to the ``eventual`` mode.
This doesn't require removing ``etcd2_params`` or ``stateboard_params``
from the configuration file either.

However, be careful: all the parameters described in the configuration file
will be applied on the Cartridge side. Thus, ``etcd2_params`` and ``stateboard_params``
from the example above will still be applied in the ``eventual`` mode,
although they are intended for use with the ``stateful`` mode.

..  _cartridge-cli_failover-status:

status
~~~~~~

..  code-block:: bash

    cartridge failover status [flags]

Checks failover status.

..  _cartridge-cli_failover-disable:

disable
~~~~~~~

..  code-block:: bash

    cartridge failover disable [flags]

Disables failover.
Another way to disable failover is to specify the ``disabled`` mode
with :ref:`set <cartridge-cli_failover-set>`
or in the :ref:`configuration file <cartridge-cli_failover-setup>` (see above).


..  // these are JSON parameters. Move to a separate file?

..  _cartridge-cli_failover-parameters:

Failover parameters
-------------------

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``mode``
            -   (Required) Failover mode.
                Possible values: ``disabled``, ``eventual``, ``stateful``.
        *   -   ``failover_timeout``
            -   Timeout in seconds used by membership to mark suspect members as dead.
        *   -   ``fencing_enabled``
            -   Abandon leadership when both the state provider quorum
                and at least one replica are lost. Works for ``stateful`` mode only.
        *   -   ``fencing_timeout``
            -   Time in seconds to actuate fencing after the check fails.
        *   -   ``fencing_pause``
            -   Period in seconds to perform the check.

Other parameters are mode-specific.


Eventual failover
~~~~~~~~~~~~~~~~~

If the ``eventual`` mode is specified, no additional parameters are required.

Read the :ref:`documentation <cartridge-failover>`
to learn more about ``eventual`` failover.


Stateful failover
~~~~~~~~~~~~~~~~~

``stateful`` failover requires the following parameters:

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``state_provider``
            -   External state provider type.
                Supported providers: ``stateboard``, ``etcd2``.
        *   -   ``stateboard_params``
            -   Stateboard configuration:

                *   ``uri`` (required): Stateboard instance URI.
                *   ``password`` (required): Stateboard instance password.
                
        *   -   ``etcd2_params``
            -   Configuration for etcd2:

                *   ``prefix``: Prefix for etcd keys (<prefix>/lock and <prefix>/leaders).
                *   ``lock_delay``: Timeout in seconds.
                    Defines the lock's time-to-live. Default value in Cartridge is ``10``.
                *   ``endpoints``: URIs used to discover and access
                    etcd cluster instances. Default value in Cartridge is
                    ``['http://localhost:2379', 'http://localhost:4001']``.
                *   ``username``
                *   ``password``

Read the :ref:`documentation <cartridge-failover>`
to learn more about ``stateful`` failover.
