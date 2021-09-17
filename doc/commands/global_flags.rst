Global flags
============

Each Cartridge CLI command supports these flags:

* ``--verbose`` - run command with verbose output.

* ``--debug`` - run command in debug mode (i.e. run with verbose output and
  don't remove temporary files).
  Is useful for debugging ``cartridge pack`` debugging.

* ``--quiet`` - hide build commands output.
  Is useful for ``cartridge pack`` and ``cartridge build`` to silence
  huge output of application build.
