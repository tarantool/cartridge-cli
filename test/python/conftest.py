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
    return {
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
    }


def get_rocks_content(project_name):
    return {
        '.rocks',
        '.rocks/share/tarantool/rocks/manifest',
        '.rocks/share/tarantool/rocks/' + project_name,
        '.rocks/share/tarantool/rocks/' + project_name + '/scm-1',
        '.rocks/share/tarantool/rocks/' + project_name + '/scm-1/rock_manifest',
        '.rocks/share/tarantool/rocks/' + project_name + '/scm-1/' + project_name + '-scm-1.rockspec',

        '.rocks/bin/luatest',
        '.rocks/share/tarantool/checks.lua',
        '.rocks/share/tarantool/luatest',
        '.rocks/share/tarantool/rocks/checks',
        '.rocks/share/tarantool/rocks/luatest',
    }


def remove_by_prefix(paths, prefix):
    return {p for p in paths if not p.startswith(prefix)}


@pytest.fixture(scope="module")
def original_project(module_tmpdir):
    project_name = 'original-project'
    project_path = create_project(module_tmpdir, project_name, 'cartridge')

    # add third-party module dependency to the rockspec
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


ignore_patterns = [
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


@pytest.fixture(scope="module")
def deprecared_project(module_tmpdir):
    project_name = 'deprecated-project'
    project_path = create_project(module_tmpdir, project_name, 'cartridge')

    def create_file(path, text=None):
        with open(path, 'w') as f:
            if text:
                f.write(text)

    # create .cartridge.ignore file
    for item in ignored_data:
        directory = os.path.join(project_path, item['dir'])
        if not os.path.exists(directory):
            os.makedirs(directory)
        create_file(os.path.join(directory, item['file']))

    create_file(
        os.path.join(project_path, ".cartridge.ignore"),
        '\n'.join(ignore_patterns)
    )

    os.remove(os.path.join(project_path, 'cartridge.pre-build'))
    os.remove(os.path.join(project_path, 'cartridge.post-build'))

    distribution_files_list = get_distribution_files(project_name)
    distribution_files_list = distribution_files_list.union([
        'ignored',  # special folder for test work cartridge ignore
        'ignored/asterisk',
        'test',
        'test/helper',
        'test/integration',
        'test/unit',
        'tmp',
        'tmp/.keep',
    ])

    return {
        'name': project_name,
        'path': project_path,
        'distribution_files_list': distribution_files_list,
        'rocks_content': get_rocks_content(project_name)
    }


@pytest.fixture(scope="module", params=['original', 'deprecated'])
def project(original_project, deprecared_project, request):
    if request.param == 'original':
        return original_project
    elif request.param == 'deprecated':
        return deprecared_project


@pytest.fixture(scope="module")
def project_without_dependencies(module_tmpdir):
    project_name = 'empty-project'
    project_path = create_project(module_tmpdir, project_name, 'cartridge')

    rockspec_path = os.path.join(project_path, '{}-scm-1.rockspec'.format(project_name))
    with open(rockspec_path, 'w') as f:
        f.write('''
                package = '{}'
                version = 'scm-1'
                source  = {{ url = '/dev/null' }}
                dependencies = {{ 'tarantool' }}
                build = {{ type = 'none' }}
            '''.format(project_name))

    return {
        'name': project_name,
        'path': project_path,
        'distribution_files_list': get_distribution_files(project_name),
        'rocks_content': {},
    }
