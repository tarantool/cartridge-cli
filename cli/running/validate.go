package running

import (
	"fmt"

	"github.com/tarantool/cartridge-cli/cli/context"
)

func Validate(ctx *context.Ctx) error {
	if !ctx.Running.Global && ctx.Running.AppsDir != "" {
		return fmt.Errorf("--apps-dir option can be used only with --global flag")
	}

	return nil
}
