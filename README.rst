Cartridge Command Line Interface
================================

.. image:: https://gitlab.com/tarantool/cartridge-cli/badges/master/pipeline.svg
   :alt: Cartridge-CLI build status on GitLab CI
   :target: https://gitlab.com/tarantool/cartridge-cli/commits/master

Installation
------------

If you have Tarantool Enterprise SDK, then just use ``cartridge`` CLI from it.

Otherwise, check the `installation guide <./doc/installation.rst>`_ to install
Cartridge CLI on your machine.

Quick start
-----------

Create your first application:

..  code-block:: bash

    cartridge create --name myapp

Go to the application directory:

..  code-block:: bash

    cd myapp

Build and start your application:

..  code-block:: bash

    cartridge build
    cartridge start -d

That's it! Now you can visit http://localhost:8081 and see your application's Admin Web UI:

.. image:: https://user-images.githubusercontent.com/11336358/75786427-52820c00-5d76-11ea-93a4-309623bda70f.png
   :align: center
   :scale: 100%

You're all set! To dive right in, follow the
`Getting started with Cartridge <https://www.tarantool.io/en/doc/latest/getting_started/getting_started_cartridge/>`__
guide.

Usage
-----

For more details, run:

.. code-block:: bash

   cartridge --help

The following commands are supported:

*   `create <./doc/commands/create.rst>`_: create a new application from template.
*   `build <./doc/commands/build.rst>`_: build an application for local development and testing.
*   `start <./doc/commands/start.rst>`_: start one or more Tarantool instances locally.
*   `stop <./doc/commands/stop.rst>`_: stop one or more Tarantool instances started locally.
*   `status <./doc/commands/status.rst>`_: get the status of one or more instances running locally.
*   `log <./doc/commands/log.rst>`_: get logs of one or more instances.
*   `clean <./doc/commands/clean.rst>`_: clean files for one or more instances.
*   `pack <./doc/commands/pack.rst>`_: pack the application into a distributable bundle.
*   `repair <./doc/commands/repair.rst>`_: patch cluster configuration files.
*   `admin <./doc/commands/admin.rst>`_: call an admin function provided by the application.
*   `replicasets<./doc/commands/replicasets.rst>`_: manage cluster replica sets running locally.
*   `enter <./doc/commands/connect.rst>`_: enter an instance running locally.
*   `connect <./doc/commands/connect.rst>`_: connect to a local instance using a specific address.
*   `failover <./doc/commands/failover.rst>`_: manage cluster failover.

All commands support `global flags <./doc/global_flags.rst>`_
that control output verbosity.
