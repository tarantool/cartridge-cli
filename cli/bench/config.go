package bench

import (
	"fmt"
)

var (
	benchSpaceName             = "__benchmark_space__"
	benchSpacePrimaryIndexName = "__bench_primary_key__"
	PreFillingCount            = 1000000
	getRandomTupleCommand      = fmt.Sprintf(
		"box.space.%s.index.%s:random",
		benchSpaceName,
		benchSpacePrimaryIndexName,
	)
)
