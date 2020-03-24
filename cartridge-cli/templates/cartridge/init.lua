local utils = require('cartridge-cli.utils')
local cartridge_template = {}

-- Cartridge application structure:
--
-------------- app_files -------------------
-- ├── ${project_name_lower}-scm-1.rockspec
-- ├── app
-- │   └── roles
-- │       └── custom.lua
-- ├── init.lua
--------------- special_files --------------
-- ├── cartridge.pre-build
-- ├── cartridge.post-build
-- ├── Dockerfile.build.cartridge
-- ├── Dockerfile.cartridge
----------------- dev_files ----------------
-- ├── deps.sh
-- ├── instances.yml
-- ├── .cartridge.yml
-- └── tmp
--     └── .keep
----------------- test_files ----------------
-- ├── test
-- │   ├── helper
-- │   │   ├── integration.lua
-- │   │   └── unit.lua
-- │   ├── helper.lua
-- │   ├── integration
-- │   │   └── api_test.lua
-- │   └── unit
-- │       └── sample_test.lua
---------------- config_files ---------------
-- ├── .luacheckrc
-- ├── .luacov
-- ├── .editorconfig
--

local app_files = require('cartridge-cli.templates.cartridge.files.app_files')
local special_files = require('cartridge-cli.templates.cartridge.files.special_files')
local dev_files = require('cartridge-cli.templates.cartridge.files.dev_files')
local test_files = require('cartridge-cli.templates.cartridge.files.test_files')
local config_files = require('cartridge-cli.templates.cartridge.files.config_files')

local files = utils.merge_lists(
    app_files,
    special_files,
    dev_files,
    test_files,
    config_files
)

for _, f in ipairs(files) do
    f.content = utils.remove_leading_spaces(f.content, 12) .. '\n'
end

cartridge_template.files = files

return cartridge_template
