import py
import pytest
import tempfile

@pytest.fixture(scope='module')
def module_tmpdir(request):
    dir = py.path.local(tempfile.mkdtemp())
    request.addfinalizer(lambda: dir.remove(rec=1))
    return str(dir)
