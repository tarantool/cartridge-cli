Repairing the cluster
=====================

The ``cartridge repair`` command repairs a running application.

Here are several simple rules you need to know before using this command:

#.  Do not use ``repair`` if you aren't sure that it's exactly what you need.
#.  Always run ``repair`` with the ``--dry-run`` flag first.
#.  Do not hesitate to use the ``--verbose`` option.
#.  Do not use the ``--force`` option if you aren't sure that it's exactly
    what you need.

..  note::

    Please look at the
    :doc:`troubleshooting documentation </book/cartridge/troubleshooting>`
    before using ``repair``.


..  code-block:: bash

    cartridge repair [subcommand]

Subcommands
-----------

Below is a list of the available repair commands.

..  contents::
    :local:

list-topology
~~~~~~~~~~~~~

..  code-block:: bash

    cartridge repair list-topology [flags]

Print the current topology summary. Requires no arguments.

remove-instance
~~~~~~~~~~~~~~~

..  code-block:: bash

    cartridge repair remove-instance UUID [flags]

Remove an instance with the specified UUID from the cluster.
If the instance isn't found, raise an error.

set-leader
~~~~~~~~~~

..  code-block:: bash

    cartridge repair set-leader REPLICASET-UUID INSTANCE-UUID [flags]

Set an instance as the leader of the replica set.
Raise an error in the following cases:

* There is no replica set or instance with that UUID.
* The instance doesn't belong to the replica set.
* The instance has been disabled or expelled.

set-uri
~~~~~~~

..  code-block:: bash

    cartridge repair set-uri INSTANCE-UUID URI-TO [flags]

Rewrite the instance's
:ref:`advertise_uri <cartridge-config-basic>`
parameter. Raise an error if the instance isn't found or is expelled.


Flags
-----

The following flags work with any ``repair`` subcommand:

..  container:: table

    ..  list-table::
        :widths: 20 80
        :header-rows: 0

        *   -   ``--name``
            -   (Required) Application name.
        *   -   ``--data-dir``
            -   The directory containing the instances' working directories.
                Defaults to ``/var/lib/tarantool``.

The following flags work with any repair command except ``list-topology``:

..  container:: table

    ..  list-table::
        :widths: 20 80
        :header-rows: 0

        *   -   ``--run-dir``
            -   The directory where PID and socket files are stored.
                Defaults to ``/var/run/tarantool``.
        *   -   ``--dry-run``
            -   Launch in dry-run mode: show changes but do not apply them.
        *   -   ``--reload``
            -   Enable instance configuration reload after the patch.

..  note::
    
    The default data and run directories for ``repair`` differ from those
    used by other ``cartridge-cli`` commands. This is because ``repair`` is
    intended for production use, while other commands are for local development.

``repair`` also supports :doc:`global flags </book/cartridge/cartridge_cli/global-flags>`.


What does repair actually do?
-----------------------------

It patches cluster-wide instance configuration files that you have on your local machine.
Note that it's not enough to *apply* the new configuration, the instance has to *reload* it.

Although ``repair`` was created for production use, it can still be applied in
local development. The command requires to specify ``--name``, the application name.
Also, remember that the default data directory is ``/var/lib/tarantool`` and
the default run directory is ``/var/run/tarantool``.
To specify other directories, use the ``data-dir`` and ``--run-dir`` options correspondingly
or provide the paths in the
:doc:`configuration file </book/cartridge/cartridge_cli/instance-paths>`.

In default mode, ``repair`` walks through all cluster-wide configurations
in the ``<data-dir>/<app-name>.*`` directories, patching all the configuration
files it locates.

With the ``--dry-run`` flag specified, files won't be patched,
and you will only see the computed configuration diff.

If different instances on the local machine use different configuration files,
``repair`` raises an error.
To patch different configuration versions independently, use the ``--force`` option.

If your application uses ``cartridge >= 2.0.0``,
you can also run ``repair`` with the ``--reload`` flag
to reload configuration for all your instances
through the console sockets in the run directory.
Make sure that you have the correct run directory specified
when you use ``--reload``.

