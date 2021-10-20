Installation
============

1.  Install third-party software:

    *   `Install <https://git-scm.com/book/en/v2/Getting-Started-Installing-Git>`__
        ``git``, the version control system.

    *   `Install <https://linuxize.com/post/how-to-unzip-files-in-linux/>`__
        the ``unzip`` utility.

    *   `Install <https://gcc.gnu.org/install/>`__
        the ``gcc`` compiler.

    *   `Install <https://cmake.org/install/>`__
        the ``cmake`` and ``make`` tools.


2.  Install Tarantool 1.10 or higher:

    You can:

    *   `Install it from a package <https://www.tarantool.io/en/download/>`__.
    *   `Build it from source <https://www.tarantool.io/en/doc/latest/dev_guide/building_from_source/>`__.

3.  [For all platforms except macOS] If you build Tarantool from source,
    you need to set up the Tarantool packages repository manually:

    ..  code-block:: bash

        curl -L https://tarantool.io/installer.sh | sudo -E bash -s -- --repo-only

4.  Install the ``cartridge-cli`` package:

    *   For CentOS, Fedora, ALT Linux (RPM package):

        ..  code-block:: bash

            sudo yum install cartridge-cli

    *   For Debian, Ubuntu (DEB package):

        ..  code-block:: bash

            sudo apt-get install cartridge-cli

    *   For MacOS X (Homebrew formula):

        ..  code-block:: bash

            brew install cartridge-cli

    *   Or build locally:

        .. code-block:: bash

           mage build

5.  Check the installation:

    ..  code-block:: bash

        cartridge version

Enable shell completion
-----------------------

Linux
~~~~~

The ``cartridge-cli`` RPM and DEB packages contain a Bash completion script,
 ``/etc/bash_completion.d/cartridge``.

To enable completion after ``cartridge-cli`` installation, open a new shell or
source the completion file at ``/etc/bash_completion.d/cartridge``.
Make sure that you have ``bash-completion`` installed.

To install Zsh completion, run:

..  code-block:: bash

    cartridge gen completion --skip-bash --zsh="${fpath[1]}/_cartridge"

Now enable shell completion:

..  code-block:: bash

    echo "autoload -U compinit; compinit" >> ~/.zshrc

OS X
~~~~

If you install ``cartridge-cli`` from ``brew``, it automatically installs both
Bash and Zsh completion.

