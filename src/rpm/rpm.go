package rpm

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tarantool/cartridge-cli/src/common"
	"github.com/tarantool/cartridge-cli/src/project"
)

func Pack(projectCtx *project.ProjectCtx) error {
	// var err error

	cpioPath := filepath.Join(projectCtx.TmpDir, "cpio")
	if err := packCpio(cpioPath, projectCtx); err != nil {
		return fmt.Errorf("Failed to pack CPIO: %s", err)
	}

	if cpioInfo, err := os.Stat(cpioPath); err != nil {
		fmt.Printf("Errr: %s\n", err.Error())
	} else {
		fmt.Printf("cpioInfo.Size: %d\n", cpioInfo.Size())
	}

	compresedCpioPath := filepath.Join(projectCtx.TmpDir, "cpio.gz")
	if err := common.CompressGzip(cpioPath, compresedCpioPath); err != nil {
		return fmt.Errorf("Failed to compress CPIO: %s", err)
	}

	lead := genRpmLead(projectCtx.Name)
	fmt.Printf("lead: %x\n", lead)

	rpmHeader, err := genRpmHeader(cpioPath, compresedCpioPath, projectCtx)
	if err != nil {
		return fmt.Errorf("Failed to gen RPM header: %s", err)
	}

	packedHeader, err := packTagSet(rpmHeader)
	if err != nil {
		return fmt.Errorf("Failed to pack RPM header: %s", err)
	}

	fmt.Printf("packedHeader: %x\n", packedHeader)

	return nil
}
