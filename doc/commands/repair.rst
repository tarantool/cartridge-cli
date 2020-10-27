===============================================================================
Repairing the cluster
===============================================================================

To repair a running application, you can use the ``cartridge repair`` command.

* Rule #1 of ``repair`` is: you do not use it if you aren't sure that
  it's exactly what you need.
* Rule #2: always use ``--dry-run`` before running ``repair``.
* Rule #3: do not hesitate to use the ``--verbose`` option.
* Rule #4: do not use the ``--force`` option if you aren't sure that it's exactly
  what you need.

.. note::
    Please, pay attention to the
    `troubleshooting documentation <https://www.tarantool.io/en/doc/2.3/book/cartridge/troubleshooting/>`_
    before using ``repair``.

.. code-block:: bash

    cartridge repair [command]

-------------------------------------------------------------------------------
Commands
-------------------------------------------------------------------------------

The following repair commands are available:

^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Topology summary
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: bash

    cartridge repair list-topology [flags]

Takes no arguments. Prints the current topology summary.

^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Remove instance
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: bash

    cartridge repair remove-instance UUID [flags]

Removes an instance with the specified UUID from cluster.
If the specified instance isn't found, raises an error.

^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Set leader
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: bash

    cartridge repair set-leader REPLICASET-UUID INSTANCE-UUID [flags]

Sets the specified instance as the leader of the specified replica set.
Raises an error if:

* a replica set or instance with the specified UUID doesn't exist;
* the specified instance doesn't belong to the specified replica set;
* the specified instance is disabled or expelled.

^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Set advertise URI
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: bash

    cartridge repair set-uri INSTANCE-UUID URI-TO [flags]

Rewrites the advertise URI for the specified instance.
If the specified instance isn't found or is expelled, raises an error.

-------------------------------------------------------------------------------
Options
-------------------------------------------------------------------------------

All repair commands have these flags:

* ``--name`` (required) is an application name.

* ``--data-dir`` is the directory where instances' working directories are placed.
  Defaults to ``/var/lib/tarantool``.

All commands, except ``list-topology``, have these flags:

* ``--run-dir`` is a directory where PID and socket files are stored.
  Defaults to ``/var/run/tarantool``.

* ``--dry-run`` runs the ``repair`` command in the dry-run mode
  (shows changes but doesn't apply them).

* ``--reload`` is a flag that enables reloading configuration on instances
  after the patch.

Command also supports `global flags <./global_flags.rst>`_.

-------------------------------------------------------------------------------
What does ``repair`` actually do?
-------------------------------------------------------------------------------

It patches the cluster-wide configuration files of application instances
placed on the local machine.
Note that it's not enough to *apply* new configuration:
the configuration should be *reloaded* by the instance.

``repair`` was created to be used on production (but it still can be used for
local development). So, it requires the application name option ``--name``.
Moreover, remember that the default data directory is ``/var/lib/tarantool`` and
the default run directory is ``/var/run/tarantool``
(both of them can be rewritten by options).

In default mode, ``repair`` walks across all cluster-wide configurations placed
in ``<data-dir>/<app-name>.*`` directories and patches all found configuration
files.

If the ``--dry-run`` flag is specified, files aren't patched, and only a computed
configuration diff is shown.

If configuration files are diverged between instances on the local machine,
``repair`` raises an error.
But you can specify the ``--force`` option to patch different versions of
configuration independently.

``repair`` can also reload configuration for all instances if the ``--reload``
flag is specified (only if the application uses ``cartridge >= 2.0.0``).
Configuration will be reloaded for all instances that are placed in the new
configuration using console sockets that are placed in the run directory.
Make sure that you specified the right run directory when using ``--reload`` flag.
