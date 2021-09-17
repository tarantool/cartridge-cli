Cartridge application lifecycle
===============================

In a nutshell:

1.  :ref:`Create an application <cartridge-cli-creating_an_application_from_template>`
    (for example, ``myapp``) from a template:

    ..  code-block:: bash

        cartridge create --name myapp
        cd ./myapp

2.  :ref:`Build the application <cartridge-cli-building-the-application>`
    for local development and testing:

    ..   code-block:: bash

        cartridge build

3.  :ref:`Run instances locally <cartridge-cli-starting-the-application-locally>`:

    ..  code-block:: bash

        cartridge start
        cartridge stop

4.  :ref:`Pack the application <cartridge-cli-packaging-the-application>`
    into a distributable (like an RPM package):

    ..  code-block:: bash

        cartridge pack rpm