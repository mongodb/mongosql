package sample

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/internal/option"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func testTable(lg log.Logger, tbl, col string,
	pipeline []bson.D, cols []*schema.Column,
	indexes []schema.Index, comment option.String) *schema.Table {
	out, err := schema.NewTable(lg, tbl, col, pipeline, cols, indexes, comment)
	if err != nil {
		panic("this table should not error")
	}
	return out
}

func TestDeserializeTableSchema(t *testing.T) {
	type test struct {
		name       string
		tableName  string
		jsonSchema string
		indexes    []indexInfo
		expected   *schema.Table
	}

	logger := log.NewComponentLogger(log.SchemaComponent, log.GlobalLogger())

	tests := []test{
		{
			name:      "test1",
			tableName: "fOo",
			jsonSchema: `{
							"bsonType": "object",
							"description": "WORLD",
							"required": ["a", "B", "c"],
							"properties": {
								"a": {"oneOf": [
										 {"bsonType": "long"},
										 {"bsonType": "null"}
									  ]
								},
								"B": {"oneOf": [
										 {"bsonType": "string"},
										 {"bsonType": "null"}
									  ],
									  "description": "fooo"
								},
								"c": {"bsonType": "string",
									  "description": "HELLO!"
								}
							}
						}`,
			indexes: []indexInfo{
				{
					Name:   "bar",
					Unique: true,
					Key: bson.D{
						{Key: "a", Value: int32(1)},
						{Key: "B", Value: int32(-1)},
					}},
				{
					Name:   "b_text_c_text",
					Unique: false,
					Key:    bson.D{},
					Weights: bson.D{
						{Key: "B", Value: int32(1)},
						{Key: "c", Value: int32(1)},
					},
				},
			},
			expected: testTable(
				logger,
				"fOo",
				"fOo",
				[]bson.D{},
				[]*schema.Column{
					schema.NewColumn("a", schema.SQLInt, "a", schema.MongoInt64, true, option.NoneString()),
					schema.NewColumn("b", schema.SQLVarchar, "B", schema.MongoString, true, option.SomeString("fooo")),
					schema.NewColumn("c", schema.SQLVarchar, "c", schema.MongoString, false, option.SomeString("HELLO!")),
				},
				[]schema.Index{
					schema.NewIndex("bar", true, false,
						[]schema.IndexPart{schema.NewIndexPart("a", 1), schema.NewIndexPart("B", -1)},
					),
					schema.NewIndex("b_text_c_text", false, true,
						[]schema.IndexPart{schema.NewIndexPart("B", 1), schema.NewIndexPart("c", 1)},
					),
				},
				option.SomeString("WORLD"),
			),
		},

		{
			name:      "remove invalid indexes",
			tableName: "fOo",
			jsonSchema: `{
							"bsonType": "object",
							"description": "WORLD",
							"required": ["a", "B", "c"],
							"properties": {
								"a": {"oneOf": [
										 {"bsonType": "long"},
										 {"bsonType": "null"}
									  ]
								},
								"B": {"oneOf": [
										 {"bsonType": "string"},
										 {"bsonType": "null"}
									  ],
									  "description": "fooo"
								},
								"c": {"bsonType": "string",
									  "description": "HELLO!"
								}
							}
						}`,
			indexes: []indexInfo{
				{
					Name:   "bar",
					Unique: true,
					Key: bson.D{
						{Key: "a", Value: int32(1)},
						{Key: "B", Value: int32(-1)},
					},
				},
				{
					Name:   "b",
					Unique: true,
					Key: bson.D{
						{Key: "aaaa", Value: int32(1)},
						{Key: "B", Value: int32(-1)},
					},
				},
				{
					Name:   "c",
					Unique: true,
					Key: bson.D{
						{Key: "a", Value: int32(1)},
						{Key: "Bdf", Value: int32(-1)},
					},
				},
				{
					Name:   "b_text_c_text",
					Unique: false,
					Key:    bson.D{},
					Weights: bson.D{
						{Key: "B", Value: int32(1)},
						{Key: "c", Value: int32(1)},
					},
				},
			},
			expected: testTable(
				logger,
				"fOo",
				"fOo",
				[]bson.D{},
				[]*schema.Column{
					schema.NewColumn("a", schema.SQLInt, "a", schema.MongoInt64, true, option.NoneString()),
					schema.NewColumn("b", schema.SQLVarchar, "B", schema.MongoString, true, option.SomeString("fooo")),
					schema.NewColumn("c", schema.SQLVarchar, "c", schema.MongoString, false, option.SomeString("HELLO!")),
				},
				[]schema.Index{
					schema.NewIndex("bar", true, false,
						[]schema.IndexPart{schema.NewIndexPart("a", 1), schema.NewIndexPart("B", -1)},
					),
					schema.NewIndex("b_text_c_text", false, true,
						[]schema.IndexPart{schema.NewIndexPart("B", 1), schema.NewIndexPart("c", 1)},
					),
				},
				option.SomeString("WORLD"),
			),
		},
	}

	sampler := NewSampler(nil, logger, nil)
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			req := require.New(t)
			bsonJSONSchema := bson.D{}
			err := bson.UnmarshalExtJSON([]byte(tst.jsonSchema), false, &bsonJSONSchema)
			req.Nil(err)
			out, err := sampler.deserializeTableSchema(tst.tableName, bsonJSONSchema, tst.indexes)
			req.Nil(err)
			req.Equal(tst.expected, out, fmt.Sprintf("Expected %#v, got %#v", tst.expected, out))
		})
	}
}

func TestDeserializeTableSchemaFailures(t *testing.T) {
	type test struct {
		tableName     string
		jsonSchema    string
		expectedError string
	}

	logger := log.NewComponentLogger(log.SchemaComponent, log.GlobalLogger())

	tests := []test{
		{
			tableName: "fOo",
			jsonSchema: `{
							"bsonType": "foo",
							"description": "WORLD",
							"required": ["a", "c"],
							"properties": {
								"a": {"oneOf": [
										 {"bsonType": "long"},
										 {"bsonType": "null"}
									  ]
								},
								"B": {"oneOf": [
										 {"bsonType": "string"},
										 {"bsonType": "null"}
									  ],
									  "description": "fooo"
								},
								"c": {"bsonType": "string",
									  "description": "HELLO!"
								}
							}
						}`,
			expectedError: "jsonSchema must have bsonType 'object'",
		},
		{
			tableName: "fOo",
			jsonSchema: `{
							"bsonType": "object",
							"description": 42,
							"required": ["a", "b", "c"],
							"properties": {
								"a": {"oneOf": [
										 {"bsonType": "long"},
										 {"bsonType": "null"}
									  ]
								},
								"B": {"oneOf": [
										 {"bsonType": "string"},
										 {"bsonType": "null"}
									  ],
									  "description": "fooo"
								},
								"c": {"bsonType": "string",
									  "description": "HELLO!"
								}
							}
						}`,
			expectedError: "jsonSchema 'description' must be a string, not int32",
		},
		{
			tableName: "fOo",
			jsonSchema: `{
							"bsonType": "object",
							"description": "WORLD",
							"required": ["a", 42, "c"],
							"properties": {
								"a": {"oneOf": [
										 {"bsonType": "long"},
										 {"bsonType": "null"}
									  ]
								},
								"B": {"oneOf": [
										 {"bsonType": "string"},
										 {"bsonType": "null"}
									  ],
									  "description": "fooo"
								},
								"c": {"bsonType": "string",
									  "description": "HELLO!"
								}
							}
						}`,
			expectedError: "jsonSchema 'required' elements must be a string, not int32",
		},
		{
			tableName: "fOo",
			jsonSchema: `{
							"bsonType": "object",
							"description": "WORLD",
							"required": [],
							"properties": {
								"a": {"oneOf": [
										 {"bsonType": "long"},
										 {"bsonType": "null"}
									  ]
								},
								"B": {"oneOf": [
										 {"bsonType": "string"},
										 {"bsonType": "null"}
									  ],
									  "description": "fooo"
								},
								"c": {"bsonType": "string",
									  "description": "HELLO!"
								}
							}
						}`,
			expectedError: "all properties must be required",
		},
		{
			tableName: "fOo",
			jsonSchema: `{
							"bsonType": "object",
							"description": "WORLD",
							"required": ["a"],
							"properties": {
							}
						}`,
			expectedError: "jsonSchema must have at least one property for writeMode schema mapping",
		},
		{
			tableName: "fOo",
			jsonSchema: `{
							"bsonType": "object",
							"description": "WORLD",
							"required": 42,
							"properties": {
							}
						}`,
			expectedError: "jsonSchema 'required' must be a primitive.A, not int32",
		},
		{
			tableName: "fOo",
			jsonSchema: `{
							"bsonType": "object",
							"description": "WORLD",
							"required": ["a", "b", "c"],
							"properties": 42
						}`,
			expectedError: "jsonSchema 'properties' must be a bson.D, not int32",
		},
		{
			tableName: "fOo",
			jsonSchema: `{
							"bsonType": "object",
							"description": "WORLD",
							"required": ["a", "c"],
							"properties": {
								"a": {"oneOf": [
										 {"bsonType": "long"},
										 {"bsonType": "null"}
									  ]
								},
								"B": {"oneOf": [
										 {"bsonType": "string"},
										 {"bsonType": "null"}
									  ],
									  "description": "fooo"
								},
								"c": {"bsonType": "string",
									  "description": "HELLO!"
								}
							}
						}`,
			expectedError: "all properties must be required",
		},
		{
			tableName: "fOo",
			jsonSchema: `{
							"bsonType": "object",
							"description": "WORLD",
							"required": ["a", "b", "c"],
							"properties": {
								"a": {"oneOf": [
										 {"bsonType": "long"},
										 {"bsonType": "null"}
									  ]
								},
								"B": {"oneOf": [
										 {"bsonType": "string"},
										 {"bsonType": "null"}
									  ],
									  "description": "fooo"
								},
								"c": {"bsonType": "string",
									  "description": "HELLO!"
								}
							}
						}`,
			expectedError: `found property named 'B' that is not in the required properties: []string{"a", "b", "c"}`,
		},
		{
			tableName: "fOo",
			jsonSchema: `{
							"bsonType": "object",
							"description": "WORLD",
							"required": ["a", "B", "c"],
							"properties": {
								"a": 42,
								"B": {"oneOf": [
										 {"bsonType": "string"},
										 {"bsonType": "null"}
									  ],
									  "description": "fooo"
								},
								"c": {"bsonType": "string",
									  "description": "HELLO!"
								}
							}
						}`,
			expectedError: "property must have a bsonType object for its value, found int32",
		},
		{
			tableName: "fOo",
			jsonSchema: `{
							"bsonType": "object",
							"description": "WORLD",
							"required": ["a", "b", "c"],
							"properties": {
								"a": {"oneOf": [
										 {"bsonType": "long"},
										 {"bsonType": "null"}
									  ]
								},
								"b": {"bsonType": 42,
									  "description": "fooo"
								},
								"c": {"bsonType": "string",
									  "description": "HELLO!"
								}
							}
						}`,
			expectedError: "bsonType must be a string denoting the type, found int32",
		},
		{
			tableName: "fOo",
			jsonSchema: `{
							"bsonType": "object",
							"description": "WORLD",
							"required": ["a", "B", "c"],
							"properties": {
								"a": {"oneOf": [
										 {"bsonType": "long"},
										 {"bsonType": "null"}
									  ]
								},
								"B": {"oneOf": 42,
									  "description": "fooo"
								},
								"c": {"bsonType": "string",
									  "description": "HELLO!"
								}
							}
						}`,
			expectedError: "'oneOf' argument must be a primitive.A, found int32",
		},
		{
			tableName: "fOo",
			jsonSchema: `{
							"bsonType": "object",
							"description": "WORLD",
							"required": ["a", "B", "c"],
							"properties": {
								"a": {"oneOf": [
										 {"bsonType": "long"},
										 {"bsonType": "null"}
									  ]
								},
								"B": {"oneOf": [
										 {"bsonType": "flibbity"},
										 {"bsonType": "null"}
									  ],
									  "description": "fooo"
								},
								"c": {"bsonType": "string",
									  "description": "HELLO!"
								}
							}
						}`,
			expectedError: "unsupported bsonType 'flibbity' for writeMode jsonSchema validator property",
		},
		{
			tableName: "fOo",
			jsonSchema: `{
							"bsonType": "object",
							"description": "WORLD",
							"required": ["a", "B", "c"],
							"properties": {
								"a": {"oneOf": [
										 {"bsonType": "long"},
										 {"bsonType": "null"}
									  ]
								},
								"B": {"oneOf": [
										 {"bsonType": "double"},
										 {"bsonType": "null"},
										 {"bsonType": "int"}
									  ],
									  "description": "fooo"
								},
								"c": {"bsonType": "string",
									  "description": "HELLO!"
								}
							}
						}`,
			expectedError: "the only supported value for 'oneOf' is a simple bsonType with null, thus there should be two elements, not 3",
		},
		{
			tableName: "fOo",
			jsonSchema: `{
							"bsonType": "object",
							"description": "WORLD",
							"required": ["a", "B", "c"],
							"properties": {
								"a": {"oneOf": [
										 {"bsonType": "long"},
										 {"bsonType": "null"}
									  ]
								},
								"B": {"oneOf": [
										 {"bsonType": "double"},
										 {"bsonType": "long"}
									  ],
									  "description": "fooo"
								},
								"c": {"bsonType": "string",
									  "description": "HELLO!"
								}
							}
						}`,
			expectedError: "'oneOf' is only supported when one of the bsonTypes is 'null', but found 'float' and 'int'",
		},
		{
			tableName: "fOo",
			jsonSchema: `{
							"bsonType": "object",
							"description": "WORLD",
							"required": ["a", "B", "c"],
							"properties": {
								"a": {"oneOf": [
										 {"bsonType": "long"},
										 {"bsonType": "null"}
									  ]
								},
								"B": {"oneOf": [
										 {"bsonType": "object"},
										 {"bsonType": "null"}
									  ],
									  "description": "fooo"
								},
								"c": {"bsonType": "string",
									  "description": "HELLO!"
								}
							}
						}`,
			expectedError: "unsupported bsonType 'object' for writeMode jsonSchema validator property",
		},
	}

	sampler := NewSampler(nil, logger, nil)
	for _, tst := range tests {
		t.Run(tst.expectedError, func(t *testing.T) {
			req := require.New(t)
			bsonJSONSchema := bson.D{}
			err := bson.UnmarshalExtJSON([]byte(tst.jsonSchema), false, &bsonJSONSchema)
			req.Nil(err)
			_, err = sampler.deserializeTableSchema(tst.tableName, bsonJSONSchema, []indexInfo{})
			req.NotNil(err)
			req.Equal(tst.expectedError, err.Error(), fmt.Sprintf("Expected %#v, got %#v", tst.expectedError, err.Error()))
		})
	}
}
