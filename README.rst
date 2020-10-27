Cartridge Command Line Interface
===============================================================================

.. image:: https://gitlab.com/tarantool/cartridge-cli/badges/master/pipeline.svg
   :alt: Cartridge-CLI build status on GitLab CI
   :target: https://gitlab.com/tarantool/cartridge-cli/commits/master

-------------------------------------------------------------------------------
Installation
-------------------------------------------------------------------------------

If you use Tarantool Enterprise SDK, then just use ``cartridge`` CLI from it.

Otherwise, check `installation guide <./doc/installation.rst>`_ to install
Cartridge CLI on your machine.

Now you can create and start your first application!


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

You can find more details in this README document or you can start with the
`getting started guide <https://www.tarantool.io/en/doc/latest/getting_started/getting_started_cartridge/>`_.

-------------------------------------------------------------------------------
Usage
-------------------------------------------------------------------------------

For more details, say:

.. code-block:: bash

   cartridge --help

The following commands are supported:

* `create <./doc/commands/create.rst>`_  — create a new application from template;
* `build <./doc/commands/build.rst>`_  — build the application for local development and testing;
* `start <./doc/commands/start.rst>`_ — start a Tarantool instance(s) locally;
* `stop <./doc/commands/stop.rst>`_ — stop a Tarantool instance(s) started locally;
* `status <./doc/commands/status.rst>`_ — get current locally running instance(s) status;
* `log <./doc/commands/log.rst>`_ — get logs of instance(s);
* `clean <./doc/commands/clean.rst>`_ - clean instance(s) files;
* `pack <./doc/commands/pack.rst>`_ — pack the application into a distributable bundle;
* `repair <./doc/commands/repair.rst>`_ — patch cluster configuration files;
* `admin <./doc/commands/admin.rst>`_ - call an admin function provided by the application.

Each command supports `global flags <./doc/global_flags.rst>`_.
