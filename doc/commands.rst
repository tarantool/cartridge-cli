Supported Cartridge CLI commands
================================

..  container:: table

    ..  list-table::
        :widths: 20 80
        :header-rows: 0

        *   -   :doc:`create <commands/create>`
            -   Create a new application from template
        *   -   :doc:`build <commands/build>`
            -   Build an application for local development and testing
        *   -   :doc:`start <commands/start>`
            -   Start one or more Tarantool instances locally
        *   -   :doc:`stop <commands/stop>`
            -   Stop one or more Tarantool instances started locally
        *   -   :doc:`status <commands/status>`
            -   Get the status of one or more instances running locally
        *   -   :doc:`enter <commands/enter>`
            -   Enter a locally running instance
        *   -   :doc:`connect <commands/connect>`
            -   Connect to a locally running instance at a specific address
        *   -   :doc:`log <commands/log>`
            -   Get the logs of one or more instances
        *   -   :doc:`clean <commands/clean>`
            -   Clean the files of one or more instances
        *   -   :doc:`pack <commands/pack>`
            -   Pack the application into a distributable bundle
        *   -   :doc:`repair <commands/repair>`
            -   Patch cluster configuration files
        *   -   :doc:`admin <commands/admin>`
            -   Сall an admin function provided by the application
        *   -   :doc:`replicasets <commands/replicasets>`
            -   Manage cluster replica sets running locally
        *   -   :doc:`failover <commands/failover>`
            -   Manage cluster failover

All commands support :doc:`global flags <global-flags>`
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

