local t = require('luatest')
local app = require('cartridge-cli')

local g = t.group('dockerfile_constructors')

local function remove_leading_spaces(s, spaces_num)
    spaces_num = spaces_num or 8
    local REMOVE_PATTERN = string.format('^%s', string.rep(' ', spaces_num))

    local res_lines = {}
    for _, line in ipairs(s:split('\n')) do
        local res_line = line:gsub(REMOVE_PATTERN, '')
        table.insert(res_lines, res_line)
    end

    return table.concat(res_lines, '\n')
end

local function check_output(command, ...)
    local cmd = string.format(command, ...)
    local res, popen_err = io.popen(string.format('((%s) 2>/dev/null) && echo OK', cmd))

    if res == nil then
        return nil, popen_err
    end

    local output = res:read("*all")
    if output:endswith('OK\n') then
        output = output:gsub('OK\n$', '')
        return output
    end

    local cmd_err = string.format('Failed to execute "%s": %s', cmd, output)
    return nil, cmd_err
end

local function get_copy_tarantool_enterprise_layers(build_sdk_dirname)
    local layers = remove_leading_spaces([=[
        ### Copy Tarantool Enterprise
        COPY <BUILD_SDK_DIRNAME> /usr/share/tarantool/tarantool-enterprise

        ENV PATH="/usr/share/tarantool/tarantool-enterprise:${PATH}"
    ]=]):strip():gsub('<BUILD_SDK_DIRNAME>', build_sdk_dirname)

    return layers
end

local function get_install_tarantool_opensource_layers(repo_version)
    local layers = remove_leading_spaces([=[
        ### Install opensource Tarantool
        RUN curl -s \
                https://packagecloud.io/install/repositories/tarantool/REPO_VERSION/script.rpm.sh | bash \
            && yum -y install tarantool tarantool-devel
    ]=]):strip():gsub('REPO_VERSION', repo_version)

    return layers
end

local function get_prepare_layers()
    local layers = remove_leading_spaces([=[
        # Create Tarantool user and directories
        RUN groupadd -r tarantool \
            && useradd -M -N -g tarantool -r -d /var/lib/tarantool -s /sbin/nologin \
                -c "Tarantool Server" tarantool \
            &&  mkdir -p /var/lib/tarantool/ --mode 755 \
            && chown tarantool:tarantool /var/lib/tarantool \
            && mkdir -p /var/run/tarantool/ --mode 755 \
            && chown tarantool:tarantool /var/run/tarantool
    ]=]):strip()

    return layers
end

local function get_install_packages_for_build_layers()
    local layers = remove_leading_spaces([=[
        ### Install packages required for build
        RUN yum install -y git-core gcc make cmake unzip
    ]=]):strip()
    return layers
end

local function get_wrap_user_layers()
    local user_id, err = check_output('id -u')
    assert(user_id ~= nil, err)
    user_id = user_id:strip()

    local layers = remove_leading_spaces([=[
        ### Wrap user
        RUN if id -u USER_ID 2>/dev/null; then \
                USERNAME=$(id -nu USER_ID); \
            else \
                USERNAME=cartridge; \
                useradd -u USER_ID ${USERNAME}; \
            fi \
            && (usermod -a -G sudo ${USERNAME} 2>/dev/null || :) \
            && (usermod -a -G wheel ${USERNAME} 2>/dev/null || :) \
            && (usermod -a -G adm ${USERNAME} 2>/dev/null || :)

        USER USER_ID
    ]=]):strip():gsub('USER_ID', user_id)

    return layers
end

local function get_dockerfile_set_path_layers(app_name)
    local layers = string.gsub(
        'ENV PATH="/usr/share/tarantool/APP_NAME:${PATH}"\n',
        'APP_NAME', app_name
    )
    return layers
end

local function get_copy_code_layers(app_name)
    local layers = string.gsub(
        'COPY . /usr/share/tarantool/APP_NAME\n',
        'APP_NAME', app_name
    )
    return layers
end

local function get_remove_build_sdk_layers(app_name, sdk_dirname)
    local layers = remove_leading_spaces([=[
        RUN rm -rf /usr/share/tarantool/<APP_NAME>/<SDK_DIRNAME>
    ]=]):strip():gsub('<APP_NAME>', app_name):gsub('<SDK_DIRNAME>', sdk_dirname)
    return layers
end

local function get_dockerfile_runtime_layers(app_name)
    local layers = remove_leading_spaces([=[
        ### Application runtime
        RUN echo 'd /var/run/tarantool 0755 tarantool tarantool' > /usr/lib/tmpfiles.d/APP_NAME.conf \
        && chmod 644 /usr/lib/tmpfiles.d/APP_NAME.conf
        USER tarantool:tarantool
        CMD TARANTOOL_WORKDIR=/var/lib/tarantool/APP_NAME.${TARANTOOL_INSTANCE_NAME:-default} \
            TARANTOOL_PID_FILE=/var/run/tarantool/APP_NAME.${TARANTOOL_INSTANCE_NAME:-default}.pid \
            TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/APP_NAME.${TARANTOOL_INSTANCE_NAME:-default}.control \
            tarantool /usr/share/tarantool/APP_NAME/init.lua
    ]=]):strip():gsub('APP_NAME', app_name)
    return layers
end

local function get_non_emply_lines(str)
    local lines = {}
    for _, line in ipairs(str:split('\n') or {}) do
        if line ~= '' then
            table.insert(lines, line)
        end
    end
    return lines
end

local function assert_lines_are_equal(actual_str, exp_str)
    t.assert_equals(
        get_non_emply_lines(actual_str:strip()),
        get_non_emply_lines(exp_str:strip())
    )
end

local function check_dockerfile(constructor, expected_dockerfile)
    local res, err = constructor()
    t.assert(res ~= nil, err)
    assert_lines_are_equal(res, expected_dockerfile)
end

g.test_install_tarantool_constructor = function()
    local constructor = app.dockerfile_constructors.install_tarantool

    -- Tarantool Enterprise (copy)
    local build_sdk_dirname = 'build-sdk-dirname'
    _G.app_state.tarantool_is_enterprise = true
    _G.app_state.sdk_path = '/path/to/sdk'
    _G.app_state.build_sdk_dirname = build_sdk_dirname

    local expected_dockerfile = get_copy_tarantool_enterprise_layers(build_sdk_dirname)
    check_dockerfile(constructor, expected_dockerfile)

    -- Tarantool 2.1
    _G.app_state.tarantool_is_enterprise = false
    _G.app_state.sdk_path = nil
    _G.app_state.build_sdk_dirname = nil
    _G.app_state.tarantool_version = '2.1.42'

    local expected_dockerfile = get_install_tarantool_opensource_layers('2x')
    check_dockerfile(constructor, expected_dockerfile)

    -- Tarantool 1.10
    _G.app_state.tarantool_is_enterprise = false
    _G.app_state.sdk_path = nil
    _G.app_state.build_sdk_dirname = nil
    _G.app_state.tarantool_version = '1.10.42'

    local expected_dockerfile = get_install_tarantool_opensource_layers('1_10')
    check_dockerfile(constructor, expected_dockerfile)
end

g.test_build_image_dockerfile_constructor = function()
    local constructor = app.dockerfile_constructors.build

    local build_base_dockerfile_layers = remove_leading_spaces([=[
        ### Base layers
        FROM centos:8
        RUN yum install -y zip
    ]=])

    -- Tarantool Enterprise
    local build_sdk_dirname = 'build-sdk-dirname'
    _G.app_state.tarantool_is_enterprise = true
    _G.app_state.sdk_path = '/path/to/sdk'
    _G.app_state.build_sdk_dirname = build_sdk_dirname
    _G.app_state.build_base_dockerfile_layers = build_base_dockerfile_layers

    local expected_dockerfile = table.concat({
        build_base_dockerfile_layers,
        get_install_packages_for_build_layers(),
        get_prepare_layers(),
        get_copy_tarantool_enterprise_layers(build_sdk_dirname),
        get_wrap_user_layers(),
    }, '\n')

    check_dockerfile(constructor, expected_dockerfile)

    -- Tarantool Opensource
    _G.app_state.tarantool_is_enterprise = false
    _G.app_state.sdk_path = nil
    _G.app_state.build_sdk_dirname = nil
    _G.app_state.tarantool_version = '2.1.42'
    _G.app_state.build_base_dockerfile_layers = build_base_dockerfile_layers

    local expected_dockerfile = table.concat({
        build_base_dockerfile_layers,
        get_install_packages_for_build_layers(),
        get_prepare_layers(),
        get_install_tarantool_opensource_layers('2x'),
        get_wrap_user_layers(),
    }, '\n')

    check_dockerfile(constructor, expected_dockerfile)

    -- app_state.build_base_dockerfile_layers is required
    _G.app_state.tarantool_is_enterprise = false
    _G.app_state.sdk_path = nil
    _G.app_state.build_sdk_dirname = nil
    _G.app_state.tarantool_version = '2.1.42'
    _G.app_state.build_base_dockerfile_layers = nil
    local res, err = constructor()
    t.assert_equals(res, nil)
    t.assert_str_icontains(err, 'build base dockerfile layers should be set')
end

g.test_runtime_image_dockerfile_constructor = function()
    local constructor = app.dockerfile_constructors.runtime
    local app_name = 'myapp'

    local runtime_base_dockerfile_layers = remove_leading_spaces([=[
        ### Base layers
        FROM centos:8
        RUN yum install -y unzip
    ]=])

    -- Tarantool Enterprise
    _G.app_state.name = app_name
    local build_sdk_dirname = 'build-sdk-dirname'
    _G.app_state.tarantool_is_enterprise = true
    _G.app_state.sdk_path = '/path/to/sdk'
    _G.app_state.build_sdk_dirname = build_sdk_dirname
    _G.app_state.runtime_base_dockerfile_layers = runtime_base_dockerfile_layers

    local expected_dockerfile = table.concat({
        runtime_base_dockerfile_layers,
        get_prepare_layers(),
        get_copy_code_layers(app_name),
        get_remove_build_sdk_layers(app_name, build_sdk_dirname),
        get_dockerfile_set_path_layers(app_name),
        get_dockerfile_runtime_layers(app_name),
    }, '\n')

    check_dockerfile(constructor, expected_dockerfile)

    -- Tarantool Opensource
    _G.app_state.name = app_name
    _G.app_state.tarantool_is_enterprise = false
    _G.app_state.sdk_path = nil
    _G.app_state.build_sdk_dirname = nil
    _G.app_state.tarantool_version = '2.1.42'
    _G.app_state.runtime_base_dockerfile_layers = runtime_base_dockerfile_layers

    local expected_dockerfile = table.concat({
        runtime_base_dockerfile_layers,
        get_prepare_layers(),
        get_install_tarantool_opensource_layers('2x'),
        get_copy_code_layers(app_name),
        get_dockerfile_runtime_layers(app_name),
    }, '\n')

    check_dockerfile(constructor, expected_dockerfile)

    -- app_state.runtime_base_dockerfile_layers is required
    _G.app_state.tarantool_is_enterprise = false
    _G.app_state.sdk_path = nil
    _G.app_state.build_sdk_dirname = nil
    _G.app_state.tarantool_version = '2.1.42'
    _G.app_state.runtime_base_dockerfile_layers = nil
    local res, err = constructor()
    t.assert_equals(res, nil)
    t.assert_str_icontains(err, 'runtime base dockerfile layers should be set')
end
