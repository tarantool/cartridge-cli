local fiber = require('fiber')
fiber.create(function()
    fiber.sleep(1)
end)

require('log').info('I am starting...')

-- ignore SIGTERM
local ffi = require('ffi')
local SIG_IGN = 1
local SIGTERM = 15
ffi.cdef[[
    void (*signal(int sig, void (*func)(int)))(int);
]]
local ignore_handler = ffi.cast("void (*)(int)", SIG_IGN)
ffi.C.signal(SIGTERM, ignore_handler)

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
