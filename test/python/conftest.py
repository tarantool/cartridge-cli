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
    return create_project(module_tmpdir, project_name, 'cartridge')


ignored_data = [
    {
        'dir': '',
        'file': 'ignored.txt'
    },
    {
        'dir': '',
        'file': 'asterisk'
    },
    {
        'dir': '',
        'file': 'ignored.lua'
    },
    {
        'dir': '',
        'file': 'ignored_by.format'
    },
    {
        'dir': 'ignored',
        'file': 'sample.txt'
    },
    {
        'dir': 'ignored/folder',
        'file': 'sample.txt'
    },
    {
        'dir': 'ignored/asterisk',
        'file': 'star.txt'
    },
    {
        'dir': 'ignored/asterisk',
        'file': 'simple'
    },
    {
        'dir': 'ignored/sample',
        'file': 'test'
    },
    {
        'dir': 'ignored',
        'file': '#test'
    }
]


patterns = [
    # patterns that match the patterns from whitelist
    '.rocks/share/tarantool/rocks/**',
    '*.lua',
    'deps.sh',
    # whitelist
    '!*.sh',
    '!.rocks/**',
    '!init.lua',
    '!app/roles/custom.lua',
    '!asterisk/',
    # for ignore
    'ignored.txt',
    '*.format',
    'ignored/*.txt',
    'ignored/folder/',
    '**/*.txt',
    'simple',
    'sample',
    'asterisk',
    # comment example
    '# /scm-1',
    # escaping \#
    '\\#test'
]


cartridge_ignore_text = '\n'.join(patterns)


@pytest.fixture(scope="module")
def prepare_ignore(project_path):
    """function creates files and directories
    to check the work .cartridge.ignore"""

    def create_file(path, text=None):
        with open(path, 'w') as f:
            if text:
                f.write(text)

    for item in ignored_data:
        directory = os.path.join(project_path, item['dir'])
        if not os.path.exists(directory):
            os.makedirs(directory)
        create_file(os.path.join(directory, item['file']))

    create_file(
        os.path.join(project_path, ".cartridge.ignore"),
        cartridge_ignore_text)
