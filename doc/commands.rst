Supported Cartridge CLI commands
================================

..  container:: table

    ..  list-table::
        :widths: 20 80
        :header-rows: 0

        *   -   `create <./commands/create.rst>`_
            -   Create a new application from template
        *   -   `build <./commands/build.rst>`_
            -   Build an application for local development and testing
        *   -   `start <./commands/start.rst>`_
            -   Start one or more Tarantool instances locally
        *   -   `stop <./commands/stop.rst>`_
            -   Stop one or more Tarantool instances started locally
        *   -   `status <./commands/status.rst>`_
            -   Get the status of one or more instances running locally
        *   -   `enter <./commands/connect.rst#cartridge-cli_enter>`_
            -   Enter a locally running instance
        *   -   `connect <./commands/connect.rst#cartridge-cli_connect>`_
            -   Connect to a locally running instance at a specific address
        *   -   `log <./commands/log.rst>`_
            -   Get the logs of one or more instances
        *   -   `clean <./commands/clean.rst>`_
            -   Clean the files of one or more instances
        *   -   `pack <./commands/pack.rst>`_
            -   Pack the application into a distributable bundle
        *   -   `repair <./commands/repair.rst>`_
            -   Patch cluster configuration files
        *   -   `admin <./commands/admin.rst>`_
            -   Сall an admin function provided by the application
        *   -   `replicasets<./commands/replicasets.rst>`_
            -   Manage cluster replica sets running locally
        *   -   `failover <./commands/failover.rst>`_
            -   Manage cluster failover

All commands support `global flags <./doc/global_flags.rst>`_
that control output verbosity.