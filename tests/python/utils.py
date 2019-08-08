#!/usr/bin/python3

import subprocess


__tarantool_version = None

def tarantool_version():
    global __tarantool_version
    if __tarantool_version is None:
        __tarantool_version = subprocess.check_output(['tarantool', '-V']).decode('ascii').split('\n')[0]

    return __tarantool_version


def tarantool_enterprise_is_used():
    return tarantool_version().startswith('Tarantool Enterprise')

