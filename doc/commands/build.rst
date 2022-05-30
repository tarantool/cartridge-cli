Building your application locally
=================================

To build your application locally (for local testing), run this in any directory:

..  code-block:: bash

    cartridge build [PATH] [flags]

Flags
-----

..  container:: table

    ..  list-table::
        :widths: 20 80
        :header-rows: 0

        *   -   ``--spec``
            -   Path to a custom ``.rockspec`` file
                that you want use for the current build.

If you run ``cartridge build`` without the ``--spec`` flag,
your application directory must contain a ``.rockspec``.
The file is already in that directory if you created your app from the default template.

``build`` also supports :doc:`global flags </book/cartridge/cartridge_cli/global-flags>`.
The ``--quiet`` flag is particularly convenient when building an application.

Details
-------

The command requires one argument -- the path to your application directory
(that is, to the build source).
The default path is ``.`` (current directory).

``cartridge build`` runs:

1.  ``./cartridge.pre-build`` (if this file exists in the application root directory)
2.  ``tarantoolctl rocks make``

During step 2 -- the key step here -- ``cartridge`` installs all dependencies
specified in the ``.rockspec`` file.

If your application depends on closed-source rocks, or if the build should contain
rocks from a project added as a submodule, **install** all these
dependencies **before** calling ``tarantoolctl rocks make``.
You can do so using a special file, ``cartridge.pre-build``,
which has to be located in your application directory.
If you created your application from template, the directory already contains the file.

In ``cartridge.pre-build``, specify all the rocks to build from submodules.
For example, add the following line:

..  code-block:: bash
    
    tarantoolctl rocks make --chdir ./third_party/proj

To learn more, read about
:doc:`pre-build and post-build scripts </book/cartridge/cartridge_cli/pre-post-build>`.

The fully built application will appear in the ``.rocks`` directory.
You can start it locally from your application directory.

Instead of using the pre-build script, you can define the build logic
by including ``cmake`` commands in your ``.rockspec``,
`like we do it in Cartridge <https://github.com/tarantool/cartridge/blob/master/cartridge-scm-1.rockspec#L26>`_.

