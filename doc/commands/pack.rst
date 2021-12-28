..  _cartridge-cli-packing-an-application:

Packaging the application
=========================


To package your application, use the ``pack`` command:

..  code-block:: bash

    cartridge pack TYPE [PATH] [flags]

where:

*   ``TYPE`` (required) is the distribution type. Supported types:

    ..  toctree::
        :maxdepth: 1

        TGZ <pack/tgz>
        RPM/DEB <pack/rpm-deb>
        Docker <pack/docker>

*   ``PATH`` (optional) is the path to the application directory.
    Defaults to ``.`` (the current directory).

Before packaging, ``cartridge pack`` builds the application.
This process is similar to what
``cartridge build`` :doc:`does <build>`.
The resulting artifact includes ``.rocks`` modules and executables
that are specific for the system where you've packaged the application.
For this reason, a distribution built on one OS can't be used on another---for
example, an RPM built on MacOS can't be installed on a CentOS machine.
However, you can work around this by enforcing package build in Docker
via the ``--use-docker`` flag.

..  toctree::
    :maxdepth: 1

    Building in Docker <pack/building-in-docker>

..  note::

    If you use open-source Tarantool, your artifact will have `tarantool` as a
    dependency. Its version will be the same as in your system's ``PATH``.
    If you use Tarantool Enterprise, your artifact will contain the
    ``tarantool`` and ``tarantoolctl`` binaries from your current SDK.

Flags
-----

The following flags control the local packaging of any distribution type,
be it RPM, DEB, TGZ, or a Docker image.

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``--name``
            -   Application name.
                The package and the systemd service will have the same name.
                The default name comes from the ``package`` field in the ``.rockspec`` file.
        *   -   ``--version``
            -   Application package version.
                By default, the version string is the output of ``git describe --tags --long``,
                normalized to ``major.minor.patch.count``.
                If the application is not a git repository,
                you have to set the ``--version`` flag explicitly.
                If you set ``--version`` flag, it will be used as provided.
        *   -   ``--suffix``
            -   The suffix of the resulting file or image name.
                For example, a ``tar.gz`` distribution is named according to the pattern:
                ``<name>-<version>[.<suffix>].<arch>.tar.gz``.
        *   -   ``--use-docker``
            -   Force Cartridge to build the application in Docker.
                Enforced if you're building a Docker image.
        *   -   ``--no-cache``
            -   Disable :ref:`path caching <cartridge-cli-path_caching>`.
                When used with ``cartridge pack docker``, also enforces
                the ``--no-cache`` ``docker`` flag.
 

To learn about distribution-specific flags,
check the documentation for creating Cartridge
:doc:`RPM/DEB distributables <pack/rpm-deb>`
and :doc:`Docker images <pack/docker>`.


Details
-------

Building the package
~~~~~~~~~~~~~~~~~~~~

By default, the package is built inside a temporary directory in
``~/.cartridge/tmp/``. In this way, the packaging process doesn't affect the contents
of your application directory.

When Cartridge copies your application files, it ignores the ``.rocks`` directory.

All file permissions are preserved in the resulting package,
and the code files owner is set to ``root:root``.

Make sure all your application files have at least ``a+r`` permissions
(``a+rx`` for directories). Otherwise, ``cartridge pack`` will raise an error.

Customizing your build directory
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

You can specify a custom build directory for your application using the
``CARTRIDGE_TEMPDIR`` environment variable.
If this directory doesn't exist, it will be created, used
for packaging the application, and then removed.

If you specify an existing directory in the ``CARTRIDGE_TEMPDIR`` environment
variable, the ``CARTRIDGE_TEMPDIR/cartridge.tmp`` directory will be used for
packaging the application and then removed.
Before the packaging starts, this nested directory will be cleaned up.

..  note::

    This may be useful if you want to use your Docker build with GitLab CI.
    Docker volumes don't work properly with the default tmp directory in this case.
    Use ``CARTRIDGE_TEMPDIR=. cartridge pack ...``.

How building works
~~~~~~~~~~~~~~~~~~

This section concern building Cartridge applications locally.
To learn about building them in Docker, check the
:doc:`corresponding documentation page <pack/building-in-docker>`.

Whether you're building a TGZ archive, an RPM/DEB distributable, or a Docker image,
your application is built in three stages.

Stage 1: Cleaning up the application directory
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

At this stage, some files are filtered out of the application directory.

*   First, ``git clean -X -d -f`` removes all untracked and
    ignored files (it works for submodules, too).
*   After that, the ``.git`` directory itself is removed.

Stage 2. Building the application
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

At this stage, ``cartridge`` runs the following:

#.  ``./cartridge.pre-build``, if it exists in the application root directory.
    Learn more about
    :doc:`pre-build and post-build scripts </book/cartridge/cartridge_cli/pre-post-build>`.
    Instead of using the pre-build script, you can define the build logic
    by including ``cmake`` commands in your ``.rockspec``,
    `like we do it in Cartridge <https://github.com/tarantool/cartridge/blob/master/cartridge-scm-1.rockspec#L26>`_.
#.  ``tarantoolctl rocks make``.
    This requires a ``.rockspec`` file in the application root directory.
    If you created your application from template, the file is already there.
    ``cartridge`` installs all dependencies specified in that file.

As a result, the fully built application will appear in the ``.rocks`` directory.
You can start it locally from your application directory.

Stage 3. Cleaning up the files before packing
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

At this stage, ``cartridge`` runs ``cartridge.post-build``, if it exists.
The post-build script removes junk files (like ``node_modules``)
generated during application build.

Learn more about
:doc:`pre-build and post-build scripts </book/cartridge/cartridge_cli/pre-post-build>`.


Versioning
~~~~~~~~~~

The package generates ``VERSION.lua``, a file that contains the current version
of the project. When you connect to an instance with
:doc:`cartridge connect <connect>`,
you can check the project version by obtaining information from this file:

..  code-block:: lua

    require('VERSION')

``VERSION.lua`` is also used when you call
:ref:`cartridge.reload_roles() <cartridge.reload_roles>`:

..  code-block:: lua

    -- Getting the project version
    require('VERSION')
    -- Reloading the instances after making some changes to VERSION.lua
    require('cartridge').reload_roles()
    -- Getting the updated project version
    require('VERSION')

..  note::

    If ``VERSION.lua`` is already in the application directory,
    it will be overwritten during packaging.

..  _cartridge-cli-path_caching:

Path caching
~~~~~~~~~~~~

You can cache paths for packaging Cartridge applications.
For example, if you package an application multiple times,
the same ``.rocks`` are installed every time over and over.
To speed up the repacking process, specify the cached paths in ``pack-cache-config.yml``,
a file located in the application root directory.

By default, the ``.rocks`` directory is cached. The standard template's
``pack-cache-config.yml`` contains the path to that directory:

..  code-block:: yaml

    - path: '.rocks':
      key-path: 'myapp-scm-1.rockspec'
    - path: 'node_modules':
      always-cache: true
    - path: 'third_party/custom_module':
      key: 'simple-hash-key'

Make sure you specify the path to ``.rocks`` from the application root directory
and provide a cache key. Let's look at the example above:

*   ``<path-to-myapp>/.rocks`` will be cached
    depending on the content of ``myapp-scm-1.rockspec``.
*   ``<path-to-myapp>/node_modules`` will always be cached.
*   ``<path-to-myapp>/third_party/custom_module`` 
    will be cached depending on ``simple-hash-key``.

You can't combine these options. For example, you can't specify ``always-cache``
and ``key-path`` at the same time.

One project path can only have one caching key.
Suppose you cached ``.rocks`` with a ``.rockspec`` file as ``key-path``.
Then you changed the contents of ``.rockspec`` and ran ``cartridge pack``.
In this case, the old cache (associated with the old key)
for the project's ``.rocks`` directory path will be deleted.
After packing, the new ``.rocks`` cache path will be saved with the new key.

There can be no more than **5** projects in the cache that have
cached paths.
If the 6th project appears, the oldest existing project is removed
from the cache directory.
However, this is not the case for cached paths within a single project.
You can cache as many paths as you like as long as they are in one project.

To disable caching, use the ``--no-cache`` flag or remove
paths from ``pack-cache-config.yml``. To completely reset the cache,
delete the ``~/.cartridge/tmp/cache`` directory.    

