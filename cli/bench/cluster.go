package bench

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/FZambia/tarantool"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

// verifyLeaders check each replica have leader.
func verifyReplicas(ctx context.BenchCtx) error {
	foundLeaders := make(map[string]int)
	for _, replica_url := range *ctx.Replicas {
		haveLeader := false
		tarantoolConnection, err := createConnection(ctx, replica_url)
		if err != nil {
			return err
		}
		defer tarantoolConnection.Close()
		command := "return box.info.replication"
		replication, err := tarantoolConnection.Exec(tarantool.Eval(command, []interface{}{}))
		if err != nil {
			return err
		}
		replicationValue := reflect.ValueOf(replication.Data).Index(0).Elem()
		replicationIterator := replicationValue.MapRange()
		for replicationIterator.Next() {
			//fmt.Println(replicationIterator.Key(), ": ", replicationIterator.Value())
			upstream := replicationIterator.Value().Elem().MapIndex(reflect.ValueOf("upstream"))
			if upstream.IsValid() && !upstream.IsZero() {
				//fmt.Println(upstream)
				peer := upstream.Elem().MapIndex(reflect.ValueOf("peer"))
				if peer.IsValid() && !peer.IsZero() {
					//fmt.Println(peer)
					if common.StringSliceContains(*ctx.Leaders, peer.Elem().String()) {
						haveLeader = true
						foundLeaders[peer.Elem().String()] += 1
					} else {
						return fmt.Errorf("Replica has a leader outside the cluster")
					}
				}
			}
		}
		if !haveLeader {
			return fmt.Errorf("Replica has no leader")
		}
	}

	if reflect.DeepEqual(reflect.ValueOf(foundLeaders).MapKeys(), *&ctx.Leaders) {
		return fmt.Errorf("There are extra leaders")
	}
	return nil
}

// verifyLeaders check each leader have replica.
func verifyLeaders(ctx context.BenchCtx) error {
	for _, leader_url := range *ctx.Leaders {
		haveReplica := false
		tarantoolConnection, err := createConnection(ctx, leader_url)
		defer tarantoolConnection.Close()
		if err != nil {
			return err
		}
		command := "return box.info.replication"
		replication, err := tarantoolConnection.Exec(tarantool.Eval(command, []interface{}{}))
		if err != nil {
			return err
		}
		replicationValue := reflect.ValueOf(replication.Data).Index(0).Elem()
		replicationIterator := replicationValue.MapRange()
		for replicationIterator.Next() {
			downstream := replicationIterator.Value().Elem().MapIndex(reflect.ValueOf("downstream"))
			if downstream.IsValid() && !downstream.IsZero() {
				haveReplica = true
			}
		}
		if !haveReplica {
			return fmt.Errorf("Leader has no replica")
		}
	}
	return nil
}

// verifyClusterTopology check cluster for wrong topology.
func verifyClusterTopology(ctx context.BenchCtx) error {
	if err := verifyReplicas(ctx); err != nil {
		return err
	}
	if err := verifyLeaders(ctx); err != nil {
		return err
	}
	return nil
}

// createNodesConnectionsPools creates connections pool for every node in cluster.
func createNodesConnectionsPools(ctx context.BenchCtx) (map[string][]RotaryConnectionsPool, error) {
	nodesConnectionsPools := make(map[string][]RotaryConnectionsPool)
	nodesConnectionsPools["leaders"] = make([]RotaryConnectionsPool, len(*ctx.Leaders))
	nodesConnectionsPools["replicas"] = make([]RotaryConnectionsPool, len(*ctx.Replicas))

	for i, leader_url := range *ctx.Leaders {
		connectionsPool, err := createConnectionsPool(ctx, leader_url)
		if err != nil {
			return nil, err
		}
		nodesConnectionsPools["leaders"][i] = RotaryConnectionsPool{
			connectionsPool: connectionsPool,
		}
	}

	for i, replica_url := range *ctx.Replicas {
		connectionsPool, err := createConnectionsPool(ctx, replica_url)
		if err != nil {
			return nil, err
		}
		nodesConnectionsPools["replicas"][i] = RotaryConnectionsPool{
			connectionsPool: connectionsPool,
		}
	}

	return nodesConnectionsPools, nil
}

// deleteNodesConnectionsPools delete all connections pools.
func deleteNodesConnectionsPools(nodesConnectionsPools map[string][]RotaryConnectionsPool) {
	for _, connectionsPools := range nodesConnectionsPools {
		for i := range connectionsPools {
			deleteConnectionsPool(connectionsPools[i].connectionsPool)
		}
	}
}

// getNextConnection retrun next connection of node connections pool.
func (rotaryConnectionsPool *RotaryConnectionsPool) getNextConnection() *tarantool.Connection {
	rotaryConnectionsPool.mutex.Lock()
	returnConnection := rotaryConnectionsPool.connectionsPool[rotaryConnectionsPool.currentIndex]
	rotaryConnectionsPool.currentIndex++
	rotaryConnectionsPool.currentIndex %= len(rotaryConnectionsPool.connectionsPool)
	rotaryConnectionsPool.mutex.Unlock()
	return returnConnection
}

// getNextConnectionsPool return next node represented by connections pool.
func (rotaryNodesConnectionsPools *RotaryNodesConnectionsPools) getNextConnectionsPool() *RotaryConnectionsPool {
	rotaryNodesConnectionsPools.mutex.Lock()
	returnConnectionsPool := &rotaryNodesConnectionsPools.rotaryConnectionsPool[rotaryNodesConnectionsPools.currentIndex]
	rotaryNodesConnectionsPools.currentIndex++
	rotaryNodesConnectionsPools.currentIndex %= len(rotaryNodesConnectionsPools.rotaryConnectionsPool)
	rotaryNodesConnectionsPools.mutex.Unlock()
	return returnConnectionsPool
}

// benchCluster execute bench algorithm for cluster.
func benchCluster(ctx context.BenchCtx, benchData *BenchmarkData) error {
	// Сreate connections pools for all nodes in cluster before starting the benchmark
	// to exclude the connection establishment time from measurements.
	nodesConnectionsPools, err := createNodesConnectionsPools(ctx)
	if err != nil {
		return err
	}
	defer deleteNodesConnectionsPools(nodesConnectionsPools)

	mutationConnections := nodesConnectionsPools["leaders"]
	selectConnections := nodesConnectionsPools["replicas"]

	if len(*ctx.Replicas) == 0 {
		selectConnections = append(selectConnections, mutationConnections...)
	}

	benchData.startTime = time.Now()

	// Start detached connections.
	for i := 0; i < ctx.Connections; i++ {
		benchData.waitGroup.Add(1)
		go func() {
			defer benchData.waitGroup.Done()
			requestsSequence := RequestsSequence{
				requests: []RequestsGenerator{
					{
						request: Request{
							operation:               clusterInsertOperation,
							ctx:                     ctx,
							clusterNodesConnections: RotaryNodesConnectionsPools{rotaryConnectionsPool: mutationConnections},
							results:                 &benchData.results,
						},
						count: ctx.InsertCount,
					},
					{
						request: Request{
							operation:               clusterSelectOperation,
							ctx:                     ctx,
							clusterNodesConnections: RotaryNodesConnectionsPools{rotaryConnectionsPool: selectConnections},
							results:                 &benchData.results,
						},
						count: ctx.SelectCount,
					},
					{
						request: Request{
							operation:               clusterUpdateOperation,
							ctx:                     ctx,
							clusterNodesConnections: RotaryNodesConnectionsPools{rotaryConnectionsPool: mutationConnections},
							results:                 &benchData.results,
						},
						count: ctx.UpdateCount,
					},
				},
				currentRequestIndex:           0,
				currentCounter:                ctx.InsertCount,
				findNewRequestsGeneratorMutex: sync.Mutex{},
			}

			// Start looped requests in connection.
			var requestsWait sync.WaitGroup
			for i := 0; i < ctx.SimultaneousRequests; i++ {
				requestsWait.Add(1)
				go func() {
					defer requestsWait.Done()
					requestsLoop(&requestsSequence, benchData.backgroundCtx)
				}()
			}
			requestsWait.Wait()
		}()
	}

	waitBenchEnd(benchData)
	return nil
}
