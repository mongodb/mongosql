package bench

import (
	"github.com/10gen/sqlproxy/internal/testutils/data"
	"gopkg.in/mgo.v2/bson"
)

var datasetsByName = map[string]data.Dataset{
	"tpch_normalized":   tpchNormalized,
	"tpch_denormalized": tpchDenormalized,
	"tpch_micro":        tpchMicro,

	"lpad_strings": lpadDataset,
}

var (
	tpchNormalized   = data.NewBSONDataset("tpch_full_normalized")
	tpchDenormalized = data.NewBSONDataset("tpch_full_denormalized")
	tpchMicro        = data.NewBSONDataset("tpch_small")
)

var lpadDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 1000
	data := []bson.D{}
	for i := 0; i < numDocs; i++ {
		data = append(data, bson.D{{"s", "abcde"}})
	}
	return "benchmark", map[string][]bson.D{"strings": data}
}
