package bench

import (
	bctx "context"
	"fmt"
	"reflect"
	"sync"

	"github.com/FZambia/tarantool"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
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

// fillBenchmarkSpace fills benchmark space with a PreFillingCount number of records
// using connectionPool for fast filling.
func fillBenchmarkSpace(ctx context.BenchCtx, connectionPool []*tarantool.Connection) (int, error) {
	var insertMutex sync.Mutex
	var waitGroup sync.WaitGroup
	filledCount := 0
	errorChan := make(chan error, ctx.Connections)
	backgroundCtx, cancel := bctx.WithCancel(bctx.Background())

	for i := 0; i < ctx.Connections; i++ {
		waitGroup.Add(1)
		go func(tarantoolConnection *tarantool.Connection) {
			defer waitGroup.Done()
			for filledCount < ctx.PreFillingCount && len(errorChan) == 0 {
				select {
				case <-backgroundCtx.Done():
					return
				default:
					// Lock mutex for checking extra iteration and increment counter.
					insertMutex.Lock()
					if filledCount == ctx.PreFillingCount {
						insertMutex.Unlock()
						return
					}
					filledCount++
					insertMutex.Unlock()
					_, err := tarantoolConnection.Exec(tarantool.Insert(
						benchSpaceName,
						[]interface{}{
							common.RandomString(ctx.KeySize),
							common.RandomString(ctx.DataSize),
						},
					))
					if err != nil {
						fmt.Println(err)
						errorChan <- err
						return
					}
				}
			}
		}(connectionPool[i])
	}

	// Goroutine for checking error in channel.
	go func() {
		for {
			select {
			case <-backgroundCtx.Done():
				return
			default:
				if len(errorChan) > 0 {
					// Stop "insert" goroutines.
					cancel()
					return
				}
			}
		}
	}()

	waitGroup.Wait()
	// Stop all goroutines.
	// If "error" goroutine stopped others "insert" goroutines, "error" goroutine stops itself.
	// If "insert" goroutine successfully completed, then need to stop "error" goroutine.
	cancel()

	// Check if we have an error.
	if len(errorChan) > 0 {
		err := <-errorChan
		close(errorChan)
		return filledCount, fmt.Errorf(
			"Error during space pre-filling: %s.",
			err.Error())
	}
	close(errorChan)

	return filledCount, nil
}
