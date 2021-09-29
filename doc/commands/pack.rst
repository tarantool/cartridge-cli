Packaging the application
=========================

To package your application, use the ``pack`` command:

..  code-block:: bash

    cartridge pack TYPE [PATH] [flags]

where:

*   ``TYPE`` (required) is the distribution type. Supported types:

    -   :doc:`TGZ <book/cartridge/cartridge_cli/commands/pack/tgz>`
    -   :doc:`RPM <book/cartridge/cartridge_cli/commands/pack/rpm_deb>`
    -   :doc:`DEB <book/cartridge/cartridge_cli/commands/pack/rpm_deb>`
    -   :doc:`Docker <book/cartridge/cartridge_cli/commands/pack/docker>`

*   ``PATH`` (optional) is the path to the application directory.
    Defaults to ``.`` (the current directory).

Before packaging, ``cartridge pack`` builds the application. This process is similar to what
``cartridge build`` :doc:`does </book/cartridge/cartridge_cli/commands/build>`.
The resulting artifact includes ``.rocks`` modules and executables
that are specific for the system where you've packaged the application.
For this reason, a distribution built on one OS can't be used on another---for
example, an RPM built on MacOS can't be installed on a CentOS machine.
However, you can work around this by enforcing package build in Docker
via the ``--use-docker`` flag.
Learn more about
:doc:`building in Docker </book/cartridge/cartridge_cli/commands/pack/building_in_docker>`.

..  note::

    If you use open-source Tarantool, your artifact will have `tarantool` as a
    dependency. Its version will be the same as in your system's ``PATH``.
    If you use Tarantool Enterprise, your artifact will contain the
    ``tarantool`` and ``tarantoolctl`` binaries from your current SDK.

Flags
-----

All distribution types
~~~~~~~~~~~~~~~~~~~~~~

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
                Expected pattern: ``major.minor.patch[-count][-commit]``.
                Input like ``major.minor.patch`` will be normalized to
                ``major.minor.patch-count``.
                The default version is the output of ``git describe --tags --long``.
                If the application is not a git repository,
                you have to set the ``--version`` flag explicitly.
        *   -   ``--suffix``
            -   The suffix of the resulting file or image name.
                For example, a tar distribution is named according to the pattern:
                ``<name>-<version>[-<suffix>].tar.gz``.
        *   -   ``--use-docker``
            -   Force Cartridge to build the application in Docker.


RPM/DEB
~~~~~~~

Use the following flags to control the local packaging of an RPM or DEB distribution.

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``--unit-template``
            -   Path to the template for the ``systemd`` unit file.
        *   -   ``--instantiated-unit-template``
            -   Path to the template for the ``systemd`` instantiated unit file.
        *   -   ``--stateboard-unit-template``
            -   Path to the template for the stateboard ``systemd`` unit file.

Learn more about the
:doc:`package contents and unit file customization <book/cartridge/cartridge_cli/pack/rpm_deb>`.

Docker image
~~~~~~~~~~~~

Use these flags to control the local packaging of a Docker image.

..  container:: table

    ..  list-table::
        :widths: 25 75
        :header-rows: 0

        *   -   ``--tag``
            -   Tag(s) of the Docker image that results from ``cartridge pack docker``.
        *   -   ``--from``
            -   Path to the base Dockerfile of the final image.
                Defaults to ``Dockerfile.cartridge`` in the application root directory.

Learn more about
:doc:`the contents of the image and how to run it <book/cartridge/cartridge_cli/pack/docker>`.

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

    This is especially useful if you want to use your Docker build with GitLab CI.
    Docker volumes don't work properly with the default tmp directory in this case.
    Use ``CARTRIDGE_TEMPDIR=. cartridge pack ...``.

How building works
~~~~~~~~~~~~~~~~~~

This section concern building Cartridge applications locally.
To learn about building them in Docker, check the
:doc:`corresponding documentation page </book/cartridge/cartridge_cli/pack/building_in_docker>`.

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

Path caching
~~~~~~~~~~~~
..  // TODO
