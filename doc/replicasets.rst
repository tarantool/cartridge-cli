.. _cartridge-cli.replicasets:

===============================================================================
Setting up replicasets
===============================================================================

The ``cartridge replicasets`` command is used to set up replica sets during local
development.

-------------------------------------------------------------------------------
Usage
-------------------------------------------------------------------------------

.. code-block:: bash

    cartridge replicasets [command] [flags] [args]

All ``replicasets`` sub-commands have these flags:

* ``--name`` - application name
* ``--run-dir`` - directory where PID and socket files are stored
  (defaults to ./tmp/run or "run-dir" in .cartridge.yml)
* ``--cfg`` - configuration file for instances
  (defaults to ./instances.yml or "cfg" in .cartridge.yml)

-------------------------------------------------------------------------------
How it works
-------------------------------------------------------------------------------

Replicasets are configured via instance console sockets placed in the run directory
using Cartridge Lua API.
All topology instances should be described in the ``instances.yml`` file (see ``--cfg``).

First, all running instances mentioned in ``instances.yml`` are connected to membership.
It's required to check if there are any instances that are already joined to cluster.
One of these instances is used to perform operations with cluster.

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Join
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    cartridge replicasets join INSTANCE_NAME... [flags]

Flags:

* ``--replicaset`` - name of replicaset

If a replica set with the specified alias isn't found in cluster, it is created.
Otherwise, instances are joined to the existing replica set.

To join an instance to the replica set, we need to know instance's advertise URI .
Advertise URIs should be described in ``instances.yml``.

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
List available roles
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    cartridge replicasets list-roles [flags]

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
List available vshard groups
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    cartridge replicasets list-vshard-groups [flags]

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Add roles to replicaset
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    cartridge replicasets add-roles ROLE_NAME... [flags]

Flags:

* ``--replicaset`` - name of replicaset
* ``--vshard-group`` - vshard group (for ``vshard-storage`` replicasets)

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Remove roles from replicaset
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    cartridge replicasets remove-roles ROLE_NAME... [flags]

Flags:

* ``--replicaset`` - name of replicaset

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Set up replica set weight
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    cartridge replicasets set-weight WEIGHT [flags]

Flags:

* ``--replicaset`` - name of replicaset

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Set up replica set failover priority
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    cartridge replicasets set-failover-priority INSTANCE_NAME... [flags]

Flags:

* ``--replicaset`` - name of replicaset

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Bootstrap vshard
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    cartridge replicasets bootstrap-vshard [flags]

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Expel instance(s)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    cartridge replicasets expel INSTANCE_NAME... [flags]

-------------------------------------------------------------------------------
Example
-------------------------------------------------------------------------------

The default application is used.
It contains the ``instances.yml`` file with the configuration of instances:

.. code-block:: yaml

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

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Create two replicasets
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Join instances:

.. code-block:: bash

    cartridge replicasets join --replicaset s-1 s1-master s1-replica

        • Join instance(s) s1-master, s1-replica to replicaset s-1
        • Instance(s) s1-master, s1-replica was successfully joined to replicaset s-1

    cartridge replicasets join --replicaset router router

        • Join instance(s) router to replicaset router
        • Instance(s) router was successfully joined to replicaset router

List available roles:

.. code-block:: bash

    cartridge replicasets list-roles

        • Available roles:
        •   failover-coordinator
        •   vshard-storage
        •   vshard-router
        •   metrics
        •   app.roles.custom

Set replicasets roles:

.. code-block:: bash

    cartridge replicasets add-roles --replicaset s-1 vshard-storage

        • Add role(s) vshard-storage to replicaset s-1
        • Replicaset s-1 now has these roles enabled:
        •   vshard-storage (default)

    cartridge replicasets add-roles \
      --replicaset router \
      vshard-router app.roles.custom failover-coordinator metrics

        • Add role(s) vshard-router, app.roles.custom, failover-coordinator, metrics to replicaset router
        • Replicaset router now has these roles enabled:
        •   failover-coordinator
        •   vshard-router
        •   metrics

Bootstrap vshard:

.. code-block:: bash

    cartridge replicasets bootstrap-vshard

        • Vshard is bootstrapped successfully

Expel instance:

.. code-block:: bash

    cartridge replicasets expel s1-replica

        • Instance(s) s1-replica has been successfully expelled
