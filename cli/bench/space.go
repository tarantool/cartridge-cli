package bench

import (
	"fmt"
	"reflect"

	"github.com/FZambia/tarantool"
)

// createBenchmarkSpace creates benchmark space with formatting and primary index.
func createBenchmarkSpace(tarantoolConnection *tarantool.Connection) error {
	// Creating space.
	createCommand := "return box.schema.space.create(...).name"
	_, err := tarantoolConnection.Exec(tarantool.Eval(createCommand, []interface{}{benchSpaceName, map[string]bool{"if_not_exists": true}}))
	if err != nil {
		return err
	}

	// Formatting space.
	formatCommand := fmt.Sprintf("box.space.%s:format", benchSpaceName)
	_, err = tarantoolConnection.Exec(tarantool.Call(formatCommand, [][]map[string]string{
		{
			{"name": "key", "type": "string"},
			{"name": "value", "type": "string"},
		},
	}))
	if err != nil {
		return err
	}

	// Creating primary index.
	createIndexCommand := fmt.Sprintf("box.space.%s:create_index", benchSpaceName)
	_, err = tarantoolConnection.Exec(tarantool.Call(createIndexCommand, []interface{}{
		benchSpacePrimaryIndexName,
		map[string]interface{}{
			"parts":         []string{"key"},
			"if_not_exists": true,
		},
	}))
	return err
}

// dropBenchmarkSpace deletes benchmark space.
func dropBenchmarkSpace(tarantoolConnection *tarantool.Connection) error {
	checkCommand := fmt.Sprintf("return box.space.%s.index[0].name", benchSpaceName)
	indexName, err := tarantoolConnection.Exec(tarantool.Eval(checkCommand, []interface{}{}))
	if err != nil {
		return err
	}
	if reflect.ValueOf(indexName.Data).Index(0).Elem().String() == benchSpacePrimaryIndexName {
		dropCommand := fmt.Sprintf("box.space.%s:drop", benchSpaceName)
		_, err := tarantoolConnection.Exec(tarantool.Call(dropCommand, []interface{}{}))
		return err
	}
	return nil
}
