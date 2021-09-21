Building your application locally
=================================

To build your application locally (for local testing), run this in any directory:

..  code-block:: bash

    cartridge build [PATH] [flags]

Flags
-----

The command supports `global flags <./global_flags.rst>`_.
For example, it's very convenient to build an application
with the ``--quiet`` flag.

Details
-------

The command requires one argument---the path to your application directory
(that is, to the build source).
The default path is ``.`` (current directory).

Your application directory must contain a ``.rockspec``.
If you created your application from template, the file is already there.

``cartridge build`` runs:

1. ``./cartridge.pre-build`` if this file exists in the application root directory
2. ``tarantoolctl rocks make``.

During step 2---the key step here---``cartridge`` installs all dependencies
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


To learn more, read about `pre-build and post-build scripts <../pre_post_build.rst>`_.

The fully built application will appear in the ``.rocks`` directory.
You can start it locally from the application directory.

An advanced alternative way o specify the build logic would be to include
``cmake`` commands in the ``.rockspec`` file, like we 
`do it <https://github.com/tarantool/cartridge/blob/master/cartridge-scm-1.rockspec#L26>`_
for ``cartridge``.
