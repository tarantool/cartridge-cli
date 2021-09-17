Application instances paths
===========================

Commands that operates with running instances computes instance files paths.
These paths are passed to instance on starting and is used by other commands
to communicate with instance (e.g. show instance logs or connect to instance
using console socket).

Paths configuration file
------------------------

For local running application commands default paths can be overriden in the
``.cartridge.yml`` file in application root.
For each ``--<flag-name>`` flag that defines some path you can specify
default value for local running in ``flag-name`` section of ``.cartridge.yml``.

For example:

.. code-block:: yaml

    run-dir: my-run-dir
    cfg: my-instances.yml
    script: my-init.lua

Instance paths
--------------

Run directory (``--run-dir``)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Is the directory where PID and socket files are stored.

Files that are stored in run directory:

* Instance PID-file: ``<run-dir>/<app-name>.<instance-name>.pid``.
* Instance console socket: ``<run-dir>/<app-name>.<instance-name>.control``.
* Instance notify socket: ``<data-dir>/<app-name>.<instance-name>.notify``

Data directory (``--data-dir``)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Is the directory where instances' working directories are placed.

Each instance's working directory is ``<data-dir>/<app-name>.<instance-name>``.

Logs directory (``--log-dir``)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Is the directory to store instances logs when running in background.
Is created on ``cartridge start -d``, can be used by ``cartridge log``.

Each instance log file is ``<log-dir>/<app-name>.<instance-name>.log``.

Instances configuration file (``--cfg``)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Is the Cartridge instances configuration file.
This path is passed to all instances as ``TARANTOOL_CFG`` environment variable.
See `configuration guide <https://www.tarantool.io/en/doc/latest/book/cartridge/topics/clusterwide-config/#configuration-basics>`_
for details.

Example:

.. code-block:: yaml

    myapp.router:
        advertise_uri: localhost:3301
        http_port: 8081

    myapp.s1-master:
        advertise_uri: localhost:3302
        http_port: 8082

    myapp-stateboard:
        listen: localhost:3310
        password: passwd
