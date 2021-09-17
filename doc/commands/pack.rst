Packing an application
======================

To pack your application, use ``pack`` command:

.. code-block:: bash

     cartridge pack TYPE [PATH] [flags]

where:

* ``TYPE`` (required) is the distribution type. Supported types:

  - `TGZ <./pack/tgz.rst>`_
  - `RPM <./pack/rpm_deb.rst>`_
  - `DEB <./pack/rpm_deb.rst>`_
  - `Docker <./pack/docker.rst>`_

* ``PATH`` (optional) is the path to the application directory to pack.
  Defaults to ``.`` (the current directory).

.. note::

    Result artifact contains application with rocks modules
    and executables specific for the system where application was built.
    By default, application is built on local machine.
    It mean that you can't install RPM built on Mac OS on some machine with
    Centos. The solution is to specify ``--use-docker`` flag that enforces
    building application in docker.

.. note::

    If you are using opensource Tarantool that your artifact has `tarantool`
    dependency (version is detected as a version of Tarantool from ``PATH``).
    For Enterprise Tarantool artifact contains ``tarantool`` and ``tarantoolctl``
    binaries from current SDK.
    When building in docker using Tarantool Enterprise, you must specify the path
    to the SDK that should be used on image. Use ``--sdk-path`` option
    (or the environment variable ``TARANTOOL_SDK_PATH``, which has lower priority)
    for that.
    If you want to use the currently activated SDK, pass ``--sdk-local`` option.

Options
-------

Options that are supported for all distribution types:

* ``--name`` is the application name.
  It coincides with the package name and the systemd-service name.
  The default name comes from the ``package`` field in the rockspec file.

* ``--version`` is the application's package
  version. The expected pattern is ``major.minor.patch[-count][-commit]``:
  if you specify ``major.minor.patch``, it is normalized to ``major.minor.patch-count``.
  The default version is determined as the result of ``git describe --tags --long``.
  If the application is not a git repository, you need to set the ``--version`` option
  explicitly.

* ``--suffix`` is the result file (or image)
  name suffix.


* ``--use-docker`` forces to build the application in Docker.

Options that are specific for RPM and DEB:

* ``--unit-template`` is the path to the template for
  the ``systemd`` unit file.

* ``--instantiated-unit-template`` is the path to the
  template for the ``systemd`` instantiated unit file.

* ``--stateboard-unit-template`` is the path to the
  template for the stateboard ``systemd`` unit file.

Options that are specific for packing in Docker:

* ``--tag`` is the tag(s) of the Docker image that results from ``pack docker``.

* ``--from`` is the path to the base Dockerfile of the result image.
  Defaults to ``Dockerfile.cartridge`` in the application root.

Options that are used on building in docker
(for ``pack docker`` or when ``--use-docker`` flag is specified):

* ``--build-from`` is
  the path to the base Dockerfile of the build image.
  Defaults to ``Dockerfile.build.cartridge`` in the application root.

* ``--no-cache`` creates build and runtime images with ``--no-cache`` docker flag.

* ``--cache-from`` images to consider as cache sources for both build and
  runtime images. See ``--cache-from`` flag for ``docker build`` command.

* ``--sdk-path`` is the path to the SDK to be delivered in the result artifact.
  Alternatively, you can pass the path via the ``TARANTOOL_SDK_PATH``
  environment variable (this variable is of lower priority).

* ``--sdk-local`` is a flag that indicates if the SDK from the local machine
  should be delivered in the result artifact.

Details
-------

By default, application build is done in a temporary directory in
``~/.cartridge/tmp/``, so the packaging process doesn't affect the contents
of your application directory.

On copying application files ``.rocks`` directory is ignored.

Files permissions are preserved, and the code files owner is set to
``root:root`` in the resulting package.

All application files should have at least ``a+r`` permissions
(``a+rx`` for directories).
Otherwise, ``cartridge pack`` command raises an error.

Customizing build directory
~~~~~~~~~~~~~~~~~~~~~~~~~~~

You can specify a custom build directory for your application using ``CARTRIDGE_TEMPDIR``
environment variable. If this directory doesn't exists, it will be created, used
for building the application, and then removed.

If you specify an existing directory in the ``CARTRIDGE_TEMPDIR`` environment
variable, the ``CARTRIDGE_TEMPDIR/cartridge.tmp`` directory will be used for
build and then removed. This directory will be cleaned up before building the
application.

.. note::

    It's useful on build in docker in GitLab CI - docker volumes don't work
    properly with default tmp directory in this case.
    Use ``CARTRIDGE_TEMPDIR=. cartridge pack ...``.


Stage 1. Cleaning up the application directory
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

On this stage, some files are filtered out of the application directory:

* First, ``git clean -X -d -f`` removes all untracked and
  ignored files (it works for submodules, too).
* After that, ``.git`` directory is removed.

Stage 2. Building the application
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Application directory should contain rockspec.

Application build can be performed in docker (for packing into ``docker`` image
or if ``--use-docker`` flag is specified).
The key steps are the same.
More details about packing in docker (e.g. what to do if your application
requires some non-standard packages for build) you can find below.

1. ``./cartridge.pre-build`` if this file exists in application root
2. ``tarantoolctl rocks make``

During step 2 -- the key step here -- ``cartridge`` installs all dependencies
specified in the rockspec file (you can find this file within the application
directory created from template).

If your application depends on closed-source rocks, or if the build should contain
rocks from a project added as a submodule, then you need to **install** all these
dependencies before calling ``tarantoolctl rocks make``.
You can do it using the file ``cartridge.pre-build`` in your application root
(again, you can find this file within the application directory created from template).
In this file, you can specify all rocks to build from submodules
(e.g. ``tarantoolctl rocks make --chdir ./third_party/proj``).
For details, see `special files <../special_files.rst>`_.

As a result, in the application's ``.rocks`` directory you will get a fully built
application that you can start locally from the application's directory.

(An advanced alternative would be to specify build logic in the
rockspec as ``cmake`` commands, like we
`do it <https://github.com/tarantool/cartridge/blob/master/cartridge-scm-1.rockspec#L26>`_
for ``cartridge``.)

Stage 3. Cleaning up the files before packing
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

On this stage, ``cartridge`` runs ``cartridge.post-build`` (if it exists) to remove
junk files (like ``node_modules``) generated during application build.

See an `special files <../special_files.rst>`_ for ``cartridge.post-build``
example.

Building application in Docker
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Building application in docker is quite simple.
The following commands are ran (just like for usual build):

1. ``./cartridge.pre-build`` if this file exists in application root
2. ``tarantoolctl rocks make``

But these commands are ran in docker image that has a volume mapped on
build directory.
As a result, build directory contents application files and rock modules that
are specific for Linux (because it was installed inside docker container).

Build image
~~~~~~~~~~~

The image where application is built has the following structure:

The base image is ``centos:8`` (see below).

All packages required for the default  ``cartridge`` application build
(``git``, ``gcc``, ``make``, ``cmake``, ``unzip``) are installed.

A proper version of Tarantool is provided:

* For opensource, Tarantool of the same version as the one used for
  local development is installed to the image.
* For Tarantool Enterprise, the SDK with Tarantool Enterprise binaries is
  copied to the image (see ``--sdk-path``, ``--sdk-local`` options).

Installing packages requied for application build
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

By default, application is build on image based on ``centos:8``.

Then, ``git``, ``gcc``, ``make``, ``cmake``, ``unzip`` packages are installed.

If your application requires some other packages for build, you
can specify base layers for build image.

Place ``Dockerfile.build.cartridge`` file in your application root (or pass a path to
the other dockerfile via ``--build-from`` opton).
The dockerfile should be started with the ``FROM centos:8``
or ``FROM centos:7`` line (except comments).

For example, if your application requires ``gcc-c++`` for build, customize the Dockerfiles as follows:

* ``Dockerfile.cartridge.build``:

  .. code-block:: dockerfile

      FROM centos:8
      RUN yum install -y gcc-c++
      # Note that git, gcc, make, cmake, unzip packages
      # will be installed anyway

.. note::

    ``git``, ``gcc``, ``make``, ``cmake``, ``unzip`` packages will be installed
    anyway on the next layer.
