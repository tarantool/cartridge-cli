Cartridge Command Line Interface
================================

..  image:: https://img.shields.io/github/v/release/tarantool/cartridge-cli?include_prereleases&label=Release&labelColor=2d3532
    :alt: Cartridge CLI latest release on GitHub
    :target: https://github.com/tarantool/cartridge-cli/releases

..  image:: https://github.com/tarantool/cartridge-cli/workflows/Tests/badge.svg
    :alt: Cartridge CLI build status on GitHub Actions
    :target: https://github.com/tarantool/cartridge-cli/actions/workflows/tests.yml


Control your Tarantool application instances via the command line.

Installation
------------

1.  Install the following third-party tools:

    *   `git <https://git-scm.com/book/en/v2/Getting-Started-Installing-Git>`__
    *   `unzip <https://linuxize.com/post/how-to-unzip-files-in-linux/>`__
    *   `gcc <https://gcc.gnu.org/install/>`__
    *   `cmake <https://cmake.org/install/>`__
        and `make <https://cmake.org/install/>`__

2.  Install Tarantool 1.10 or higher. You have two options here:

    *   `Install from a package <https://www.tarantool.io/en/download/>`__
    *   `Build from source <https://www.tarantool.io/en/doc/latest/dev_guide/building_from_source/>`__

3.  [For all platforms except macOS] If you build Tarantool from source,
    you need to set up the Tarantool packages repository manually:

    ..  code-block:: bash

        curl -L https://tarantool.io/installer.sh | sudo -E bash -s -- --repo-only

4.  Install the ``cartridge-cli`` package:

    *   For CentOS, Fedora, or ALT Linux (RPM package):

        ..  code-block:: bash

            sudo yum install cartridge-cli

    *   For Debian or Ubuntu (DEB package):

        ..  code-block:: bash

            sudo apt-get install cartridge-cli

    *   For macOS (Homebrew formula):

        ..  code-block:: bash

            brew install cartridge-cli

    *   Or build locally:

        .. code-block:: bash

           mage build

5.  Check the installation:

    ..  code-block:: bash
        
        cartridge version

    You may see a warning: ``Project path is not a project``.
    Don't worry, it only means there is no Cartridge application yet.

Now you can
`create and run <https://www.tarantool.io/en/doc/latest/getting_started/getting_started_cartridge/>`__
your first application!


Quick start
-----------

To create your first application, run:

..  code-block:: bash

    cartridge create --name myapp

Go to the application directory:

..  code-block:: bash

    cd myapp

Build and start your application:

..  code-block:: bash

    cartridge build
    cartridge start

Now open http://localhost:8081 and see your application's Admin Web UI:

..  image:: https://user-images.githubusercontent.com/11336358/75786427-52820c00-5d76-11ea-93a4-309623bda70f.png
    :align: center

You're all set! To dive right in, follow the
`Getting started with Cartridge <https://www.tarantool.io/en/doc/latest/getting_started/getting_started_cartridge/>`__
guide.

Usage
-----

For details about how to use Cartridge CLI, see the documentation links below.

*   `Enabling shell completion for Cartridge CLI <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/installation/#enable-shell-completion>`__
*   `Supported Cartridge CLI commands <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/>`__
*   `Cartridge application lifecycle <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/lifecycle/>`__

*   `Creating a Cartridge application from template <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/create/>`__
*   `Building the application locally <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/build/>`__
*   `Starting the application locally <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/start/>`__
*   `Stopping the application locally <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/stop/>`__
*   `Checking instance status <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/status/>`__
*   `Entering a locally running instance <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/enter/>`__
*   `Connecting to a locally running instance at a specific address <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/connect/>`__
*   `Displaying logs <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/log/>`__
*   `Cleaning instance files <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/clean/>`__
*   `Repairing the cluster <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/repair/>`__
*   `Setting up replica sets <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/replicasets/>`__
*   `Configuring failover <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/failover/>`__
*   `Running admin functions <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/admin/>`__
*   `Packaging your application <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/pack/>`__

    -   `Building a distribution <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/pack/#building-the-package>`__
    -   `Packing a TGZ <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/pack/tgz/>`__
    -   `Packing an RPM or DEB distribution <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/pack/rpm-deb/>`__
    -   `Creating a Docker image of your app <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/pack/docker/>`__
    -   `Building in Docker <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/commands/pack/building-in-docker/>`__

*   `Global flags <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/global-flags/>`__
*   `Application instance paths <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/instance-paths/>`__
*   `Pre-build and post-build scripts <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/pre-post-build/>`__
