require('strict').on()

local log = require('log')

log.info(os.getenv('TARANTOOL_CONSOLE_SOCK'))
log.info(os.getenv('TARANTOOL_WORKDIR'))
log.info(os.getenv('TARANTOOL_PID_FILE'))
