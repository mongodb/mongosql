package results

import (
	"strings"
	"testing"

	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/schema"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func TestColumnsMarshalBSON(t *testing.T) {
	// This tests that structs with a results.Columns field can
	// successfully marshal and unmarshal BSON.

	type testType struct {
		Columns Columns `bson:"columns"`
	}

	expected := testType{
		Columns: []*Column{
			{
				ColumnType: &ColumnType{
					EvalType:    types.EvalObjectID,
					MongoType:   schema.MongoObjectID,
					UUIDSubType: types.EvalBinary,
				},
				SelectID:            1,
				Table:               "foo",
				OriginalTable:       "foo",
				Database:            "db",
				Name:                "_id",
				OriginalName:        "_id",
				MappingRegistryName: "_id",
				MongoName:           "_id",
				PrimaryKey:          true,
				Comments:            "",
				IsPolymorphic:       false,
				HasAlteredType:      true,
				Nullable:            false,
			},
			{
				ColumnType: &ColumnType{
					EvalType:    types.EvalString,
					MongoType:   schema.MongoString,
					UUIDSubType: types.EvalBinary,
				},
				SelectID:            1,
				Table:               "foo",
				OriginalTable:       "foo",
				Database:            "db",
				Name:                "a",
				OriginalName:        "a",
				MappingRegistryName: "a",
				MongoName:           "a",
				PrimaryKey:          false,
				Comments:            "",
				IsPolymorphic:       true,
				HasAlteredType:      false,
				Nullable:            true,
			},
			{
				ColumnType: &ColumnType{
					EvalType:    types.EvalInt64,
					MongoType:   schema.MongoInt64,
					UUIDSubType: types.EvalBinary,
				},
				SelectID:            1,
				Table:               "foo",
				OriginalTable:       "foo",
				Database:            "db",
				Name:                "b",
				OriginalName:        "b",
				MappingRegistryName: "b",
				MongoName:           "b",
				PrimaryKey:          false,
				Comments:            "this is b",
				IsPolymorphic:       false,
				HasAlteredType:      false,
				Nullable:            false,
			},
		},
	}

	actualBytes, err := bson.Marshal(&expected)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	actualColumnsArray := bsoncore.Document(actualBytes).Lookup("columns").Array()
	actualColumns, err := actualColumnsArray.Values()
	if err != nil {
		t.Fatalf("failed to get \"columns\" array values: %v", err)
	}

	if len(actualColumns) != len(expected.Columns) {
		t.Fatalf("actual number of columns does not match expected number of columns (+++ actual, --- expected)\n+++ %v\n--- %v",
			len(actualColumns), len(expected.Columns))
	}

	expectedKeys := []string{"columnType.sqlType", "columnType.mongoType", "columnType.uuidSubType",
		"selectID", "originalTable", "database", "name", "mongoName", "primaryKey", "comments", "isPolymorphic", "hasAlteredType", "nullable"}
	expectedValues := [][]bsoncore.Value{
		{
			makeStringValue("objectID"), makeStringValue("bson.ObjectId"), makeStringValue("binary"),
			makeInt32Value(1), makeStringValue("foo"), makeStringValue("db"), makeStringValue("_id"),
			makeStringValue("_id"), makeBoolValue(true), makeStringValue(""), makeBoolValue(false),
			makeBoolValue(true), makeBoolValue(false),
		},
		{
			makeStringValue("string"), makeStringValue("string"), makeStringValue("binary"),
			makeInt32Value(1), makeStringValue("foo"), makeStringValue("db"), makeStringValue("a"),
			makeStringValue("a"), makeBoolValue(false), makeStringValue(""), makeBoolValue(true),
			makeBoolValue(false), makeBoolValue(true),
		},
		{
			makeStringValue("int64"), makeStringValue("int64"), makeStringValue("binary"),
			makeInt32Value(1), makeStringValue("foo"), makeStringValue("db"), makeStringValue("b"),
			makeStringValue("b"), makeBoolValue(false), makeStringValue("this is b"),
			makeBoolValue(false), makeBoolValue(false), makeBoolValue(false),
		},
	}

	// Check that the values are correctly set.
	for i, actualColumn := range actualColumns {
		actualColumnDoc, ok := actualColumn.DocumentOK()
		if !ok {
			t.Fatalf("actual column %v is not a document", i)
		}

		var actualElements []bsoncore.Element
		actualElements, err = actualColumnDoc.Elements()
		if err != nil {
			t.Fatalf("failed to get actual document elements: %v", err)
		}

		// Check that exactly the expected number of elements are present.
		// We subtract 2 from length of expected keys because the "columnType"
		// element will be one single element with a document value.
		if len(actualElements) != len(expectedKeys)-2 {
			t.Fatalf("actual number of elements does not match expected number of elements (+++ actual, --- expected)\n+++ %v\n--- %v",
				len(actualElements), len(expectedKeys)-2)
		}

		expectedColumnValues := expectedValues[i]
		for j, key := range expectedKeys {
			var actualVal bsoncore.Value
			actualVal, err = actualColumnDoc.LookupErr(strings.Split(key, ".")...)
			if err != nil {
				t.Fatalf("column %v: error looking up key %q: %v", i, key, err)
			}

			if !actualVal.Equal(expectedColumnValues[j]) {
				t.Fatalf("column %v: actual value does match expected value for %q (+++ actual, --- expected)\n+++ %v\n--- %v",
					i, key, actualVal, expectedColumnValues[j])
			}
		}
	}

	actual := testType{}
	err = bson.Unmarshal(actualBytes, &actual)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(actual.Columns) != len(expected.Columns) {
		t.Fatalf("actual number of columns does not match expected number of columns (+++ actual, --- expected)\n+++ %v\n--- %v",
			len(actual.Columns), len(expected.Columns))
	}

	for i, actualColumn := range actual.Columns {
		expectedColumn := expected.Columns[i]

		// check that tagged fields are set
		if actualColumn.ColumnType.EvalType != expectedColumn.ColumnType.EvalType {
			t.Fatalf("column %v: actual EvalType does not match expected EvalType (+++ actual, --- expected)\n+++ %v\n--- %v",
				i, actualColumn.ColumnType.EvalType, expectedColumn.ColumnType.EvalType)
		}
		if actualColumn.ColumnType.MongoType != expectedColumn.ColumnType.MongoType {
			t.Fatalf("column %v: actual MongoType does not match expected MongoType (+++ actual, --- expected)\n+++ %v\n--- %v",
				i, actualColumn.ColumnType.MongoType, expectedColumn.ColumnType.MongoType)
		}
		if actualColumn.ColumnType.UUIDSubType != expectedColumn.ColumnType.UUIDSubType {
			t.Fatalf("column %v: actual UUIDSubType does not match expected UUIDSubType (+++ actual, --- expected)\n+++ %v\n--- %v",
				i, actualColumn.ColumnType.UUIDSubType, expectedColumn.ColumnType.UUIDSubType)
		}
		if actualColumn.SelectID != expectedColumn.SelectID {
			t.Fatalf("column %v: actual SelectID does not match expected SelectID (+++ actual, --- expected)\n+++ %v\n--- %v",
				i, actualColumn.SelectID, expectedColumn.SelectID)
		}
		if actualColumn.OriginalTable != expectedColumn.OriginalTable {
			t.Fatalf("column %v: actual OriginalTable does not match expected OriginalTable (+++ actual, --- expected)\n+++ %v\n--- %v",
				i, actualColumn.OriginalTable, expectedColumn.OriginalTable)
		}
		if actualColumn.Database != expectedColumn.Database {
			t.Fatalf("column %v: actual Database does not match expected Database (+++ actual, --- expected)\n+++ %v\n--- %v",
				i, actualColumn.Database, expectedColumn.Database)
		}
		if actualColumn.Name != expectedColumn.Name {
			t.Fatalf("column %v: actual Name does not match expected Name (+++ actual, --- expected)\n+++ %v\n--- %v",
				i, actualColumn.Name, expectedColumn.Name)
		}
		if actualColumn.MongoName != expectedColumn.MongoName {
			t.Fatalf("column %v: actual MongoName does not match expected MongoName (+++ actual, --- expected)\n+++ %v\n--- %v",
				i, actualColumn.MongoName, expectedColumn.MongoName)
		}
		if actualColumn.PrimaryKey != expectedColumn.PrimaryKey {
			t.Fatalf("column %v: actual PrimaryKey does not match expected PrimaryKey (+++ actual, --- expected)\n+++ %v\n--- %v",
				i, actualColumn.PrimaryKey, expectedColumn.PrimaryKey)
		}
		if actualColumn.Comments != expectedColumn.Comments {
			t.Fatalf("column %v: actual Comments does not match expected Comments (+++ actual, --- expected)\n+++ %v\n--- %v",
				i, actualColumn.Comments, expectedColumn.Comments)
		}
		if actualColumn.IsPolymorphic != expectedColumn.IsPolymorphic {
			t.Fatalf("column %v: actual IsPolymorphic does not match expected IsPolymorphic (+++ actual, --- expected)\n+++ %v\n--- %v",
				i, actualColumn.IsPolymorphic, expectedColumn.IsPolymorphic)
		}
		if actualColumn.HasAlteredType != expectedColumn.HasAlteredType {
			t.Fatalf("column %v: actual HasAlteredType does not match expected HasAlteredType (+++ actual, --- expected)\n+++ %v\n--- %v",
				i, actualColumn.HasAlteredType, expectedColumn.HasAlteredType)
		}
		if actualColumn.Nullable != expectedColumn.Nullable {
			t.Fatalf("column %v: actual Nullable does not match expected Nullable (+++ actual, --- expected)\n+++ %v\n--- %v",
				i, actualColumn.Nullable, expectedColumn.Nullable)
		}

		// check that skipped fields are default valued
		if actualColumn.Table != "" {
			t.Fatalf("column %v: actual Table is not default/empty value: %v", i, actualColumn.Table)
		}
		if actualColumn.OriginalName != "" {
			t.Fatalf("column %v: actual OriginalName is not default/empty value: %v", i, actualColumn.Table)
		}
		if actualColumn.MappingRegistryName != "" {
			t.Fatalf("column %v: actual MappingRegistryName is not default/empty value: %v", i, actualColumn.Table)
		}
	}
}

func makeStringValue(s string) bsoncore.Value {
	return bsoncore.Value{
		Type: bsontype.String,
		Data: bsoncore.AppendString(nil, s),
	}
}

func makeInt32Value(i int32) bsoncore.Value {
	return bsoncore.Value{
		Type: bsontype.Int32,
		Data: bsoncore.AppendInt32(nil, i),
	}
}

func makeBoolValue(b bool) bsoncore.Value {
	return bsoncore.Value{
		Type: bsontype.Boolean,
		Data: bsoncore.AppendBoolean(nil, b),
	}
}
