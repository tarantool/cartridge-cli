package bench

import (
	"math/rand"
	"reflect"

	"github.com/FZambia/tarantool"
	"github.com/tarantool/cartridge-cli/cli/common"
)

// insertOperationOnConnection execute insert operation with specified connection.
func insertOperationOnConnection(tarantoolConnection *tarantool.Connection, request *Request) {
	_, err := tarantoolConnection.Exec(
		tarantool.Insert(
			benchSpaceName,
			[]interface{}{
				common.RandomString(request.ctx.KeySize),
				common.RandomString(request.ctx.DataSize),
			}))
	request.results.incrementRequestsCounters(err)
}

// insertOperation execute insert operation.
func insertOperation(request *Request) {
	insertOperationOnConnection(request.tarantoolConnection, request)
}

// clusterInsertOperation execute insert operation on cluster topology.
func clusterInsertOperation(request *Request) {
	connectionsPool := request.clusterNodesConnections.getNextConnectionsPool()
	tarantoolConnection := connectionsPool.getNextConnection()
	insertOperationOnConnection(tarantoolConnection, request)
}

// selectOperationOnConnection execute select operation with specified connection.
func selectOperationOnConnection(tarantoolConnection *tarantool.Connection, request *Request) {
	_, err := tarantoolConnection.Exec(tarantool.Call(
		getRandomTupleCommand,
		[]interface{}{rand.Int()}))
	request.results.incrementRequestsCounters(err)
}

// selectOperation execute select operation.
func selectOperation(request *Request) {
	selectOperationOnConnection(request.tarantoolConnection, request)
}

// clusterSelectOperation execute select operation on cluster topology.
func clusterSelectOperation(request *Request) {
	connectionsPool := request.clusterNodesConnections.getNextConnectionsPool()
	tarantoolConnection := connectionsPool.getNextConnection()
	selectOperationOnConnection(tarantoolConnection, request)
}

// updateOperationOnConnection execute update operation with specified connection.
func updateOperationOnConnection(tarantoolConnection *tarantool.Connection, request *Request) {
	getRandomTupleResponse, err := tarantoolConnection.Exec(
		tarantool.Call(getRandomTupleCommand,
			[]interface{}{rand.Int()}))
	if err == nil {
		data := getRandomTupleResponse.Data
		if len(data) > 0 {
			key := reflect.ValueOf(data[0]).Index(0).Elem().String()
			_, err := request.tarantoolConnection.Exec(
				tarantool.Update(
					benchSpaceName,
					benchSpacePrimaryIndexName,
					[]interface{}{key},
					[]tarantool.Op{tarantool.Op(
						tarantool.OpAssign(
							2,
							common.RandomString(request.ctx.DataSize)))}))
			request.results.incrementRequestsCounters(err)
		}
	}
}

// updateOperation execute update operation.
func updateOperation(request *Request) {
	updateOperationOnConnection(request.tarantoolConnection, request)
}

// clusterUpdateOperation execute update operation on cluster topology.
func clusterUpdateOperation(request *Request) {
	connectionsPool := request.clusterNodesConnections.getNextConnectionsPool()
	tarantoolConnection := connectionsPool.getNextConnection()
	updateOperationOnConnection(tarantoolConnection, request)
}

// getNext return next operation in operations sequence.
func (requestsSequence *RequestsSequence) getNext() *Request {
	// If at the moment the number of remaining requests = 0,
	// then find a new generator, which requests count > 0.
	// If new generator has requests count = 0, then repeat.
	requestsSequence.findNewRequestsGeneratorMutex.Lock()
	defer requestsSequence.findNewRequestsGeneratorMutex.Unlock()
	for requestsSequence.currentCounter == 0 {
		// Increase the index, which means logical switching to a new generator.
		requestsSequence.currentRequestIndex++
		requestsSequence.currentRequestIndex %= len(requestsSequence.requests)
		// Get new generator by index.
		nextRequestsGenerator := &requestsSequence.requests[requestsSequence.currentRequestIndex]
		// Get requests count for new operation.
		requestsSequence.currentCounter = nextRequestsGenerator.count
	}
	// Logical taking of a single request.
	requestsSequence.currentCounter--
	return &requestsSequence.requests[requestsSequence.currentRequestIndex].request
}
