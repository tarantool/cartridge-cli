.. _cartridge-cli:

===============================================================================
Cartridge Command Line Interface
===============================================================================

.. image:: https://github.com/tarantool/cartridge-cli/workflows/Tests/badge.svg
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

   .. code-block:: bash

       curl -L https://tarantool.io/installer.sh | sudo -E bash -s -- --repo-only

4. Install the ``cartridge-cli`` package:

   * for CentOS, Fedora, ALT Linux (RPM package):

     .. code-block:: bash

         sudo yum install cartridge-cli

   * for Debian, Ubuntu (DEB package):

     .. code-block:: bash

         sudo apt-get install cartridge-cli

   * for MacOS X (Homebrew formula):

     .. code-block:: bash

         brew install cartridge-cli

5. Check the installation:

   .. code-block:: bash

      cartridge version

Now you can
`create and start <https://www.tarantool.io/en/doc/latest/getting_started/getting_started_cartridge/>`_
your first application!

-------------------------------------------------------------------------------
Quick start
-------------------------------------------------------------------------------

To create your first application:

.. code-block:: bash

    cartridge create --name myapp

Let's go inside:

.. code-block:: bash

    cd myapp

Now build the application and start it:

.. code-block:: bash

    cartridge build
    cartridge start

That's it! Now you can visit http://localhost:8081 and see your application's Admin Web UI:

.. image:: https://user-images.githubusercontent.com/11336358/75786427-52820c00-5d76-11ea-93a4-309623bda70f.png
   :align: center
   :scale: 100%

You can find more details in this README document or you can start with the
`getting started guide <https://www.tarantool.io/en/doc/latest/getting_started/getting_started_cartridge/>`_.

.. _cartridge-cli-usage:

-------------------------------------------------------------------------------
Command-line completion
-------------------------------------------------------------------------------

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Linux
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

RPM and DEB ``cartridge-cli`` packages contain ``/etc/bash_completion.d/cartridge``
Bash completion script.
To enable completion after ``cartridge-cli`` installation start a new shell or
source ``/etc/bash_completion.d/cartridge`` completion file.
Make sure that you have bash completion installed.

To install Zsh completion, say

.. code-block:: bash

    cartridge gen completion --skip-bash --zsh="${fpath[1]}/_cartridge"

To enable shell completion:

.. code-block:: bash

    echo "autoload -U compinit; compinit" >> ~/.zshrc

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
OS X
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

If you install ``cartridge-cli`` from ``brew``, it automatically installs both
Bash and Zsh completions.

-------------------------------------------------------------------------------
Usage
-------------------------------------------------------------------------------

For more details, say:

.. code-block:: bash

   cartridge --help

The following commands are supported:

* ``create`` — create a new application from template;
* ``build`` — build the application for local development and testing;
* ``start`` — start a Tarantool instance(s);
* ``stop`` — stop a Tarantool instance(s);
* ``status`` — get current instance(s) status;
* ``log`` — get logs of instance(s);
* ``clean`` - clean instance(s) files;
* ``pack`` — pack the application into a distributable bundle;
* ``repair`` — patch cluster configuration files;
* `admin <doc/admin.rst>`_ - call an admin function provided by the application;
* `replicasets <doc/replicasets.rst>`_ - manage cluster replica sets running locally;
* `enter and connect <doc/connect.rst>`_ - connect to running instance.

The following global flags are supported:

* ``verbose`` — verbose mode, additional log messages are shown as well as
  commands/docker output (such as `tarantoolctl rocks make` or `docker build` output);
* ``debug`` — debug mode (the same as verbose, but temporary files and
  directories aren't removed);
* ``quiet`` — the mode that hides all logs; only errors are shown.

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
An application lifecycle
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

In a nutshell:

1. `Create <Creating an application from template_>`_
   an application (e.g. ``myapp``) from template:

   .. code-block:: bash

       cartridge create --name myapp
       cd ./myapp

2. `Build <Building an application_>`_ the application
   for local development and testing:

   .. code-block:: bash

       cartridge build

3. `Run <Starting/stopping an application locally_>`_
   instances locally:

   .. code-block:: bash

       cartridge start
       cartridge stop

4. `Pack <Packing an application_>`_ the application into
   a distributable (e.g. into an RPM package):

   .. code-block:: bash

       cartridge pack rpm

.. _cartridge_cli_creating_an_application_from_template:

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Creating an application from template
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

To create an application from the Cartridge template, say this in any directory:

.. code-block:: bash

    cartridge create [PATH] [flags]

The following options (``[flags]``) are supported:

.. // Please, update the doc in cli/commands on updating this section

* ``--name strin`` is an application name.

* ``--from DIR`` is a path to the application template (see details below).

* ``--template string`` is a name of application template to be used.
  Currently only ``cartridge`` template is supported.

Application is created in the ``<path>/<app-name>/`` directory.

By default, ``cartridge`` template is used.
It contains a simple Cartridge application with:

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

You can create your own application template and use it with ``cartridge create``
with ``--from`` flag.

If template directory is a git repository, the `.git/` files would be ignored on
instantiating template.
In the created application a new git repo is initialized.

Template application shouldn't contain `.rocks` directory.
To specify application dependencies use rockspec and `cartridge.pre-build` files.

Filenames and content can contain `text templates <Templates_>`_.

.. _Templates: https://golang.org/pkg/text/template/

Available variables are:

* ``Name`` — the application name;
* ``StateboardName`` — the application stateboard name (``<app-name>-stateboard``);
* ``Path`` - an absolute path to the application.

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

    print("Hi, I am {{ .Name }} application")
    print("I also have a stateboard named {{ .StateboardName }}")

.. _cartridge-cli-building-an-application:

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Building an application
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

To build your application locally (for local testing), say this in any directory:

.. code-block:: bash

    cartridge build [PATH] [flags]

.. // Please, update the doc in cli/commands on updating this section

This command requires one argument — the path to your application directory
(i.e. to the build source). The default path is ``.`` (the current directory).

This command runs:

1. ``cartridge.pre-build`` if the
   `pre-build file <Special files_>`_ exists.
   This builds the application in the ``[PATH]`` directory.
2. ``tarantoolctl rocks make`` if the
   `rockspec file <Special files_>`_ exists.
   This installs all Lua rocks to the ``[PATH]`` directory.

During step 1 of the ``cartridge build`` command, ``cartridge`` builds the application
inside the application directory -- unlike when building the application as part
of the ``cartridge pack`` command, when the application is built in a temporary
`build directory <Build directory_>`_ and no build artifacts
remain in the application directory.

During step 2 -- the key step here -- ``cartridge`` installs all dependencies
specified in the rockspec file (you can find this file within the application
directory created from template).

(An advanced alternative would be to specify build logic in the
rockspec as ``cmake`` commands, like we
`do it <https://github.com/tarantool/cartridge/blob/master/cartridge-scm-1.rockspec#L26>`_
for ``cartridge``.)

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

.. _cartridge-cli-starting-stopping-an-application-locally:

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Starting/stopping an application locally
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

**********
``start``
**********

Now, after the application is `built <Building an application_>`_,
you can run it locally:

.. code-block:: bash

    cartridge start [INSTANCE_NAME...] [flags]

where ``[INSTANCE_NAME...]`` means that several instances can be specified.

If no ``INSTANCE_NAME`` is provided, all the instances from the
Cartridge instances configuration file are taken as arguments (see the ``--cfg``
option below).

We also need an application name (``APP_NAME``) to pass it to the instances while
started and to define paths to the instance files (for example, ``<run-dir>/<APP_NAME>.<INSTANCE_NAME>.pid``).
By default, the ``APP_NAME`` is taken from the application rockspec in the current
directory, but also it can be defined explicitly via the ``--name`` option
(see description below).

^^^^^^^^
Options
^^^^^^^^

The following options (``[flags]``) are supported:

.. // Please, update the doc in cli/commands on updating this section

* ``--script FILE`` is the application's entry point.
  It should be a relative path to the entry point in the project directory
  or an absolute path.
  Defaults to ``init.lua`` (or to the value of the "script"
  parameter in the Cartridge `configuration file <Overriding default options_>`_).

* ``--run-dir DIR`` is the directory where PID and socket files are stored.
  Defaults to ``./tmp/run`` (or to the value of the "run-dir"
  parameter in the Cartridge `configuration file <Overriding default options_>`_).

* ``--data-dir DIR`` is the directory where instances' data is stored.
  Each instance's working directory is ``<data-dir>/<app-name>.<instance-name>``.
  Defaults to ``./tmp/data`` (or to the value of the "data-dir"
  parameter in the Cartridge `configuration file <Overriding default options_>`_).

* ``--log-dir DIR`` is the directory to store instances logs
  when running in background.
  Defaults to ``./tmp/log`` (or to the value of the "log-dir"
  parameter in the Cartridge `configuration file <Overriding default options_>`_).

* ``--cfg FILE`` is the configuration file for Cartridge instances.
  Defaults to ``./instances.yml`` (or to the value of the "cfg"
  parameter in the Cartridge `configuration file <Overriding default options_>`_).

  The ``instances.yml`` file contains parameters for starting Cartridge
  application instances and is placed in the application root directory.
  These parameters are parsed on the `cartridge.cfg() <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_api/modules/cartridge/#cfg-opts-box-opts>`_
  call.

  Example of the ``instances.yml`` file:

  .. code-block:: yaml

      myapp.router:
        advertise_uri: localhost:3301
        http_port: 8081

      myapp.s1-master:
        advertise_uri: localhost:3302
        http_port: 8082

  Parameters that can be specified in ``instances.yml`` are listed
  `here <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_api/modules/cartridge/#cfg-opts-box-opts>`_.
  The ``advertise_uri`` parameter is required.

  .. note::

     The following parameters, if specified in ``instances.yml``, will be
     overwritten by Cartridge CLI environment variables on
     ``cartridge start``:

     * ``workdir``
     * ``console_sock``
     * ``pid_file``.

  You can also specify custom parameters in ``instances.yml``, but they
  should be defined in the application code. The example below shows the usage
  of the ``my_param`` custom parameter:

  ``instances.yml``:

  .. code-block:: yaml

      myapp.router:
        advertise_uri: localhost:3301
        http_port: 8081
        my_param: 'Hello, world'

  ``init.lua``:

  .. code-block:: lua

     local argparse = require('cartridge.argparse')
     local my_param = argparse.get_opts({my_param='string'})

* ``--daemonize, -d`` starts the instance in background.
  With this option, Tarantool also waits until the application's main script is
  finished.
  For example, it is useful if the ``init.lua`` requires time-consuming startup
  from snapshot, and Tarantool waits for the startup to complete.
  This is also useful if the application's main script generates errors, and
  Tarantool can handle them.

* ``--stateboard`` starts the application stateboard as well as instances.
  Ignored if ``--stateboard-only`` is specified. Value can be specified in "cfg"
  parameter in the Cartridge `configuration file <Overriding default options_>`_).

* ``--stateboard-only`` starts only the application stateboard.
  If specified, ``INSTANCE_NAME...`` are ignored.

* ``--name string`` defines the application name.
  By default, it is taken from the application rockspec.

* ``--timeout string`` Time to wait for instance(s) start in background.
  Can be specified in seconds or in the duration form (``72h3m0.5s``).
  Timeout can't be negative.
  Timeout ``0`` means no timeout (wait for instance(s) start forever).
  The default timeout is 60 seconds (``1m0s``).

^^^^^^^^^^^^^^^^^^^^^^
Environment variables
^^^^^^^^^^^^^^^^^^^^^^

The ``cartridge start`` command starts a Tarantool instance with enforced
**environment variables**:

.. code-block:: bash

    TARANTOOL_APP_NAME="<instance-name>"
    TARANTOOL_INSTANCE_NAME="<app-name>"
    TARANTOOL_CFG="<cfg>"
    TARANTOOL_PID_FILE="<run-dir>/<app-name>.<instance-name>.pid"
    TARANTOOL_CONSOLE_SOCK="<run-dir>/<app-name>.<instance-name>.control"
    TARANTOOL_WORKDIR="<data-dir>/<app-name>.<instance-name>.control"

When started in background, a notify socket path is passed additionally:

.. code-block:: bash

    NOTIFY_SOCKET="<data-dir>/<app-name>.<instance-name>.notify"

``cartridge.cfg()`` uses  ``TARANTOOL_APP_NAME`` and ``TARANTOOL_INSTANCE_NAME``
to read the instance's configuration from the file provided in ``TARANTOOL_CFG``.

^^^^^^^^^^^^^^^^^^^^^^^^^^^
Overriding default options
^^^^^^^^^^^^^^^^^^^^^^^^^^^

You can override default options for the ``cartridge`` command in the
``./.cartridge.yml`` configuration file.

Here is an example of ``.cartridge.yml``:

.. code-block:: yaml

    run-dir: my-run-dir
    cfg: my-instances.yml
    script: my-init.lua
    stateboard: true

**Note:** the config of the `standard application template <Creating an application from template_>`_ initially has the ``stateboard: true`` parameter.

.. // Please, update the doc in cli/commands on updating this section

*********
``stop``
*********

To stop one or more running instances, say:

.. code-block:: bash

    cartridge stop [INSTANCE_NAME...] [flags]

By default, SIGTERM is sent to instances.

The following options (``[flags]``) are supported:

* ``-f, --force`` indicates if instance(s) stop should be forced (sends SIGKILL).

The following `options <Options_>`_ from the ``start`` command
are supported:

* ``--run-dir DIR``
* ``--cfg FILE``
* ``--stateboard``
* ``--stateboard-only``

.. note::

   ``run-dir`` should be exactly the same as used in the ``cartridge start``
   command. PID files stored there are used to stop the running instances.

.. // Please, update the doc in cli/commands on updating this section

***********
``status``
***********

To check the current instance status, use the ``status`` command:

.. code-block:: bash

    cartridge status [INSTANCE_NAME...] [flags]

The following `options <Options_>`_ from the ``start`` command
are supported:

* ``--run-dir DIR``
* ``--cfg FILE``
* ``--stateboard``
* ``--stateboard-only``

.. // Please, update the doc in cli/commands on updating this section

*******
``log``
*******

To get logs of the instance running in background, use the ``log`` command:

.. code-block:: bash

    cartridge log [INSTANCE_NAME...] [flags]

The following options (``[flags]``) are supported:

* ``-f, --follow`` outputs appended data as the log grows.

* ``-n, --lines int`` is the number of lines to output (from the end).
  Defaults to 15.

The following `options <Options_>`_ from the ``start`` command
are supported:

* ``--log-dir DIR``
* ``--run-dir DIR``
* ``--cfg FILE``
* ``--stateboard``
* ``--stateboard-only``

.. // Please, update the doc in cli/commands on updating this section

.. _cartridge-cli-packing-an-application:

*********
``clean``
*********

To remove instance(s) files (log, workdir, console socket, PID-file and notify socket),
use the ``clean`` command:

.. code-block:: bash

    cartridge clean [INSTANCE_NAME...] [flags]

`cartridge clean` for running instance(s) causes an error.

The following `options <Options_>`_ from the ``start`` command
are supported:

* ``--log-dir DIR``
* ``--data-dir DIR``
* ``--run-dir DIR``
* ``--cfg FILE``
* ``--stateboard``
* ``--stateboard-only``

.. // Please, update the doc in cli/commands on updating this section

.. _cartridge-cli-packing-an-application:

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Packing an application
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

To pack your application, say this in any directory:

.. code-block:: bash

     cartridge pack TYPE [PATH] [flags]

where:

* ``TYPE`` (required) is the distribution type. Supported types:

  * `TGZ <TGZ_>`_
  * `RPM <RPM and DEB_>`_
  * `DEB <RPM and DEB_>`_
  * `Docker <Docker_>`_

* ``PATH`` (optional) is the path to the application directory to pack.
  Defaults to ``.`` (the current directory).

.. note::

  If you pack application into RPM or DEB on MacOS without `--use-docker`
  flag, the result artifact is broken - it contains rocks and executables
  that can't be used on Linux. In this case packing fails.

The options (``[flags]``) are as follows:

.. // Please, update cmd_pack usage in cartridge-cli.lua file on updating the doc

* ``--name string`` (common for all distribution types) is the application name.
  It coincides with the package name and the systemd-service name.
  The default name comes from the ``package`` field in the rockspec file.

* ``--version string`` (common for all distribution types) is the application's package
  version. The expected pattern is ``major.minor.patch[-count][-commit]``:
  if you specify ``major.minor.patch``, it is normalized to ``major.minor.patch-count``.
  The default version is determined as the result of ``git describe --tags --long``.
  If the application is not a git repository, you need to set the ``--version`` option
  explicitly.

* ``--suffix string`` (common for all distribution types) is the result file (or image)
  name suffix.

* ``--unit-template string`` (used for ``rpm`` and ``deb``) is the path to the template for
  the ``systemd`` unit file.

* ``--instantiated-unit-template string`` (used for ``rpm`` and ``deb``) is the path to the
  template for the ``systemd`` instantiated unit file.

* ``--stateboard-unit-template string`` (used for ``rpm`` and ``deb``) is the path to the
  template for the stateboard ``systemd`` unit file.

* ``--use-docker`` (enforced for ``docker``) forces to build the application in Docker.

* ``--tag strings`` (used for ``docker``) is the tag(s) of the Docker image that results from
  ``pack docker``.

* ``--from string`` (used for ``docker``) is the path to the base Dockerfile of the runtime
  image. Defaults to ``Dockerfile.cartridge`` in the application root.

* ``--build-from string`` (common for all distribution types, used for building in Docker) is
  the path to the base Dockerfile of the build image.
  Defaults to ``Dockerfile.build.cartridge`` in the application root.

* ``--no-cache`` creates build and runtime images with ``--no-cache`` docker flag.

* ``--cache-from strings`` images to consider as cache sources for both build and
  runtime images. See ``--cache-from`` flag for ``docker build`` command.

* ``--sdk-path string`` (common for all distribution types, used for building in Docker) is the
  path to the SDK to be delivered in the result artifact.
  Alternatively, you can pass the path via the ``TARANTOOL_SDK_PATH``
  environment variable (this variable is of lower priority).

* ``--sdk-local`` (common for all distribution types, used for building in Docker) is a
  flag that indicates if the SDK from the local machine should be delivered in the
  result artifact.

For Tarantool Enterprise, you must specify one (and only one)
of the ``--sdk-local`` and ``--sdk-path`` options.

For ``rpm``, ``deb``, and ``tgz``, we also deliver rocks modules and executables
specific for the system where the ``cartridge pack`` command is running.

For ``docker``, the resulting runtime image will contain rocks modules
and executables specific for the base image (``centos:8``).

Next, we dive deeper into the packaging process.

.. _cartridge-cli-build-directory:

****************
Build directory
****************

The first step of the packaging process is to
`build the application <Building an application_>`_.

By default, application build is done in a temporary directory in
``~/.cartridge/tmp/``, so the packaging process doesn't affect the contents
of your application directory.

You can specify a custom build directory for your application in the ``CARTRIDGE_TEMPDIR``
environment variable. If this directory doesn't exists, it will be created, used
for building the application, and then removed.

If you specify an existing directory in the ``CARTRIDGE_TEMPDIR`` environment
variable, the ``CARTRIDGE_TEMPDIR/cartridge.tmp`` directory will be used for
build and then removed. This directory will be cleaned up before building the
application.

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

See an `example <Example: cartridge.post-build_>`_
in `special files <Special files_>`_.

.. cartridge-cli-repair:

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Repairing a cluster
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

To repair a running application, you can use the ``cartridge repair`` command.

There are several simple rules you need to know before using this command:

* Rule #1 of ``repair`` is: you do not use it if you aren't sure that
  it's exactly what you need.
* Rule #2: always use ``--dry-run`` before running ``repair``.
* Rule #3: do not hesitate to use the ``--verbose`` option.
* Rule #4: do not use the ``--force`` option if you aren't sure that it's exactly
  what you need.

Please, pay attention to the
`troubleshooting documentation <https://www.tarantool.io/en/doc/2.3/book/cartridge/troubleshooting/>`_
before using ``repair``.

What does ``repair`` actually do?

It patches the cluster-wide configuration files of application instances
placed on the local machine.
Note that it's not enough to *apply* new configuration:
the configuration should be *reloaded* by the instance.

``repair`` was created to be used on production (but it still can be used for
local development). So, it requires the application name option ``--name``.
Moreover, remember that the default data directory is ``/var/lib/tarantool`` and
the default run directory is ``/var/run/tarantool``
(both of them can be rewritten by options).

In default mode, ``repair`` walks across all cluster-wide configurations placed
in ``<data-dir>/<app-name>.*`` directories and patches all found configuration
files.

If the ``--dry-run`` flag is specified, files aren't patched, and only a computed
configuration diff is shown.

If configuration files are diverged between instances on the local machine,
``repair`` raises an error.
But you can specify the ``--force`` option to patch different versions of
configuration independently.

``repair`` can also reload configuration for all instances if the ``--reload``
flag is specified (only if the application uses ``cartridge >= 2.0.0``).
Configuration will be reloaded for all instances that are placed in the new
configuration using console sockets that are placed in the run directory.
Make sure that you specified the right run directory when using ``--reload`` flag.

.. code-block:: bash

    cartridge repair [command]

The following ``repair`` commands are available
(see `details <Repair commands_>`_ below):

* ``list-topology`` - shows the current topology summary;
* ``remove-instance`` - removes an instance from the cluster;
* ``set-leader`` - changes a replica set leader;
* ``set-uri`` - changes an instance's advertise URI.

All repair commands have these flags:

* ``--name`` (required) is an application name.

* ``--data-dir`` is a directory where the instances' data is stored
  (defaults to ``/var/lib/tarantool``).

All commands, except ``list-topology``, have these flags:

* ``--run-dir`` is a directory where PID and socket files are stored
  (defaults to ``/var/run/tarantool``).

* ``--dry-run`` runs the ``repair`` command in the dry-run mode
  (shows changes but doesn't apply them).

* ``--reload`` is a flag that enables reloading configuration on instances
  after the patch.

.. cartridge-cli-repair-commands:

***************
Repair commands
***************

^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Topology summary
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: bash

    cartridge repair list-topology [flags]

Takes no arguments. Prints the current topology summary.

^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Remove instance
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: bash

    cartridge repair remove-instance UUID [flags]

Removes an instance with the specified UUID from cluster.
If the specified instance isn't found, raises an error.

^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Set leader
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: bash

    cartridge repair set-leader REPLICASET-UUID INSTANCE-UUID [flags]

Sets the specified instance as the leader of the specified replica set.
Raises an error if:

* a replica set or instance with the specified UUID doesn't exist;
* the specified instance doesn't belong to the specified replica set;
* the specified instance is disabled or expelled.

^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Set advertise URI
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: bash

    cartridge repair set-uri INSTANCE-UUID URI-TO [flags]

Rewrites the advertise URI for the specified instance.
If the specified instance isn't found or is expelled, raises an error.

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

.. code-block:: bash

    systemctl start myapp@instance-1
    systemctl start myapp@instance-2

If you use stateful failover, you need to start application stateboard.

(Remember that your application should contain ``stateboard.init.lua`` in its
root.)

Add the ``myapp-stateboard`` section to ``/etc/tarantool/conf.d/myapp.yml``:

.. code-block:: yaml

    myapp-stateboard:
      listen: localhost:3310
      password: passwd

Then, start the stateboard service:

.. code-block:: bash

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

* for both RPM and DEB:

  .. code-block:: bash

      curl -L https://tarantool.io/installer.sh | VER=${TARANTOOL_VERSION} bash

The package contents is as follows:

* the contents of the distribution directory, placed in the
  ``/usr/share/tarantool/<app-name>`` directory
  (for Tarantool Enterprise, this directory also contains ``tarantool`` and
  ``tarantoolctl`` binaries);

* unit files for running the application as a ``systemd`` service:
  ``/etc/systemd/system/<app-name>.service`` and
  ``/etc/systemd/system/<app-name>@.service``;

* application stateboard unit file:
  ``/etc/systemd/system/<app-name>-stateboard.service``
  (will be packed only if the application contains ``stateboard.init.lua`` in its root);

* the file ``/usr/lib/tmpfiles.d/<app-name>.conf`` that allows the instance to restart
  after server restart.

The following directories are created:

* ``/etc/tarantool/conf.d/`` — directory for instances configuration;
* ``/var/lib/tarantool/`` — directory to store instances snapshots;
* ``/var/run/tarantool/`` — directory to store PID-files and console sockets.

See the `documentation <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_dev/#deploying-an-application>`_
for details about deploying a Tarantool Cartridge application.

To start the ``instance-1`` instance of the ``myapp`` service, say:

.. code-block:: bash

    systemctl start myapp@instance-1

To start the application stateboard service, say:

.. code-block:: bash

    systemctl start myapp-stateboard

This instance will look for its
`configuration <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_dev/#configuring-instances>`_
across all sections of the YAML file(s) stored in ``/etc/tarantool/conf.d/*``.

Use the options ``--unit-template``, ``--instantiated-unit-template`` and
``--stateboard-unit-template`` to customize standard unit files.

You may need it first of all for DEB packages, if your build platform
is different from the deployment platform. In this case, ``ExecStartPre`` may
contain an incorrect path to `mkdir`. As a hotfix, we suggest customizing the
unit files.

Example of an instantiated unit file:

.. code-block:: kconfig

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

* ``Name`` — the application name;
* ``StateboardName`` — the application stateboard name (``<app-name>-stateboard``);

* ``DefaultWorkDir`` — default instance working directory (``/var/lib/tarantool/<app-name>.default``);
* ``InstanceWorkDir`` — application instance working directory (``/var/lib/tarantool/<app-name>.<instance-name>``);
* ``StateboardWorkDir`` — stateboard working directory (``/var/lib/tarantool/<app-name>-stateboard``);

* ``DefaultPidFile`` — default instance pid file (``/var/run/tarantool/<app-name>.default.pid``);
* ``InstancePidFile`` — application instance pid file (``/var/run/tarantool/<app-name>.<instance-name>.pid``);
* ``StateboardPidFile`` — stateboard pid file (``/var/run/tarantool/<app-name>-stateboard.pid``);

* ``DefaultConsoleSock`` — default instance console socket (``/var/run/tarantool/<app-name>.default.control``);
* ``InstanceConsoleSock`` — application instance console socket (``/var/run/tarantool/<app-name>.<instance-name>.control``);
* ``StateboardConsoleSock`` — stateboard console socket (``/var/run/tarantool/<app-name>-stateboard.control``);

* ``ConfPath`` — path to the application instances config (``/etc/tarantool/conf.d``);

* ``AppEntrypointPath`` — path to the application entrypoint (``/usr/share/tarantool/<app-name>/init.lua``);
* ``StateboardEntrypointPath`` — path to the stateboard entrypoint (``/usr/share/tarantool/<app-name>/stateboard.init.lua``);

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

* `Stage 1. Cleaning up the application directory <Stage 1. Cleaning up the application directory_>`_
* `Stage 2. Building the application <Stage 2. Building the application_>`_
  (the build is always done `in Docker <Building in Docker_>`_)
* `Stage 3. Cleaning up the files before packaging <Stage 3. Cleaning up the files before packing_>`_

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
or ``FROM centos:7`` line (except comments).

We expect the base docker image to be ``centos:8`` or ``centos:7``, but you can use any other.

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

You can pass ``--cache-from`` and ``--no-cache`` options of ``docker build``
command on building application in docker.

************************
Using the runtime image
************************

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

.. _cartridge-cli-special-files:

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Special files
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

You can put these files in your application root to control the application
packaging process (see examples below):

* ``cartridge.pre-build``: a script to be run before ``tarantoolctl rocks make``.
  The main purpose of this script is to build some non-standard rocks modules
  (for example, from a submodule).
  Should be executable.

* ``cartridge.post-build``: a script to be run after ``tarantoolctl rocks make``.
  The main purpose of this script is to remove build artifacts from result package.
  Should be executable.

.. _cartridge-cli-example-cartridge-prebuild

*****************************
Example: cartridge.pre-build
*****************************

.. code-block:: bash

    #!/bin/sh

    # The main purpose of this script is to build some non-standard rocks modules.
    # It will be run before `tarantoolctl rocks make` on application build

    tarantoolctl rocks make --chdir ./third_party/my-custom-rock-module

.. _cartridge-cli-example-cartridge-postbuild

******************************
Example: cartridge.post-build
******************************

.. code-block:: bash

    #!/bin/sh

    # The main purpose of this script is to remove build artifacts from resulting package.
    # It will be ran after `tarantoolctl rocks make` on application build.

    rm -rf third_party
    rm -rf node_modules
    rm -rf doc
