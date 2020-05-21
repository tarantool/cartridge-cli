package rpm

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/sassoftware/go-rpmutils"
	"github.com/tarantool/cartridge-cli/src/common"
	"github.com/tarantool/cartridge-cli/src/project"
)

type filesInfoType struct {
	BaseNames      []string
	DirNames       []string
	DirIndexes     []int
	FileUserNames  []string
	FileGroupNames []string
	FileSizes      []int64
	FileModes      []os.FileMode
	FileInodes     []uint64
	FileDevices    []int32
	FileMtimes     []int64
	FileLangs      []string
	FileRdevs      []int32
	FileLinkTos    []string
	FileFlags      []uint32
	FileDigests    []string
}

func genRpmHeader(cpioPath string, projectCtx *project.ProjectCtx) ([]byte, error) {
	var err error

	// compute payload digest
	payloadDigestAlgo := rpmutils.PGPHASHALGO_SHA256
	payloadDigest, err := common.FileSHA256Hex(cpioPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to get payload digest: %s", err)
	}

	fmt.Printf("%d: %s\n", payloadDigestAlgo, payloadDigest)

	// gen fileinfo
	filesInfo, err := generateFilesInfo(projectCtx.PackageFilesDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to get files info: %s", err)
	}

	fmt.Printf("%#v\n", filesInfo)

	// err = packHeaderValues()

	return nil, nil
}

func generateFilesInfo(dirPath string) (filesInfoType, error) {
	filesInfo := filesInfoType{}

	err := filepath.Walk(dirPath, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relFilePath, err := filepath.Rel(dirPath, filePath)
		if err != nil {
			return err
		}

		// skip known files
		if _, known := knownFiles[relFilePath]; known {
			return nil
		}

		if fileInfo.Mode().IsRegular() {
			filesInfo.FileFlags = append(filesInfo.FileFlags, rpmutils.RPMFILE_NOREPLACE) // XXX

			fileDigest, err := common.FileMD5Hex(filePath)
			if err != nil {
				return fmt.Errorf("Failed to get file MD5 hex: %s", err)
			}

			filesInfo.FileDigests = append(filesInfo.FileDigests, fileDigest)
		} else {
			filesInfo.FileFlags = append(filesInfo.FileFlags, rpmutils.RPMFILE_NONE) // XXX
			filesInfo.FileDigests = append(filesInfo.FileDigests, emptyDigest)
		}

		fileDir := filepath.Dir(relFilePath)
		dirIndex := addDirAngGetIndex(&filesInfo.DirNames, fileDir)
		filesInfo.DirIndexes = append(filesInfo.DirIndexes, dirIndex)

		filesInfo.BaseNames = append(filesInfo.BaseNames, filepath.Base(filePath))
		filesInfo.FileModes = append(filesInfo.FileModes, fileInfo.Mode())
		filesInfo.FileMtimes = append(filesInfo.FileMtimes, fileInfo.ModTime().Unix())

		filesInfo.FileUserNames = append(filesInfo.FileUserNames, defaultFileUser)
		filesInfo.FileGroupNames = append(filesInfo.FileGroupNames, defaultFileGroup)
		filesInfo.FileLangs = append(filesInfo.FileLangs, defaultFileLang)
		filesInfo.FileLinkTos = append(filesInfo.FileLinkTos, defaultFileLinkTo)

		sysFileInfo, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("Failed to get file info: %s", err)
		}

		filesInfo.FileSizes = append(filesInfo.FileSizes, sysFileInfo.Size)
		filesInfo.FileInodes = append(filesInfo.FileInodes, sysFileInfo.Ino)
		filesInfo.FileDevices = append(filesInfo.FileDevices, sysFileInfo.Dev)
		filesInfo.FileRdevs = append(filesInfo.FileRdevs, sysFileInfo.Rdev)

		return nil
	})

	if err != nil {
		return filesInfo, err
	}

	return filesInfo, nil
}

func addDirAngGetIndex(dirNames *[]string, fileDir string) int {
	for i, dirName := range *dirNames {
		if dirName == fileDir {
			return i
		}
	}

	*dirNames = append(*dirNames, fileDir)
	return len(*dirNames) - 1
}
