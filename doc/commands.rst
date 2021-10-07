Supported Cartridge CLI commands
================================

..  container:: table

    ..  list-table::
        :widths: 20 80
        :header-rows: 0

        *   -   :doc:`create </book/cartridge/cartridge_cli/commands/create>`
            -   Create a new application from template
        *   -   :doc:`build </book/cartridge/cartridge_cli/commands/build>`
            -   Build an application for local development and testing
        *   -   :doc:`start </book/cartridge/cartridge_cli/commands/start>`
            -   Start one or more Tarantool instances locally
        *   -   :doc:`stop </book/cartridge/cartridge_cli/commands/stop>`
            -   Stop one or more Tarantool instances started locally
        *   -   :doc:`status </book/cartridge/cartridge_cli/commands/status>`
            -   Get the status of one or more instances running locally
        *   -   :doc:`enter </book/cartridge/cartridge_cli/commands/enter>`
            -   Enter a locally running instance
        *   -   :doc:`connect </book/cartridge/cartridge_cli/commands/connect>`
            -   Connect to a locally running instance at a specific address
        *   -   :doc:`log </book/cartridge/cartridge_cli/commands/log>`
            -   Get the logs of one or more instances
        *   -   :doc:`clean </book/cartridge/cartridge_cli/commands/clean>`
            -   Clean the files of one or more instances
        *   -   :doc:`pack </book/cartridge/cartridge_cli/commands/pack>`
            -   Pack the application into a distributable bundle
        *   -   :doc:`repair </book/cartridge/cartridge_cli/commands/repair>`
            -   Patch cluster configuration files
        *   -   :doc:`admin </book/cartridge/cartridge_cli/commands/admin>`
            -   Сall an admin function provided by the application
        *   -   :doc:`replicasets</book/cartridge/cartridge_cli/commands/replicasets>`
            -   Manage cluster replica sets running locally
        *   -   :doc:`failover </book/cartridge/cartridge_cli/commands/failover>`
            -   Manage cluster failover

All commands support :doc:`global flags </book/cartridge/cartridge_cli/commands/global_flags>`
that control output verbosity.

..  toctree::
    :hidden:

    create <commands/create>
    build <commands/build>
    start <commands/start>
    stop <commands/stop>
    status <commands/status>
    enter <commands/enter>
    connect <commands/connect>
    log <commands/log>
    clean <commands/clean>
    pack <commands/pack>
    repair <commands/repair>
    admin <commands/admin>
    replicasets <commands/replicasets>
    failover <commands/failover>

test
