package bench

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/10gen/sqlproxy/internal/testutil/data"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var complexDocumentsDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 1000
	data := make([]bson.D, numDocs)
	for i := 0; i < numDocs; i++ {
		data[i] = bson.D{
			{Key: "a", Value: bson.D{
				{Key: "b", Value: bson.D{{
					Key: "c", Value: rand.Int31n(15),
				}}},
			}},
			{Key: "b", Value: bson.D{
				{Key: "a", Value: rand.Int31n(15)}},
			},
			{Key: "c", Value: rand.Int31n(15)},
			{Key: "d", Value: docsArrayWithNesting(2, 5)},
		}
	}
	return "benchmark", map[string][]bson.D{
		"complex_docs": data,
	}
}

var complexDocumentsIndexes = []bson.D{
	{
		{Key: "key", Value: bson.D{
			{Key: "c", Value: 1}},
		},
		{Key: "name", Value: "c_index"},
	},
	{
		{Key: "key", Value: bson.D{
			{Key: "d.b", Value: 1}},
		},
		{Key: "name", Value: "d_b_index"},
	},
	{
		{Key: "key", Value: bson.D{
			{Key: "d.a.a", Value: 1}},
		},
		{Key: "name", Value: "d_a_a_index"},
	},
}

var dateDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 25000
	dates := make([]primitive.DateTime, 20)
	for i := 0; i < 20; i++ {
		dates[i] = primitive.NewDateTimeFromTime(time.Now())
	}
	data := make([]bson.D, numDocs)
	for i := 0; i < numDocs; i++ {
		data[i] = bson.D{
			{Key: "random", Value: "abc"},
			{Key: "date_field", Value: dates[i%20]},
		}
	}
	return "benchmark", map[string][]bson.D{
		"foo_date": data,
	}
}

var filterDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 1000
	data := make([]bson.D, 0, numDocs*2)
	for i := 0; i < numDocs; i++ {
		data = append(data, bson.D{
			{Key: "a", Value: "abc"},
			{Key: "b", Value: bson.A{"abc", "abc ", " abc"}},
		})
		data = append(data, bson.D{
			{Key: "a", Value: " abc"},
			{Key: "b", Value: bson.A{" abc", "abc ", " abc"}},
		})
	}
	return "benchmark", map[string][]bson.D{
		"foo": data,
	}
}

var intDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 25000
	data := make([]bson.D, numDocs)
	rand.Seed(17)
	for i := 0; i < numDocs; i++ {
		data[i] = bson.D{{
			Key:   "int_field",
			Value: int32(rand.Intn(100) * int(math.Pow(-1, float64(i)))),
		}}
	}
	return "benchmark", map[string][]bson.D{
		"foo_int": data,
	}
}

var joinDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 1000
	data := make([]bson.D, numDocs)
	for i := 0; i < numDocs; i++ {
		data[i] = bson.D{{Key: "s", Value: "abcde"}}
	}
	return "benchmark", map[string][]bson.D{
		"foo": data,
		"bar": data,
	}
}

var scalarConflictDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 10000
	data := make([]bson.D, 0, numDocs)
	for i := 0; i < numDocs/4; i++ {
		data = append(data, bson.D{
			{
				Key: "a", Value: "hello",
			},
			{
				Key: "b", Value: 3.14,
			},
		},
		)
		data = append(data, bson.D{
			{
				Key: "a", Value: 3.14,
			},
			{
				Key: "b", Value: true,
			},
		},
		)
		data = append(data, bson.D{
			{
				Key: "a", Value: int32(1),
			},
			{
				Key: "b", Value: false,
			},
		},
		)
		data = append(data, bson.D{
			{
				Key: "a", Value: int32(1),
			},
			{
				Key: "b", Value: true,
			},
		},
		)
	}

	return "benchmark", map[string][]bson.D{
		"scalar_conflict": data,
	}
}

var scalarNoConflictDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 10000
	data := make([]bson.D, numDocs)
	for i := 0; i < numDocs; i++ {
		data[i] = bson.D{
			{
				Key: "a", Value: "hello",
			},
			{
				Key: "b", Value: 3.14,
			},
		}
	}

	return "benchmark", map[string][]bson.D{
		"scalar_noconflict": data,
	}
}

var objectConflictDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 9999
	data := make([]bson.D, 0, numDocs)
	for i := 0; i < numDocs/3; i++ {
		data = append(data, bson.D{
			{
				Key: "a", Value: bson.D{{
					Key: "b", Value: bson.D{{
						Key: "c", Value: int32(42),
					}},
				}},
			},
			{
				Key: "d", Value: bson.D{{
					Key: "e", Value: bson.D{{
						Key: "f", Value: int32(43),
					}},
				}},
			},
		})

		data = append(data, bson.D{
			{
				Key: "a", Value: bson.D{{
					Key: "b", Value: int32(40),
				}},
			},
			{
				Key: "d", Value: bson.D{{
					Key: "e", Value: int32(41),
				}},
			},
		})

		data = append(data, bson.D{
			{
				Key: "a", Value: int32(38),
			},
			{
				Key: "d", Value: int32(39),
			},
		})
	}
	return "benchmark", map[string][]bson.D{
		"nested_object_conflict": data,
	}
}

var objectNoConflictDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 9999
	data := make([]bson.D, numDocs)
	for i := 0; i < numDocs; i++ {
		data[i] = bson.D{
			{
				Key: "a", Value: bson.D{{
					Key: "b", Value: bson.D{{
						Key: "c", Value: int32(42),
					}},
				}},
			},
			{
				Key: "d", Value: bson.D{{
					Key: "e", Value: bson.D{{
						Key: "f", Value: int32(43),
					}},
				}},
			},
		}
	}
	return "benchmark", map[string][]bson.D{
		"nested_object_noconflict": data,
	}
}

var objectConflictArrayDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 10000
	data := make([]bson.D, 0, numDocs)
	for i := 0; i < numDocs/2; i++ {
		data = append(data, bson.D{
			{
				Key: "a", Value: bson.D{{
					Key: "b", Value: bson.A{
						bson.D{{
							Key: "c", Value: int32(42),
						}},
						bson.D{{
							Key: "c", Value: int32(43),
						}},
					},
				}},
			},
		})

		data = append(data, bson.D{
			{
				Key: "a", Value: bson.A{
					bson.D{{
						Key: "e", Value: bson.D{{
							Key: "f", Value: int32(44),
						}},
					}},
					bson.D{{
						Key: "e", Value: bson.D{{
							Key: "f", Value: int32(45),
						}},
					}},
					bson.D{{
						Key: "e", Value: int32(46),
					}},
				},
			},
		})
	}
	return "benchmark", map[string][]bson.D{
		"nested_object_array_conflict": data,
	}
}

var objectNoConflictArrayDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 10000
	data := make([]bson.D, numDocs)
	for i := 0; i < numDocs; i++ {
		data[i] = bson.D{
			{
				Key: "a", Value: bson.D{{
					Key: "b", Value: bson.A{
						bson.D{{
							Key: "c", Value: int32(42),
						}},
						bson.D{{
							Key: "c", Value: int32(43),
						}},
					},
				}},
			},
		}
	}
	return "benchmark", map[string][]bson.D{
		"nested_object_array_noconflict": data,
	}
}

var lpadDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 1000
	data := make([]bson.D, numDocs)
	for i := 0; i < numDocs; i++ {
		data[i] = bson.D{{Key: "s", Value: "abcde"}}
	}
	return "benchmark", map[string][]bson.D{"strings": data}
}

var tupleDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 1000000
	data := make([]bson.D, numDocs)
	for i := 0; i < numDocs; i++ {
		data[i] = bson.D{
			{Key: "a", Value: "abcde"},
			{Key: "b", Value: "bcdef"},
			{Key: "c", Value: "cdefg"},
		}
	}
	return "benchmark", map[string][]bson.D{"strings": data}
}

var orderAndGroupDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 100000
	data := make([]bson.D, numDocs)
	rand.Seed(13)
	for i := 0; i < numDocs; i++ {
		data[i] = bson.D{
			{Key: "a", Value: rand.Int31n(15)},
			{Key: "b", Value: rand.Int31n(15)},
			{Key: "c", Value: rand.Int31n(15)},
		}
	}
	return "benchmark", map[string][]bson.D{
		"foo": data,
	}
}

var unionDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 100000
	threeFieldsLargeData := make([]bson.D, numDocs)
	rand.Seed(13)
	for i := 0; i < numDocs; i++ {
		threeFieldsLargeData[i] = bson.D{
			{Key: "a", Value: rand.Int31n(15)},
			{Key: "b", Value: rand.Int31n(15)},
			{Key: "c", Value: rand.Int31n(15)},
		}
	}

	mkDocElem := func(n int) bson.D {
		b := make(bson.D, n)
		for i := 0; i < n; i++ {
			b[i] = bson.E{Key: "a" + strconv.Itoa(i),
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
	threeFieldsSmallData := make([]bson.D, numDocs)
	rand.Seed(13)
	for i := 0; i < numDocs; i++ {
		threeFieldsSmallData[i] = bson.D{
			{Key: "a", Value: rand.Int31n(15)},
			{Key: "b", Value: rand.Int31n(15)},
			{Key: "c", Value: rand.Int31n(15)},
		}
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
	data := make([]bson.D, numDocs)
	rand.Seed(23)
	strings := []string{"babucket1", "bucket2", "bucket3", "bucket4"}
	for i := 0; i < numDocs; i++ {
		data[i] = bson.D{
			{Key: "string_field", Value: strings[i%4]},
			{Key: "string_field_number", Value: strconv.Itoa(rand.Int())},
			{Key: "string_field_date", Value: time.Now().String()},
		}
	}
	return "benchmark", map[string][]bson.D{
		"foo_string": data,
	}
}

var subqueriesDataset data.DynamicDataset = func() (string, map[string][]bson.D) {
	numDocs := 1000
	data := make([]bson.D, numDocs)
	for i := 0; i < numDocs; i++ {
		data[i] = bson.D{
			{Key: "a", Value: int32(i)},
			{Key: "b", Value: int32(999 - i)},
		}
	}
	return "benchmark", map[string][]bson.D{
		"foo": data,
		"bar": data,
	}
}

var tableauDataset = data.Once(
	data.DatasetGroup{
		data.NewBSONDataset("attendees-new"),
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
	doc := make(bson.D, n)
	for i := 0; i < n; i++ {
		fieldName := fmt.Sprintf("field%d", i)
		doc[i] = bson.E{Key: fieldName, Value: "value"}
	}
	return doc
}

func docsArrayWithNesting(depth, length int) bson.A {
	// base case
	if depth <= 1 {
		docsArray := make(bson.A, length)
		for i := 0; i < length; i++ {
			docsArray[i] = bson.D{
				{Key: "a", Value: rand.Int31n(10)},
				{Key: "b", Value: rand.Int31n(10)},
				{Key: "c", Value: rand.Int31n(10)},
			}
		}
		return docsArray
	}

	docsArray := make(bson.A, length)
	for i := 0; i < length; i++ {
		docsArray[i] = bson.D{
			{Key: "a", Value: docsArrayWithNesting(depth-1, length)},
			{Key: "b", Value: rand.Int31n(10)},
			{Key: "c", Value: rand.Int31n(10)},
		}
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

	if strings.Contains(name, "tuples") {
		return data.Resample(tupleDataset)
	}

	if strings.Contains(name, "simple_conversions") {
		doc := bson.D{
			{Key: "non_numeric_string", Value: "value"},
			{Key: "bool", Value: false},
			{Key: "double", Value: 2.3},
			{Key: "int", Value: int32(2)},
			{Key: "numeric_string", Value: "2.5"},
		}
		return repeatDoc("conversions", doc, 100000)
	}

	if strings.Contains(name, "simple_count") {
		doc := bson.D{
			{Key: "a", Value: "value"},
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
			bson.D{{Key: "a", Value: "arb"}}, 100000))
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

	if strings.HasPrefix(name, "simple_select_nested_object_array_conflict") {
		return data.Resample(objectConflictArrayDataset)
	}

	if strings.HasPrefix(name, "simple_select_nested_object_array_noconflict") {
		return data.Resample(objectNoConflictArrayDataset)
	}

	if strings.HasPrefix(name, "simple_select_nested_object_conflict") {
		return data.Resample(objectConflictDataset)
	}

	if strings.HasPrefix(name, "simple_select_nested_object_noconflict") {
		return data.Resample(objectNoConflictDataset)
	}

	if strings.HasPrefix(name, "simple_select_scalar_conflict") {
		return data.Resample(scalarConflictDataset)
	}

	if strings.HasPrefix(name, "simple_select_scalar_noconflict") {
		return data.Resample(scalarNoConflictDataset)
	}

	switch name {
	case "overhead_select_thousand_simple_docs":
		return repeatDoc("items", bson.D{{Key: "key", Value: "value"}}, 1000)
	case "overhead_select_million_simple_docs":
		return repeatDoc("items", bson.D{{Key: "key", Value: "value"}}, 1000000)
	case "overhead_select_one_doc_thousand_fields":
		doc := docWithManyFields(1000)
		return repeatDoc("items", doc, 1)
	case "overhead_select_one_doc_ten_thousand_fields":
		doc := docWithManyFields(10000)
		return repeatDoc("items", doc, 1)
	case "overhead_select_thousand_docs_ten_fields":
		doc := docWithManyFields(10)
		return repeatDoc("items", doc, 1000)
	case "overhead_select_thousand_docs_hundred_fields",
		"simple_select_thousand_docs_hundred_fields":
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
			{Key: "a", Value: "value"},
			{Key: "b", Value: "value"},
		}
		return repeatDoc("items", doc, 1000)
	case "overhead_select_docs_with_deeply_nested_arrays":
		doc := bson.D{
			{Key: "a", Value: docsArrayWithNesting(4, 4)},
		}
		return repeatDoc("items", doc, 1000)
	case "simple_select_nested_object_conflict":
		return data.Resample(objectConflictDataset)
	case "simple_select_scalar_conflict":
		return data.Resample(scalarConflictDataset)
	case "simple_no_null_checks_for_literals":
		return data.Resample(tupleDataset)
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
