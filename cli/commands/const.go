package commands

import "time"

// DEFAULT VALUES
const (
	defaultStartTimeout = 1 * time.Minute
	defaultLogLines     = 15
	defaultNetMsgMax    = 768
)

// ENV
const (
	cartridgeTmpDirEnv = "CARTRIDGE_TEMPDIR"
)
