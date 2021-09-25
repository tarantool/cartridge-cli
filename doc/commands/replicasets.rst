.. _cartridge-cli.replicasets:

Setting up replicasets
======================

The ``cartridge replicasets`` command is used to configure replica sets on local start.

Usage
-----

..  code-block:: bash

    cartridge replicasets [command] [flags] [args]

The following flags work with any ``replicasets`` subcommand:

..  container:: table

    ..  list-table::
        :widths: 20 80
        :header-rows: 0

        *   -   ``--name``
            -   Application name.
        *   -   ``--run-dir``
            -   The directory where PID and socket files are stored.
                Defaults to ``./tmp/run`` or the ``run-dir`` value in ``.cartridge.yml``.
        *   -   ``--cfg``
            -   Instances' configuration file.
                Defaults to ``./instances.yml`` or the ``cfg`` value in ``.cartridge.yml``.


How it works
------------

Replicasets are configured using the Cartridge Lua API.
All instances in the topology are described in a single file,
``instances.yml`` (see the ``--cfg`` flag).
The instances receive their configuration through instance console sockets
that can be found in the run directory.

First, all the running instances mentioned in ``instances.yml`` are organized into a
:ref:`membership <https://www.tarantool.io/en/doc/latest/reference/reference_rock/membership/>`
network.
In this way, Cartridge checks if there are any instances that have already joined the cluster.
One of these instances is then used to perform cluster operations.


Subcommands
-----------

..  contents::
    :local:

setup
~~~~~

..  code-block:: bash

    cartridge replicasets setup [flags]

Setup replica sets using a file.

Flags:

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``--file``
            -   File with replica set configuration.
                Defaults to ``replicasets.yml``.
        *   -   ``--bootstrap-vshard``
            -   Bootstrap vshard upon setup.

Example configuration:

..  code-block:: yaml

    router:
      instances:
      - router
      roles:
      - vshard-router
      - app.roles.custom
    s-1:
      instances:
      - s1-master
      - s1-replica
      roles:
      - vshard-storage
      weight: 11
      all_rw: false
      vshard_group: default

All the instances should be described in ``instances.yml`` (or another file passed via
``--cfg``).


save
~~~~

..  code-block:: bash

    cartridge replicasets save [flags]

Saves the current replica set configuration to a file.

Flags:

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -  ``--file``
            -  The file to save the configuration to.
               Defaults to ``replicasets.yml``.

list
~~~~

..  code-block:: bash

    cartridge replicasets list [flags]

Lists the current cluster topology.

..  _cartridge-cli_replicasets-join:

join
~~~~

..  code-block:: bash

    cartridge replicasets join [INSTANCE_NAME...] [flags]

Joins an instance to a cluster.

Flags:

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``--replicaset``
            -   Name of the replica set

If a replica set with the specified alias isn't found in cluster, it is created.
Otherwise, instances are joined to an existing replica set.

To join an instance to a replica set, Cartridge requires the instance to have an
`advertise_uri <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_dev/#configuration-basics>`__.
These parameters should be described in ``instances.yml``.

list-roles
~~~~~~~~~~

..  code-block:: bash

    cartridge replicasets list-roles [flags]

List the available roles.
..  // what does this mean?

list-vshard-groups
~~~~~~~~~~~~~~~~~~

..  code-block:: bash

    cartridge replicasets list-vshard-groups [flags]

List the available vshard groups.

..  _cartridge-cli_replicasets-add-roles:

add-roles
~~~~~~~~~

..  code-block:: bash

    cartridge replicasets add-roles [ROLE_NAME...] [flags]

Add roles to the replica set.

Flags:

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``--replicaset``
            -   Name of the replica set
        *   -   ``--vshard-group``
            -   Vshard group for ``vshard-storage`` replica sets



remove-roles
~~~~~~~~~~~~

..  code-block:: bash

    cartridge replicasets remove-roles [ROLE_NAME...] [flags]

Remove roles from the replica set.

Flags:

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``--replicaset``
            -   Name of the replica set

..  _cartridge-cli_replicasets-set-weight:

set-weight
~~~~~~~~~~

..  code-block:: bash

    cartridge replicasets set-weight WEIGHT [flags]

Specify the weight of the replica set.

Flags:

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``--replicaset``
            -   Name of the replica set

set-failover-priority
~~~~~~~~~~~~~~~~~~~~~

..  code-block:: bash

    cartridge replicasets set-failover-priority INSTANCE_NAME... [flags]

Configure replica set failover priority.

Flags:

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``--replicaset``
            -   Name of the replica set

bootstrap-vshard
~~~~~~~~~~~~~~~~

..  code-block:: bash

    cartridge replicasets bootstrap-vshard [flags]

Bootstrap vshard.

expel
~~~~~

..  code-block:: bash

    cartridge replicasets expel [INSTANCE_NAME...] [flags]

Expel one or more instances.


Examples
--------

We'll use an application created via ``cartridge create``.
Here is its ``instances.yml`` file:

..  code-block:: yaml

    ---
    myapp.router:
    advertise_uri: localhost:3301
    http_port: 8081

    myapp.s1-master:
    advertise_uri: localhost:3302
    http_port: 8082

    myapp.s1-replica:
    advertise_uri: localhost:3303
    http_port: 8083

    # other instances are hidden in this example

Create two replicasets
~~~~~~~~~~~~~~~~~~~~~~

Join instances:

..  code-block:: bash

    cartridge replicasets join --replicaset s-1 s1-master s1-replica

        • Join instance(s) s1-master, s1-replica to replica set s-1
        • Instance(s) s1-master, s1-replica have been successfully joined to replica set s-1

    cartridge replicasets join --replicaset router router

        • Join instance(s) router to replica set router
        • Instance(s) router have been successfully joined to replica set router

List the available roles:

..  code-block:: bash

    cartridge replicasets list-roles

        •   Available roles:
        •   failover-coordinator
        •   vshard-storage
        •   vshard-router
        •   metrics
        •   app.roles.custom

Set roles:

..  code-block:: bash

    cartridge replicasets add-roles --replicaset s-1 vshard-storage

        • Add role(s) vshard-storage to replica set s-1
        • Replica set s-1 now has these roles enabled:
        •   vshard-storage (default)

    cartridge replicasets add-roles \
      --replicaset router \
      vshard-router app.roles.custom failover-coordinator metrics

        • Add role(s) vshard-router, app.roles.custom, failover-coordinator, metrics to replica set router
        • Replica set router now has these roles enabled:
        •   failover-coordinator
        •   vshard-router
        •   metrics
        •   app.roles.custom

Bootstrap vshard:

..  code-block:: bash

    cartridge replicasets bootstrap-vshard

        • Vshard is bootstrapped successfully

List current replica sets:

..  code-block:: bash

    cartridge replicasets list

        • Current replica sets:
    • router
    Role: failover-coordinator | vshard-router | metrics | app.roles.custom
        ★ router localhost:3301
    • s-1                    default | 1
    Role: vshard-storage
        ★ s1-master localhost:3302
        • s1-replica localhost:3303

Expel an instance:

..  code-block:: bash

    cartridge replicasets expel s1-replica

        • Instance(s) s1-replica have been successfully expelled
