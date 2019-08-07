#!/usr/bin/python3

import subprocess


__tarantool_version = None

def tarantool_enterprise_is_used():
    global __tarantool_version
    if __tarantool_version is None:
        __tarantool_version = subprocess.check_output(['tarantool', '-V']).decode('ascii')

    return __tarantool_version.startswith('Tarantool Enterprise')
