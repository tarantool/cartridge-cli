.. _cartridge-cli:

Cartridge Command Line Interface
================================

.. image:: https://img.shields.io/github/v/release/tarantool/cartridge-cli?include_prereleases&label=Release&labelColor=2d3532
   :alt: Cartridge CLI latest release on GitHub
   :target: https://github.com/tarantool/cartridge-cli/releases

.. image:: https://github.com/tarantool/cartridge-cli/workflows/Tests/badge.svg
   :alt: Cartridge CLI build status on GitHub Actions
   :target: https://github.com/tarantool/cartridge-cli/actions/workflows/tests.yml

.. contents::

Installation
------------

1. Install third-party software:

   * `Install <https://git-scm.com/book/en/v2/Getting-Started-Installing-Git>`__
     ``git``, the version control system.

   * `Install <https://linuxize.com/post/how-to-unzip-files-in-linux/>`__
     the ``unzip`` utility.

   * `Install <https://gcc.gnu.org/install/>`__
     the ``gcc`` compiler.

   * `Install <https://cmake.org/install/>`__
     the ``cmake`` and ``make`` tools.

2. Install Tarantool 1.10 or higher.

   You can:

   * `Install it from a package <https://www.tarantool.io/en/download/>`__.
   * :doc:`Build it from source </dev_guide/building_from_source/>`.

3. [For all platforms except macOS] If you build Tarantool from source,
   you need to set up the Tarantool packages repository manually:

   .. code-block:: bash

       curl -L https://tarantool.io/installer.sh | sudo -E bash -s -- --repo-only

4. Install the ``cartridge-cli`` package:

   * For CentOS, Fedora, or ALT Linux (RPM package):

     .. code-block:: bash

         sudo yum install cartridge-cli

   * For Debian or Ubuntu (DEB package):

     .. code-block:: bash

         sudo apt-get install cartridge-cli

   * For macOS (Homebrew formula):

     .. code-block:: bash

         brew install cartridge-cli

5. Check the installation:

   .. code-block:: bash

      cartridge version

Now you can
:doc:`create and run </getting_started/getting_started_cartridge>`
your first application!

Quick start
-----------

To create your first application, run:

.. code-block:: bash

    cartridge create --name myapp

Go to the app directory:

.. code-block:: bash

    cd myapp

Finally, build and start your application:

.. code-block:: bash

    cartridge build
    cartridge start

You can now open http://localhost:8081 and see your application's Admin Web UI:

.. image:: https://user-images.githubusercontent.com/11336358/75786427-52820c00-5d76-11ea-93a4-309623bda70f.png
   :align: center

You're all set! Keep reading this document or follow the
:doc:`getting started with Cartridge </getting_started/getting_started_cartridge>` guide.

.. _cartridge-cli-usage:

Command-line completion
-----------------------

Linux
~~~~~

The ``cartridge-cli`` RPM and DEB packages contain a Bash completion script
for ``/etc/bash_completion.d/cartridge``.

To enable completion after ``cartridge-cli`` installation, open a new shell or
source the completion file at ``/etc/bash_completion.d/cartridge``.
Make sure that you have ``bash-completion`` installed.

To install Zsh completion, run:

.. code-block:: bash

    cartridge gen completion --skip-bash --zsh="${fpath[1]}/_cartridge"

Now enable shell completion:

.. code-block:: bash

    echo "autoload -U compinit; compinit" >> ~/.zshrc

OS X
~~~~

If you install ``cartridge-cli`` from ``brew``, it automatically installs both
Bash and Zsh completion.

Usage
-----

For more details, use the ``--help`` flag:

.. code-block:: bash

   cartridge --help

Here is a list of supported Cartridge CLI commands:

* ``create``: create a new application from template.
* ``build``: build an application for local development and testing.
* ``start``: start one or more Tarantool instances.
* ``stop``: stop one or more Tarantool instances.
* ``status``: get the status of one or more current instances.
* ``log``: get logs for one or more instances.
* ``clean``: clean files for one or more instances.
* ``pack``: pack the application into a distributable bundle.
* ``repair``: patch cluster configuration files.
* `admin <https://github.com/tarantool/cartridge-cli/blob/master/doc/admin.rst>`__:
  call an admin function provided by the application.
* `replicasets <https://github.com/tarantool/cartridge-cli/blob/master/doc/replicasets.rst>`__:
  manage cluster replica sets running locally.
* `enter <https://github.com/tarantool/cartridge-cli/blob/master/doc/connect.rst>`__
  and `connect <https://github.com/tarantool/cartridge-cli/blob/master/doc/connect.rst>`__:
  connect to a running instance.
* `failover <https://github.com/tarantool/cartridge-cli/blob/master/doc/failover.rst>`__:
  manage cluster failover.

You can control output verbosity with these global flags:

* ``verbose``: displays additional log messages as well as
  commands/docker output, such as the output of ``tarantoolctl rocks make`` or ``docker build``.
* ``debug``: works the same as verbose, but temporary files and
  directories aren't removed during command execution.
* ``quiet``: hides all logs, only displays error messages.

Application lifecycle
~~~~~~~~~~~~~~~~~~~~~

In a nutshell:

1. :ref:`Create an application <cartridge-cli-creating_an_application_from_template>`
   (for example, ``myapp``) from a template:

   .. code-block:: bash

       cartridge create --name myapp
       cd ./myapp

2. :ref:`Build the application <cartridge-cli-building-the-application>`
   for local development and testing:

   .. code-block:: bash

       cartridge build

3. :ref:`Run instances locally <cartridge-cli-starting-the-application-locally>`:

   .. code-block:: bash

       cartridge start
       cartridge stop

4. :ref:`Pack the application <cartridge-cli-packaging-the-application>`
   into a distributable (like an RPM package):

   .. code-block:: bash

       cartridge pack rpm

.. _cartridge-cli-creating_an_application_from_template:

Creating an application from a template
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

To create an application from a Cartridge template, run ``cartridge create`` in any directory:

..  code-block:: bash

    cartridge create [PATH] [flags]

Supported options (``[flags]``):

.. // Please update the doc in cli/commands on updating this section

*   ``--name string``: application name.

*   ``--from DIR``: path to the application template. For more about templates, see below.

*   ``--template string``: name of the application template to be used.
    Only the template ``cartridge`` is supported at the time.

Your application will appear in the ``<path>/<app-name>/`` directory.

If you have ``git`` installed, a Git repository with
a ``.gitignore`` file will be also set up in the project root directory.
The initial commit will be created, tagged with
:ref:`version <cartridge-versioning>` 0.1.0.

.. _cartridge-cli-project-directory:

Project directory
^^^^^^^^^^^^^^^^^

Let's take a closer look at the files inside the ``<app_name>/`` directory.

* Application files:

  * ``app/roles/custom-role.lua`` is a sample
    :ref:`custom role <cartridge-roles>`
    with a simple HTTP API. Can be enabled as ``app.roles.custom``.
  * ``<app_name>-scm-1.rockspec`` contains application
    dependencies.
  * ``init.lua`` is the application entry point.
  * ``stateboard.init.lua`` is the application
    :ref:`stateboard <cartridge-failover>` entry point.

* :ref:`Build and packaging files <cartridge-cli-special-files>`:

  * ``cartridge.pre-build``
  * ``cartridge.post-build``
  * ``Dockerfile.build.cartridge``
  * ``Dockerfile.cartridge``
  * ``package-deps.txt``
  * ``pack-cache-config.yml``

* Development files:

  * ``deps.sh`` resolves dependencies listed in the ``.rockspec`` file
    and installs test dependencies (like ``luatest``).
  * ``instances.yml`` contains the configuration of instances and is used by ``cartridge start``.
  * ``.cartridge.yml`` contains the Cartridge configuration and is also used by ``cartridge start``.
  * ``systemd-unit-params.yml`` contains systemd parameters.
  * ``tmp`` is a directory for temporary files, used as a run directory (see ``.cartridge.yml``).
  * ``.git`` is the directory responsible for the Git repository.
  * ``.gitignore`` is a file where you can specify the files for Git to ignore.
  * ``env.lua`` is a file that sets common rock paths,
    which allows you to start the application from any directory.

* Test files (with sample tests):

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

* Configuration files:

  * ``.luacheckrc``
  * ``.luacov``
  * ``.editorconfig``

.. _cartridge-cli-using-a-custom-template:

Using a custom template
^^^^^^^^^^^^^^^^^^^^^^^

The template used by default is ``cartridge``.
It produces a simple Cartridge application that includes:

* One custom role with an HTTP endpoint.
* Sample tests and basic test helpers.
* Files required for development (like ``.luacheckrc``).

To create an application based on your own custom template, run ``cartridge create`` with the ``--from`` flag.

If the template directory is a Git repository, all files in the ``.git`` directory will be ignored on
instantiating the template.
Instead, a new git repo will be initialized for the newly created application.

Don't include the ``.rocks`` directory in your template application.
To specify application dependencies, use the ``.rockspec`` and ``cartridge.pre-build`` files.

.. _cartridge-cli-text-variables:

Text variables
^^^^^^^^^^^^^^

File names and messages can include `text templates <https://golang.org/pkg/text/template/>`_.
You can use the following variables:

* ``Name``: application name.
* ``StateboardName``: application stateboard name (``<app-name>-stateboard``).
* ``Path``: absolute path to the application.

For example:

.. code-block:: text

    my-template
    ├── {{ .Name }}-scm-1.rockspec
    └── init.lua
    └── stateboard.init.lua
    └── test
        └── sample_test.lua

``init.lua``:

.. code-block:: lua

    print("Hi, I am the {{ .Name }} application")
    print("I also have a stateboard named {{ .StateboardName }}")

.. _cartridge-cli-building-the-application:

Building the application
~~~~~~~~~~~~~~~~~~~~~~~~

To build your application locally (for local testing), run this in any directory:

.. code-block:: bash

    cartridge build [PATH] [flags]

The following options (``[flags]``) are supported:

* ``--spec`` is the path to the ``.rockspec`` to use for the current build.
  *Note* that the ``.rockspec`` file name should be in lowercase.

.. // Please update the doc in cli/commands on updating this section

The command requires one argument---the path to your application directory
(that is, to the build source).
The default path is ``.`` (the current directory).

``cartridge build`` is executed in two steps:

1.  If there is a  :ref:`pre-build file <cartridge-cli-special-files>`,
    ``cartridge.pre-build`` builds the application in the ``[PATH]`` directory.
2.  If there is a :ref:`rockspec file <cartridge-cli-special-files>`,
    ``tarantoolctl rocks make`` installs all Lua rocks to the ``[PATH]`` directory.

First, ``cartridge`` builds the application inside the application directory.
This is different from ``cartridge pack``, which builds the application inside the
:ref:`build directory <cartridge-cli-build-directory>`.
No build artifacts remain in the application directory.

Second, ``cartridge`` installs all dependencies specified in the ``.rockspec`` file.
That file is located in the application directory created from template.

Alternatively, you can define the build logic in the rockspec in the form of ``cmake`` commands,
`like we do in Cartridge <https://github.com/tarantool/cartridge/blob/master/cartridge-scm-1.rockspec#L26>`_.

If your application depends on closed-source rocks, or if your build contains
rocks from a project added as a submodule, install all those
dependencies **before** calling ``tarantoolctl rocks make``.
You can do so using the file ``cartridge.pre-build`` in your application root.
That file is also located in the application directory created from template.
In ``cartridge.pre-build``, you can specify all the rocks to build
(for example, add ``tarantoolctl rocks make --chdir ./third_party/proj``).
For details, see :ref:`build and packaging files <cartridge-cli-special-files>`.

As a result, a fully built application will appear in the ``.rocks`` directory.
You can start it locally from the application directory.

.. _cartridge-cli-starting-the-application-locally:

Starting the application locally
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

After you've :ref:`built your application <cartridge-cli-building-the-application>`,
you can run it locally:

.. code-block:: bash

    cartridge start [INSTANCE_NAME...] [flags]

where ``[INSTANCE_NAME...]`` stands for one or multiple instance names.

If no ``INSTANCE_NAME`` is provided, all the instances in the
Cartridge instances configuration file will be taken as arguments (see the ``--cfg``
option below).

The application name, ``APP_NAME``, is passed to the instances during startup and
used in instance file paths,
for example: ``<run-dir>/<APP_NAME>.<INSTANCE_NAME>.pid``).
By default, ``APP_NAME`` is derived from the application rockspec in the current
directory. However, the variable also can be defined explicitly via the ``--name`` option
(see below).

.. _cartridge-cli-options:

Options
^^^^^^^

Supported options (``[flags]``):

.. // Please update the doc in cli/commands on updating this section

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``--script FILE``
            -   Application entry point.
                Can be an absolute or relative path to the entry point
                in the project directory.
                Defaults to ``init.lua`` or the value of the ``script`` parameter
                in the Cartridge `configuration file <cartridge-cli-overriding-default-options>`__.
        *   -   ``--run-dir DIR``
            -   The directory where PID and socket files are stored.
                Defaults to ``./tmp/run`` or the value of the ``run-dir`` parameter
                in the Cartridge `configuration file <cartridge-cli-overriding-default-options>`__.
        *   -   ``--data-dir DIR``
            -   The directory where instance data are stored.
                Each instance's working directory is named
                ``<data-dir>/<app-name>.<instance-name>``.
                Defaults to ``./tmp/data`` or the value of the ``data-dir`` parameter
                in the Cartridge :ref:`configuration file <cartridge-cli-overriding-default-options>`.
        *   -   ``--log-dir DIR``
            -   The directory to store instances logs when running in the background.
                Defaults to ``./tmp/log`` or the value of the ``log-dir`` parameter
                in the Cartridge :ref:`configuration file <cartridge-cli-overriding-default-options>`.
        *   -   ``--cfg FILE``
            -   Cartridge instance configuration file.
                Defaults to ``./instances.yml`` or the value of the ``cfg`` parameter.
                Read more about :ref:`using configuration files <cartridge-cli-configuration-files>`
                below.
        *   -   ``--daemonize, -d``
            -   Starts the instance(s) in the background.
                With this option, Tarantool also waits until the application's init script
                finishes evaluating.
                This is useful if ``init.lua`` requires time-consuming startup
                from a snapshot. Another use case would be if your application's init script
                generates errors, so Tarantool can handle them.
        *   -   ``--stateboard``
            -   Starts the application stateboard and the instances.
                Ignored if ``--stateboard-only`` is specified.
                The value can be indicated via the ``cfg`` parameter in the Cartridge
                :ref:`configuration file <cartridge-cli-overriding-default-options>`).
        *   -   ``--stateboard-only``
            -   Starts only the application stateboard.
                If specified, the ``INSTANCE_NAME...`` parameters are ignored.

        *   -   ``--name string``
            -   Defines the application name.
                By default, it is taken from the application rockspec.
        *   -   ``--timeout string``
            -   Time to wait for the instance(s) to start in the background.
                Can be specified in seconds or in the duration form (``72h3m0.5s``).
                Can't be negative.
                A ``0`` timeout means that Tarantool will wait forever for instance(s) to start.
                The default timeout is 60 seconds (``1m0s``).

.. _cartridge-cli-configuration-files:

Configuration files
^^^^^^^^^^^^^^^^^^^

The ``instances.yml`` file in the application directory contains parameters
for starting Cartridge application instances. These parameters are parsed on
:ref:`cartridge.cfg() <cartridge.cfg>`
call.

Example ``instances.yml`` file:

..  code-block:: yaml

    myapp.router:
        advertise_uri: localhost:3301
        http_port: 8081

    myapp.s1-master:
        advertise_uri: localhost:3302
        http_port: 8082

For the full list of parameters that can be specified in ``instances.yml``, read the
:ref:`cartridge.cfg() documentation <cartridge.cfg>`.
``advertise_uri`` is a required parameter.

..  note::

    The following parameters, if specified in ``instances.yml``, will be
    overwritten by Cartridge CLI environment variables on
    ``cartridge start``:

    * ``workdir``
    * ``console_sock``
    * ``pid_file``.

You can specify custom parameters in ``instances.yml``, but they also
have to be defined in your application code.
See the following example, where ``my_param`` is a custom parameter:

``instances.yml``:

..  code-block:: yaml

    myapp.router:
        advertise_uri: localhost:3301
        http_port: 8081
        my_param: 'Hello, world'

``init.lua``:

..  code-block:: lua

    local argparse = require('cartridge.argparse')
    local my_param = argparse.get_opts({my_param='string'})

.. _cartridge-cli-environment-variables:

Environment variables
^^^^^^^^^^^^^^^^^^^^^

The ``cartridge start`` command starts a Tarantool instance with enforced
**environment variables**:

..  code-block:: bash

    TARANTOOL_APP_NAME="<instance-name>"
    TARANTOOL_INSTANCE_NAME="<app-name>"
    TARANTOOL_CFG="<cfg>"
    TARANTOOL_PID_FILE="<run-dir>/<app-name>.<instance-name>.pid"
    TARANTOOL_CONSOLE_SOCK="<run-dir>/<app-name>.<instance-name>.control"
    TARANTOOL_WORKDIR="<data-dir>/<app-name>.<instance-name>.control"

When started in background, a notify socket path is passed additionally:

..  code-block:: bash

    NOTIFY_SOCKET="<data-dir>/<app-name>.<instance-name>.notify"

``cartridge.cfg()`` uses  ``TARANTOOL_APP_NAME`` and ``TARANTOOL_INSTANCE_NAME``
to read the instance's configuration from the file provided in ``TARANTOOL_CFG``.

.. _cartridge-cli-overriding-default-options:

Overriding default options
^^^^^^^^^^^^^^^^^^^^^^^^^^

You can override default options for the ``cartridge`` command in the
``./.cartridge.yml`` configuration file.

Here is an example of ``.cartridge.yml``:

.. code-block:: yaml

    run-dir: my-run-dir
    cfg: my-instances.yml
    script: my-init.lua
    stateboard: true

**Note:** the config of the
:ref:`standard application template <cartridge-cli-creating_an_application_from_template>`
initially has the ``stateboard`` parameter set to ``true``.

..  // Please update the doc in cli/commands on updating this section

..  _cartridge-cli-stopping-the-application-locally:

Stopping the application locally
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

To stop one or more instances, run:

.. code-block:: bash

    cartridge stop [INSTANCE_NAME...] [flags]

By default, the instances receive a SIGTERM.

Supported options (``[flags]``):

* ``-f, --force`` allows force-stopping the instance(s) with a SIGKILL.

`Some options <Options_>`_ are identical to those of the ``start`` command:

* ``--run-dir DIR``
* ``--cfg FILE``
* ``--stateboard``
* ``--stateboard-only``

.. note::

   When you call ``cartridge stop``, use the exact same ``run-dir`` as when
   you called ``cartridge start``.
   The PID files stored in that directory are used to stop the running instances.

.. // Please update the doc in cli/commands on updating this section

.. _cartridge-cli-checking-instance-status:

Checking instance status
~~~~~~~~~~~~~~~~~~~~~~~~

Use the ``status`` command to check the current instance status:

.. code-block:: bash

    cartridge status [INSTANCE_NAME...] [flags]

:ref:`Some options <cartridge-cli-options>` are identical to those of the ``start`` command:

* ``--run-dir DIR``
* ``--cfg FILE``
* ``--stateboard``
* ``--stateboard-only``

.. // Please update the doc in cli/commands on updating this section

.. _cartridge-cli-displaying-logs:

Displaying logs
~~~~~~~~~~~~~~~

Use the ``log`` command to display the logs of one or more instances running in the background:

.. code-block:: bash

    cartridge log [INSTANCE_NAME...] [flags]

Supported options (``[flags]``):

* ``-f, --follow`` outputs appended data as the log grows.

* ``-n, --lines int`` is the number of last lines to be displayed.
  Defaults to 15.

:ref:`Some options <cartridge-cli-options>` are identical to those of the ``start`` command:

* ``--log-dir DIR``
* ``--run-dir DIR``
* ``--cfg FILE``
* ``--stateboard``
* ``--stateboard-only``

.. // Please update the doc in cli/commands on updating this section

.. _cartridge-cli-cleaning-instance-files:

Cleaning instance files
~~~~~~~~~~~~~~~~~~~~~~~

Use the ``clean`` command to remove the files associated with one or more instances
(such as the log file, workdir, console socket, PID file and notify socket):

.. code-block:: bash

    cartridge clean [INSTANCE_NAME...] [flags]

Executing ``cartridge clean`` for running instance(s) causes an error.

:ref:`Some options <cartridge-cli-options>` are identical to those of the ``start`` command:

* ``--log-dir DIR``
* ``--data-dir DIR``
* ``--run-dir DIR``
* ``--cfg FILE``
* ``--stateboard``
* ``--stateboard-only``

.. // Please update the doc in cli/commands on updating this section

.. _cartridge-cli-packaging-the-application:

Packaging the application
~~~~~~~~~~~~~~~~~~~~~~~~~

To pack your application, run this in any directory:

..  code-block:: bash

    cartridge pack TYPE [PATH] [flags]

where:

* ``TYPE`` (required) is the distribution type. Supported types:

  * `TGZ <TGZ_>`_
  * `RPM <RPM and DEB_>`_
  * `DEB <RPM and DEB_>`_
  * `Docker <Docker_>`_

* ``PATH`` (optional) is the path to the application directory that you want to pack.
  Defaults to ``.`` (the current directory).

.. note::

  If you pack your application into an RPM or DEB on MacOS without the ``--use-docker``
  flag, the final artifact will be broken, because it will contain rocks and executables
  that can't be used on Linux. In this case packing will fail.

Supported options (``[flags]``):

.. // Please update cmd_pack usage in cartridge-cli.lua file on updating the doc

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``--name string``
            -   Application name. Common for all distribution types.
                Same as the package name and the systemd service name.
                Derived from the ``package`` field in the ``.rockspec`` file by default.
        *   -   ``--spec``
            -   Path to the ``.rockspec`` file to use for the current build.
                Note that the file name should be in *lowercase*.
        *   -   ``--version string``
            -   Package version. Common for all distribution types.
                Expected pattern: ``major.minor.patch[-count][-commit]``.
                If you specify the version in the ``major.minor.patch``format,
                it will be normalized to ``major.minor.patch-count``.
                By default, the version string is the output of ``git describe --tags --long``.
                If your application is not a git repository,
                you have to set the ``--version`` option explicitly.
        *   -   ``--suffix string``
            -   Suffix of the resulting  file (or image) name.
                Common for all distribution types.
        *   -   ``--unit-template string``
            -   Used for ``rpm`` and ``deb`` packages only.
                Path to the template for the ``systemd`` unit file.
        *   -   ``--instantiated-unit-template string``
            -   Used for ``rpm`` and ``deb`` packages only.
                Path to the template for the ``systemd`` instantiated unit file.
        *   -   ``--stateboard-unit-template string``
            -   Used for ``rpm`` and ``deb`` packages only.
                Path to the template for the stateboard ``systemd`` unit file.
        *   -   ``--use-docker``
            -   Enforced for ``docker``.
                Forces to build the application in Docker.
        *   -   ``--tag strings``
            -   Used for ``docker`` only.
                Tag(s) of the Docker image that results from ``pack docker``.
        *   -   ``--from string``
            -   Used for ``docker`` only.
                Path to the base Dockerfile of the runtime image.
                Defaults to ``Dockerfile.cartridge`` in the application directory.
        *   -   ``--build-from string``
            -   Common for all distribution types, used for building in Docker.
                Path to the base Dockerfile of the build image.
                Defaults to ``Dockerfile.build.cartridge`` in the application directory.
        *   -   ``--no-cache``
            -   Creates build and runtime images with the ``--no-cache`` Docker flag.
        *   -   ``--cache-from strings``
            -   Images that work as cache sources for both build and runtime images.
                See the ``--cache-from`` flag for the ``docker build`` command.
        *   -   ``--sdk-path string``
            -   Common for all distribution types.
                Path to the SDK that will be delivered in the final artifact.
                Alternatively, you can pass the path via the ``TARANTOOL_SDK_PATH``
                environment variable. However, this variable has lower priority.
        *   -   ``--sdk-local``
            -   Common for all distribution types, used for building in Docker.
                Indicates that the SDK from the local machine
                should be delivered in the final artifact.
        *   -   ``--deps``
            -   Used for ``rpm`` and ``deb`` packages only.
                Defines the dependencies of the package.
        *   -   ``--deps-file``
            -   Used for ``rpm`` and ``deb`` packages only.
                Path to the file that contains package dependencies.
                Defaults to ``package-deps.txt`` in the application directory.
        *   -   ``--preinst``
            -   Used for ``rpm`` and ``deb`` packages only.
                Path to the pre-install script for RPM and DEB packages.
        *   -   ``--postinst``
            -   Used for ``rpm`` and ``deb`` packages only.
                Path to the post-install script for RPM and DEB packages.
        *   -   ``--unit-params-file``
            -   Used for ``rpm`` and ``deb`` packages only.
                Path to the file that contains unit parameters for ``systemd`` unit files.

Example of the file containing package dependencies:

..  code-block:: text

    dependency_01 >= 2.5
    dependency_01 <
    dependency_02 >= 1, < 5
    dependency_03==2
    dependency_04<5,>=1.5.3

Each line must describe a single dependency.
You can specify both the major and minor version of the dependency:

..  code-block:: bash

    dependency_05 >= 4, < 5

The ``--deps`` and ``--deps-file`` flags require similar formats of dependency information.
However, ``--deps`` does not allow you to specify major and minor versions:

..  code-block:: bash

    # You can't do that:
    cartridge pack rpm --deps dependency_06>=4,<5 appname

    # Instead, do this:
    cartridge pack rpm --deps dependency_06>=4,dependency_06<5 appname

    # Or this:
    cartridge pack rpm --deps dependency_06>=4 --deps dependency_06<5 appname

For Tarantool Enterprise, specify either ``--sdk-local`` or ``--sdk-path``
(not both at the same time).

For ``rpm``, ``deb``, and ``tgz``, rocks and executables are also included in the build.
The executables are specific for the system where you run ``cartridge pack``.

For ``docker``, the resulting runtime image will contain rocks modules
and executables specific for the base image (``centos:8``).

The default pre-install script for ``rpm`` and ``deb`` packages is ``preinst.sh``,
and the default post-install script for those packages is ``postinst.sh``.
Both files are located in the project directory.
If your project directory contains a pre- or post-install script with that default name,
you don't have to use ``--preinst`` or ``--postinst``.

Provide absolute paths to executables in the pre- and post-install scripts
or use ``/bin/sh -c ''`` instead.

Example of a pre/post-install script:

..  code-block:: bash

    /bin/sh -c 'touch file-path'
    /bin/sh -c 'mkdir dir-path'
    # or
    /bin/mkdir dir-path

The package generates ``VERSION.lua``, a file that contains the current version
of the project. When you connect to an instance with
`cartridge connect <https://github.com/tarantool/cartridge-cli/blob/master/doc/connect.rst>`__,
you can check the project version by obtaining information from this file:

..  code-block:: lua

    require('VERSION')

This file is also used when you call
:ref:`cartridge.reload_roles() <cartridge.reload_roles>`:

..  code-block:: lua

    -- Getting the project version
    require('VERSION')
    -- Reloading the instances after making some changes to VERSION.lua
    require('cartridge').reload_roles()
    -- Getting the updated project version
    require('VERSION')

..  note::

    If there is already a ``VERSION.lua`` file in the application directory,
    it will be overwritten during packaging.

You can pass parameters to unit files. To do so,
specify the file containing the parameters using the ``--unit-params-file`` flag.
The ``fd-limit`` option allows limiting the number of file descriptors
determined by the ``LimitNOFILE`` parameter in the ``systemd`` unit file and
the ``systemd`` instantiated unit file.
The ``stateboard-fd-limit`` allows setting the file descriptor limit
in the stateboard ``systemd`` unit file.

..  TODO PROOFREAD next paragraph

You can pass parameters by env with systemd unit file by specifying instance and
stateboard arguments in ``systemd-unit-params.yml``. Parameter from
``systemd-unit-params.yml`` converts to ``Environment=TARANTOOL_<PARAM>: <value>``
in the unit file. Note that these variables have higher priority than variables
specified later in the instance configuration file.

..  code-block:: yaml

    fd-limit: 1024
    stateboard-fd-limit: 2048
    instance-env:
        app-name: 'my-app'
        net_msg_max: 1024
        pid_file: '/some/special/dir/my-app.%i.pid'
        my-param: 'something'
        # or
        # TARANTOOL_MY_PARAM: 'something'
    stateboard-env:
        app-name: 'my-app-stateboard'
        pid_file: '/some/special/dir/my-app-stateboard.pid'

You can pass parameters to the systemd unit file by env.
To do so, specify the instance and stateboard arguments in ``systemd-unit-params.yml``.
Each parameter from ``systemd-unit-params.yml`` converts to
``Environment=TARANTOOL_<PARAM>: <value>`` in the unit file.
Note that these variables have higher priority than the variables
specified later in the instance configuration file.

.. code-block:: yaml

    instance-env:
        app-name: 'my-app'
        net_msg_max: 1024
        pid_file: '/some/special/dir/my-app.%i.pid'
        my-param: 'something'
        # or
        # TARANTOOL_MY_PARAM: 'something'
    stateboard-env:
        app-name: 'my-app-stateboard'
        pid_file: '/some/special/dir/my-app-stateboard.pid'

Some ``systemd`` unit parameters can be listed in the ``systemd-unit-params.yml``
file in the project directory. You can also use a file with a different name,
specifying it in the ``--unit-params-file`` option.

Supported options:

* ``fd-limit``: the ``LimitNOFILE`` option for an application instance.

* ``stateboard-fd-limit``: the ``LimitNOFILE`` option for a stateboard instance.

* ``stateboard-fd-limit`` - ``LimitNOFILE`` option for stateboard instance;

* ``instance-env``: environment variables for
  :doc:`cartridge.argparse </book/cartridge/cartridge_api/modules/cartridge.argparse>`
  (like ``net-msg-max``) for an application instance.

* ``stateboard-env``: environment variables for
  :doc:`cartridge.argparse </book/cartridge/cartridge_api/modules/cartridge.argparse>`
  (like ``net-msg-max``) for the stateboard instance.

We provide the ability to cache paths for packaged applications. For example, you
package an application multiple times, and the same rocks are installed each time.
You can speed up the repack process by specifying cached paths in the ``pack-cache-config.yml``
file. By default, we suggest caching the ``.rocks`` directory - we put this path in
the standard application template.

..  code-block:: yaml

    - path: '.rocks':
      key-path: 'myapp-scm-1.rockspec'
    - path: 'node_modules':
      always-cache: true
    - path: 'third_party/custom_module':
      key: 'simple-hash-key'

You must specify the path to the directory from the root of the application
and specify the cache key. In the example above:

* ``<path-to-myapp>/.rocks`` path will be cached depending on the content of the ``myapp-scm-1.rockspec`` file
* ``<path-to-myapp>/node_modules`` path will always be cached
* ``<path-to-myapp>/third_party/custom_module`` path will be cached depending on the ``simple-hash-key`` key

You can't combine these options. For example, you can't specify the ``always-cache``
and ``key-path`` flags at the same time.

One project path can only store one caching key. For example, you have cached ``.rocks``
with ``key-path`` as a ``.rockspec`` file. You have changed the contents of the ``.rockspec``
file and run the ``cartridge pack``. In such case, old cache (for the old key) for the
``.rocks`` path of this project will be deleted. After packing, current ``.rocks`` cache
path will be saved with the new key.

In addition, there can be no more than **5** projects in the cache that have
cached paths. If a 6th project appears, oldest existing project is removed
from cache directory. But this is not the case for cached project paths: you can
cache as many paths as you like for one project.

You can always disable caching by using the ``--no-cache`` flag or by removing
paths from the ``pack-cache-config.yml`` file. To completely reset the cache,
delete ``~/.cartridge/tmp/cache`` directory.

* ``stateboard-env``: environment variables for
  :doc:`cartridge.argparse </book/cartridge/cartridge_api/modules/cartridge.argparse>`
  (like ``net-msg-max``) for a stateboard instance.

The paths you use in packaged applications can be cached. This can be useful if you
package your application multiple times, so that each time the same rocks are installed.
To speed up the repackaging process, list the paths you want to cache
in the ``pack-cache-config.yml`` file.
We suggest caching the ``.rocks`` directory and did so in the default application template:

..  code-block:: yaml

    - path: '.rocks':
      key-path: 'myapp-scm-1.rockspec'
    - path: 'node_modules':
      always-cache: true
    - path: 'third_party/custom_module':
      key: 'simple-hash-key'

Specify every path as related to the application directory and provide caching keys.
In the example above:

* ``<path-to-myapp>/.rocks`` will be cached depending on the content of the ``myapp-scm-1.rockspec`` file.
* ``<path-to-myapp>/node_modules`` will always be cached.
* ``<path-to-myapp>/third_party/custom_module`` will be cached depending on the ``simple-hash-key`` key.

You can't combine these options---for example, you can't use ``always-cache``
and ``key-path`` at the same time.

Every project path can only store a single caching key. Suppose that you cached
``.rocks`` in your project and provided a ``.rockspec`` file as the ``key-path``.
Then, if you change the contents of your ``.rockspec`` file and run ``cartridge pack``,
the old ``.rocks`` cache will be deleted, because it depended on the old key.
After packaging, the current ``.rocks`` cache path will be saved with the new key.

In addition, the cache cannot contain more than **5** projects that have cached paths.
If the 6th project appears, the oldest project will be removed from the cache directory.
However, this is not the case for cached project paths:
you can cache as many paths as you like for one project.

You can always disable caching by using the ``--no-cache`` flag or by removing
paths from ``pack-cache-config.yml``. To completely reset the cache,
delete the ``~/.cartridge/tmp/cache`` directory.

Next, let's dive deeper into the packaging process.

.. _cartridge-cli-build-directory:

Build directory
^^^^^^^^^^^^^^^

The first step of the packaging process is to
:ref:`build the application <cartridge-cli-building-the-application>`.

By default, the application is built inside a temporary directory in
``~/.cartridge/tmp/``, so that packaging doesn't affect the contents
of your application directory.
All the application source files are copied to that temporary directory.

You can specify a custom build directory for your application in the ``CARTRIDGE_TEMPDIR``
environment variable. If that directory doesn't exist yet, it will be created, used
for building the application, and then removed.

If you specify an existing directory in the ``CARTRIDGE_TEMPDIR`` environment
variable, the temporary ``CARTRIDGE_TEMPDIR/cartridge.tmp`` directory will be created in it.
That nested directory will be cleaned up before building the application.

The temporary build directory is what becomes the distribution package,
so it will be referred to as `distribution directory` from now on.

The build process has three stages.

.. _stage-1-cleaning-up-the-distribution-directory:

Build stage 1: Cleaning up the distribution directory
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Some files are filtered out of the directory:

* First, ``git clean -X -d -f`` removes all untracked and
  ignored files (in submodules too).
* After that, the ``.rocks`` and ``.git`` directories are removed.

All file permissions are preserved,
and the code files owner is set to ``root:root`` in the resulting package.

All application files must have at least ``a+r`` permissions
(``a+rx`` for directories).
Otherwise, the ``cartridge pack`` command will raise an error.

.. _stage-2-building-the-application:

Build stage 2. Building the application
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

At this stage, ``cartridge`` :ref:`builds <cartridge-cli-building-the-application>`
the application in the cleaned-up distribution directory.

.. _stage-3-cleaning-up-before-packaging:

Build stage 3. Cleaning up before packaging
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Cartridge runs ``cartridge.post-build``, if it exists, to remove
junk files generated during application build (such as ``node_modules``).

See the :ref:`example <cartridge-cli-example-cartridge-postbuild>`
in the section on :ref:`build and packaging files <cartridge-cli-special-files>`.

.. _cartridge-cli-repair:

Repairing a cluster
~~~~~~~~~~~~~~~~~~~

The ``cartridge repair`` command repairs a running application.

Here are several simple rules you need to know before using this command:

1. Don't use the ``repair`` command if you aren't sure it's exactly what you need.
2. Always use ``--dry-run`` before running ``repair``.
3. Do not hesitate to use the ``--verbose`` option.
4. Do not use the ``--force`` option if you aren't sure it's exactly what you need.

Please look at the
:doc:`troubleshooting documentation </book/cartridge/troubleshooting>`
before using ``repair``.

What does ``repair`` actually do?

It patches cluster-wide instance configuration files that you have on your local machine.
Note that it's not enough to *apply* the new configuration, the instance has to *reload* it.

``repair`` was created for production use, but it still can be applied in
local development. The command requires the application name option, ``--name``.
Remember also that the default data directory is ``/var/lib/tarantool`` and
the default run directory is ``/var/run/tarantool``.
You can define other directories using the corresponding options.

In default mode, ``repair`` walks through all cluster-wide configurations
in ``<data-dir>/<app-name>.*`` directories and patches all the configuration
files it locates.

With the ``--dry-run`` flag specified, files won't be patched,
and you will only see the computed configuration diff.

If configuration files differ between instances on the local machine,
``repair`` raises an error.
To patch different versions of configuration independently,
specify the ``--force`` option.

If your application uses ``cartridge >= 2.0.0``,
you can also run ``repair`` with the ``--reload`` flag
to reload configuration for all your instances.
Configuration will be reloaded for all instances
using the console sockets in the run directory.
When using the ``--reload`` flag, make sure that you specify the right run directory.

.. code-block:: bash

    cartridge repair [command]

Here is a list of ``repair`` commands
(see :ref:`details <cartridge-cli-repair-commands>` below):

* ``list-topology``: show the current topology summary.
* ``remove-instance``: remove an instance from the cluster.
* ``set-leader``: change a replica set leader.
* ``set-uri``: change an instance's advertise_uri parameter.

All repair commands have the following flags:

* ``--name`` (required) is the application name.

* ``--data-dir`` is the directory storing instance data (defaults to ``/var/lib/tarantool``).

All commands except ``list-topology`` have the following flags:

* ``--run-dir`` is the directory storing PID and socket files (defaults to ``/var/run/tarantool``).

* ``--dry-run`` runs the ``repair`` command in the dry run mode,
  displaying changes without applying them.

* ``--reload`` enables reloading configuration on instances after the patch.

.. _cartridge-cli-repair-commands:

Repair commands
^^^^^^^^^^^^^^^

**Topology summary**

.. code-block:: bash

    cartridge repair list-topology [flags]

Takes no arguments. Prints the current topology summary.

**Remove instance**

.. code-block:: bash

    cartridge repair remove-instance UUID [flags]

Removes an instance with the specified UUID from the cluster.
If the specified instance isn't found, raises an error.

**Set leader**

.. code-block:: bash

    cartridge repair set-leader REPLICASET-UUID INSTANCE-UUID [flags]

Sets the specified instance as the leader of the specified replica set.
Raises an error in the following cases:

* There is no replica set or instance with that UUID.
* The instance doesn't belong to the replica set.
* The instance has been disabled or expelled.

**Set advertise_uri**

.. code-block:: bash

    cartridge repair set-uri INSTANCE-UUID URI-TO [flags]

Rewrites the advertise_uri parameter for the specified instance.
If the instance isn't found or is expelled, the command raises an error.


.. _cartridge-cli-tgz:

TGZ
~~~

``cartridge pack tgz ./myapp`` creates a .tgz archive. It contains all files from the
:ref:`distribution directory <cartridge-cli-build-directory>` --
the application source code and rocks modules described in the application ``.rockspec``.

The final artifact name is ``<name>-<version>[-<suffix>].tar.gz``.


.. _cartridge-cli-rpm-and-deb:

RPM and DEB
~~~~~~~~~~~

``cartridge pack rpm|deb ./myapp`` creates an RPM or DEB package.

The final artifact name is ``<name>-<version>[-<suffix>].{rpm,deb}``.

Usage example
^^^^^^^^^^^^^

After the package is installed, you have to provide configuration for instances to start.

For example, to start two instances of your application, ``myapp``,
put the ``myapp.yml`` file in the ``/etc/tarantool/conf.d`` directory:

.. code-block:: yaml

    myapp:
      cluster_cookie: secret-cookie

    myapp.instance-1:
      http_port: 8081
      advertise_uri: localhost:3301

    myapp.instance-2:
      http_port: 8082
      advertise_uri: localhost:3302

For more about instance configuration, see the
:ref:`documentation <cartridge-config>`.

Now start the configured instances:

.. code-block:: bash

    systemctl start myapp@instance-1
    systemctl start myapp@instance-2

If you use stateful failover, start the application stateboard, too.
Remember that in this case, you must have ``stateboard.init.lua`` in the application directory.

Add the ``myapp-stateboard`` section to ``/etc/tarantool/conf.d/myapp.yml``:

.. code-block:: yaml

    myapp-stateboard:
      listen: localhost:3310
      password: passwd

Then, start the stateboard service:

.. code-block:: bash

    systemctl start myapp-stateboard

Package details
^^^^^^^^^^^^^^^

The installed package name will be ``<name>`` no matter what the artifact name is.

The package contains metadata, specifically its name (which is the application name)
and version.

If you use an open source version of Tarantool, the package has a ``tarantool``
dependency (version >= ``<major>.<minor>`` and < ``<major+1>``, where
``<major>.<minor>`` is the version of Tarantool used for packaging the application).
Enable the Tarantool repo so that your package manager installs the dependency correctly:

* for both RPM and DEB:

  .. code-block:: bash

      curl -L https://tarantool.io/installer.sh | VER=${TARANTOOL_VERSION} bash

After unpacking, the contents of the package are placed in specific locations:

* The contents of the distribution directory are placed at
  ``/usr/share/tarantool/<app-name>``.
  In case of Tarantool Enterprise, this directory also contains the ``tarantool`` and
  ``tarantoolctl`` binaries.

* The unit files for running the application as a ``systemd`` service
  are unpacked as ``/etc/systemd/system/<app-name>.service`` and
  ``/etc/systemd/system/<app-name>@.service``.

* The application stateboard unit file is unpacked as
  ``/etc/systemd/system/<app-name>-stateboard.service``.
  It is packed only if there is a ``stateboard.init.lua`` file
  in the application directory.

* The file ``/usr/lib/tmpfiles.d/<app-name>.conf`` allows the instance to restart
  after server reboot.

The following directories are created:

* ``/etc/tarantool/conf.d/`` stores instance configuration.
* ``/var/lib/tarantool/`` stores instance snapshots.
* ``/var/run/tarantool/`` stores PID files and console sockets.

See the :ref:`documentation <cartridge-deploy>`
for details about deploying a Tarantool Cartridge application.

To start ``instance-1`` of the ``myapp`` service, run:

.. code-block:: bash

    systemctl start myapp@instance-1

To start the application stateboard service, run:

.. code-block:: bash

    systemctl start myapp-stateboard

The instance will look for its :ref:`configuration <cartridge-config>`
across all YAML files stored in ``/etc/tarantool/conf.d/``.

Use the options ``--unit-template``, ``--instantiated-unit-template`` and
``--stateboard-unit-template`` to customize standard unit files.
This may be especially useful for DEB packages, if your build platform
is different from the deployment platform. In this case, ``ExecStartPre`` may
contain an incorrect path to `mkdir`. As a hotfix, we suggest customizing the
unit files.

Example of an instantiated unit file:

..  code-block:: kconfig

    [Unit]
    Description=Tarantool Cartridge app {{ .Name }}@%i
    After=network.target

    [Service]
    Type=simple
    ExecStartPre=/bin/sh -c 'mkdir -p {{ .InstanceWorkDir }}'
    ExecStart={{ .Tarantool }} {{ .AppEntrypointPath }}
    Restart=on-failure
    RestartSec=2
    User=tarantool
    Group=tarantool

    Environment=TARANTOOL_APP_NAME={{ .Name }}
    Environment=TARANTOOL_WORKDIR={{ .InstanceWorkDir }}
    Environment=TARANTOOL_CFG={{ .ConfPath }}
    Environment=TARANTOOL_PID_FILE={{ .InstancePidFile }}
    Environment=TARANTOOL_CONSOLE_SOCK={{ .InstanceConsoleSock }}
    Environment=TARANTOOL_INSTANCE_NAME=%i

    LimitCORE=infinity
    # Disable OOM killer
    OOMScoreAdjust=-1000
    # Increase fd limit for Vinyl
    LimitNOFILE=65535

    # Systemd waits until all xlogs are recovered
    TimeoutStartSec=86400s
    # Give a reasonable amount of time to close xlogs
    TimeoutStopSec=10s

    [Install]
    WantedBy=multi-user.target
    Alias={{ .Name }}.%i

Supported variables:

..  container:: table

    ..  list-table::
        :widths: 20 80
        :header-rows: 0

        *   -   ``Name``
            -   Application name.
        *   -   ``StateboardName``
            -   Application stateboard name (``<app-name>-stateboard``).
        *   -   ``DefaultWorkDir``
            -   Default instance working directory
                (``/var/lib/tarantool/<app-name>.default``).
        *   -   ``InstanceWorkDir``
            -   Application instance working directory
                (``/var/lib/tarantool/<app-name>.<instance-name>``).
        *   -   ``StateboardWorkDir``
            -   Stateboard working directory
                (``/var/lib/tarantool/<app-name>-stateboard``).
        *   -   ``DefaultPidFile``
            -   Default instance PID file (``/var/run/tarantool/<app-name>.default.pid``).
        *   -   ``InstancePidFile``
            -   Application instance PID file
                (``/var/run/tarantool/<app-name>.<instance-name>.pid``).
        *   -   ``StateboardPidFile``
            -   Stateboard PID file (``/var/run/tarantool/<app-name>-stateboard.pid``).
        *   -   ``DefaultConsoleSock``
            -   Default instance console socket
                (``/var/run/tarantool/<app-name>.default.control``).
        *   -   ``InstanceConsoleSock``
            -   Application instance console socket
                (``/var/run/tarantool/<app-name>.<instance-name>.control``).
        *   -   ``StateboardConsoleSock``
            -   Stateboard console socket (``/var/run/tarantool/<app-name>-stateboard.control``).
        *   -   ``ConfPath``
            -   Path to the application instances config (``/etc/tarantool/conf.d``).
        *   -   ``AppEntrypointPath``
            -   Path to the application entrypoint
                (``/usr/share/tarantool/<app-name>/init.lua``).
        *   -   ``StateboardEntrypointPath``
            -   Path to the stateboard entrypoint
                (``/usr/share/tarantool/<app-name>/stateboard.init.lua``).

.. _cartridge-cli-docker:

Docker
~~~~~~

``cartridge pack docker ./myapp`` builds a Docker image where you can start
one instance of the application.

Usage example
^^^^^^^^^^^^^

To start ``instance-1`` of the ``myapp`` application, run:

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

You can set ``CARTRIDGE_RUN_DIR``, ``CARTRIDGE_DATA_DIR`` environment variables.

.. code-block:: bash

    docker run -d \
                    --name instance-1 \
                    -e CARTRIDGE_RUN_DIR=my-custom-run-dir \
                    -e CARTRIDGE_DATA_DIR=my-custom-data-dir \
                    -e TARANTOOL_ADVERTISE_URI=3302 \
                    -e TARANTOOL_CLUSTER_COOKIE=secret \
                    -e TARANTOOL_HTTP_PORT=8082 \
                    -p 127.0.0.1:8082:8082 \
                    myapp:1.0.0

The variable ``CARTRIDGE_DATA_DIR`` is the working directory
that contains the pid file and the console socket.
It is set to ``/var/lib/tarantool`` by default.

You can also set the variables ``TARANTOOL_WORKDIR``, ``TARANTOOL_PID_FILE``,
and ``TARANTOOL_CONSOLE_SOCK``.

.. code-block:: bash

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

.. code-block:: bash

    docker logs instance-1

Runtime image tag
^^^^^^^^^^^^^^^^^

The final image is tagged as follows:

* ``<name>:<detected_version>[-<suffix>]`` by default.
* ``<name>:<version>[-<suffix>]`` if the ``--version`` parameter is specified.
* ``<tag>`` if the ``--tag`` parameter is specified.

.. _cartridge-cli-build-and-runtime-images:

Build and runtime images
^^^^^^^^^^^^^^^^^^^^^^^^

In fact, two images are created during the packing process:
the build image and the runtime image.

First, the build image is used to build the application.
The building stages here are exactly the same as for other distribution types:

* :ref:`Stage 1. Cleaning up the distribution directory <stage-1-cleaning-up-the-distribution-directory>`.
* :ref:`Stage 2. Building the application <stage-2-building-the-application>`.
  It is always performed :ref:`in Docker <cartridge-cli-docker>`.
* :ref:`Stage 3. Cleaning up before packaging <stage-3-cleaning-up-before-packaging>`.

Second, the files are copied to the resulting runtime image. This is similar
to packaging the application as an archive. The runtime image is the direct result
of running ``cartridge pack docker``.

Both images are based on ``centos:8``.

All packages required for the default  ``cartridge`` application build
(``git``, ``gcc``, ``make``, ``cmake``, ``unzip``) are installed in the build image.

The proper version of Tarantool is provided in the runtime image.

* If you use open-source Tarantool, the image will contain the same version of Tarantool
  that you used for local development.
* If you use Tarantool Enterprise, the bundle with Tarantool Enterprise binaries
  will be copied to the image.

If your application build or runtime requires other applications,
you can specify the base layers for your build and runtime images:

* Build image: ``Dockerfile.build.cartridge`` (default) or ``--build-from``.
* Runtime image: ``Dockerfile.cartridge`` (default) or ``--from``.

The Dockerfile of your base image must start with ``FROM centos:8``
or ``FROM centos:7`` (apart from comments).

We expect the base docker image to be ``centos:8`` or ``centos:7``,
but you can use any other distribution.

For example, if your application requires ``gcc-c++`` for the build and ``zip`` for
the runtime, customize your Dockerfiles as follows:

* ``Dockerfile.cartridge.build``:

  .. code-block:: dockerfile

      FROM centos:8
      RUN yum install -y gcc-c++
      # Note that git, gcc, make, cmake, and unzip
      # will be installed anyway

* ``Dockerfile.cartridge``:

  .. code-block:: dockerfile

      FROM centos:8
      RUN yum install -y zip

.. _cartridge-cli-tarantool-enterprise-sdk:

Tarantool Enterprise SDK
^^^^^^^^^^^^^^^^^^^^^^^^

If you use Tarantool Enterprise, you have to explicitly specify the Tarantool SDK
to be delivered in the runtime image.

To use the SDK from your local machine, pass the ``--sdk-local``
flag to the ``cartridge pack docker`` command.

Alternatively, specify a local path to another SDK using the ``--sdk-path``
option or the environment variable ``TARANTOOL_SDK_PATH``, which has lower priority.

Customizing the Docker application build
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

When building your application with ``docker build``,
you can use the options ``--cache-from`` and ``--no-cache``.

Using the runtime image
^^^^^^^^^^^^^^^^^^^^^^^

The application code is placed in the ``/usr/share/tarantool/<app-name>``
directory. An open source version of Tarantool is installed to the image.

The run directory is ``/var/run/tarantool/<app-name>``.
The working directory is ``/var/lib/tarantool/<app-name>``.

The runtime image also contains the file ``/usr/lib/tmpfiles.d/<app-name>.conf``
that allows the instance to reboot after container restart.

It is the user's responsibility to set up the proper ``advertise_uri`` parameter
(``<host>:<port>``) if the containers are deployed on different machines.
Each instance's ``advertise_uri`` must be the same on all machines,
because all other instances use it to connect to that instance.
Suppose that you start an instance with ``advertise_uri`` set to
``localhost:3302``. Addressing that instance as ``<instance-host>:3302`` from a different
instance won't work, because other instances will only recognize it as ``localhost:3302``.

If you specify only a port, ``cartridge`` will use an auto-detected IP.
In this case you have to configure Docker networks to set up inter-instance communication.

You can use Docker volumes to store instance snapshots and xlogs on the
host machine. If you updated your application code, you can create a new image for it,
stop the old container, and start a new one using the new image.

.. _cartridge-cli-special-files:

Build and packaging files
~~~~~~~~~~~~~~~~~~~~~~~~~

Put these files in your application directory to control the packaging process.
See the examples below.

* ``cartridge.pre-build`` is a script that runs before ``tarantoolctl rocks make``.
  The main purpose of this script is to build non-standard rocks modules
  (for example, from a submodule).
  Must be executable.

* ``cartridge.post-build`` is a script that runs after ``tarantoolctl rocks make``.
  The main purpose of this script is to remove build artifacts from the final package.
  Must be executable.

.. _cartridge-cli-example-cartridge-prebuild:

Example: cartridge.pre-build
^^^^^^^^^^^^^^^^^^^^^^^^^^^^

..  code-block:: bash

    #!/bin/sh

    # The main purpose of this script is to build non-standard rocks modules.
    # It will run before `tarantoolctl rocks make` during application build.

    tarantoolctl rocks make --chdir ./third_party/my-custom-rock-module

.. _cartridge-cli-example-cartridge-postbuild:

Example: cartridge.post-build
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

..  code-block:: bash

    #!/bin/sh

    # The main purpose of this script is to remove build artifacts from resulting package.
    # It will run after `tarantoolctl rocks make` during application build.

    rm -rf third_party
    rm -rf node_modules
    rm -rf doc
