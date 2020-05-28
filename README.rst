.. _cartridge-cli:

===============================================================================
Cartridge Command Line Interface
===============================================================================

.. image:: https://gitlab.com/tarantool/cartridge-cli/badges/master/pipeline.svg
   :alt: Cartridge-CLI build status on GitLab CI
   :target: https://gitlab.com/tarantool/cartridge-cli/commits/master

.. contents::

-------------------------------------------------------------------------------
Installation
-------------------------------------------------------------------------------

1. Install third-party software:

   * `Install <https://git-scm.com/book/en/v2/Getting-Started-Installing-Git>`_
     ``git``, a version control system.

   * `Install <https://linuxize.com/post/how-to-unzip-files-in-linux/>`_
     the ``unzip`` utility.

   * `Install <https://gcc.gnu.org/install/>`_
     the ``gcc`` compiler.

   * `Install <https://cmake.org/install/>`_
     the ``cmake`` and ``make`` tools.

2. Install Tarantool 1.10 or higher.

   You can:

   * Install it from a package (see https://www.tarantool.io/en/download/
     for OS-specific instructions).
   * Build it from sources (see
     https://www.tarantool.io/en/download/os-installation/building-from-source/).

3. [On all platforms except MacOS X] If you built Tarantool from sources,
   you need to manually set up the Tarantool packages repository:

   .. code-block:: console

       curl -L https://tarantool.io/installer.sh | sudo -E bash -s -- --repo-only

4. Install the ``cartridge-cli`` package:

   * for CentOS, Fedora, ALT Linux (RPM package):

     .. code-block:: console

         sudo yum install cartridge-cli

   * for Debian, Ubuntu (DEB package):

     .. code-block:: console

         sudo apt-get install cartridge-cli

   * for MacOS X (Homebrew formula):

     .. code-block:: console

         brew install cartridge-cli

   * for any OS (from luarocks):

     .. code-block:: console

         tarantoolctl rocks install cartridge-cli

     This installs the rock to the application's directory.
     The executable is available at ``.rocks/bin/cartridge``.
     Optionally, you can add ``.rocks/bin`` to the executable path:

     .. code-block:: console

         export PATH=$PWD/.rocks/bin/:$PATH

5. Check the installation:

   .. code-block:: console

       cartridge --version

Now you can
`create and start <https://www.tarantool.io/en/doc/latest/getting_started/getting_started_cartridge/>`_
your first application!

-------------------------------------------------------------------------------
Quick start
-------------------------------------------------------------------------------

To create your first application:

.. code-block:: console

    cartridge create --name myapp

Let's go inside:

.. code-block:: console

    cd myapp

Now build the application and start it:

.. code-block:: console

    cartridge build
    cartridge start

That's all! You can visit http://localhost:8081 and see your application's Admin Web UI:

.. image:: https://user-images.githubusercontent.com/11336358/75786427-52820c00-5d76-11ea-93a4-309623bda70f.png
   :align: center
   :scale: 100%

You can find more details in this documentation -- or start with the
`getting started guide <https://www.tarantool.io/en/doc/latest/getting_started/getting_started_cartridge/>`_.

.. _cartridge-cli-usage:

-------------------------------------------------------------------------------
Usage
-------------------------------------------------------------------------------

For more details, say:

.. code-block:: console

    cartridge --help

These commands are supported:

* ``create`` -- create a new application from template;
* ``build`` -- build the application for local development and testing;
* ``start`` -- start a Tarantool instance(s);
* ``stop`` -- stop a Tarantool instance(s);
* ``status`` -- get current instance(s) status;
* ``pack`` -- pack the application into a distributable bundle.

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
An application's lifecycle
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

In a nutshell:

1. `Create <Creating an application from template_>`_
   an application (e.g. ``myapp``) from template:

   .. code-block:: console

       cartridge create --name myapp
       cd ./myapp

2. `Build <Building an application_>`_ the application
   for local development and testing:

   .. code-block:: console

       cartridge build

3. `Run <Starting stopping an application locally_>`_
   instances locally:

   .. code-block:: console

       cartridge start
       cartridge stop

4. `Pack <Packing an application_>`_ the application into
   a distributable (e.g. into an RPM package):

   .. code-block:: console

       cartridge pack rpm

.. _cartridge cli creating an application from template:

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Creating an application from template
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

To create an application from the Cartridge template, say this in any directory:

.. code-block:: console

    cartridge create --name <app_name> /path/to/

This will create a simple Cartridge application in the ``/path/to/<app_name>/``
directory with:

* one custom role with an HTTP endpoint;
* sample tests and basic test helpers;
* files required for development (like ``.luacheckrc``).

If you have ``git`` installed, this will also set up a Git repository with the
initial commit, tag it with
`version <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_dev/#application-versioning>`_
0.1.0, and add a ``.gitignore`` file to the project root.

Let's take a closer look at the files inside the ``<app_name>/`` directory:

* application files:

  * ``app/roles/custom-role.lua`` a sample
    `custom role <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_dev/#cluster-roles>`_
    with simple HTTP API; can be enabled as ``app.roles.custom``
  * ``<app_name>-scm-1.rockspec`` file where you can specify application
    dependencies
  * ``init.lua`` file which is the entry point for your application
  * ``stateboard.init.lua`` file which is the entry point for the application
    `stateboard <https://github.com/tarantool/cartridge/blob/master/topics/failover.md>`_

* `special files <Special files_>`_ (used to build and pack
  the application):

  * ``cartridge.pre-build``
  * ``cartridge.post-build``
  * ``Dockerfile.build.cartridge``
  * ``Dockerfile.cartridge``

* development files:

  * ``deps.sh`` script that resolves the dependencies from the ``.rockspec`` file
    and installs test dependencies (like ``luatest``)
  * ``instances.yml`` file with instances configuration (used by ``cartridge start``)
  * ``.cartridge.yml`` file with Cartridge configuration (used by ``cartridge start``)
  * ``tmp`` directory for temporary files (used as a run dir, see ``.cartridge.yml``)
  * ``.git`` file necessary for a Git repository
  * ``.gitignore`` file where you can specify the files for Git to ignore
  * ``env.lua`` file that sets common rock paths so that the application can be
    started from any directory.

* test files (with sample tests):

  .. code-block:: text

      test
      ├── helper
      │   ├── integration.lua
      │   └── unit.lua
      │   ├── helper.lua
      │   ├── integration
      │   │   └── api_test.lua
      │   └── unit
      │       └── sample_test.lua

* configuration files:

  * ``.luacheckrc``
  * ``.luacov``
  * ``.editorconfig``

.. _cartridge-cli-building-an-application:

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Building an application
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

*****************
Building locally
*****************

To build your application locally (for local testing), say this in any directory:

.. code-block:: console

    cartridge build [<path>]

.. // Please, update cmd_build usage in cartridge-cli.lua file on updating the doc

This command requires one argument -- the path to your application directory
(i.e. to the build source). The default path is ``.`` (the current directory).

This command runs:

1. ``cartridge.pre-build`` (or [DEPRECATED] ``.cartridge.pre``), if the
   `pre-build file <Special files_>`_ exists.
   This builds the application in the ``path`` directory.
2. ``tarantoolctl rocks make``, if the
   `rockspec file <Special files_>`_ exists.
   This installs all Lua rocks to the `path` directory.

During step 1 of the ``cartridge build`` command, ``cartridge`` builds the application
inside the application directory -- unlike when building the application as part
of the ``cartridge pack`` command, when the application is built in a temporary
`build directory <Build directory_>`_ and no build artifacts
remain in the application directory.

During step 2 -- the key step here -- ``cartridge`` installs all dependencies
specified in the rockspec file (you can find this file within the application
directory created from template).

.. NOTE::

   An advanced alternative would be to specify build logic in the
   rockspec as ``cmake`` commands, like we
   `do it <https://github.com/tarantool/cartridge/blob/master/cartridge-scm-1.rockspec#L26>`_
   for ``cartridge``.

If your application depends on closed-source rocks, or if the build should contain
rocks from a project added as a submodule, then you need to **install** all these
dependencies before calling ``tarantoolctl rocks make``.
You can do it using the file ``cartridge.pre-build`` in your application root
(again, you can find this file within the application directory created from template).
In this file, you can specify all rocks to build
(e.g. ``tarantoolctl rocks make --chdir ./third_party/proj``).
For details, see `special files <Special files_>`_.

As a result, in the application's ``.rocks`` directory you will get a fully built
application that you can start locally from the application's directory.

.. _cartridge-cli-building-in-docker:

*******************
Building in Docker
*******************

By default, ``cartridge build`` is building an application locally.

However, if you build it in OS X, all rocks and executables in the resulting
package will be specific for OS X, so the application won't work in Linux.
To build an application in OS X and run it in Linux, call ``cartridge build``
with the flag ``--use-docker`` and get the application built in a Docker container.

This image is created similarly to the
`build image <Build and runtime images_>`_
created during ``cartridge pack``.

.. _cartridge-cli-starting-stopping-an-application-locally:

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Starting/stopping an application locally
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Now that the application is `built <Building an application_>`_,
you can run it locally:

.. code-block:: console

    cartridge start [APP_NAME[.INSTANCE_NAME]] [options]

The options are:

.. // Please, update cmd_start usage in cartridge-cli.lua file on updating the doc

* ``--script FILE`` is the application's entry point. Defaults to:

  * TARANTOOL_SCRIPT,
  * or ``./init.lua`` when running from the app's directory,
  * or ``:apps_path/:app_name/init.lua`` in a multi-app environment.

* ``--apps-path PATH`` is the path to the application directory when running
  in a multi-app environment. Defaults to ``/usr/share/tarantool``.

* ``--run-dir DIR`` is the directory with pid and sock files.
  Defaults to TARANTOOL_RUN_DIR or `/var/run/tarantool`.

* ``--cfg FILE`` is the configuration file for Cartridge instances.
  Defaults to TARANTOOL_CFG or ``./instances.yml``.

* ``--daemonize / -d`` starts the instance in background.
  With this option, Tarantool also waits until the app's main script is finished.
  For example, this is useful if ``init.lua`` requires time-consuming startup from
  snapshot, and Tarantool waits for the startup to complete.
  This is also useful if the app's main script generates errors, and Tarantool
  can handle them.

* ``--stateboard`` starts the application stateboard as well as instances.
  Defaults to TARANTOOL_STATEBOARD or ``false``.
  Ignored if ``--stateboard-only`` is specified.

* ``--stateboard-only`` starts only the application stateboard.
  Defaults to TARANTOOL_STATEBOARD_ONLY or ``false``.
  If specified, ``INSTANCE_NAME`` is ignored.

The ``cartridge start`` command starts a ``tarantool`` instance with enforced
**environment variables**:

.. code-block:: text

    TARANTOOL_INSTANCE_NAME
    TARANTOOL_CFG
    TARANTOOL_PID_FILE - %run_dir%/%instance_name%.pid
    TARANTOOL_CONSOLE_SOCK - %run_dir%/%instance_name%.sock

``cartridge.cfg()`` uses ``TARANTOOL_INSTANCE_NAME`` to read the instance's
configuration from the file provided in ``TARANTOOL_CFG``.

You can override default options for the ``cartridge`` command in
``./.cartridge.yml`` or ``~/.cartridge.yml``.

You can also override ``.cartridge.yml`` options
in corresponding environment variables (``TARANTOOL_*``).

Here is an example of ``.cartridge.yml``:

.. code-block:: yaml

    run_dir: tmp/run
    cfg: cartridge.yml
    apps_path: /usr/local/share/tarantool
    script: init.lua

When ``APP_NAME`` is not provided, it is parsed from the ``./*.rockspec`` filename.

When ``INSTANCE_NAME`` is not provided, ``cartridge`` reads the ``cfg`` file and starts
all defined instances:

.. code-block:: console

    # in the application directory
    cartridge start # starts all instances
    cartridge start .router_1 # start single instance
    cartridge start .router_1 --stateboard # start single instance and stateboard
    cartridge start --stateboard-only # start stateboard only

    # in a multi-application environment
    cartridge start app_1 # starts all instances of app_1
    cartridge start app_1 --stateboard # starts all instances of app_1 and stateboard
    cartridge start app_1.router_1 # start single instance

.. // Please, update cmd_stop usage in cartridge-cli.lua file on updating the doc

To stop one or more running instances, say:

.. code-block:: console

    cartridge stop [APP_NAME[.INSTANCE_NAME]] [options]

These options from the ``start`` command are supported:

* ``--run-dir DIR``
* ``--cfg FILE``
* ``--apps-path PATH``
* ``--stateboard``
* ``--stateboard-only``

.. // Please, update cmd_status usage in cartridge-cli.lua file on updating the doc

To check current instances status use ``status`` command:

.. code-block:: console

    cartridge status [APP_NAME[.INSTANCE_NAME]] [options]

These options from the ``start`` command are supported:

* ``--run-dir DIR``
* ``--cfg FILE``
* ``--apps-path PATH``
* ``--stateboard``
* ``--stateboard-only``

.. _cartridge-cli-packing-an-application:

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Packing an application
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

To pack your application, say this in any directory:

.. code-block:: console

    cartridge pack [options] <type> [<path>]

where:

* ``type`` [REQUIRED] is the distribution type. The supported types are:
  ``rpm``, ``tgz``, ``docker``, ``deb``. See details below.

* ``path`` [OPTIONAL] is the path to the application directory to pack.
  Defaults to ``.`` (the current directory).

All types of distribution are described below:

* `TGZ <TGZ_>`_
* `RPM <RPM and DEB_>`_
* `DEB <RPM and DEB_>`_
* `Docker <Docker_>`_

The options are:

.. // Please, update cmd_pack usage in cartridge-cli.lua file on updating the doc

* ``--name`` (common for all distribution types) is the application name.
  It coincides with the package name and the systemd-service name.
  The default name comes from the ``package`` field in the rockspec file.

* ``--version`` (common for all distribution types) is the application's package
  version. The expected pattern is ``major.minor.patch[-count][-commit]``:
  if you specify ``major.minor.patch``, it is normalized to ``major.minor.patch-count``.
  The default version is determined as the result of ``git describe --tags --long``.
  If the application is not a git repository, you need to set the ``--version`` option
  explicitly.

* ``--suffix`` (common for all distribution types) is the result file (or image)
  name suffix.

* ``--unit-template`` (used for ``rpm`` and ``deb``) is the path to the template for
  the ``systemd`` unit file.

* ``--instantiated-unit-template`` (used for ``rpm`` and ``deb``) is the path to the
  template for the ``systemd`` instantiated unit file.

* ``--from`` (used for ``docker``) is the path to the base Dockerfile of the runtime
  image. Defaults to ``Dockerfile.cartridge`` in the application root.

* ``--use-docker`` (ignored for ``docker``) forces to build the application in Docker.

* ``--tag`` (used for ``docker``) is the tag of the Docker image that results from
  ``pack docker``.

* ``--build-from`` (common for all distribution types, used for building in Docker) is
  the path to the base Dockerfile of the build image.
  Defaults to ``Dockerfile.build.cartridge`` in the application root.

* ``--sdk-local`` (common for all distribution types, used for building in Docker) is a
  flag that indicates if the SDK from the local machine should be delivered in the
  result artifact.

* ``--sdk-path`` (common for all distribution types, used for building in Docker) is the
  path to the SDK to be delivered in the result artifact.
  Alternatively, you can pass the path via the ``TARANTOOL_SDK_PATH``
  environment variable (this variable is of lower priority).

.. NOTE::

    For Tarantool Enterprise, you must specify one (and only one)
    of the ``--sdk-local`` and ``--sdk-path`` options.

For ``rpm``, ``deb``, and ``tgz``, we also deliver rocks modules and executables
specific for the system where the ``cartridge pack`` command is running.

For ``docker``, the resulting runtime image will contain rocks modules
and executables specific for the base image (``centos:8``).

Further on we dive deeper into the packaging process.

.. _cartridge-cli-build-directory:

****************
Build directory
****************

The first step of the packaging process is to
`build the application <Building an application_>`_.

By default, application build is done in a temporary directory in
``~/.cartridge/tmp/``, so the packaging process doesn't affect the contents
of your application directory.

You can specify a custom build directory for your application in the ``CARTRIDGE_BUILDDIR``
environment variable. If this directory doesn't exists, it will be created, used
for building the application, and then removed.

If you specify an existing directory in the ``CARTRIDGE_BUILDDIR`` environment
variable, the ``CARTRIDGE_BUILDDIR/build.cartridge`` repository will be used for
build and then removed. This directory will be cleaned up before building the
application.

.. NOTE::

    The specified directory cannot be an application subdirectory.

.. _cartridge-cli-distribution-directory:

***********************
Distribution directory
***********************

For each distribution type, a temporary directory with application source files
is created (further on we address it as *application directory*).
This includes 3 stages.

.. _stage-1-cleaning-up-the-application-directory:

^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Stage 1. Cleaning up the application directory
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

On this stage, some files are filtered out of the application directory:

* First, ``git clean -X -d -f`` removes all untracked and
  ignored files (it works for submodules, too).
* After that, ``.rocks`` and ``.git`` directories are removed.

Files permissions are preserved, and the code files owner is set to
``root:root`` in the resulting package.

.. NOTE::

    All application files should have at least ``a+r`` permissions
    (``a+rx`` for directories).
    Otherwise, ``cartridge pack`` command raises an error.

.. _stage-2-building-the-application:

^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Stage 2. Building the application
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

On this stage, ``cartridge`` `builds <Building an application_>`_
the application in the cleaned up application directory.

.. _stage-3-cleaning-up-the-files-before-packing:

^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Stage 3. Cleaning up the files before packing
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

On this stage, ``cartridge`` runs ``cartridge.post-build`` (if it exists) to remove
junk files (like ``node_modules``) generated during application build.

See an `example <Example cartridge postbuild_>`_
in `special files <Special files_>`_.

.. cartridge-cli-tgz:

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
TGZ
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

``cartridge pack tgz ./myapp`` creates a .tgz archive. It contains all files from the
`distribution directory <Distribution directory_>`_
(i.e. the application source code and rocks modules described in the application
rockspec).

The result artifact name is ``<name>-<version>[-<suffix>].tar.gz``.

.. cartridge-cli-rpm-and-deb:

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
RPM and DEB
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

``cartridge pack rpm|deb ./myapp`` creates an RPM or DEB package.

The result artifact name is ``<name>-<version>[-<suffix>].{rpm,deb}``.

**************
Usage example
**************

After package installation you need to specify configuration for instances to start.

For example, if your application is named ``myapp`` and you want to start two
instances, put the ``myapp.yml`` file into the ``/etc/tarantool/conf.d`` directory.

.. code-block:: yaml

    myapp:
      cluster_cookie: secret-cookie

    myapp.instance-1:
      http_port: 8081
      advertise_uri: localhost:3301

    myapp.instance-2:
      http_port: 8082
      advertise_uri: localhost:3302

For more details about instances configuration see the
`documentation <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_dev/#configuring-instances>`_.

Now, start the configured instances:

.. code-block:: console

    systemctl start myapp@instance-1
    systemctl start myapp@instance-2

If you use stateful failover, you need to start application stateboard.

.. NOTE::

    Your application should contain ``stateboard.init.lua`` in its root.

Add the ``myapp-stateboard`` section to ``/etc/tarantool/conf.d/myapp.yml``:

.. code-block:: yaml

    myapp-stateboard:
      listen: localhost:3310
      password: passwd

Then, start the stateboard service:

.. code-block:: console

    systemctl start myapp-stateboard

****************
Package details
****************

The installed package name will be ``<name>`` no matter what the artifact name is.

It contains meta information: the package name (which is the application name),
and the package version.

If you use an opensource version of Tarantool, the package has a ``tarantool``
dependency (version >= ``<major>.<minor>`` and < ``<major+1>``, where
``<major>.<minor>`` is the version of Tarantool used for packing the application).
You should enable the Tarantool repo to allow your package manager install
this dependency correctly:

* for RPM:

  .. code-block:: console

      curl -s \
              https://packagecloud.io/install/repositories/tarantool/${tarantool_repo_version}/script.rpm.sh | bash \
          && yum -y install tarantool tarantool-devel

* for DEB:

  .. code-block:: console

      curl -s \
              https://packagecloud.io/install/repositories/tarantool/${tarantool_repo_version}/script.deb.sh | bash \
          && apt-get -y install tarantool

The package contents is as follows:

* the contents of the distribution directory, placed in the
  ``/usr/share/tarantool/<app_name>`` directory
  (for Tarantool Enterprise, this directory also contains ``tarantool`` and
  ``tarantoolctl`` binaries);

* unit files for running the application as a ``systemd`` service:
  ``/etc/systemd/system/<app_name>.service`` and
  ``/etc/systemd/system/<app_name>@.service``;

* application stateboard unit file:
  ``/etc/systemd/system/<app_name>-stateboard.service``
  (will be packed only if the application contains ``stateboard.init.lua`` in its root);

* the file ``/usr/lib/tmpfiles.d/<app_name>.conf`` that allows the instance to restart
  after server restart.

These directories are created:

* ``/etc/tarantool/conf.d/`` -- directory for instances configuration;
* ``/var/lib/tarantool/`` -- directory to store instances snapshots;
* ``/var/run/tarantool/`` -- directory to store PID-files and console sockets.

See the `documentation <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_dev/#deploying-an-application>`_
for details about deploying a Tarantool Cartridge application.

To start the ``instance-1`` instance of the ``myapp`` service, say:

.. code-block:: console

    systemctl start myapp@instance-1

To start the application stateboard service, say:

.. code-block:: console

    systemctl start myapp-stateboard

This instance will look for its
`configuration <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_dev/#configuring-instances>`_
across all sections of the YAML file(s) stored in ``/etc/tarantool/conf.d/*``.

Use the options ``--unit-template`` and ``--instantiated-unit-template`` to customize
standard unit files.

.. NOTE::

    You may need it first of all for DEB packages, if your build platform
    is different from the deployment platform. In this case, ``ExecStartPre`` may
    contain an incorrect path to `mkdir`. As a hotfix, we suggest customizing the
    unit files.

Example of an instantiated unit file:

.. code-block:: kconfig

    [Unit]
    Description=Tarantool Cartridge app ${name}@%i
    After=network.target

    [Service]
    Type=simple
    ExecStartPre=/bin/sh -c 'mkdir -p ${workdir}.default'
    ExecStart=${bindir}/tarantool ${app_dir}/init.lua
    User=tarantool
    Group=tarantool

    Environment=TARANTOOL_WORKDIR=${workdir}.%i
    Environment=TARANTOOL_CFG=/etc/tarantool/conf.d/
    Environment=TARANTOOL_PID_FILE=/var/run/tarantool/${app_name}.%i.pid
    Environment=TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/${app_name}.%i.control
    Environment=TARANTOOL_INSTANCE_NAME=%i

In this file, you can use the following environment variables:

* ``app_name`` -- the application name;
* ``app_dir `` -- application files directory (by default, ``/usr/share/tarantool/<app_name>``)
* ``workdir`` -- path to the work directory (by default, ``/var/lib/tarantool/<app_name>``);
* ``bindir`` -- the directory, where Tarantool executable is placed.

.. _cartridge-cli-docker:

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Docker
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

``cartridge pack docker ./myapp`` builds a Docker image where you can start
one instance of the application.

**************
Usage example
**************

To start the ``instance-1`` instance of the ``myapp`` application, say:

.. code-block:: console

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

.. code-block:: console

    docker logs instance-1

******************
Runtime image tag
******************

The result image is tagged as follows:

* ``<name>:<detected_version>[-<suffix>]``: by default;
* ``<name>:<version>[-<suffix>]``: if the ``--version`` parameter is specified;
* ``<tag>``: if the ``--tag`` parameter is specified.

.. _cartridge-cli-build-and-runtime-images:

*************************
Build and runtime images
*************************

In fact, two images are created during the packing process:
build image and runtime image.

First, the build image is used to perform application build.
The build stages here are exactly the same as for other distribution types:

* `Stage 1. Cleaning up the application directory <Stage 1 Cleaning up the application directory_>`_
* `Stage 2. Building the application <Stage 2 Building the application_>`_
  (the build is always done `in Docker <Building in Docker_>`_)
* `Stage 3. Cleaning up the files before packaging <Stage 3 Cleaning up the files before packing_>`_

Second, the files are copied to the resulting (runtime) image, similarly
to packing an application as an archive. This image is exactly the
result of running ``cartridge pack docker``).

Both images are based on ``centos:8``.

All packages required for the default  ``cartridge`` application build
(``git``, ``gcc``, ``make``, ``cmake``, ``unzip``) are installed on the build image.

A proper version of Tarantool is provided on the runtime image:

* For opensource, Tarantool of the same version as the one used for
  local development is installed to the image.
* For Tarantool Enterprise, the bundle with Tarantool Enterprise binaries is
  copied to the image.

If your application requires some other applications for build or runtime, you
can specify base layers for build and runtime images:

* build image: ``Dockerfile.build.cartridge`` (default) or ``--build-from``;
* runtime image: ``Dockerfile.cartridge`` (default) or ``--from``.

The Dockerfile of the base image should be started with the ``FROM centos:8``
line (except comments).

For example, if your application requires ``gcc-c++`` for build and ``zip`` for
runtime, customize the Dockerfiles as follows:

* ``Dockerfile.cartridge.build``:

  .. code-block:: dockerfile

      FROM centos:8
      RUN yum install -y gcc-c++
      # Note that git, gcc, make, cmake, unzip packages
      # will be installed anyway

* `Dockerfile.cartridge`:

  .. code-block:: dockerfile

      FROM centos:8
      RUN yum install -y zip

*************************
Tarantool Enterprise SDK
*************************

If you use Tarantool Enterprise, you should explicitly specify the Tarantool SDK
to be delivered on the runtime image.

If you want to use the SDK from your local machine, just pass the ``--sdk-local``
flag to the ``cartridge pack docker`` command.

Alternatively, you can specify a local path to another SDK using the ``--sdk-path``
option (or the environment variable ``TARANTOOL_SDK_PATH``, which has lower priority).

********************************************
Customizing the application build in Docker
********************************************

You can pass custom arguments for the ``docker build`` command via the
``TARANTOOL_DOCKER_BUILD_ARGS`` environment variable.
For example, ``TARANTOOL_DOCKER_BUILD_ARGS='--no-cache --quiet'``

************************
Using the runtime image
************************

The application code is placed in the ``/usr/share/tarantool/${app_name}``
directory. An opensource version of Tarantool is installed to the image.

The run directory is ``/var/run/tarantool/${app_name}``,
the workdir is ``/var/lib/tarantool/${app_name}``.

The runtime image also contains the file ``/usr/lib/tmpfiles.d/<name>.conf``
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

.. _cartridge-cli-special-files:

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Special files
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

You can put these files in your application root to control the application
packaging process (see examples below):

* ``cartridge.pre-build``: a script to be run before ``tarantoolctl rocks make``.
  The main purpose of this script is to build some non-standard rocks modules
  (for example, from a submodule).

* ``cartridge.post-build``: a script to be run after ``tarantoolctl rocks make``.
  The main purpose of this script is to remove build artifacts from result package.

* [DEPRECATED] ``.cartridge.ignore``: here you can specify some files and directories
  to be excluded from the package build. See the
  `documentation <https://www.tarantool.io/ru/doc/latest/book/cartridge/cartridge_dev/#using-cartridge-ignore-files>`_
  for details.

* [DEPRECATED] ``.cartridge.pre``: a script to be run before ``tarantoolctl rocks make``.
  The main purpose of this script is to build some non-standard rocks modules
  (for example, from a submodule).

.. NOTE::

    * You can use any of these approaches (just take care not to mix them):

      * ``cartridge.pre-build`` + ``cartridge.post-build``, or
      * [deprecated] ``.cartridge.ignore`` + ``.cartridge.pre``.

    * Packing to a Docker image isn't compatible with the deprecated
      packaging process.

.. _cartridge-cli-example-cartridge-prebuild

*****************************
Example: cartridge.pre-build
*****************************

.. code-block:: console

    #!/bin/sh

    # The main purpose of this script is to build some non-standard rocks modules.
    # It will be run before `tarantoolctl rocks make` on application build

    tarantoolctl rocks make --chdir ./third_party/my-custom-rock-module

.. _cartridge-cli-example-cartridge-postbuild

******************************
Example: cartridge.post-build
******************************

.. code-block:: console

    #!/bin/sh

    # The main purpose of this script is to remove build artifacts from resulting package.
    # It will be ran after `tarantoolctl rocks make` on application build.

    rm -rf third_party
    rm -rf node_modules
    rm -rf doc
