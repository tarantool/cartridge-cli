===============================================================================
Special files
===============================================================================

You can put these files in your application root to control the application
packaging process (see examples below):

* ``cartridge.pre-build``: a script to be run before ``tarantoolctl rocks make``.
  The main purpose of this script is to build some non-standard rocks modules
  (for example, from a submodule).
  Should be executable.

* ``cartridge.post-build``: a script to be run after ``tarantoolctl rocks make``.
  The main purpose of this script is to remove build artifacts from result package.
  Should be executable.

*****************************
Example: cartridge.pre-build
*****************************

.. code-block:: bash

    #!/bin/sh

    # The main purpose of this script is to build some non-standard rocks modules.
    # It will be run before `tarantoolctl rocks make` on application build

    tarantoolctl rocks make --chdir ./third_party/my-custom-rock-module

******************************
Example: cartridge.post-build
******************************

.. code-block:: bash

    #!/bin/sh

    # The main purpose of this script is to remove build artifacts from resulting package.
    # It will be ran after `tarantoolctl rocks make` on application build.

    rm -rf third_party
    rm -rf node_modules
    rm -rf doc
