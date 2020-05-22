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
	DirIndexes     []int32
	FileUserNames  []string
	FileGroupNames []string
	FileSizes      []int32
	FileModes      []int16
	FileInodes     []int32
	FileDevices    []int32
	FileMtimes     []int32
	FileLangs      []string
	FileRdevs      []int16
	FileLinkTos    []string
	FileFlags      []int32
	FileDigests    []string
}

func genRpmHeader(cpioPath, compresedCpioPath string, projectCtx *project.ProjectCtx) (rpmTagSetType, error) {
	// var err error

	rmpHeader := rpmTagSetType{}

	// compute payload digest
	payloadDigestAlgo := rpmutils.PGPHASHALGO_SHA256
	// payloadDigest, err := common.FileSHA256Hex(compresedCpioPath)
	// if err != nil {
	// 	return nil, fmt.Errorf("Failed to get payload digest: %s", err)
	// }

	cpioFileInfo, err := os.Stat(cpioPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to get payload size: %s", err)
	}
	payloadSize := cpioFileInfo.Size()

	// gen fileinfo
	filesInfo, err := getFilesInfo(projectCtx.PackageFilesDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to get files info: %s", err)
	}

	rmpHeader.addTags([]rpmTagType{
		{ID: tagName, Type: rpmTypeString, Value: projectCtx.Name},
		{ID: tagVersion, Type: rpmTypeString, Value: projectCtx.Version},
		{ID: tagRelease, Type: rpmTypeString, Value: projectCtx.Release},
		{ID: tagSummary, Type: rpmTypeString, Value: ""},
		{ID: tagDescription, Type: rpmTypeString, Value: ""},

		{ID: tagLicense, Type: rpmTypeString, Value: "N/A"},
		{ID: tagGroup, Type: rpmTypeString, Value: "None"},
		{ID: tagOs, Type: rpmTypeString, Value: "linux"},
		{ID: tagArch, Type: rpmTypeString, Value: "x86_64"},

		{ID: tagPayloadFormat, Type: rpmTypeString, Value: "cpio"},
		{ID: tagPayloadCompressor, Type: rpmTypeString, Value: "gzip"},
		{ID: tagPayloadFlags, Type: rpmTypeString, Value: "5"},

		{ID: tagPrein, Type: rpmTypeString, Value: ""},            // XXX
		{ID: tagPreinProg, Type: rpmTypeString, Value: "/bin/sh"}, // XXX

		{ID: tagDirNames, Type: rpmTypeStringArray, Value: filesInfo.DirNames},
		{ID: tagBaseNames, Type: rpmTypeStringArray, Value: filesInfo.BaseNames},
		{ID: tagDirIndexes, Type: rpmTypeInt32, Value: filesInfo.DirIndexes},

		{ID: tagFileUsernames, Type: rpmTypeStringArray, Value: filesInfo.FileUserNames},
		{ID: tagFileGroupnames, Type: rpmTypeStringArray, Value: filesInfo.FileGroupNames},
		{ID: tagFileSizes, Type: rpmTypeInt32, Value: filesInfo.FileSizes},
		{ID: tagFileModes, Type: rpmTypeInt16, Value: filesInfo.FileModes},
		// {ID: tagFileInodes, Type: rpmTypeInt32, Value: filesInfo.FileInodes},
		{ID: tagFileDevices, Type: rpmTypeInt32, Value: filesInfo.FileDevices},
		{ID: tagFileRdevs, Type: rpmTypeInt16, Value: filesInfo.FileRdevs},
		// {ID: tagFileMtimes, Type: rpmTypeInt16, Value: filesInfo.FileMtimes},
		{ID: tagFileFlags, Type: rpmTypeInt32, Value: filesInfo.FileFlags},
		{ID: tagFileLangs, Type: rpmTypeStringArray, Value: filesInfo.FileLangs},
		{ID: tagFileDigests, Type: rpmTypeStringArray, Value: filesInfo.FileDigests},
		{ID: tagFileLinkTos, Type: rpmTypeStringArray, Value: filesInfo.FileLinkTos},

		{ID: tagSize, Type: rpmTypeInt32, Value: []int32{int32(payloadSize)}},
		// {ID: tagPayloadDigest, Type: rpmTypeStringArray, Value: []string{payloadDigest}},
		{ID: tagPayloadDigestAlgo, Type: rpmTypeInt32, Value: []int32{int32(payloadDigestAlgo)}},
	}...)

	// XXX: add tarantool dependency tags

	return rmpHeader, nil
}

func getFilesInfo(dirPath string) (filesInfoType, error) {
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
		fileDir = fmt.Sprintf("/%s/", fileDir)
		dirIndex := addDirAngGetIndex(&filesInfo.DirNames, fileDir)
		filesInfo.DirIndexes = append(filesInfo.DirIndexes, int32(dirIndex))

		filesInfo.BaseNames = append(filesInfo.BaseNames, filepath.Base(filePath))
		filesInfo.FileMtimes = append(filesInfo.FileMtimes, int32(fileInfo.ModTime().Unix()))

		filesInfo.FileUserNames = append(filesInfo.FileUserNames, defaultFileUser)
		filesInfo.FileGroupNames = append(filesInfo.FileGroupNames, defaultFileGroup)
		filesInfo.FileLangs = append(filesInfo.FileLangs, defaultFileLang)
		filesInfo.FileLinkTos = append(filesInfo.FileLinkTos, defaultFileLinkTo)

		sysFileInfo, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("Failed to get file info: %s", err)
		}

		filesInfo.FileSizes = append(filesInfo.FileSizes, int32(sysFileInfo.Size))
		filesInfo.FileModes = append(filesInfo.FileModes, int16(sysFileInfo.Mode))
		filesInfo.FileInodes = append(filesInfo.FileInodes, int32(sysFileInfo.Ino))
		filesInfo.FileDevices = append(filesInfo.FileDevices, sysFileInfo.Dev)
		filesInfo.FileRdevs = append(filesInfo.FileRdevs, int16(sysFileInfo.Rdev))

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
