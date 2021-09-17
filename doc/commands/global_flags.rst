Global flags
============

All Cartridge CLI commands support these flags:

* ``--verbose``: run command with verbose output,
  including the output of nested commands like
  ``tarantoolctl rocks make`` or ``docker build``.

* ``--debug``: run command in debug mode---that is,
  with verbose output and without removing temporary files.
  Useful for debugging ``cartridge pack``.

* ``--quiet``: hide build commands output.
  Useful for silencing the huge output
  of ``cartridge pack`` and ``cartridge build``.
