Building in Docker
==================

To build your application in Docker, run this:

..  code-block:: bash

    cartridge pack TYPE --use-docker

For ``TYPE``, indicate ``rpm``, ``deb``, or ``tgz``.

You might want to perform application build in Docker
if your distributable is intended for a system different than the one you use.

In this case, ``cartridge.pre-build``, ``tarantoolctl rocks make``,
and ``cartridge.post-build`` run inside a Docker image
that has a volume mapped onto the build directory.
As a result, the build directory will contain Linux-specific application files
and rocks modules.

If you want to package a distribution on your local machine without using Docker,
check the :doc:`packaging overview </book/cartridge/cartridge_cli/commands/pack>`.

Flags
-----

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``--build-from``
            -   Path to the base Dockerfile of the build image.
                Defaults to ``Dockerfile.build.cartridge`` in the application root directory.
        *   -   ``--cache-from``
            -   Images that work as cache sources for both build and runtime images.
                See the ``--cache-from`` flag for ``docker build``.
        *   -   ``--sdk-path``
            -   Enterprise only.
                Path to the SDK to be delivered in the final artifact.
                Alternatively, you can pass the path via the ``TARANTOOL_SDK_PATH``
                environment variable, which is of lower priority.
        *   -   ``--sdk-local``
            -   Enterprise only.
                Deliver the SDK from the local machine in the final artifact.

..  note::

    If you're building a Tarantool Enterprise application in Docker,
    make sure you specify the path to the SDK you want to include in the image.
    Do that using the ``--sdk-path`` flag
    or the environment variable ``TARANTOOL_SDK_PATH``, which has lower priority.
    To specify the currently activated SDK, pass the ``--sdk-local`` flag.

Build image
-----------

The image where the package is built
will be referred to as the build image. It has the following structure:

*   Base image: ``centos:7`` (see below).
*   Pre-installed packages: ``git``, ``gcc``, ``make``, ``cmake``, ``unzip``.
    These are the packages required for building the default  ``cartridge`` application.
*   The image includes a version of Tarantool:

    -   If you use open-source Tarantool, the image contains
        the same version you've used for local development.
    -   If you use Tarantool Enterprise, the SDK with Tarantool Enterprise binaries
        is copied to the image.
        See the ``--sdk-path`` and ``--sdk-local`` flags.

To customize your build image, use the ``Dockerfile.build.cartridge`` file
in the application root directory.

Installing packages required for application build
--------------------------------------------------

By default, the build image is based on ``centos:7``.
``git``, ``gcc``, ``make``, ``cmake``, and ``unzip`` packages are installed in that image.
If your application requires other packages for building, you
can specify more base layers for the build image.

To do that, place the file ``Dockerfile.build.cartridge`` in your application root directory
or pass a path to another Dockerfile with the ``--build-from`` flag.
Make sure your Dockerfile starts with the line ``FROM centos:7`` (except comments).

For example, if your application build requires ``gcc-c++``,
customize the Dockerfile as follows:

*   ``Dockerfile.build.cartridge``:

    ..  code-block:: dockerfile

        FROM centos:7
        RUN yum install -y gcc-c++
        # Note that git, gcc, make, cmake, and unzip
        # will be installed anyway

..  note::

    ``git``, ``gcc``, ``make``, ``cmake``, and ``unzip`` will be installed
    anyway on the next layer.

