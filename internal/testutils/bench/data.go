package bench

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/internal/testutils/data"
	"gopkg.in/mgo.v2/bson"
)

var tpchAlterations = []string{
	"alter table partsupp change `_id.ps_partkey` `ps_partkey`",
	"alter table partsupp change `_id.ps_suppkey` `ps_suppkey`",
	"alter table part change `_id` `p_partkey`",
	"alter table supplier change `_id` `s_suppkey`",
	"alter table nation change `_id` `n_nationkey`",
	"alter table region change `_id` `r_regionkey`",
	"alter table lineitem change `_id.l_orderkey` `l_orderkey`",
	"alter table lineitem change `_id.l_linenumber` `l_linenumber`",
	"alter table orders change `_id` `o_orderkey`",
	"alter table customer change `_id` `c_custkey`",
	"alter table mongo_orders_o_lineitems rename to mongo_lineitem",
	"alter table mongo_lineitem change " +
		"`o_lineitems.l_comment` `l_comment`, " +
		"change `o_lineitems.l_commitdate` `l_commitdate`, " +
		"change `o_lineitems.l_discount` `l_discount`, " +
		"change `o_lineitems.l_extendedprice` `l_extendedprice`, " +
		"change `o_lineitems.l_linenumber` `l_linenumber`, " +
		"change `o_lineitems.l_linestatus` `l_linestatus`, " +
		"change `o_lineitems.l_p_brand` `l_p_brand`, " +
		"change `o_lineitems.l_p_container` `l_p_container`, " +
		"change `o_lineitems.l_p_name` `l_p_name`, " +
		"change `o_lineitems.l_p_size` `l_p_size`, " +
		"change `o_lineitems.l_p_type` `l_p_type`, " +
		"change `o_lineitems.l_partkey` `l_partkey`, " +
		"change `o_lineitems.l_quantity` `l_quantity`, " +
		"change `o_lineitems.l_receiptdate` `l_receiptdate`, " +
		"change `o_lineitems.l_returnflag` `l_returnflag`, " +
		"change `o_lineitems.l_s_n_name` `l_s_n_name`, " +
		"change `o_lineitems.l_s_name` `l_s_name`, " +
		"change `o_lineitems.l_s_nationkey` `l_s_nationkey`, " +
		"change `o_lineitems.l_s_r_name` `l_s_r_name`, " +
		"change `o_lineitems.l_s_regionkey` `l_s_regionkey`, " +
		"change `o_lineitems.l_shipdate` `l_shipdate`, " +
		"change `o_lineitems.l_shipinstruct` `l_shipinstruct`, " +
		"change `o_lineitems.l_shipmode` `l_shipmode`, " +
		"change `o_lineitems.l_suppkey` `l_suppkey`, " +
		"change `o_lineitems.l_tax` `l_tax`",
	"alter table mongo_part_p_suppliers rename to mongo_partsupp",
	"alter table mongo_partsupp " +
		"change `p_suppliers.ps_availqty` `ps_availqty`, " +
		"change `p_suppliers.ps_comment` `ps_comment`, " +
		"change `p_suppliers.ps_s_n_name` `ps_s_n_name`, " +
		"change `p_suppliers.ps_s_name` `ps_s_name`, " +
		"change `p_suppliers.ps_s_nationkey` `ps_s_nationkey`, " +
		"change `p_suppliers.ps_s_r_name` `ps_s_r_name`, " +
		"change `p_suppliers.ps_s_regionkey` `ps_s_regionkey`, " +
		"change `p_suppliers.ps_suppkey` `ps_suppkey`, " +
		"change `p_suppliers.ps_supplycost` `ps_supplycost`",
	"alter table mongo_supplier change `_id` `s_suppkey`",
	"alter table mongo_partsupp change `_id` `ps_partkey`",
	"alter table mongo_part change `_id` `p_partkey`",
	"alter table mongo_orders change `_id` `o_orderkey`",
	"alter table mongo_lineitem change `_id` `l_orderkey`",
	"alter table mongo_customer change `_id` `c_custkey`",
}

func getDatasetForBenchmark(name string) data.Dataset {
	if strings.Contains(name, "tpch_micro") {
		return data.WithAlterations(
			data.Resample(tpchMicro),
			"tpch", tpchAlterations...,
		)
	}

	if strings.Contains(name, "tpch_full") {
		return data.WithAlterations(
			data.Resample(tpchFull),
			"tpch", tpchAlterations...,
		)
	}

	if strings.Contains(name, "tableau_") {
		return data.Resample(tableauDataset)
	}

	if strings.Contains(name, "simple_lpad") {
		return data.Resample(lpadDataset)
	}

	switch name {
	case "overhead_select_thousand_simple_docs":
		return repeatDoc("items", bson.D{{"key", "value"}}, 1000)
	case "overhead_select_million_simple_docs":
		return repeatDoc("items", bson.D{{"key", "value"}}, 1000000)
	case "overhead_select_one_doc_thousand_fields":
		doc := docWithManyFields(1000)
		return repeatDoc("items", doc, 1)
	case "overhead_select_one_doc_ten_thousand_fields":
		doc := docWithManyFields(10000)
		return repeatDoc("items", doc, 1)
	case "overhead_select_thousand_docs_ten_fields":
		doc := docWithManyFields(10)
		return repeatDoc("items", doc, 1000)
	case "overhead_select_thousand_docs_hundred_fields":
		doc := docWithManyFields(100)
		return repeatDoc("items", doc, 1000)
	case "overhead_select_ten_docs_ten_fields":
		doc := docWithManyFields(10)
		return repeatDoc("items", doc, 10)
	case "overhead_select_hundred_docs_ten_fields":
		doc := docWithManyFields(10)
		return repeatDoc("items", doc, 100)
	case "overhead_select_ten_thousand_docs_ten_fields":
		doc := docWithManyFields(10)
		return repeatDoc("items", doc, 10000)
	case "simple_complex_predicate_expr":
		doc := bson.D{
			{"a", "value"},
			{"b", "value"},
		}
		return repeatDoc("items", doc, 1000)
	default:
		panic(fmt.Errorf("no dataset for benchmark %s", name))
	}
}

var (
	tpchNormalized   = data.Once(data.NewBSONDataset("tpch_full_normalized"))
	tpchDenormalized = data.Once(data.NewBSONDataset("tpch_full_denormalized"))

	tpchMicro = data.Once(data.NewBSONDataset("tpch_small"))
	tpchFull  = data.DatasetGroup{tpchNormalized, tpchDenormalized}
)

var tableauDataset = data.Once(
	data.DatasetGroup{
		data.NewBSONDataset("attendees"),
		data.NewBSONDataset("flights201406"),
	},
)

var lpadDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 1000
	data := []bson.D{}
	for i := 0; i < numDocs; i++ {
		data = append(data, bson.D{{"s", "abcde"}})
	}
	return "benchmark", map[string][]bson.D{"strings": data}
}

func repeatDoc(collection string, doc bson.D, count int) data.Dataset {
	dynData := data.DynamicDataset(func() (string, map[string][]bson.D) {
		data := make([]bson.D, 0, count)
		for i := 0; i < count; i++ {
			data = append(data, doc)
		}
		return "benchmark", map[string][]bson.D{collection: data}
	})
	return data.Resample(dynData)
}

func docWithManyFields(n int) bson.D {
	doc := bson.D{}
	for i := 0; i < n; i++ {
		fieldName := fmt.Sprintf("field%d", i)
		doc = append(doc, bson.DocElem{fieldName, "value"})
	}
	return doc
}
