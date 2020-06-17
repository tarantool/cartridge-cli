package rpm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"
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

func genRpmHeader(relPaths []string, cpioPath, compresedCpioPath string, projectCtx *project.ProjectCtx) (rpmTagSetType, error) {
	rmpHeader := rpmTagSetType{}

	// compute payload digest
	payloadDigestAlgo := hashAlgoSHA256
	payloadDigest, err := common.FileSHA256Hex(compresedCpioPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to get payload digest: %s", err)
	}

	cpioFileInfo, err := os.Stat(cpioPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to get payload size: %s", err)
	}
	payloadSize := cpioFileInfo.Size()

	// gen fileinfo
	filesInfo, err := getFilesInfo(relPaths, projectCtx.PackageFilesDir)
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

		{ID: tagPrein, Type: rpmTypeString, Value: project.PreInstScriptContent},
		{ID: tagPreinProg, Type: rpmTypeString, Value: "/bin/sh"},

		{ID: tagDirNames, Type: rpmTypeStringArray, Value: filesInfo.DirNames},
		{ID: tagBaseNames, Type: rpmTypeStringArray, Value: filesInfo.BaseNames},
		{ID: tagDirIndexes, Type: rpmTypeInt32, Value: filesInfo.DirIndexes},

		{ID: tagFileUsernames, Type: rpmTypeStringArray, Value: filesInfo.FileUserNames},
		{ID: tagFileGroupnames, Type: rpmTypeStringArray, Value: filesInfo.FileGroupNames},
		{ID: tagFileSizes, Type: rpmTypeInt32, Value: filesInfo.FileSizes},
		{ID: tagFileModes, Type: rpmTypeInt16, Value: filesInfo.FileModes},
		{ID: tagFileInodes, Type: rpmTypeInt32, Value: filesInfo.FileInodes},
		{ID: tagFileDevices, Type: rpmTypeInt32, Value: filesInfo.FileDevices},
		{ID: tagFileRdevs, Type: rpmTypeInt16, Value: filesInfo.FileRdevs},
		{ID: tagFileMtimes, Type: rpmTypeInt32, Value: filesInfo.FileMtimes},
		{ID: tagFileFlags, Type: rpmTypeInt32, Value: filesInfo.FileFlags},
		{ID: tagFileLangs, Type: rpmTypeStringArray, Value: filesInfo.FileLangs},
		{ID: tagFileDigests, Type: rpmTypeStringArray, Value: filesInfo.FileDigests},
		{ID: tagFileLinkTos, Type: rpmTypeStringArray, Value: filesInfo.FileLinkTos},

		{ID: tagSize, Type: rpmTypeInt32, Value: []int32{int32(payloadSize)}},
		{ID: tagPayloadDigest, Type: rpmTypeStringArray, Value: []string{payloadDigest}},
		{ID: tagPayloadDigestAlgo, Type: rpmTypeInt32, Value: []int32{int32(payloadDigestAlgo)}},
	}...)

	if !projectCtx.TarantoolIsEnterprise {
		// add Tarantool dependency
		tarantoolDepName := "tarantool"
		flagGreaterOrEqual := int32(rpmSenseGreater | rpmSenseEqual)

		minVersion := strings.SplitN(projectCtx.TarantoolVersion, "-", 2)[0]
		maxVersion, err := common.GetNextMajorVersion(minVersion)
		if err != nil {
			return nil, fmt.Errorf("Failed to get next major version of Tarantool %s", err)
		}

		rmpHeader.addTags([]rpmTagType{
			{ID: tagRequireName, Type: rpmTypeStringArray,
				Value: []string{tarantoolDepName, tarantoolDepName}},
			{ID: tagRequireFlags, Type: rpmTypeInt32,
				Value: []int32{flagGreaterOrEqual, int32(rpmSenseLess)}},
			{ID: tagRequireVersion, Type: rpmTypeStringArray,
				Value: []string{minVersion, maxVersion}},
		}...)
	}

	return rmpHeader, nil
}

func getFilesInfo(relPaths []string, dirPath string) (filesInfoType, error) {
	filesInfo := filesInfoType{}

	for _, relPath := range relPaths {
		fullFilePath := filepath.Join(dirPath, relPath)
		fileInfo, err := os.Stat(fullFilePath)
		if err != nil {
			return filesInfo, err
		}

		if fileInfo.Mode().IsRegular() {
			filesInfo.FileFlags = append(filesInfo.FileFlags, fileFlag) // XXX

			fileDigest, err := common.FileMD5Hex(fullFilePath)
			if err != nil {
				return filesInfo, fmt.Errorf("Failed to get file MD5 hex: %s", err)
			}

			filesInfo.FileDigests = append(filesInfo.FileDigests, fileDigest)
		} else {
			filesInfo.FileFlags = append(filesInfo.FileFlags, dirFlag) // XXX
			filesInfo.FileDigests = append(filesInfo.FileDigests, emptyDigest)
		}

		fileDir := filepath.Dir(relPath)
		fileDir = fmt.Sprintf("/%s/", fileDir)
		dirIndex := addDirAndGetIndex(&filesInfo.DirNames, fileDir)
		filesInfo.DirIndexes = append(filesInfo.DirIndexes, int32(dirIndex))

		filesInfo.BaseNames = append(filesInfo.BaseNames, filepath.Base(relPath))
		filesInfo.FileMtimes = append(filesInfo.FileMtimes, int32(fileInfo.ModTime().Unix()))

		filesInfo.FileUserNames = append(filesInfo.FileUserNames, defaultFileUser)
		filesInfo.FileGroupNames = append(filesInfo.FileGroupNames, defaultFileGroup)
		filesInfo.FileLangs = append(filesInfo.FileLangs, defaultFileLang)
		filesInfo.FileLinkTos = append(filesInfo.FileLinkTos, defaultFileLinkTo)

		sysFileInfo, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return filesInfo, fmt.Errorf("Failed to get file info")
		}

		filesInfo.FileSizes = append(filesInfo.FileSizes, int32(sysFileInfo.Size))
		filesInfo.FileModes = append(filesInfo.FileModes, int16(sysFileInfo.Mode))
		filesInfo.FileInodes = append(filesInfo.FileInodes, int32(sysFileInfo.Ino))
		filesInfo.FileDevices = append(filesInfo.FileDevices, int32(sysFileInfo.Dev))
		filesInfo.FileRdevs = append(filesInfo.FileRdevs, int16(sysFileInfo.Rdev))
	}

	return filesInfo, nil
}

func addDirAndGetIndex(dirNames *[]string, fileDir string) int {
	for i, dirName := range *dirNames {
		if dirName == fileDir {
			return i
		}
	}

	*dirNames = append(*dirNames, fileDir)
	return len(*dirNames) - 1
}
