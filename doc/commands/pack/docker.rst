Packaging an application into a Docker image
============================================

``cartridge pack docker`` builds a Docker image that can be used to start
containers of application instances.

Flags
-----

Use these flags to control the local packaging of a Docker image.
For flags applicable for packaging any distribution type,
check the :doc:`packaging overview </book/cartridge/cartridge_cli/commands/pack>`.

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``--tag``
            -   Tag(s) of the Docker image that results from ``cartridge pack docker``.
        *   -   ``--from``
            -   Path to the base Dockerfile of the result image.
                Defaults to ``Dockerfile.cartridge`` in the application root directory.

Result image tag
----------------

The result image is tagged as follows:

*   ``<name>:<detected-version>[-<suffix>]``: by default.
*   ``<name>:<version>[-<suffix>]``: if the ``--version`` parameter is specified.
*   ``<tag>``: if the ``--tag`` parameter is specified.

Starting application instances
------------------------------

To start ``instance-1`` of the application ``myapp``, run:

..  code-block:: bash

    docker run -d \
                    --name instance-1 \
                    -e TARANTOOL_INSTANCE_NAME=instance-1 \
                    -e TARANTOOL_ADVERTISE_URI=3302 \
                    -e TARANTOOL_CLUSTER_COOKIE=secret \
                    -e TARANTOOL_HTTP_PORT=8082 \
                    -p 127.0.0.1:8082:8082 \
                    myapp:1.0.0

By default, ``TARANTOOL_INSTANCE_NAME`` is set to ``default``.

You can specify the environment variables ``CARTRIDGE_RUN_DIR`` and ``CARTRIDGE_DATA_DIR``:

..  code-block:: bash

    docker run -d \
                    --name instance-1 \
                    -e CARTRIDGE_RUN_DIR=my-custom-run-dir \
                    -e CARTRIDGE_DATA_DIR=my-custom-data-dir \
                    -e TARANTOOL_ADVERTISE_URI=3302 \
                    -e TARANTOOL_CLUSTER_COOKIE=secret \
                    -e TARANTOOL_HTTP_PORT=8082 \
                    -p 127.0.0.1:8082:8082 \
                    myapp:1.0.0

``CARTRIDGE_DATA_DIR`` is the working directory
that contains the instance's PID file and console socket.
By default, it is set to ``/var/lib/tarantool``.

You can also set variables like ``TARANTOOL_WORKDIR``, ``TARANTOOL_PID_FILE``,
and ``TARANTOOL_CONSOLE_SOCK``:

..  code-block:: bash

    docker run -d \
                    --name instance-1 \
                    -e TARANTOOL_WORKDIR=custom-workdir \
                    -e TARANTOOL_PID_FILE=custom-pid-file \
                    -e TARANTOOL_CONSOLE_SOCK=custom-console-sock \
                    -e TARANTOOL_ADVERTISE_URI=3302 \
                    -e TARANTOOL_CLUSTER_COOKIE=secret \
                    -e TARANTOOL_HTTP_PORT=8082 \
                    -p 127.0.0.1:8082:8082 \
                    myapp:1.0.0


To check the instance logs, run:

..  code-block:: bash

    docker logs instance-1

Image details
-------------

The base image is ``centos:7`` (see below).

The application code is placed in ``/usr/share/tarantool/<app-name>``.
An open-source version of Tarantool is installed to the image.

The run directory is ``/var/run/tarantool/<app-name>``.
The workdir is ``/var/lib/tarantool/<app-name>``.

The runtime image also contains the file ``/usr/lib/tmpfiles.d/<app-name>.conf``
that allows the instance to restart after container reboot.

It is the user's responsibility to set up the proper ``advertise_uri`` parameter
(``<host>:<port>``) if the containers are deployed on different machines.
Make sure each instance's ``advertise_uri`` is the same on all machines,
because all other instances will use it to connect to that instance.
Suppose you start an instance with ``advertise_uri`` set to
``localhost:3302``. Addressing that instance as ``<instance-host>:3302`` from a different
instance won't work, because other instances will only recognize it as ``localhost:3302``.

If you specify only a port, ``cartridge`` will use an auto-detected IP.
In this case you have to configure Docker networks to set up inter-instance communication.

You can use Docker volumes to store instance snapshots and xlogs on the
host machine. If you updated your application code, you can create a new image for it,
stop the old container, and start a new one using the new image.

Installing packages required by the application in runtime
----------------------------------------------------------

By default, the result image is based on ``centos:7``.

If your application requires some other packages in runtime, you
can specify base layers for result image.

Place the file ``Dockerfile.cartridge`` in your application root directory
or pass a path to another Dockerfile with the ``--from`` flag.
Make sure your Dockerfile starts with the line ``FROM centos:7``
or ``FROM centos:8`` (except comments).

For example, if your application requires ``zip``
for runtime, customize the Dockerfile as follows:

*   `Dockerfile.cartridge`:

    ..  code-block:: dockerfile

        FROM centos:8

        RUN yum install -y zip
