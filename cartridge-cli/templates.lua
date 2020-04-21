local fio = require('fio')

local utils = require('cartridge-cli.utils')

local templates = {}

local known_templates = {
    cartridge = require('cartridge-cli.templates.cartridge'),
}

function templates.instantiate(dest_dir, template_name, app_name)
    assert(fio.path.exists(dest_dir))

    if known_templates[template_name] == nil then
        return nil, string.format('Template %q does not exists', template_name)
    end

    local template = known_templates[template_name]

    local expand_params = {
        project_name=app_name,
        project_name_lower=string.lower(app_name),
        stateboard_name=utils.get_stateboard_name(app_name)
    }

    for _, file in ipairs(template.files) do
        local filename = utils.expand(file.name, expand_params)

        local filepath = fio.pathjoin(dest_dir, filename)
        local filedir = fio.dirname(filepath)
        if not fio.path.exists(filedir) then
            local ok, err = utils.make_tree(filedir)
            if not ok then return false, err end
        end

        local file_content = utils.expand(file.content, expand_params)

        local ok, err = utils.write_file(filepath, file_content, file.mode)
        if not ok then return false, err end
    end

    return true
end

return templates
