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

5.  Check the installation:

    ..  code-block:: bash
        
        cartridge version


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

*   `Enabling shell completion for Cartridge CLI <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#command-line-completion>`__
*   `List of supported Cartridge CLI commands <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#usage>`__
*   `Cartridge application lifecycle <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#application-lifecycle>`__

*   `Creating a Cartridge application from a template <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#creating-an-application-from-a-template>`__
*   `Building the application locally <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#building-the-application>`__
*   `Starting the application locally <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#starting-the-application-locally>`__

    -   `Configuration files <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#configuration-files>`__
    -   `Environment variables <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#environment-variables>`__
    -   `Overriding default options <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#overriding-default-options>`__

*   `Stopping the application locally <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#stopping-the-application-locally>`__
*   `Checking instance status <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#checking-instance-status>`__
*   `Displaying logs <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#displaying-logs>`__
*   `Cleaning instance files <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#cleaning-instance-files>`__
*   `Repairing a cluster <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#repairing-a-cluster>`__
*   `Packaging your application <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#packaging-the-application>`__

    -   `Building a distribution <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#build-directory>`__
    -   `Packing a TGZ <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#repair-commands>`__
    -   `Packing an RPM or DEB distribution <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#rpm-and-deb>`__
    -   `Creating a Docker image of your app <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#docker>`__

*   `Configuring an installed package <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#usage-example>`__
*   `Files to control build and packaging <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_cli/#build-and-packaging-files>`__

