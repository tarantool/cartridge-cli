package rpm

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

/**
 *
 *  Many thanks to @knazarov, who wrote packing in RPM in Lua a long time ago
 *  This code can be found here
 *  https://github.com/tarantool/cartridge-cli/blob/cafd75a5c8ddfdb93ef8290d6e4ebd5d83e4c46e/cartridge-cli.lua#L1814
 *
 *  RPM file is a binary file format, consisting of metadata in the form of
 *  key-value pairs and then a gzipped cpio archive (of SVR-4 variety).
 *
 *  Documentation on the binary format can be found here:
 *  - http://ftp.rpm.org/max-rpm/s1-rpm-file-format-rpm-file-format.html
 *  - https://docs.fedoraproject.org/ro/Fedora_Draft_Documentation/0.1/html/RPM_Guide/ch-package-structure.html
 *
 *  Also I've found this explanatory blog post to be of great help:
 *  - https://blog.bethselamin.de/posts/argh-pm.html
 *
 *  Here's what the layout looks like:
 *
 *  +-----------------------+
 *  |                       |
 *  |     Lead (legacy)     |
 *  |                       |
 *  +-----------------------+
 *  |                       |
 *  |   Signature Header    |
 *  |                       |
 *  +-----------------------+
 *  |                       |
 *  |        Header         |
 *  |                       |
 *  +-----------------------+
 *  |                       |
 *  |                       |
 *  |    Data (cpio.gz)     |
 *  |                       |
 *  |                       |
 *  +-----------------------+
 *
 *  Both signature sections have the same format: a set of typed
 *  key-value pairs.
 *
 *  While debugging, I used rpm-dissecting tool from mkrepo:
 *  - https://github.com/tarantool/mkrepo/blob/master/mkrepo.py
 *
 */

// Pack creates an RPM archive ctx.Pack.ResPackagePath
// that contains files from ctx.Pack.PackageFilesDir
func Pack(ctx *context.Ctx) error {
	var err error

	relPaths, err := getSortedRelPaths(ctx.Pack.PackageFilesDir)
	if err != nil {
		return fmt.Errorf("Failed to get sorted package files list: %s", err)
	}

	cpioPath := filepath.Join(ctx.Cli.TmpDir, "cpio")
	if err := packCpio(relPaths, cpioPath, ctx); err != nil {
		return fmt.Errorf("Failed to pack CPIO: %s", err)
	}

	compresedCpioPath := filepath.Join(ctx.Cli.TmpDir, "cpio.gz")
	if err := common.CompressGzip(cpioPath, compresedCpioPath); err != nil {
		return fmt.Errorf("Failed to compress CPIO: %s", err)
	}

	rpmHeader, err := genRpmHeader(relPaths, cpioPath, compresedCpioPath, ctx)
	if err != nil {
		return fmt.Errorf("Failed to gen RPM header: %s", err)
	}

	packedHeader, err := packTagSet(rpmHeader, headerImmutable)
	if err != nil {
		return fmt.Errorf("Failed to pack RPM header: %s", err)
	}

	// write header to file
	rpmHeaderFilePath := filepath.Join(ctx.Cli.TmpDir, "header")
	rpmHeaderFile, err := os.Create(rpmHeaderFilePath)
	if err != nil {
		return fmt.Errorf("Failed to create RPM body file: %s", err)
	}
	defer rpmHeaderFile.Close()

	if _, err := io.Copy(rpmHeaderFile, packedHeader); err != nil {
		return fmt.Errorf("Failed to write RPM lead to file: %s", err)
	}

	// create body file = header + compressedCpio
	rpmBodyFilePath := filepath.Join(ctx.Cli.TmpDir, "body")
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
	lead := genRpmLead(ctx.Project.Name)
	if err := common.ConcatBuffers(lead, packedSignature); err != nil {
		return err
	}

	// create lead file
	leadFilePath := filepath.Join(ctx.Cli.TmpDir, "lead")
	leadFile, err := os.Create(leadFilePath)
	if err != nil {
		return fmt.Errorf("Failed to create RPM lead file: %s", err)
	}

	if _, err := io.Copy(leadFile, lead); err != nil {
		return fmt.Errorf("Failed to write RPM lead to file: %s", err)
	}

	// create RPM file
	err = common.MergeFiles(ctx.Pack.ResPackagePath,
		leadFilePath,
		rpmBodyFilePath,
	)
	if err != nil {
		return fmt.Errorf("Failed to write result RPM file: %s", err)
	}

	return nil
}
