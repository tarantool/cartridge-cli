Creating an application from template
=====================================

To create an application from a Cartridge template, run this in any directory:

..  code-block:: bash

    cartridge create [path] [flags]

Flags
-----

..  container:: table

    ..  list-table::
        :widths: 20 80
        :header-rows: 0

        *   -   ``--name``
            -   Application name.
        *   -   ``--from``
            -   Path to the application template. See details below.
        *   -   ``--template``
            -   Name of the application template.
                Currently, only the ``cartridge`` template is supported.

The command also supports `global flags <./global_flags.rst>`_.

Details
-------

Your application will appear in the ``<path>/<app-name>/`` directory.

The template used by default is ``cartridge``.
It produces a simple Cartridge application that includes:

* One custom role with an HTTP endpoint.
* Sample tests and basic test helpers.
* Development files like ``.luacheckrc``.

If you have ``git`` installed, a Git repository with
a ``.gitignore`` file will be also set up in the project root directory.
The initial commit will be created and tagged with
`version <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_dev/#application-versioning>`_.

Project directory
^^^^^^^^^^^^^^^^^

Let's take a closer look at the files inside the ``<app_name>/`` directory:

*   Application files:

    -   ``app/roles/custom-role.lua``: a sample
        `custom role <https://www.tarantool.io/en/doc/latest/book/cartridge/cartridge_dev/#cluster-roles>`_
        with a simple HTTP API. Can be enabled as ``app.roles.custom``.
    -   ``<app_name>-scm-1.rockspec``: contains application dependencies.
    -   ``init.lua``: application entry point.
    -   ``stateboard.init.lua`` application `stateboard <./failover.rst>`_ entry point.

*   Build and packaging files:

    -   ``cartridge.pre-build``
    -   ``cartridge.post-build``
    -   ``Dockerfile.build.cartridge``
    -   ``Dockerfile.cartridge``

  To learn more, check the documentation
  on `pre-build and post-build scripts <../pre_post_build.rst>`_.

*   Development files:

    -   ``deps.sh`` resolves dependencies listed in the ``.rockspec`` file
        and installs test dependencies (like ``luatest``).
    -   ``instances.yml`` contains the configuration of instances and is used by ``cartridge start``.
    -   ``.cartridge.yml`` contains the Cartridge configuration and is also used by ``cartridge start``.
    -   ``systemd-unit-params.yml`` contains systemd parameters.
    -   ``tmp`` is a directory for temporary files, used as a run directory (see ``.cartridge.yml``).
    -   ``.git`` is the directory responsible for the Git repository.
    -   ``.gitignore`` is a file where you can specify the files for Git to ignore.
    -   ``env.lua`` is a file that sets common rock paths,
        which allows you to start the application from any directory.

*   Test files (with sample tests):

  ..  code-block:: text

      test
      ├── helper
      │   ├── integration.lua
      │   └── unit.lua
      │   ├── helper.lua
      │   ├── integration
      │   │   └── api_test.lua
      │   └── unit
      │       └── sample_test.lua

*   Configuration files:

    -   ``.luacheckrc``
    -   ``.luacov``
    -   ``.editorconfig``

Using a custom template
^^^^^^^^^^^^^^^^^^^^^^^

You can also make your own custom template. To create an application from it,
run ``cartridge create`` with the ``--from`` flag, specifying the path to your template.

If the template directory is a Git repository,
all files in the ``.git`` directory will be ignored upon instantiating the template.
Instead, a new git repo will be initialized for the newly created application.

Don't include the ``.rocks`` directory in your template application.
To specify application dependencies, use the ``.rockspec`` and ``cartridge.pre-build`` files.

Text variables
^^^^^^^^^^^^^^

Filenames and content can contain `text templates <https://golang.org/pkg/text/template/>`_.

You can use the following variables:

* ``Name``: application name.
* ``StateboardName``: application stateboard name (``<app-name>-stateboard``).
* ``Path``: absolute path to the application.

For example:

..  code-block:: text

    my-template
    ├── {{ .Name }}-scm-1.rockspec
    └── init.lua
    └── stateboard.init.lua
    └── test
        └── sample_test.lua

``init.lua``:

..  code-block:: lua

    print("Hi, I am {{ .Name }} application")
    print("I also have a stateboard named {{ .StateboardName }}")
