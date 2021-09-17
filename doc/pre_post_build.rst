Pre-build and post-build scripts
================================

Put these files in your application directory to control the packaging process.
See the examples below.

*   ``cartridge.pre-build`` is a script that runs before ``tarantoolctl rocks make``.
    The main purpose of this script is to build non-standard rocks modules
    (for example, from a submodule).
    Must be executable.

*   ``cartridge.post-build`` is a script that runs after ``tarantoolctl rocks make``.
    The main purpose of this script is to remove build artifacts from the final package.
    Must be executable.


Example: cartridge.pre-build
----------------------------

..  code-block:: bash

    #!/bin/sh

    # The main purpose of this script is to build non-standard rocks modules.
    # It will run before `tarantoolctl rocks make` during application build.

    tarantoolctl rocks make --chdir ./third_party/my-custom-rock-module


Example: cartridge.post-build
-----------------------------

..  code-block:: bash

    #!/bin/sh

    # The main purpose of this script is to remove build artifacts from resulting package.
    # It will run after `tarantoolctl rocks make` during application build.

    rm -rf third_party
    rm -rf node_modules
    rm -rf doc
