#!/usr/bin/env tarantool

local fiber = require('fiber')
fiber.create(function()
    fiber.sleep(1)
end)
