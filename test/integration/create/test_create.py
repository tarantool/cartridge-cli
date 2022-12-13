import os
import stat
import subprocess

import pytest
from project import Project, copy_project
from utils import recursive_listdir, run_command_and_get_output


@pytest.fixture(scope="module")
def default_project(cartridge_cmd, module_tmpdir):
    project = Project(cartridge_cmd, 'default-project', module_tmpdir, 'cartridge')
    return project


def test_project(default_project):
    project = default_project

    copy_project("default_project", project)

    process = subprocess.run(['./deps.sh'], cwd=project.path)
    assert process.returncode == 0, "Installing deps failed"

    process = subprocess.run(['.rocks/bin/luacheck', '.'], cwd=project.path)
    assert process.returncode == 0, "luacheck failed"

    process = subprocess.run(['.rocks/bin/luatest'], cwd=project.path)
    assert process.returncode == 0, "luatest failed"


def test_project_recreation(cartridge_cmd, default_project):
    project = default_project

    # try to create project with the same name in the same place
    cmd = [
        cartridge_cmd, "create",
        "--name", project.name,
        "--template", project.template,
        project.basepath
    ]

    rc, output = run_command_and_get_output(cmd)
    assert rc == 1
    assert "Application already exists in {}".format(project.path[:-6]) in output

    # check that project directory wasn't deleted
    assert os.path.exists(project.path)


def test_both_template_and_from_specified(cartridge_cmd, tmpdir):
    cmd = [
        cartridge_cmd, "create",
        "--name", "myapp",
        "--template", "cartridge",
        "--from", "path/to",
        tmpdir,
    ]

    rc, output = run_command_and_get_output(cmd)
    assert rc == 1
    assert "You can specify only one of --from and --template options" in output


def test_from_is_not_a_directory(cartridge_cmd, tmpdir):
    filepath = os.path.join(tmpdir, 'template-file')
    with open(filepath, 'w') as f:
        f.write('{{ .Name }}\n')

    cmd = [
        cartridge_cmd, "create",
        "--name", "myapp",
        "--from", filepath,
        tmpdir,
    ]

    rc, output = run_command_and_get_output(cmd)
    assert rc == 1
    assert "Specified path is not a directory" in output


@pytest.fixture(scope="function")
def app_template(tmpdir):
    template_dir = os.path.join(tmpdir, 'simple-app-template')
    os.makedirs(template_dir)

    # initialize git to check that .git files are ignored
    assert subprocess.run(['git', 'init'], cwd=template_dir).returncode == 0

    template_files = {
        'init.lua': {
            'mode': 0o755,
            'content': '''#!/usr/bin/env tarantool
print('Hi, I am {{ .Name }} app entry point.')
print('I also have stateboard instance named {{ .StateboardName }}')
''',
        },
        '{{ .Name }}-scm-1.rockspec': {
            'mode': 0o644,
            'content': '''package = '{{ .Name }}'
version = 'scm-1'
source  = {
    url = '/dev/null',
}
-- Put any modules your app depends on here
dependencies = {
    'tarantool',
    'lua >= 5.1',
    'checks == 3.1.0-1',
    'cartridge == 2.7.7-1',
    'metrics == 0.15.1-1',
}
build = {
    type = 'none';
}
'''
        },
        '.luacheckrc': {
            'mode': 0o644,
            'content': "include_files = {'**/*.lua', '*.luacheckrc', '*.rockspec'}\n"
        },
        'test': {
            'mode': 0o755,
        },
        'test/helper': {
            'mode': 0o755,
        },
        'test/helper/integration.lua': {
            'mode': 0o644,
            'content': '''local t = require('luatest')

local cartridge_helpers = require('cartridge.test-helpers')
local shared = require('test.helper')

local helper = {shared = shared}
'''
        },
    }

    for filepath, fileinfo in template_files.items():
        fullpath = os.path.join(template_dir, filepath)
        dirname = os.path.dirname(fullpath)

        if not os.path.exists(dirname):
            os.makedirs(dirname)

        if fileinfo.get('content') is not None:
            with open(fullpath, 'w') as f:
                f.write(fileinfo['content'])
        else:
            os.makedirs(fullpath)

        os.chmod(fullpath, fileinfo['mode'])

    return {
        'path': template_dir,
        'files': template_files,
    }


def test_from(cartridge_cmd, app_template, tmpdir):
    APPNAME = "myapp"

    cmd = [
        cartridge_cmd, "create",
        "--from", app_template['path'],
        "--name", APPNAME,
        tmpdir,
    ]

    rc, output = run_command_and_get_output(cmd)
    assert rc == 0

    app_path = os.path.join(tmpdir, APPNAME)
    assert not os.path.exists(os.path.join(app_path, '.rocks'))

    files = recursive_listdir(app_path)
    files = {f for f in files if not f.startswith('.git')}

    def subst(s):
        return s.replace('{{ .Name }}', APPNAME).replace('{{ .StateboardName }}', '%s-stateboard' % APPNAME)

    exp_files = {}
    for filepath, fileinfo in app_template['files'].items():
        exp_files.update({
            subst(filepath): {
                'mode': fileinfo['mode'],
                'content': subst(fileinfo['content']) if fileinfo.get('content') is not None else None,
            }
        })

    assert set(exp_files.keys()) == files

    for filepath, fileinfo in exp_files.items():
        created_file_path = os.path.join(app_path, filepath)
        created_file_mode = os.stat(created_file_path)[stat.ST_MODE] & 0o777
        assert created_file_mode == fileinfo['mode']

        if fileinfo.get('content') is None:
            continue

        with open(created_file_path, 'r') as f:
            created_file_content = f.read()

        assert created_file_content == fileinfo['content']


def test_from_with_rocks(cartridge_cmd, app_template, tmpdir):
    APPNAME = "myapp"

    rocks_dir_path = os.path.join(app_template['path'], '.rocks')
    os.makedirs(rocks_dir_path)

    cmd = [
        cartridge_cmd, "create",
        "--from", app_template['path'],
        "--name", APPNAME,
        tmpdir,
    ]

    rc, output = run_command_and_get_output(cmd)
    assert rc == 1

    assert "Project template shouldn't contain .rocks directory" in output
