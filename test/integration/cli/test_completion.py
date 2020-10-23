import subprocess
import os


def test_completion(cartridge_cmd, tmpdir):
    cmd = [
        cartridge_cmd, "gen", "completion",
    ]

    process = subprocess.run(cmd, cwd=tmpdir)
    assert process.returncode == 0

    comp_names = [
        "completion/bash/cartridge",
        "completion/zsh/_cartridge",
    ]

    for comp_name in comp_names:
        comp_path = os.path.join(tmpdir, comp_name)
        assert os.path.exists(comp_path)

        filemode = os.stat(comp_path).st_mode & 0o777
        assert filemode == 0o644
