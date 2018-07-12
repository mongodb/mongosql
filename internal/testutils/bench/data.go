package bench

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/internal/testutils/data"
	"github.com/mongodb/mongo-tools/common/json"
	"gopkg.in/mgo.v2/bson"
)

var complexDocumentsDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 1000
	data := []bson.D{}
	for i := 0; i < numDocs; i++ {
		data = append(data, bson.D{
			{Name: "a", Value: bson.D{
				{Name: "b", Value: bson.D{{
					Name: "c", Value: rand.Int31n(15),
				}}},
			}},
			{Name: "b", Value: bson.D{
				{Name: "a", Value: rand.Int31n(15)}},
			},
			{Name: "c", Value: rand.Int31n(15)},
			{Name: "d", Value: docsArrayWithNesting(2, 5)},
		})
	}
	return "benchmark", map[string][]bson.D{
		"complex_docs": data,
	}
}

var complexDocumentsIndexes = []bson.D{
	{
		{Name: "key", Value: bson.D{
			{Name: "c", Value: 1}},
		},
		{Name: "name", Value: "c_index"},
	},
	{
		{Name: "key", Value: bson.D{
			{Name: "d.b", Value: 1}},
		},
		{Name: "name", Value: "d_b_index"},
	},
	{
		{Name: "key", Value: bson.D{
			{Name: "d.a.a", Value: 1}},
		},
		{Name: "name", Value: "d_a_a_index"},
	},
}

var dateDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 25000
	dates := []json.Date{}
	for i := 0; i < 20; i++ {
		dates = append(dates, json.Date(time.Now().Unix()))
	}
	data := []bson.D{}
	for i := 0; i < numDocs; i++ {
		data = append(data, bson.D{
			{Name: "random", Value: "abc"},
			{Name: "date_field", Value: dates[i%20]},
		})
	}
	return "benchmark", map[string][]bson.D{
		"foo_date": data,
	}
}

var filterDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 1000
	data := []bson.D{}
	for i := 0; i < numDocs; i++ {
		data = append(data, bson.D{
			{Name: "a", Value: "abc"},
			{Name: "b", Value: []interface{}{"abc", "abc ", " abc"}},
		})
		data = append(data, bson.D{
			{Name: "a", Value: " abc"},
			{Name: "b", Value: []interface{}{" abc", "abc ", " abc"}},
		})
	}
	return "benchmark", map[string][]bson.D{
		"foo": data,
	}
}

var intDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 25000
	data := []bson.D{}
	rand.Seed(17)
	for i := 0; i < numDocs; i++ {
		doc := bson.D{{
			Name:  "int_field",
			Value: int32(rand.Intn(100) * int(math.Pow(-1, float64(i)))),
		}}
		data = append(data, doc)
	}
	return "benchmark", map[string][]bson.D{
		"foo_int": data,
	}
}

var joinDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 1000
	data := []bson.D{}
	for i := 0; i < numDocs; i++ {
		data = append(data, bson.D{{Name: "s", Value: "abcde"}})
	}
	return "benchmark", map[string][]bson.D{
		"foo": data,
		"bar": data,
	}
}

var lpadDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 1000
	data := []bson.D{}
	for i := 0; i < numDocs; i++ {
		data = append(data, bson.D{{Name: "s", Value: "abcde"}})
	}
	return "benchmark", map[string][]bson.D{"strings": data}
}

var orderAndGroupDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 100000
	data := []bson.D{}
	rand.Seed(13)
	for i := 0; i < numDocs; i++ {
		data = append(data, bson.D{
			{Name: "a", Value: rand.Int31n(15)},
			{Name: "b", Value: rand.Int31n(15)},
			{Name: "c", Value: rand.Int31n(15)},
		})
	}
	return "benchmark", map[string][]bson.D{
		"foo": data,
	}
}

var unionDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 100000
	threeFieldsLargeData := []bson.D{}
	rand.Seed(13)
	for i := 0; i < numDocs; i++ {
		threeFieldsLargeData = append(threeFieldsLargeData, bson.D{
			{Name: "a", Value: rand.Int31n(15)},
			{Name: "b", Value: rand.Int31n(15)},
			{Name: "c", Value: rand.Int31n(15)},
		})
	}

	mkDocElem := func(n int) bson.D {
		b := make(bson.D, n)
		for i := 0; i < n; i++ {
			b[i] = bson.DocElem{Name: "a" + strconv.Itoa(i),
				Value: rand.Int31n(15),
			}
		}
		return b
	}

	mkCollection := func(numDocs int, colNum int) []bson.D {
		ret := make([]bson.D, numDocs)
		rand.Seed(13)
		for i := 0; i < numDocs; i++ {
			ret[i] = mkDocElem(colNum)
		}
		return ret
	}

	tenFieldsLargeData := mkCollection(numDocs, 10)
	fifteenFieldsLargeData := mkCollection(numDocs, 15)
	hundredFieldsLargeData := mkCollection(numDocs, 100)

	numDocs = 10
	threeFieldsSmallData := []bson.D{}
	rand.Seed(13)
	for i := 0; i < numDocs; i++ {
		threeFieldsSmallData = append(threeFieldsSmallData, bson.D{
			{Name: "a", Value: rand.Int31n(15)},
			{Name: "b", Value: rand.Int31n(15)},
			{Name: "c", Value: rand.Int31n(15)},
		})
	}

	return "benchmark", map[string][]bson.D{
		"foo":    threeFieldsLargeData,
		"bar":    threeFieldsSmallData,
		"baz":    tenFieldsLargeData,
		"car":    fifteenFieldsLargeData,
		"barbaz": hundredFieldsLargeData,
	}
}

var stringDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 25000
	data := []bson.D{}
	rand.Seed(23)
	strings := []string{"babucket1", "bucket2", "bucket3", "bucket4"}
	for i := 0; i < numDocs; i++ {
		data = append(data, bson.D{
			{Name: "string_field", Value: strings[i%4]},
			{Name: "string_field_number", Value: strconv.Itoa(rand.Int())},
			{Name: "string_field_date", Value: time.Now().String()},
		})
	}
	return "benchmark", map[string][]bson.D{
		"foo_string": data,
	}
}

var subqueriesDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 1000
	data := []bson.D{}
	for i := 0; i < numDocs; i++ {
		data = append(data, bson.D{
			{Name: "a", Value: int32(i)},
			{Name: "b", Value: int32(999 - i)},
		})
	}
	return "benchmark", map[string][]bson.D{
		"foo": data,
		"bar": data,
	}
}

var tableauDataset = data.Once(
	data.DatasetGroup{
		data.NewBSONDataset("attendees"),
		data.NewBSONDataset("flights201406"),
	},
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
	"alter table mongo_partsupp change `_id` `p_partkey`",
	"alter table mongo_part change `_id` `p_partkey`",
	"alter table mongo_orders change `_id` `o_orderkey`",
	"alter table mongo_lineitem change `_id` `o_orderkey`",
	"alter table mongo_customer change `_id` `c_custkey`",
}

var (
	tpchNormalized   = data.Once(data.NewBSONDataset("tpch_full_normalized"))
	tpchDenormalized = data.Once(data.NewBSONDataset("tpch_full_denormalized"))

	tpchMicro = data.Once(data.NewBSONDataset("tpch_small"))
	tpchFull  = data.DatasetGroup{tpchNormalized, tpchDenormalized}
)

func docWithManyFields(n int) bson.D {
	doc := bson.D{}
	for i := 0; i < n; i++ {
		fieldName := fmt.Sprintf("field%d", i)
		doc = append(doc, bson.DocElem{Name: fieldName, Value: "value"})
	}
	return doc
}

func docsArrayWithNesting(depth, length int) []interface{} {
	// base case
	if depth <= 1 {
		docsArray := []interface{}{}
		for i := 0; i < length; i++ {
			docsArray = append(docsArray, bson.D{
				{Name: "a", Value: rand.Int31n(10)},
				{Name: "b", Value: rand.Int31n(10)},
				{Name: "c", Value: rand.Int31n(10)},
			})
		}
		return docsArray
	}

	docsArray := []interface{}{}
	for i := 0; i < length; i++ {
		docsArray = append(docsArray, bson.D{
			{Name: "a", Value: docsArrayWithNesting(depth-1, length)},
			{Name: "b", Value: rand.Int31n(10)},
			{Name: "c", Value: rand.Int31n(10)},
		})
	}
	return docsArray
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

	if strings.Contains(name, "simple_conversions") {
		doc := bson.D{
			{Name: "non_numeric_string", Value: "value"},
			{Name: "bool", Value: false},
			{Name: "double", Value: 2.3},
			{Name: "int", Value: int32(2)},
			{Name: "numeric_string", Value: "2.5"},
		}
		return repeatDoc("conversions", doc, 100000)
	}

	if strings.Contains(name, "simple_count") {
		doc := bson.D{
			{Name: "a", Value: "value"},
		}

		if strings.Contains(name, "_million") {
			return repeatDoc("count", doc, 1000000)
		}

		if strings.Contains(name, "_hundred_thousand") {
			return repeatDoc("count", doc, 100000)
		}

		if strings.Contains(name, "_thousand") {
			return repeatDoc("count", doc, 1000)
		}
	}

	if name[:7] == "filter_" {
		return data.Resample(filterDataset)
	}

	if name[:5] == "join_" {
		if strings.Contains(name, "parent_child") || strings.Contains(name, "grandchild_child") {
			return data.WithIndexes(
				complexDocumentsDataset,
				complexDocumentsIndexes,
				"benchmark",
				"complex_docs",
			)
		}
		return data.Resample(joinDataset)
	}

	if name[:11] == "subqueries_" {
		return data.Resample(subqueriesDataset)
	}

	if name[:6] == "limit_" {
		return data.Resample(repeatDoc("foo",
			bson.D{{Name: "a", Value: "arb"}}, 100000))
	}

	if name[:6] == "union_" {
		return data.Resample(unionDataset)
	}

	if name[:6] == "order_" || name[:6] == "group_" {
		return data.Resample(orderAndGroupDataset)
	}

	if strings.Contains(name, "document_struct") {
		return data.Resample(complexDocumentsDataset)
	}

	if strings.HasPrefix(name, "field_types_int") ||
		strings.HasPrefix(name, "overhead_select_int") {

		return data.Resample(intDataset)
	}

	if strings.HasPrefix(name, "field_types_string") ||
		strings.HasPrefix(name, "overhead_select_string") {

		return data.Resample(stringDataset)
	}

	if strings.HasPrefix(name, "field_types_date") ||
		strings.HasPrefix(name, "overhead_select_date") {
		return data.Resample(dateDataset)
	}

	switch name {
	case "overhead_select_thousand_simple_docs":
		return repeatDoc("items", bson.D{{Name: "key", Value: "value"}}, 1000)
	case "overhead_select_million_simple_docs":
		return repeatDoc("items", bson.D{{Name: "key", Value: "value"}}, 1000000)
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
	case "simple_complex_predicate_expr", "overhead_select_with_large_pipeline":
		doc := bson.D{
			{Name: "a", Value: "value"},
			{Name: "b", Value: "value"},
		}
		return repeatDoc("items", doc, 1000)
	case "overhead_select_docs_with_deeply_nested_arrays":
		doc := bson.D{
			{Name: "a", Value: docsArrayWithNesting(4, 4)},
		}
		return repeatDoc("items", doc, 1000)
	default:
		panic(fmt.Errorf("no dataset for benchmark %s", name))
	}
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
