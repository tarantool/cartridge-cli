Packing an application into docker image
========================================

``cartridge pack docker`` builds a Docker image that can be used to start
application instances containers.

Result image tag
----------------

The result image is tagged as follows:

* ``<name>:<detected-version>[-<suffix>]``: by default;
* ``<name>:<version>[-<suffix>]``: if the ``--version`` parameter is specified;
* ``<tag>``: if the ``--tag`` parameter is specified.

Starting application instances
------------------------------

To start the ``instance-1`` instance of the ``myapp`` application, say:

.. code-block:: bash

    docker run -d \
                    --name instance-1 \
                    -e TARANTOOL_INSTANCE_NAME=instance-1 \
                    -e TARANTOOL_ADVERTISE_URI=3302 \
                    -e TARANTOOL_CLUSTER_COOKIE=secret \
                    -e TARANTOOL_HTTP_PORT=8082 \
                    -p 127.0.0.1:8082:8082 \
                    myapp:1.0.0

By default, ``TARANTOOL_INSTANCE_NAME`` is set to ``default``.

To check the instance logs, say:

.. code-block:: bash

    docker logs instance-1

Image details
-------------

The base image is ``centos:8`` (see below).

The application code is placed in the ``/usr/share/tarantool/<app-name>``
directory. An opensource version of Tarantool is installed to the image.

The run directory is ``/var/run/tarantool/<app-name>``,
the workdir is ``/var/lib/tarantool/<app-name>``.

The runtime image also contains the file ``/usr/lib/tmpfiles.d/<app-name>.conf``
that allows the instance to restart after container restart.

It is the user's responsibility to set up a proper advertise URI
(``<host>:<port>``) if the containers are deployed on different machines.
The problem here is that an instance's advertise URI must be the same on all
machines, because it will be used by all the other instances to connect to this
one. For example, if you start an instance with an advertise URI set to
``localhost:3302``, and then address it as ``<instance-host>:3302`` from other
instances, this won't work: the other instances will be recognizing it only as
``localhost:3302``.

If you specify only a port, ``cartridge`` will use an auto-detected IP,
so you need to configure Docker networks to set up inter-instance communication.

You can use Docker volumes to store instance snapshots and xlogs on the
host machine. To start an image with a new application code, just stop the
old container and start a new one using the new image.

Installing packages requied by application in runtime
-----------------------------------------------------

By default, the result image is based on ``centos:8``.

If your application requires some other packages in runtime, you
can specify base layers for result image.

Place ``Dockerfile.cartridge`` file in your application root (or pass a path to
the other dockerfile via ``--from`` opton).
The dockerfile should be started with the ``FROM centos:8``
or ``FROM centos:7`` line (except comments).

For example, if your application requires ``zip`` for runtime, customize the dockerfile as follows:

* `Dockerfile.cartridge`:

  .. code-block:: dockerfile

      FROM centos:8

      RUN yum install -y zip
