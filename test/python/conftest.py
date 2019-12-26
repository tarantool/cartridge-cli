import py
import os
import pytest
import tempfile

from utils import project_name
from utils import create_project


@pytest.fixture(scope='module')
def module_tmpdir(request):
    dir = py.path.local(tempfile.mkdtemp())
    request.addfinalizer(lambda: dir.remove(rec=1))
    return str(dir)


@pytest.fixture(scope="module")
def project_path(module_tmpdir):
    path = create_project(module_tmpdir, project_name, 'cartridge')

    # add cartridge.post-build file to remove test/ and tmp/ contents
    with open(os.path.join(path, 'cartridge.post-build'), 'w') as f:
        f.write('''
        #!/bin/sh

        rm -rf test
        rm -rf tmp
        ''')
    return path
