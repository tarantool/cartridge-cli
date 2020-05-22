package rpm

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
	Data  []byte
}

func (tagSet *rpmTagSetType) addTags(tags ...rpmTagType) {
	*tagSet = append(*tagSet, tags...)
}

func packValues(values ...interface{}) []byte {
	buf := &bytes.Buffer{}

	for _, v := range values {
		binary.Write(buf, binary.BigEndian, v)
	}

	return buf.Bytes()
}

func packTag(tag rpmTagType) (*packedTagType, error) {
	var packed packedTagType

	switch tag.Type {
	case rpmTypeNull: // NULL
		// XXX: It should be array of nil's ??
	case rpmTypeChar: // CHAR
		// XXX: It should be array of rune's or bytes ??
	case rpmTypeBin: // BIN
		boolValue, ok := tag.Value.(bool)
		if !ok {
			return nil, fmt.Errorf("BIN value should be bool")
		}

		packed.Count = 1
		packed.Data = packValues(boolValue)
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
			packed.Data = append(packed.Data, packValues(bytedString)...)
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
		packed.Data = packValues(bytedString)

	case rpmTypeInt8: // INT8
		// value should be []int8
		int8Values, ok := tag.Value.([]int8)
		if !ok {
			return nil, fmt.Errorf("INT8 value should be []int8")
		}

		packed.Count = len(int8Values)

		for _, value := range int8Values {
			packed.Data = append(packed.Data, packValues(value)...)
		}

	case rpmTypeInt16: // INT16
		// value should be []int16
		int16Values, ok := tag.Value.([]int16)
		if !ok {
			return nil, fmt.Errorf("INT16 value should be []int16")
		}

		packed.Count = len(int16Values)

		for _, value := range int16Values {
			packed.Data = append(packed.Data, packValues(value)...)
		}

	case rpmTypeInt32: // INT32
		// value should be []int32
		int32Values, ok := tag.Value.([]int32)
		if !ok {
			return nil, fmt.Errorf("INT32 value should be []int32")
		}

		packed.Count = len(int32Values)

		for _, value := range int32Values {
			packed.Data = append(packed.Data, packValues(value)...)
		}

	case rpmTypeInt64: // INT64
		// value should be []int64
		int64Values, ok := tag.Value.([]int64)
		if !ok {
			return nil, fmt.Errorf("INT64 value should be []int64")
		}

		packed.Count = len(int64Values)

		for _, value := range int64Values {
			packed.Data = append(packed.Data, packValues(value)...)
		}

	default:
		return nil, fmt.Errorf("Unknown tag type: %d", tag.Type)
	}

	return &packed, nil
}

func alignData(data *[]byte, padding int) {
	dataLen := len(*data)

	if dataLen%padding != 0 {
		alignedDataLen := (dataLen/padding + 1) * padding

		missedBytesNum := alignedDataLen - dataLen

		paddingBytes := make([]byte, missedBytesNum)
		*data = append(*data, paddingBytes...)
	}
}

func packTagSet(tagSet rpmTagSetType) ([]byte, error) {
	var resData []byte

	for _, tag := range tagSet {
		fmt.Printf("tag.ID: %d\n", tag.ID)

		fmt.Printf("tag.Value: %v\n", tag.Value)
		fmt.Printf("tag.Type: %d\n", tag.Type)

		packed, err := packTag(tag)

		if err != nil {
			return nil, err
		}

		fmt.Printf("packed.Data: %x\n", packed.Data)
		fmt.Printf("packed.Count: %d\n", packed.Count)
		fmt.Println()

		if padding, ok := padByType[tag.Type]; !ok {
			return nil, fmt.Errorf("Padding for type %d is not set", tag.Type)
		} else if padding > 0 {
			alignData(&resData, padding)
		}

		resData = append(resData, packed.Data...)
	}

	return resData, nil
}
