Migration from Cartridge CLI to tt
==================================

.. note::

    The migration instruction is also available in the
    `tt repository <https://github.com/tarantool/tt/blob/master/doc/examples.md#transition-from-cartridge-cli-to-tt>`_
    on GitHub.

To start managing a Cartridge application with ``tt`` instead of Cartridge CLI,
run ``tt init`` in the application directory:

.. code-block:: bash

    $ tt init
    • Found existing config '.cartridge.yml'
    • Environment config is written to 'tt.yaml'

This creates a ``tt`` environment based on the existing Cartridge configuration.
Now you're ready to manage the application with ``tt``:

.. code-block:: bash

    $ tt start
    • Starting an instance [app:s1-master]...
    • Starting an instance [app:s1-replica]...
    • Starting an instance [app:s2-master]...
    • Starting an instance [app:s2-replica]...
    • Starting an instance [app:stateboard]...
    • Starting an instance [app:router]...
    $ tt status
    INSTANCE           STATUS          PID
    app:s1-replica     RUNNING         112645
    app:s2-master      RUNNING         112646
    app:s2-replica     RUNNING         112647
    app:stateboard     RUNNING         112655
    app:router         RUNNING         112656
    app:s1-master      RUNNING         112644

Commands difference
-------------------

Most Cartridge CLI commands look the same in ``tt``: ``cartridge start`` and
``tt start``, ``cartridge create`` and ``tt create``, and so on. To migrate such
calls, it is usually enough to replace the utility name. There can be slight differences
in command flags and format. For details on ``tt`` commands, see the
:ref:`tt commands reference <tt-commands>`.

The following commands are different in ``tt``:

*   Cartridge CLI commands ``admin``, ``bench``, ``failover``, ``repair``, ``replicasets``
    are implemented as subcommands of ``tt cartridge``. Example, ``tt cartridge repair``.
*   ``cartridge enter`` and ``cartridge connect`` are covered by ``tt connect``.
*   The analog of ``cartridge gen completion`` is ``tt completion``
*   ``cartridge log`` and ``cartridge pack docker`` functionality is not supported in ``tt``.

