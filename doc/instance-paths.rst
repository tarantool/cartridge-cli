Application instance paths
==========================

The commands that operate with running instances compute instance file paths.
Default paths are passed to every instance on start. Other commands use them
to communicate with the instance -- for example, to show the logs
or connect to the instance through its console socket.

Path configuration file
-----------------------

The file ``.cartridge.yml``, located in the application root directory,
lets you override default paths for a locally running application.
Instead of indicating default paths with flags, you can do so by
defining values for similarly named keys in ``.cartridge.yml``.

For example, instead of

..  code-block:: bash

    cartridge start --run-dir my-run-dir --cfg my-instances.yml --script my-init.lua

you can write the following in your ``.cartridge.yml``:

..  code-block:: yaml

    run-dir: my-run-dir
    cfg: my-instances.yml
    script: my-init.lua

In ``.cartridge.yml``, you can also enable or disable the ``stateboard`` parameter.
It is initially set to ``true`` in the template application.

Directory paths
---------------

Run directory
^^^^^^^^^^^^^

The run directory (``--run-dir``) is where PID and socket files are stored.
More specifically, it contains:

*   Instance PID files: ``<run-dir>/<app-name>.<instance-name>.pid``
*   Instance console sockets: ``<run-dir>/<app-name>.<instance-name>.control``
*   Instance notify sockets: ``<run-dir>/<app-name>.<instance-name>.notify``.

Data directory
^^^^^^^^^^^^^^

The data directory (``--data-dir``) contains the instances'
working directories.

Each instance's working directory is
``<data-dir>/<app-name>.<instance-name>``.

Logs directory
^^^^^^^^^^^^^^

The logs directory (``--log-dir``) is where instance logs are stored
when the instances run in the background.
This directory is created on ``cartridge start -d`` and can be used by ``cartridge log``.

Each instance's log file is ``<log-dir>/<app-name>.<instance-name>.log``.

Instance configuration file
^^^^^^^^^^^^^^^^^^^^^^^^^^^

This file (``--cfg``) lets you configure Cartridge instances.
The path to the file is passed to all instances
as the environment variable ``TARANTOOL_CFG``.
See the :ref:`configuration guide <cartridge-config-basic>`
for details.

Example:

..  code-block:: yaml

    myapp.router:
        advertise_uri: localhost:3301
        http_port: 8081

    myapp.s1-master:
        advertise_uri: localhost:3302
        http_port: 8082

    myapp-stateboard:
        listen: localhost:3310
        password: passwd

