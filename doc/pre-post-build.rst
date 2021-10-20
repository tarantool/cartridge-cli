Pre-build and post-build scripts
================================

Put the files ``cartridge.pre-build`` and ``cartridge.post-build``
in your application directory to control the packaging process.

..  note::

    These files are not to be confused with
    :ref:`pre-install and post-install scripts <cartridge-cli-preinst_postinst>`,
    which can be added to an RPM/DEB package of your Cartridge application.


cartridge.pre-build
-------------------

If your application depends on closed-source rocks, or if the build should contain
rocks from a project added as a submodule, then you need to **install** all these
dependencies before calling ``tarantoolctl rocks make``. 
To avoid doing it manually, use the file ``cartridge.pre-build``.

``cartridge.pre-build`` is a script that runs before ``tarantoolctl rocks make``.
The main purpose of this script is to build non-standard rocks modules
(for example, from a submodule). Specify in it all the ``.rocks`` to build from submodules.
For example: ``tarantoolctl rocks make --chdir ./third_party/proj``.

The file must be executable.

If you created your application from template,
``cartridge.pre-build`` is already in your application directory.


Example
~~~~~~~

..  code-block:: bash

    #!/bin/sh

    # The main purpose of this script is to build non-standard rocks modules.
    # It will run before `tarantoolctl rocks make` during application build.

    tarantoolctl rocks make --chdir ./third_party/my-custom-rock-module


cartridge.post-build
--------------------

``cartridge.post-build`` is a script that runs after ``tarantoolctl rocks make``.
The main purpose of this script is to remove build artifacts from the final package.
Must be executable.

Example
~~~~~~~

..  code-block:: bash

    #!/bin/sh

    # The main purpose of this script is to remove build artifacts from resulting package.
    # It will run after `tarantoolctl rocks make` during application build.

    rm -rf third_party
    rm -rf node_modules
    rm -rf doc

