===============================================================================
Building an application in local directory
===============================================================================

To build your application locally (for local testing), say this in any directory:

.. code-block:: bash

    cartridge build [PATH] [flags]

-------------------------------------------------------------------------------
Options
-------------------------------------------------------------------------------

There is no special options for building an application locally.

Command also supports `global flags <./global_flags.rst>`_.

It's very convenient to build application with ``--quiet`` flag.

-------------------------------------------------------------------------------
Details
-------------------------------------------------------------------------------

This command requires one argument — the path to your application directory
(i.e. to the build source). The default path is ``.`` (the current directory).

Application directory should contain rockspec.

This command runs:

1. ``./cartridge.pre-build`` if this file exists in application root
2. ``tarantoolctl rocks make``

During step 2 -- the key step here -- ``cartridge`` installs all dependencies
specified in the rockspec file (you can find this file within the application
directory created from template).

If your application depends on closed-source rocks, or if the build should contain
rocks from a project added as a submodule, then you need to **install** all these
dependencies before calling ``tarantoolctl rocks make``.
You can do it using the file ``cartridge.pre-build`` in your application root
(again, you can find this file within the application directory created from template).
In this file, you can specify all rocks to build from submodules
(e.g. ``tarantoolctl rocks make --chdir ./third_party/proj``).
For details, see `special files <../special_files.rst>`_.

As a result, in the application's ``.rocks`` directory you will get a fully built
application that you can start locally from the application's directory.

(An advanced alternative would be to specify build logic in the
rockspec as ``cmake`` commands, like we
`do it <https://github.com/tarantool/cartridge/blob/master/cartridge-scm-1.rockspec#L26>`_
for ``cartridge``.)
