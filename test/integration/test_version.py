#!/usr/bin/python3

from utils import run_command_and_get_output


def test_version_command(cartridge_cmd):
    rc, output = run_command_and_get_output([cartridge_cmd, "--version"])
    assert rc == 0
    assert 'Tarantool Cartridge CLI v' in output
