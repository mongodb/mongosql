package catalog

import (
	"strconv"
	"testing"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/mongoast/astprint"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func TestMongoTableMarshalBSON(t *testing.T) {
	// This tests that a catalog.MongoTable can
	// successfully marshal and unmarshal BSON.

	req := require.New(t)

	expectedName := "foo"
	expectedCollation := collation.Default
	expectedColumn1 := &results.Column{
		ColumnType: &results.ColumnType{
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
	}
	expectedColumn2 := &results.Column{
		ColumnType: &results.ColumnType{
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
	}
	expectedColumn3 := &results.Column{
		ColumnType: &results.ColumnType{
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
	}
	expectedColumns := results.Columns{expectedColumn1, expectedColumn2, expectedColumn3}
	expectedColumnMap := map[string]*results.Column{
		expectedColumn1.Name: expectedColumn1,
		expectedColumn2.Name: expectedColumn2,
		expectedColumn3.Name: expectedColumn3,
	}
	expectedPrimaryKeys := results.Columns{expectedColumn1}
	expectedIndexes := []Index{
		{
			columns:        results.Columns{expectedColumn1},
			unique:         true,
			fullText:       false,
			constraintName: "_id_",
		},
	}
	expectedForeignKeys := []ForeignKey{
		{
			columns:         results.Columns{expectedColumn3},
			constraintName:  "foreign_b",
			foreignDatabase: "foreignDB",
			foreignTable:    "foreignTable",
			localToForeignColumn: map[string]string{
				"b": "b",
			},
		},
	}
	expectedComments := "comments"
	expectedTableType := BaseTable
	expectedCollectionName := "foo"
	expectedPipeline := ast.NewPipeline(
		ast.NewProjectStage(
			ast.NewAssignProjectItem(
				"x",
				ast.NewUnary(
					"$abs",
					ast.NewConstant(bsoncore.Value{Type: bsontype.Int32, Data: bsoncore.AppendInt32(nil, -10)}),
				),
			),
		),
	)

	expected := &MongoTable{
		name:           expectedName,
		collation:      expectedCollation,
		columns:        expectedColumns,
		columnMap:      expectedColumnMap,
		primaryKeys:    expectedPrimaryKeys,
		indexes:        expectedIndexes,
		foreignKeys:    expectedForeignKeys,
		comments:       expectedComments,
		tableType:      expectedTableType,
		isSharded:      true,
		collectionName: expectedCollectionName,
		pipeline:       expectedPipeline,
	}

	// Check marshaling
	actualBytes, err := bson.Marshal(expected)
	req.NoError(err, "failed to marshal")

	actualDoc := bsoncore.Document(actualBytes)
	actualElements, err := actualDoc.Elements()
	req.NoError(err, "failed to get elements from marshaled document")

	expectedValues := map[string]*bsoncore.Value{
		"collectionName": makeStringValue(expectedCollectionName),
		"tableName":      makeStringValue(expectedName),
		"collation":      makeStringValue(string(expectedCollation.Name)),
		"columns":        nil, // we will not check columns here since results.Columns marshaling is tested separately
		"primaryKeys":    makeArrayValue([]bsoncore.Value{*makeStringValue("_id")}),
		"indexes":        makeIndexesArrayValue(expectedIndexes),
		"foreignKeys":    makeForeignKeysArrayValue(expectedForeignKeys),
		"comments":       makeStringValue(expectedComments),
		"tableType":      makeStringValue(expectedTableType),
		"pipeline":       makeStringValue(astprint.String(expectedPipeline)),
	}

	// check that number of key-value pairs is what's expected
	req.Equal(len(expectedValues), len(actualElements), "number of elements")

	// check that each value is correctly named and set
	for expectedKey, expectedValue := range expectedValues {
		var actualValue bsoncore.Value
		actualValue, err = actualDoc.LookupErr(expectedKey)
		req.NoError(err, "error looking up key %q", expectedKey)

		if expectedValue == nil {
			continue
		}

		req.True(actualValue.Equal(*expectedValue), "value for %q", expectedKey)
	}

	// Check unmarshaling
	actual := MongoTable{}
	err = bson.Unmarshal(actualBytes, &actual)
	req.NoError(err, "failed to unmarshal")

	req.Equal(expected.name, actual.name, "name")
	req.Equal(expected.collation, actual.collation, "collation")
	req.Equal(expected.comments, actual.comments, "comments")
	req.Equal(expected.tableType, actual.tableType, "tableType")
	req.Equal(expected.isSharded, actual.isSharded, "isSharded")
	req.Equal(expected.collectionName, actual.collectionName, "collectionName")
	req.Equal(expected.columns, actual.columns, "columns")
	req.Equal(expected.columnMap, actual.columnMap, "columnMap")
	req.Equal(expected.primaryKeys, actual.primaryKeys, "primaryKeys")
	req.Equal(expected.indexes, actual.indexes, "indexes")
	req.Equal(expected.foreignKeys, actual.foreignKeys, "foreignKeys")
	req.Equal(expected.pipeline, actual.pipeline, "pipeline")
}

func TestMongoTableMarshalBSONWithNilValues(t *testing.T) {
	// This tests that a catalog.MongoTable can
	// successfully marshal and unmarshal BSON.

	req := require.New(t)

	expectedName := "foo"
	expectedCollation := collation.Default
	expectedColumn := &results.Column{
		ColumnType: &results.ColumnType{
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
	}
	expectedColumns := results.Columns{expectedColumn}
	expectedColumnMap := map[string]*results.Column{
		expectedColumn.Name: expectedColumn,
	}
	var expectedPrimaryKeys results.Columns
	var expectedIndexes []Index
	var expectedForeignKeys []ForeignKey
	expectedComments := "comments"
	expectedTableType := BaseTable
	expectedCollectionName := "foo"
	expectedPipeline := ast.NewPipeline(
		ast.NewProjectStage(
			ast.NewAssignProjectItem(
				"x",
				ast.NewUnary(
					"$abs",
					ast.NewConstant(bsoncore.Value{Type: bsontype.Int32, Data: bsoncore.AppendInt32(nil, -10)}),
				),
			),
		),
	)

	expected := &MongoTable{
		name:           expectedName,
		collation:      expectedCollation,
		columns:        expectedColumns,
		columnMap:      expectedColumnMap,
		primaryKeys:    expectedPrimaryKeys,
		indexes:        expectedIndexes,
		foreignKeys:    expectedForeignKeys,
		comments:       expectedComments,
		tableType:      expectedTableType,
		isSharded:      true,
		collectionName: expectedCollectionName,
		pipeline:       expectedPipeline,
	}

	// Check marshaling
	actualBytes, err := bson.Marshal(expected)
	req.NoError(err, "failed to marshal")

	actualDoc := bsoncore.Document(actualBytes)
	actualElements, err := actualDoc.Elements()
	req.NoError(err, "failed to get elements from marshaled document")

	expectedValues := map[string]*bsoncore.Value{
		"collectionName": makeStringValue(expectedCollectionName),
		"tableName":      makeStringValue(expectedName),
		"collation":      makeStringValue(string(expectedCollation.Name)),
		"columns":        nil, // we will not check columns here since results.Columns marshaling is tested separately
		"primaryKeys":    makeNullValue(),
		"indexes":        makeNullValue(),
		"foreignKeys":    makeNullValue(),
		"comments":       makeStringValue(expectedComments),
		"tableType":      makeStringValue(expectedTableType),
		"pipeline":       makeStringValue(astprint.String(expectedPipeline)),
	}

	// check that number of key-value pairs is what's expected
	req.Equal(len(expectedValues), len(actualElements), "number of elements")

	// check that each value is correctly named and set
	for expectedKey, expectedValue := range expectedValues {
		var actualValue bsoncore.Value
		actualValue, err = actualDoc.LookupErr(expectedKey)
		req.NoError(err, "error looking up key %q", expectedKey)

		if expectedValue == nil {
			continue
		}

		req.True(actualValue.Equal(*expectedValue), "value for %q", expectedKey)
	}

	// Check unmarshaling
	actual := MongoTable{}
	err = bson.Unmarshal(actualBytes, &actual)
	req.NoError(err, "failed to unmarshal")

	req.Equal(expected.name, actual.name, "name")
	req.Equal(expected.collation, actual.collation, "collation")
	req.Equal(expected.comments, actual.comments, "comments")
	req.Equal(expected.tableType, actual.tableType, "tableType")
	req.Equal(expected.isSharded, actual.isSharded, "isSharded")
	req.Equal(expected.collectionName, actual.collectionName, "collectionName")
	req.Equal(expected.columns, actual.columns, "columns")
	req.Equal(expected.columnMap, actual.columnMap, "columnMap")
	req.Equal(expected.primaryKeys, actual.primaryKeys, "primaryKeys")
	req.Equal(expected.indexes, actual.indexes, "indexes")
	req.Equal(expected.foreignKeys, actual.foreignKeys, "foreignKeys")
	req.Equal(expected.pipeline, actual.pipeline, "pipeline")
}

func makeStringValue(s string) *bsoncore.Value {
	return &bsoncore.Value{
		Type: bsontype.String,
		Data: bsoncore.AppendString(nil, s),
	}
}

func makeNullValue() *bsoncore.Value {
	return &bsoncore.Value{
		Type: bsontype.Null,
	}
}

func makeArrayValue(values []bsoncore.Value) *bsoncore.Value {
	_, arr := bsoncore.AppendArrayStart(nil)
	for i, value := range values {
		arr = bsoncore.AppendValueElement(arr, strconv.Itoa(i), value)
	}
	arr, _ = bsoncore.AppendArrayEnd(arr, 0)

	return &bsoncore.Value{
		Type: bsontype.Array,
		Data: arr,
	}
}

func makeIndexesArrayValue(indexes []Index) *bsoncore.Value {
	docValues := make([]bsoncore.Value, len(indexes))
	for i, index := range indexes {
		docValues[i] = makeIndexDocValue(index)
	}
	return makeArrayValue(docValues)
}

func makeIndexDocValue(index Index) bsoncore.Value {
	idxColumnNames := make([]bsoncore.Value, len(index.columns))
	for i, col := range index.columns {
		idxColumnNames[i] = *makeStringValue(col.Name)
	}

	_, doc := bsoncore.AppendDocumentStart(nil)
	doc = bsoncore.AppendArrayElement(doc, "columns", makeArrayValue(idxColumnNames).Data)
	doc = bsoncore.AppendBooleanElement(doc, "unique", index.unique)
	doc = bsoncore.AppendBooleanElement(doc, "fullText", index.fullText)
	doc = bsoncore.AppendStringElement(doc, "constraintName", index.constraintName)
	doc, _ = bsoncore.AppendDocumentEnd(doc, 0)

	return bsoncore.Value{
		Type: bsontype.EmbeddedDocument,
		Data: doc,
	}
}

func makeForeignKeysArrayValue(fks []ForeignKey) *bsoncore.Value {
	docValues := make([]bsoncore.Value, len(fks))
	for i, fk := range fks {
		docValues[i] = makeForeignKeyDocValue(fk)
	}
	return makeArrayValue(docValues)
}

func makeForeignKeyDocValue(fk ForeignKey) bsoncore.Value {
	fkColumnNames := make([]bsoncore.Value, len(fk.columns))
	for i, col := range fk.columns {
		fkColumnNames[i] = *makeStringValue(col.Name)
	}

	_, localToForeignColumnDoc := bsoncore.AppendDocumentStart(nil)
	for local, foreign := range fk.localToForeignColumn {
		localToForeignColumnDoc = bsoncore.AppendStringElement(localToForeignColumnDoc, local, foreign)
	}
	localToForeignColumnDoc, _ = bsoncore.AppendDocumentEnd(localToForeignColumnDoc, 0)

	_, doc := bsoncore.AppendDocumentStart(nil)
	doc = bsoncore.AppendArrayElement(doc, "columns", makeArrayValue(fkColumnNames).Data)
	doc = bsoncore.AppendStringElement(doc, "constraintName", fk.constraintName)
	doc = bsoncore.AppendStringElement(doc, "foreignDatabase", fk.foreignDatabase)
	doc = bsoncore.AppendStringElement(doc, "foreignTable", fk.foreignTable)
	doc = bsoncore.AppendDocumentElement(doc, "localToForeignColumn", localToForeignColumnDoc)
	doc, _ = bsoncore.AppendDocumentEnd(doc, 0)

	return bsoncore.Value{
		Type: bsontype.EmbeddedDocument,
		Data: doc,
	}
}
