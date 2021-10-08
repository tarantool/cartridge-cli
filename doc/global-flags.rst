Global flags
============

All Cartridge CLI commands support these flags:

..  container:: table

    ..  list-table::
        :widths: 20 80
        :header-rows: 0

        *   -   ``--verbose``
            -   Run commands with verbose output,
                including the output of nested commands like
                ``tarantoolctl rocks make`` or ``docker build``.
        *   -   ``--debug``
            -   Run command in debug mode---that is,
                with verbose output and without removing temporary files.
                Useful for debugging ``cartridge pack``.
        *   -   ``--quiet``
            -   Hide command output, only display error messages.
                Useful for suppressing the huge output
                of ``cartridge pack`` and ``cartridge build``.

test

