package rpm

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tarantool/cartridge-cli/src/common"
	"github.com/tarantool/cartridge-cli/src/project"
)

func Pack(projectCtx *project.ProjectCtx) error {
	var err error

	cpioPath := filepath.Join(projectCtx.TmpDir, "cpio")
	if err := packCpio(cpioPath, projectCtx); err != nil {
		return fmt.Errorf("Failed to pack CPIO: %s", err)
	}

	compresedCpioPath := filepath.Join(projectCtx.TmpDir, "cpio.gz")
	if err := common.CompressGzip(cpioPath, compresedCpioPath); err != nil {
		return fmt.Errorf("Failed to compress CPIO: %s", err)
	}

	rpmHeader, err := genRpmHeader(cpioPath, compresedCpioPath, projectCtx)
	if err != nil {
		return fmt.Errorf("Failed to gen RPM header: %s", err)
	}

	packedHeader, err := packTagSet(rpmHeader, headerImmutable)
	if err != nil {
		return fmt.Errorf("Failed to pack RPM header: %s", err)
	}

	// write header to file
	rpmHeaderFilePath := filepath.Join(projectCtx.TmpDir, "header")
	rpmHeaderFile, err := os.Create(rpmHeaderFilePath)
	if err != nil {
		return fmt.Errorf("Failed to create RPM body file: %s", err)
	}
	defer rpmHeaderFile.Close()

	if _, err := rpmHeaderFile.Write(*packedHeader); err != nil {
		return fmt.Errorf("Failed to write RPM header to file: %s", err)
	}

	// create body file = header + compressedCpio
	rpmBodyFilePath := filepath.Join(projectCtx.TmpDir, "body")
	if err := common.MergeFiles(rpmBodyFilePath, rpmHeaderFilePath, compresedCpioPath); err != nil {
		return fmt.Errorf("Failed to concat RPM header with compressed payload: %s", err)
	}

	// compute signature
	signature, err := genSignature(rpmBodyFilePath, rpmHeaderFilePath, cpioPath)
	if err != nil {
		return fmt.Errorf("Failed to gen RPM signature: %s", err)
	}

	packedSignature, err := packTagSet(*signature, headerSignatures)
	if err != nil {
		return fmt.Errorf("Failed to pack RPM header: %s", err)
	}
	alignData(packedSignature, 8)

	// compute lead
	lead := genRpmLead(projectCtx.Name)
	lead = append(lead, *packedSignature...)

	// create lead file
	leadFilePath := filepath.Join(projectCtx.TmpDir, "lead")
	leadFile, err := os.Create(leadFilePath)
	if err != nil {
		return fmt.Errorf("Failed to create RPM lead file: %s", err)
	}

	if _, err := leadFile.Write(lead); err != nil {
		return fmt.Errorf("Failed to write RPM lead to file: %s", err)
	}

	// create RPM file
	err = common.MergeFiles(projectCtx.ResPackagePath,
		leadFilePath,
		rpmBodyFilePath,
	)
	if err != nil {
		return fmt.Errorf("Failed to write result RPM file: %s", err)
	}

	return nil
}
