import py
import os
import pytest
import tempfile
import re

from utils import create_project


@pytest.fixture(scope='module')
def module_tmpdir(request):
    dir = py.path.local(tempfile.mkdtemp())
    request.addfinalizer(lambda: dir.remove(rec=1))
    return str(dir)


def get_distribution_files(project_name):
    return set([
        'Dockerfile.cartridge',
        '.cartridge.yml',
        '.editorconfig',
        '.gitignore',
        '.luacheckrc',
        'deps.sh',
        'init.lua',
        'instances.yml',
        'app',
        'app/roles',
        'app/roles/custom.lua',
        project_name + '-scm-1.rockspec',
        'VERSION',
    ])


def get_rocks_content(project_name):
    return set([
        '.rocks',
        '.rocks/share/tarantool/rocks/manifest',
        '.rocks/share/tarantool/rocks/' + project_name,
        '.rocks/share/tarantool/rocks/' + project_name + '/scm-1',
        '.rocks/share/tarantool/rocks/' + project_name + '/scm-1/rock_manifest',
        '.rocks/share/tarantool/rocks/' + project_name + '/scm-1/' + project_name + '-scm-1.rockspec',

        '.rocks/bin/luatest',
        '.rocks/share/tarantool/checks.lua',
        '.rocks/share/tarantool/luarocks/test/luatest.lua',
        '.rocks/share/tarantool/luatest',
        '.rocks/share/tarantool/rocks/checks',
        '.rocks/share/tarantool/rocks/luatest',
    ])


@pytest.fixture(scope="module")
def project(module_tmpdir):
    project_name = 'original-project'
    project_path = create_project(module_tmpdir, project_name, 'cartridge')

    # add third-party module dependency to the rockspec
    current_rockspec = None
    with open(os.path.join(project_path, '{}-scm-1.rockspec'.format(project_name)), 'r') as f:
        current_rockspec = f.read()

    with open(os.path.join(project_path, '{}-scm-1.rockspec'.format(project_name)), 'w') as f:
        f.write(re.sub(
            r"dependencies = {",
            "dependencies = {\n    'custom-module == scm-1',",
            current_rockspec
        ))

    # create custom-module itself
    dependency_module_path = os.path.join(project_path, 'third_party', 'custom-module')
    os.makedirs(dependency_module_path)
    with open(os.path.join(dependency_module_path, 'custom-module-scm-1.rockspec'), 'w') as f:
        rockspec_lines = [
            "package = 'custom-module'",
            "version = 'scm-1'",
            "source  = { url = '/dev/null' }",
            "build = { type = 'none'}",
        ]
        f.write('\n'.join(rockspec_lines))

    # add cartridge.pre-build file to install custom-module dependency
    with open(os.path.join(project_path, 'cartridge.pre-build'), 'w') as f:
        prebuild_script_lines = [
            "#!/bin/sh",
            "tarantoolctl rocks make --chdir ./third_party/custom-module",
        ]
        f.write('\n'.join(prebuild_script_lines))

    # add cartridge.post-build file to remove test/ and tmp/ contents
    with open(os.path.join(project_path, 'cartridge.post-build'), 'w') as f:
        postbuild_script_lines = [
            "#!/bin/sh",
            "rm -rf test tmp third_party"
        ]
        f.write('\n'.join(postbuild_script_lines))

    # add custom-module to rocks content
    rocks_content = get_rocks_content(project_name)
    rocks_content.add('.rocks/share/tarantool/rocks/custom-module')

    return {
        'name': project_name,
        'path': project_path,
        'distribution_files_list': get_distribution_files(project_name),
        'rocks_content': rocks_content,
    }
