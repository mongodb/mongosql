package bench

import (
	"fmt"

	"github.com/10gen/sqlproxy/internal/testutils/data"
	"gopkg.in/mgo.v2/bson"
)

func getDatasetForBenchmark(name string) data.Dataset {
	switch name {
	case "lpad", "lpad_long", "lpad_many_column_args", "lpad_long_many_column_args",
		"lpad_long_nested":
		return lpadDataset
	case "select_thousand_simple_docs":
		return repeatDoc("items", bson.D{{"key", "value"}}, 1000)
	case "select_million_simple_docs":
		return repeatDoc("items", bson.D{{"key", "value"}}, 1000000)
	case "select_one_doc_thousand_fields":
		doc := docWithManyFields(1000)
		return repeatDoc("items", doc, 1)
	case "select_one_doc_ten_thousand_fields":
		doc := docWithManyFields(10000)
		return repeatDoc("items", doc, 1)
	default:
		panic(fmt.Errorf("no dataset for benchmark %s", name))
	}
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

func repeatDoc(collection string, doc bson.D, count int) data.DynamicDataset {
	return func() (string, map[string][]bson.D) {
		data := make([]bson.D, 0, count)
		for i := 0; i < count; i++ {
			data = append(data, doc)
		}
		return "benchmark", map[string][]bson.D{collection: data}
	}
}

func docWithManyFields(n int) bson.D {
	doc := bson.D{}
	for i := 0; i < n; i++ {
		fieldName := fmt.Sprintf("field%d", i)
		doc = append(doc, bson.DocElem{fieldName, "value"})
	}
	return doc
}
