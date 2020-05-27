package rpm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/tarantool/cartridge-cli/src/common"
)

type rpmValueType int32

type rpmTagType struct {
	ID    int
	Type  rpmValueType
	Value interface{}
}

type rpmTagSetType []rpmTagType

type packedTagType struct {
	Count int
	Data  *bytes.Buffer
}

func (tagSet *rpmTagSetType) addTags(tags ...rpmTagType) {
	*tagSet = append(*tagSet, tags...)
}

func packValues(values ...interface{}) *bytes.Buffer {
	buf := bytes.NewBuffer(nil)

	for _, v := range values {
		binary.Write(buf, binary.BigEndian, v)
	}

	return buf
}

func packTag(tag rpmTagType) (*packedTagType, error) {
	var packed packedTagType
	packed.Data = bytes.NewBuffer(nil)

	switch tag.Type {
	case rpmTypeNull: // NULL
		if tag.Value != nil {
			return nil, fmt.Errorf("NULL value should be nil")
		}

		packed.Count = 1
	case rpmTypeChar: // CHAR
		// XXX: It should be array of rune's or bytes ??
	case rpmTypeBin: // BIN
		byteArray, ok := tag.Value.([]byte)
		if !ok {
			return nil, fmt.Errorf("BIN value should be []byte")
		}

		packed.Count = len(byteArray)
		for _, byteValue := range byteArray {
			if _, err := io.Copy(packed.Data, packValues(byteValue)); err != nil {
				return nil, err
			}
		}
	case rpmTypeStringArray: // STRING_ARRAY
		// value should be strings array
		stringsArray, ok := tag.Value.([]string)
		if !ok {
			return nil, fmt.Errorf("STRING_ARRAY value should be []string")
		}

		packed.Count = len(stringsArray)

		for _, v := range stringsArray {
			bytedString := []byte(v)
			bytedString = append(bytedString, 0)
			if _, err := io.Copy(packed.Data, packValues(bytedString)); err != nil {
				return nil, err
			}
		}
	case rpmTypeString: // STRING
		// value should be string
		stringValue, ok := tag.Value.(string)
		if !ok {
			return nil, fmt.Errorf("STRING value should be string")
		}

		packed.Count = 1

		bytedString := []byte(stringValue)
		bytedString = append(bytedString, 0)
		if _, err := io.Copy(packed.Data, packValues(bytedString)); err != nil {
			return nil, err
		}

	case rpmTypeInt8: // INT8
		// value should be []int8
		int8Values, ok := tag.Value.([]int8)
		if !ok {
			return nil, fmt.Errorf("INT8 value should be []int8")
		}

		packed.Count = len(int8Values)

		if _, err := io.Copy(packed.Data, packValues(int8Values)); err != nil {
			return nil, err
		}

	case rpmTypeInt16: // INT16
		// value should be []int16
		int16Values, ok := tag.Value.([]int16)
		if !ok {
			return nil, fmt.Errorf("INT16 value should be []int16")
		}

		packed.Count = len(int16Values)

		for _, value := range int16Values {
			if _, err := io.Copy(packed.Data, packValues(value)); err != nil {
				return nil, err
			}
		}

	case rpmTypeInt32: // INT32
		// value should be []int32
		int32Values, ok := tag.Value.([]int32)
		if !ok {
			return nil, fmt.Errorf("INT32 value should be []int32")
		}

		packed.Count = len(int32Values)

		for _, value := range int32Values {
			if _, err := io.Copy(packed.Data, packValues(value)); err != nil {
				return nil, err
			}
		}

	case rpmTypeInt64: // INT64
		// value should be []int64
		int64Values, ok := tag.Value.([]int64)
		if !ok {
			return nil, fmt.Errorf("INT64 value should be []int64")
		}

		packed.Count = len(int64Values)

		for _, value := range int64Values {
			if _, err := io.Copy(packed.Data, packValues(value)); err != nil {
				return nil, err
			}
		}

	default:
		return nil, fmt.Errorf("Unknown tag type: %d", tag.Type)
	}

	return &packed, nil
}

func alignData(data *bytes.Buffer, padding int) {
	dataLen := data.Len()

	if dataLen%padding != 0 {
		alignedDataLen := (dataLen/padding + 1) * padding

		missedBytesNum := alignedDataLen - dataLen

		paddingBytes := make([]byte, missedBytesNum)
		data.Write(paddingBytes)
	}
}

func getPackedTagIndex(offset int, tagID int, tagType rpmValueType, count int) *bytes.Buffer {
	tagIndex := packValues(
		int32(tagID),
		int32(tagType),
		int32(offset),
		int32(count),
	)

	return tagIndex
}

func getTagSetHeader(tagsNum int, dataLen int) *bytes.Buffer {
	tagSetHeader := packValues(
		headerMagic[0], headerMagic[1], headerMagic[2],
		byte(versionMagic),
		int32(reservedMagic),
		int32(tagsNum),
		int32(dataLen),
	)

	return tagSetHeader
}

func packTagSet(tagSet rpmTagSetType, regionTagID int) (*bytes.Buffer, error) {
	var resData = bytes.NewBuffer(nil)
	var tagsIndex = bytes.NewBuffer(nil)
	var resIndex = bytes.NewBuffer(nil)

	// tags index
	for _, tag := range tagSet {
		packed, err := packTag(tag)

		if err != nil {
			return nil, err
		}
		if padding, ok := padByType[tag.Type]; !ok {
			return nil, fmt.Errorf("Padding for type %d is not set", tag.Type)
		} else if padding > 0 {
			alignData(resData, padding)
		}

		tagIndex := getPackedTagIndex(resData.Len(), tag.ID, tag.Type, packed.Count)

		if err := common.ConcatBuffers(resData, packed.Data); err != nil {
			return nil, err
		}

		if err := common.ConcatBuffers(tagsIndex, tagIndex); err != nil {
			return nil, err
		}
	}

	// regionTag index
	regionTagIndex := getPackedTagIndex(resData.Len(), regionTagID, rpmTypeBin, 16)

	// resIndex is regionTagIndex + tagsIndex
	if err := common.ConcatBuffers(resIndex, regionTagIndex, tagsIndex); err != nil {
		return nil, err
	}

	// regionTag data
	tagsNum := len(tagSet) + 1
	regionTagData := getPackedTagIndex(-tagsNum*16, regionTagID, rpmTypeBin, 16)

	// resData is tagsData + regionTagData
	if err := common.ConcatBuffers(resData, regionTagData); err != nil {
		return nil, err
	}

	// tagSetHeader
	tagSetHeader := getTagSetHeader(tagsNum, resData.Len())

	// res is tagSetHeader + resIndex + resData
	var res = bytes.NewBuffer(nil)
	if err := common.ConcatBuffers(res, tagSetHeader, resIndex, resData); err != nil {
		return nil, err
	}

	return res, nil
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

		if _, known := knownFiles[filePath]; !known {
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
