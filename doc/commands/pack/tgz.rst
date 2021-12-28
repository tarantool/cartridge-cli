Packaging an application into a TGZ archive
===========================================

``cartridge pack tgz`` creates a ``.tgz`` archive.
It contains the directory ``<app-name>``
with the application source code and the ``.rocks`` modules
described in the application's ``.rockspec`` file.

The resulting artifact name is ``<app-name>-<version>[.<suffix>].<arch>.tar.gz``.

