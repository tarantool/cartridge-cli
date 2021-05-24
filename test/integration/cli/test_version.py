from utils import run_command_and_get_output


def test_version_command(cartridge_cmd):
    for version_cmd in ["version", "-v", "--version"]:
        rc, output = run_command_and_get_output([cartridge_cmd, version_cmd])
        assert rc == 0
        assert 'Tarantool Cartridge CLI\n Version:\t2' in output
