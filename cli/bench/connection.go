package bench

import (
	"fmt"

	"github.com/FZambia/tarantool"
	"github.com/tarantool/cartridge-cli/cli/context"
)

// createConnection creates connection to tarantool,
// using specified url if necessary.
func createConnection(ctx context.BenchCtx, url_optional ...string) (*tarantool.Connection, error) {
	connect_url := ctx.URL
	if len(url_optional) > 1 {
		return nil, fmt.Errorf("Otpional url is more than one")
	}
	if len(url_optional) == 1 {
		connect_url = url_optional[0]
	}
	tarantoolConnection, err := tarantool.Connect(connect_url, tarantool.Opts{
		User:     ctx.User,
		Password: ctx.Password,
	})
	if err != nil {
		return nil, fmt.Errorf(
			"Couldn't connect to Tarantool %s.",
			connect_url)
	}
	return tarantoolConnection, nil
}

// createConnection creates connections pool to tarantool,
// using specified url if necessary.
func createConnectionsPool(ctx context.BenchCtx, url_optional ...string) ([]*tarantool.Connection, error) {
	connectionPool := make([]*tarantool.Connection, ctx.Connections)
	var err error
	for i := 0; i < ctx.Connections; i++ {
		connectionPool[i], err = createConnection(ctx, url_optional...)
		if err != nil {
			return nil, err
		}
	}
	return connectionPool, nil
}

// deleteConnectionsPool delete connection pool.
func deleteConnectionsPool(connectionPool []*tarantool.Connection) {
	for _, connection := range connectionPool {
		connection.Close()
	}
}
