Cartridge application lifecycle
===============================

In a nutshell:

1.  :doc:`Create an application</book/cartridge/cartridge_cli/commands/create/>`
    (for example, ``myapp``) from a template:

    ..  code-block:: bash

        cartridge create --name myapp
        cd ./myapp

2.  :doc:`Build the application </book/cartridge/cartridge_cli/commands/build/>`
    for local development and testing:

    ..  code-block:: bash

        cartridge build

3.  :doc:`Run instances locally </book/cartridge/cartridge_cli/commands/start/>`:

    ..  code-block:: bash

        cartridge start
        cartridge stop

4.  :doc:`Pack the application </book/cartridge/cartridge_cli/commands/pack/>`
    into a distributable (like an RPM package):

    ..  code-block:: bash

        cartridge pack rpm

test

