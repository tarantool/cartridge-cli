===============================================================================
Creating an application from template
===============================================================================

To create an application from the Cartridge template, say this in any directory:

.. code-block:: bash

    cartridge create [PATH] [flags]

-------------------------------------------------------------------------------
Options
-------------------------------------------------------------------------------

* ``--name`` is an application name.

* ``--from`` is a path to the application template (see details below).

* ``--template`` is a name of application template to be used.
  Currently only ``cartridge`` template is supported.

Command also supports `global flags <./global_flags.rst>`_.

-------------------------------------------------------------------------------
Details
-------------------------------------------------------------------------------

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

* `special files <./special_files.rst>`_ (used to build and pack
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
