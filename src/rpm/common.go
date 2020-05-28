package rpm

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"sort"
)

func packValues(values ...interface{}) *bytes.Buffer {
	buf := bytes.NewBuffer(nil)

	for _, v := range values {
		binary.Write(buf, binary.BigEndian, v)
	}

	return buf
}

func alignData(data *bytes.Buffer, boundaries int) {
	dataLen := data.Len()

	if dataLen%boundaries != 0 {
		alignedDataLen := (dataLen/boundaries + 1) * boundaries

		missedBytesNum := alignedDataLen - dataLen

		paddingBytes := make([]byte, missedBytesNum)
		data.Write(paddingBytes)
	}
}

func getSortedRelPaths(srcDir string) ([]string, error) {
	var files []string

	err := filepath.Walk(srcDir, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		filePath, err = filepath.Rel(srcDir, filePath)
		if err != nil {
			return err
		}

		// system dirs shouldn't be added to the paths list
		if _, isSystem := systemDirs[filePath]; !isSystem {
			files = append(files, filePath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Strings(files)
	return files, nil
}
