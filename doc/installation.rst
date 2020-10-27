===============================================================================
Installation
===============================================================================

1. Install third-party software:

   * `git <https://git-scm.com/book/en/v2/Getting-Started-Installing-Git>`_,
     a version control system.

   * `unzip <https://linuxize.com/post/how-to-unzip-files-in-linux/>`_ utility.

   * `gcc <https://gcc.gnu.org/install/>`_ compiler.

   * `cmake and make <https://cmake.org/install/>`_ tools.

2. Install Tarantool 1.10 or higher.

   You can:

   * Install it from a package (see https://www.tarantool.io/en/download/
     for OS-specific instructions).
   * Build it from sources (see
     https://www.tarantool.io/en/download/os-installation/building-from-source/).

3. [On all platforms except MacOS X] If you built Tarantool from sources,
   you need to manually set up the Tarantool packages repository:

   .. code-block:: bash

       curl -L https://tarantool.io/installer.sh | sudo -E bash -s -- --repo-only

4. Install the ``cartridge-cli`` package:

   * for CentOS, Fedora, ALT Linux (RPM package):

     .. code-block:: bash

         sudo yum install cartridge-cli

   * for Debian, Ubuntu (DEB package):

     .. code-block:: bash

         sudo apt-get install cartridge-cli

   * for MacOS X (Homebrew formula):

     .. code-block:: bash

         brew install cartridge-cli

5. Check the installation:

   .. code-block:: bash

      cartridge version

-------------------------------------------------------------------------------
Enable shell completion
-------------------------------------------------------------------------------

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Linux
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

RPM and DEB ``cartridge-cli`` packages contain ``/etc/bash_completion.d/cartridge``
Bash completion script.
To enable completion after ``cartridge-cli`` installation start a new shell or
source ``/etc/bash_completion.d/cartridge`` completion file.
Make sure that you have bash completion installed.

To install Zsh completion, say

.. code-block:: bash

    cartridge gen completion --skip-bash --zsh="${fpath[1]}/_cartridge"

To enable shell completion:

.. code-block:: bash

    echo "autoload -U compinit; compinit" >> ~/.zshrc

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
OS X
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

If you install ``cartridge-cli`` from ``brew``, it automatically installs both
Bash and Zsh completions.
