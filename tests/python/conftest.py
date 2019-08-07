import pytest
import py
import tempfile
import subprocess
import os

from utils import tarantool_enterprise_is_used


@pytest.fixture(scope='session')
def session_tmpdir(request):
    dir = py.path.local(tempfile.mkdtemp())
    request.addfinalizer(lambda: dir.remove(rec=1))
    return str(dir)


@pytest.fixture(scope="session", params=['plain', 'cluster'])
def project_path(request, session_tmpdir):
    if request.param == 'cluster' and not tarantool_enterprise_is_used():
        pytest.skip('Skip cluster template test for Opensource Tarantool')

    project_name = 'project-{}'.format(request.param)
    cmd = ["tarantoolapp", "create",
        "--name", project_name,
        "--template", request.param]
    process = subprocess.run(cmd, cwd=session_tmpdir)
    assert process.returncode == 0, \
        "Error during creating the project"
    return os.path.join(session_tmpdir, project_name)
