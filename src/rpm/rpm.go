package rpm

import (
	"fmt"
	"path/filepath"

	"github.com/tarantool/cartridge-cli/src/project"
)

func Pack(projectCtx *project.ProjectCtx) error {
	var err error

	cpioPath := filepath.Join(projectCtx.TmpDir, "cpio")
	if err := packCpio(cpioPath, projectCtx); err != nil {
		return fmt.Errorf("Failed to pack CPIO: %s", err)
	}

	_, err = genRpmHeader(cpioPath, projectCtx)
	if err != nil {
		return fmt.Errorf("Failed to gen RPM header: %s", err)
	}

	// lead := genRpmLead(projectCtx.Name)
	// fmt.Printf("lead: %x\n", lead)

	return nil
}
