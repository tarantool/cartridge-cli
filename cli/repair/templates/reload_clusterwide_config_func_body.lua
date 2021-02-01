
local ClusterwideConfig = require('cartridge.clusterwide-config')
local confapplier = require('cartridge.confapplier')

local conf_path, wish_state_timeout = ...

local cfg, err = ClusterwideConfig.load(conf_path)
assert(err == nil, string.format('Failed to load new config: %s', err))

local current_uuid = box.info().uuid
if cfg:get_readonly().topology.servers[current_uuid] == nil then
	return false
end

local roles_configured_state = 'RolesConfigured'
local connecting_fullmesh_state = 'ConnectingFullmesh'

local state = confapplier.wish_state(roles_configured_state, wish_state_timeout)

if state == connecting_fullmesh_state then
	error(string.format(
		'Failed to reach %s config state. Stuck in %s. ' ..
			'Call "box.cfg({replication_connect_quorum = 0})" in instance console and try again',
		roles_configured_state, state
	))
end

if state ~= roles_configured_state then
	error(string.format(
		'Failed to reach %s config state. Stuck in %s',
		roles_configured_state, state
	))
end

cfg:lock()
local ok, err = confapplier.apply_config(cfg)
assert(ok, string.format('Failed to apply new config: %s', err))

return true
