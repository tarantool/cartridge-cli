import os
import re
import subprocess

from utils import create_project
from utils import recursive_listdir
from utils import tarantool_enterprise_is_used


CARTRIDGE_PACK_SPECIAL_FILES = {
    # pre and post build hooks
    'cartridge.pre-build',
    'cartridge.post-build',

    # deprecated files
    '.cartridge.ignore',
    '.cartridge.pre',
}


# ###############
# Helpers
# ###############
def get_base_project_rocks(project_name, rockspec_name):
    return {
        '.rocks',
        '.rocks/share/tarantool/rocks/manifest',
        os.path.join('.rocks/share/tarantool/rocks', project_name),
        os.path.join('.rocks/share/tarantool/rocks', project_name, 'scm-1'),
        os.path.join('.rocks/share/tarantool/rocks', project_name, 'scm-1/rock_manifest'),
        os.path.join('.rocks/share/tarantool/rocks', project_name, 'scm-1', rockspec_name),
    }


def filter_out_files_removed_on_pack(project_files):
    # remove .git files and special files (pre and post build hooks)
    def is_removed_on_pack(filename):
        if filename in CARTRIDGE_PACK_SPECIAL_FILES:
            return True

        if filename.startswith('.git') and filename != '.gitignore':
            return True

    return set(filter(
        lambda x: not is_removed_on_pack(x),
        project_files
    ))


###############
# Class Project
###############
class Project:
    def __init__(self, name, basepath, template='cartridge'):
        self.name = name
        self.basepath = basepath
        self.template = template
        # create project and save its path
        self.path = create_project(basepath, name, template)

        # save tarantool_enterprise_is_used() result to variable
        tarantool_is_enterprise = tarantool_enterprise_is_used()

        # files that should be delivered in the result package
        project_files = recursive_listdir(self.path)
        self.distribution_files = filter_out_files_removed_on_pack(project_files)
        self.distribution_files.add('VERSION')
        if tarantool_is_enterprise:
            self.distribution_files.update({'tarantool', 'tarantoolctl'})

        # project rockspec name and path
        self.rockspec_name = '{}-scm-1.rockspec'.format(self.name)
        self.rockspec_path = os.path.join(self.path, self.rockspec_name)

        # rocks that should be delivered in the result package
        self.rocks_content = get_base_project_rocks(self.name, self.rockspec_name)

        # keys that should be mentioned in the package VERSION file
        self.version_file_keys = {
            'TARANTOOL',
            self.name,
            # default application dependencies
            'cartridge',
            'luatest',
        }
        if tarantool_is_enterprise:
            self.version_file_keys.add('TARANTOOL_SDK')


# ###############
# Project helpers
# ###############
def remove_dependency(project, dependency_name):
    with open(project.rockspec_path, 'r') as f:
        current_rockspec = f.read()

    new_rockspec = re.sub(
        r"'{}\s+==\s+\S+,\n".format(dependency_name),
        '',
        current_rockspec
    )

    with open(project.rockspec_path, 'w') as f:
        f.write(new_rockspec)

    project.version_file_keys.difference_update({dependency_name})


def add_dependency(project, dependency_name, dependency_version='scm-1'):
    with open(project.rockspec_path, 'r') as f:
        current_rockspec = f.read()
    new_rockspec = re.sub(
        r"dependencies\s+=\s+{\s*\n",
        '\n'.join([
            "dependencies = {",
            "   '{} == {}',\n".format(dependency_name, dependency_version)
        ]),
        current_rockspec
    )

    with open(project.rockspec_path, 'w') as f:
        f.write(new_rockspec)

    project.version_file_keys.update({dependency_name})


def add_dependency_submodule(project):
    SUBMODULE_NAME = 'custom-module'

    # create submodule itself
    submodule_path = os.path.join(project.path, 'third_party', SUBMODULE_NAME)
    os.makedirs(submodule_path)
    with open(os.path.join(submodule_path, '{}-scm-1.rockspec'.format(SUBMODULE_NAME)), 'w') as f:
        rockspec_lines = [
            "package = '{}'".format(SUBMODULE_NAME),
            "version = 'scm-1'",
            "source  = { url = '/dev/null' }",
            "build = { type = 'none'}",
        ]
        f.write('\n'.join(rockspec_lines))

    # init git repo and add to project as a submodule
    process = subprocess.run(['git', 'init'], cwd=submodule_path)
    assert process.returncode == 0, "Failed to init git repo for project submodule"

    process = subprocess.run(['git', 'add', '-A'], cwd=submodule_path)
    assert process.returncode == 0, "Failed to add project files to git"
    process = subprocess.run(['git', 'commit', '-m', '"Init"'], cwd=submodule_path)
    assert process.returncode == 0, "Failed to add initial commin"

    submodule_relpath = os.path.join('.', os.path.relpath(submodule_path, project.path))
    process = subprocess.run(
        ['git', 'submodule', 'add', submodule_relpath, submodule_relpath],
        cwd=project.path
    )
    assert process.returncode == 0, "Failed to add a submodule"

    project.distribution_files.add('.gitmodules')

    # add third-party module dependency to the rockspec
    add_dependency(project, SUBMODULE_NAME)

    # add submodule to rocks content
    project.rocks_content.add('.rocks/share/tarantool/rocks/{}'.format(SUBMODULE_NAME))

    # add cartridge.pre-build file to install submodule dependency
    with open(os.path.join(project.path, 'cartridge.pre-build'), 'w') as f:
        prebuild_script_lines = [
            "#!/bin/sh",
            "tarantoolctl rocks make --chdir ./third_party/{}".format(SUBMODULE_NAME),
        ]
        f.write('\n'.join(prebuild_script_lines))

    # add cartridge.post-build file to remove test/, tmp/ and third_party/ contents
    # and remove test/ and tmp/ from project.distribution_files
    with open(os.path.join(project.path, 'cartridge.post-build'), 'w') as f:
        postbuild_script_lines = [
            "#!/bin/sh",
            "rm -rf test tmp third_party"
        ]
        f.write('\n'.join(postbuild_script_lines))

        project.distribution_files = set(filter(
            lambda x: not any([x.startswith(p) for p in ['test', 'tmp']]),
            project.distribution_files
        ))

    # add custom-project to version_file_keys
    project.version_file_keys.add(SUBMODULE_NAME)


ignored_data = [
    {'dir': '',                  'file': 'ignored.txt'},
    {'dir': '',                  'file': 'asterisk'},
    {'dir': '',                  'file': 'ignored.lua'},
    {'dir': '',                  'file': 'ignored_by.format'},
    {'dir': 'ignored',           'file': 'sample.txt'},
    {'dir': 'ignored/folder',    'file': 'sample.txt'},
    {'dir': 'ignored/asterisk',  'file': 'star.txt'},
    {'dir': 'ignored/asterisk',  'file': 'simple'},
    {'dir': 'ignored/sample',    'file': 'test'},
    {'dir': 'ignored',           'file': '#test'}
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
    # an alternative for post-build hook in original
    'third_party',
    'test',
    'tmp',
    # comment example
    '# /scm-1',
    # escaping \#
    '\\#test'
]


def use_deprecated_files(project):
    def create_file(path, text=None):
        with open(path, 'w') as f:
            if text:
                f.write(text)

    # create .cartridge.ignore file
    for item in ignored_data:
        directory = os.path.join(project.path, item['dir'])
        if not os.path.exists(directory):
            os.makedirs(directory)
        create_file(os.path.join(directory, item['file']))

    create_file(
        os.path.join(project.path, ".cartridge.ignore"),
        '\n'.join(ignore_patterns)
    )

    # use deprecated pre-build hook name
    os.rename(
        os.path.join(project.path, 'cartridge.pre-build'),
        os.path.join(project.path, '.cartridge.pre')
    )

    # remove post-build hook
    os.remove(os.path.join(project.path, 'cartridge.post-build'))

    # update distribution files
    project.distribution_files = project.distribution_files.union({
        'ignored',  # special folder for test work cartridge ignore
        'ignored/asterisk',
    })


def remove_all_dependencies(project):
    with open(project.rockspec_path, 'w') as f:
        f.write('''
                package = '{}'
                version = 'scm-1'
                source  = {{ url = '/dev/null' }}
                dependencies = {{ 'tarantool' }}
                build = {{ type = 'none' }}
            '''.format(project.name))
