#!/usr/bin/env tarantool

local fiber = require('fiber')
fiber.create(function()
    fiber.sleep(1)
end)

assert(os.getenv('TARANTOOL_APP_NAME'))
assert(os.getenv('TARANTOOL_INSTANCE_NAME'))
assert(os.getenv('TARANTOOL_CFG'))
assert(os.getenv('TARANTOOL_PID_FILE'))
assert(os.getenv('TARANTOOL_CONSOLE_SOCK'))

fiber.sleep(0.01) -- let `cartridge start` write pid_file and start listening socket
-- Copied from cartridge.cfg to provide support for NOTIFY_SOCKET in old tarantool
local tnt_version = string.split(_TARANTOOL, '.')
local tnt_major = tonumber(tnt_version[1])
local tnt_minor = tonumber(tnt_version[2])
if tnt_major < 2 or (tnt_major == 2 and tnt_minor < 2) then
  local notify_socket = os.getenv('NOTIFY_SOCKET')
  if notify_socket then
      local socket = require('socket')
      local sock = assert(socket('AF_UNIX', 'SOCK_DGRAM', 0), 'Can not create socket')
      sock:sendto('unix/', notify_socket, 'READY=1')
  end
end
