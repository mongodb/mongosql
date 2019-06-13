package mongo_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/schema/mapping"
	"github.com/10gen/sqlproxy/schema/mongo"
	. "github.com/smartystreets/goconvey/convey"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMergeSchema(t *testing.T) {
	Convey("Given an empty object schema", t, func() {
		schema := mongo.NewCollectionSchema()

		Convey("Merging it with a schema of any other type should fail", func() {
			otherTypes := []mongo.BSONType{
				mongo.Int, mongo.Long, mongo.Double, mongo.Decimal, mongo.BinData,
				mongo.Boolean, mongo.String, mongo.Date, mongo.ObjectID, mongo.Array,
			}
			for _, typ := range otherTypes {
				other := &mongo.Schema{
					BSONType: typ,
				}
				err := schema.Merge(other)
				So(err, ShouldNotBeNil)
			}
		})

		Convey("Merging it with a non-empty object schema", func() {
			other, err := mongo.NewSchemaFromValue(bsonutil.NewD(
				bsonutil.NewDocElem("a", int32(1)),
				bsonutil.NewDocElem("b", int32(2)),
			))
			So(err, ShouldBeNil)

			err = schema.Merge(other)
			So(err, ShouldBeNil)

			Convey("Should give a schema resembling the non-empty one", func() {
				jsonActual, err := schema.JSONSchema()
				So(err, ShouldBeNil)
				jsonExpected, err := other.JSONSchema()
				So(err, ShouldBeNil)
				So(jsonActual, ShouldEqual, jsonExpected)
			})

			Convey("Merging it with another non-empty object schema", func() {
				other, err := mongo.NewSchemaFromValue(bsonutil.NewD(
					bsonutil.NewDocElem("b", int32(1)),
					bsonutil.NewDocElem("c", int32(5)),
					bsonutil.NewDocElem("d", int64(2)),
				))
				So(err, ShouldBeNil)

				err = schema.Merge(other)
				So(err, ShouldBeNil)

				Convey("Should give a schema resembling the union of all three objects", func() {
					expected := &mongo.Schema{
						BSONType: mongo.Object,
						Properties: map[string]*mongo.Schemata{
							"a": mongo.NewSchemata(&mongo.Schema{BSONType: mongo.Int}),
							"b": mongo.NewSchemata(&mongo.Schema{BSONType: mongo.Int}),
							"c": mongo.NewSchemata(&mongo.Schema{BSONType: mongo.Int}),
							"d": mongo.NewSchemata(&mongo.Schema{BSONType: mongo.Long}),
						},
					}
					expected.Properties["b"].Counts["int"] = 2

					jsonActual, err := schema.JSONSchema()
					So(err, ShouldBeNil)
					jsonExpected, err := expected.JSONSchema()
					So(err, ShouldBeNil)
					So(jsonActual, ShouldEqual, jsonExpected)
				})
			})
		})
	})

	Convey("Given an empty array schema", t, func() {
		schema, err := mongo.NewArraySchema([]interface{}{})
		So(err, ShouldBeNil)

		Convey("Merging it with a schema of any other type should fail", func() {
			otherTypes := []mongo.BSONType{
				mongo.Int, mongo.Long, mongo.Double, mongo.Decimal, mongo.BinData,
				mongo.Boolean, mongo.String, mongo.Date, mongo.ObjectID, mongo.Object,
			}
			for _, typ := range otherTypes {
				other := &mongo.Schema{
					BSONType: typ,
				}
				err := schema.Merge(other)
				So(err, ShouldNotBeNil)
			}
		})

		Convey("Merging it with a non-empty array schema", func() {
			other, err := mongo.NewArraySchema([]interface{}{true, true, false})
			So(err, ShouldBeNil)

			err = schema.Merge(other)
			So(err, ShouldBeNil)

			jsonActual, err := schema.JSONSchema()
			So(err, ShouldBeNil)
			jsonExpected, err := other.JSONSchema()
			So(err, ShouldBeNil)

			Convey("Should give a schema resembling the non-empty one", func() {
				So(jsonActual, ShouldEqual, jsonExpected)
			})

			Convey("And merging with an array schema of another type", func() {
				other, err := mongo.NewArraySchema([]interface{}{"abc"})
				So(err, ShouldBeNil)

				err = schema.Merge(other)
				So(err, ShouldBeNil)

				Convey("Should result in a superposition of the two array schemas", func() {
					newExpected := &mongo.Schema{
						BSONType: mongo.Array,
						Items: &mongo.Schemata{
							Schemas: map[mongo.BSONType]*mongo.Schema{
								mongo.Boolean: {BSONType: mongo.Boolean},
								mongo.String:  {BSONType: mongo.String},
							},
							Counts: map[mongo.BSONType]int{
								mongo.Boolean: 3,
								mongo.String:  1,
							},
						},
					}

					newJSONExpected, err := newExpected.JSONSchema()
					So(err, ShouldBeNil)

					newJSONActual, err := schema.JSONSchema()
					So(err, ShouldBeNil)

					So(newJSONActual, ShouldEqual, newJSONExpected)
				})
			})

			Convey("And merging it with an array schema of a different type", func() {
				other, err := mongo.NewArraySchema([]interface{}{"abc", "def", "ghi", "jkl"})
				So(err, ShouldBeNil)

				err = schema.Merge(other)
				So(err, ShouldBeNil)

				Convey("Should result in a superposition of the two array schemas", func() {
					newExpected := &mongo.Schema{
						BSONType: mongo.Array,
						Items: &mongo.Schemata{
							Schemas: map[mongo.BSONType]*mongo.Schema{
								mongo.Boolean: {BSONType: mongo.Boolean},
								mongo.String:  {BSONType: mongo.String},
							},
							Counts: map[mongo.BSONType]int{
								mongo.Boolean: 3,
								mongo.String:  4,
							},
						},
					}

					newJSONExpected, err := newExpected.JSONSchema()
					So(err, ShouldBeNil)

					newJSONActual, err := schema.JSONSchema()
					So(err, ShouldBeNil)

					So(newJSONActual, ShouldEqual, newJSONExpected)
				})
			})
		})
	})
}

func TestRenderJSONSchema(t *testing.T) {
	schema := mongo.NewCollectionSchema()

	err := schema.IncludeSample(bsonutil.NewD(
		bsonutil.NewDocElem("scalar", int32(1)),
		bsonutil.NewDocElem("array", bsonutil.NewArray(
			"a",
			"b",
			"c",
		)),
		bsonutil.NewDocElem("object", bsonutil.NewD(
			bsonutil.NewDocElem("bool", true),
		)),
		bsonutil.NewDocElem("nil", nil),
		bsonutil.NewDocElem("unique_id", primitive.Binary{Subtype: 0x03}),
		bsonutil.NewDocElem("nestedarray", bsonutil.NewArray(
			bsonutil.NewArray(true, false, false),
			bsonutil.NewArray(true, true, true),
		)),
	))
	if err != nil {
		t.Fatalf("Failed including sample: %v", err)
	}

	json, err := schema.JSONSchema()
	if err != nil {
		t.Fatalf("Failed rendering JSON Schema: %v", err)
	}

	actual := fmt.Sprintf("\n%s", json)
	expected := `
{
    "bsonType": "object",
    "properties": {
        "array": {
            "schemas": {
                "array": {
                    "bsonType": "array",
                    "items": {
                        "schemas": {
                            "string": {
                                "bsonType": "string"
                            }
                        },
                        "counts": {
                            "string": 3
                        }
                    }
                }
            },
            "counts": {
                "array": 1
            }
        },
        "nestedarray": {
            "schemas": {
                "array": {
                    "bsonType": "array",
                    "items": {
                        "schemas": {
                            "array": {
                                "bsonType": "array",
                                "items": {
                                    "schemas": {
                                        "bool": {
                                            "bsonType": "bool"
                                        }
                                    },
                                    "counts": {
                                        "bool": 6
                                    }
                                }
                            }
                        },
                        "counts": {
                            "array": 2
                        }
                    }
                }
            },
            "counts": {
                "array": 1
            }
        },
        "nil": {
            "schemas": {
                "null": {
                    "bsonType": "null"
                }
            },
            "counts": {
                "null": 1
            }
        },
        "object": {
            "schemas": {
                "object": {
                    "bsonType": "object",
                    "properties": {
                        "bool": {
                            "schemas": {
                                "bool": {
                                    "bsonType": "bool"
                                }
                            },
                            "counts": {
                                "bool": 1
                            }
                        }
                    }
                }
            },
            "counts": {
                "object": 1
            }
        },
        "scalar": {
            "schemas": {
                "int": {
                    "bsonType": "int"
                }
            },
            "counts": {
                "int": 1
            }
        },
        "unique_id": {
            "schemas": {
                "binData": {
                    "bsonType": "binData",
                    "specialType": "uuid3"
                }
            },
            "counts": {
                "binData": 1
            }
        }
    }
}`

	spacelessExpected := strings.Replace(expected, " ", "", -1)
	spacelessActual := strings.Replace(actual, " ", "", -1)
	if spacelessActual != spacelessExpected {
		t.Fatalf("Expected:\n%s\n\nActual:\n%s\n", expected, actual)
	}
}

func TestSampling(t *testing.T) {

	Convey("Given an empty collection", t, func() {
		collection := mongo.NewCollectionSchema()
		So(collection, shouldHaveType, mongo.Object)

		Convey("Including a flat document", func() {
			err := collection.IncludeSample(bsonutil.NewD(
				bsonutil.NewDocElem("a", int32(1)),
				bsonutil.NewDocElem("b", int32(1)),
			))
			So(err, ShouldBeNil)

			Convey("Should result in 2 properties", func() {
				So(collection.Properties, ShouldHaveLength, 2)

				So(collection.Properties["a"], shouldHaveSampleCount, 1)
				So(collection.Properties["a"], shouldHaveCandidateTypes, mongo.Int)

				So(collection.Properties["b"], shouldHaveSampleCount, 1)
				So(collection.Properties["b"], shouldHaveCandidateTypes, mongo.Int)
			})

			Convey("And including an additional flat document with the same fields", func() {
				err = collection.IncludeSample(bsonutil.NewD(
					bsonutil.NewDocElem("a", int32(1)),
					bsonutil.NewDocElem("b", int32(1)),
				))
				So(err, ShouldBeNil)

				Convey("Should increase the 2 properties' sample counts", func() {
					So(collection.Properties, ShouldHaveLength, 2)

					So(collection.Properties["a"], shouldHaveSampleCount, 2)
					So(collection.Properties["a"], shouldHaveCandidateTypes, mongo.Int)

					So(collection.Properties["b"], shouldHaveSampleCount, 2)
					So(collection.Properties["b"], shouldHaveCandidateTypes, mongo.Int)
				})
			})

			Convey("And including an additional flat document with different fields", func() {
				err = collection.IncludeSample(bsonutil.NewD(
					bsonutil.NewDocElem("c", int32(2)),
					bsonutil.NewDocElem("d", int32(2)),
				))
				So(err, ShouldBeNil)

				Convey("Should result in 4 properties", func() {
					So(collection.Properties, ShouldHaveLength, 4)

					So(collection.Properties["a"], shouldHaveSampleCount, 1)
					So(collection.Properties["a"], shouldHaveCandidateTypes, mongo.Int)

					So(collection.Properties["b"], shouldHaveSampleCount, 1)
					So(collection.Properties["b"], shouldHaveCandidateTypes, mongo.Int)

					So(collection.Properties["c"], shouldHaveSampleCount, 1)
					So(collection.Properties["c"], shouldHaveCandidateTypes, mongo.Int)

					So(collection.Properties["d"], shouldHaveSampleCount, 1)
					So(collection.Properties["d"], shouldHaveCandidateTypes, mongo.Int)
				})
			})

			Convey("And including another flat document with same fields but other types", func() {
				err = collection.IncludeSample(bsonutil.NewD(
					bsonutil.NewDocElem("a", "string"),
					bsonutil.NewDocElem("b", 3.2),
				))
				So(err, ShouldBeNil)
				err = collection.IncludeSample(bsonutil.NewD(
					bsonutil.NewDocElem("a", bsonutil.NewArray(
						"string",
						int32(1),
					)),
					bsonutil.NewDocElem("b", bsonutil.NewD(bsonutil.NewDocElem("c", int64(1)))),
				))
				So(err, ShouldBeNil)

				Convey("Should result in 2 properties", func() {
					So(collection.Properties, ShouldHaveLength, 2)
				})

				Convey("Should result in each field having multiple candidate types", func() {
					So(
						collection.Properties["a"],
						shouldHaveCandidateTypes,
						mongo.Array, mongo.Int, mongo.String,
					)
					So(
						collection.Properties["b"],
						shouldHaveCandidateTypes,
						mongo.Object, mongo.Int, mongo.Double,
					)
				})

			})
		})

		Convey("Including a document with a nested document", func() {
			err := collection.IncludeSample(bsonutil.NewD(
				bsonutil.NewDocElem("a", int32(1)),
				bsonutil.NewDocElem("b", bsonutil.NewD(
					bsonutil.NewDocElem("c", int32(1)),
					bsonutil.NewDocElem("d", int32(1)),
				)),
			))
			So(err, ShouldBeNil)

			Convey("Should result in 2 properties", func() {
				So(collection.Properties, ShouldHaveLength, 2)

				So(collection.Properties["a"], shouldHaveSampleCount, 1)
				So(collection.Properties["a"], shouldHaveCandidateTypes, mongo.Int)

				So(collection.Properties["b"], shouldHaveSampleCount, 1)
				So(collection.Properties["b"], shouldHaveCandidateTypes, mongo.Object)

				Convey("And the nested document should have 2 properties", func() {
					doc := mapping.PolymorphicMajorityCountHeuristic(collection.Properties["b"])
					So(doc[0], shouldHaveType, mongo.Object)

					So(doc[0].Properties, ShouldHaveLength, 2)

					So(doc[0].Properties["c"], shouldHaveSampleCount, 1)
					So(doc[0].Properties["c"], shouldHaveCandidateTypes, mongo.Int)

					So(doc[0].Properties["d"], shouldHaveSampleCount, 1)
					So(doc[0].Properties["d"], shouldHaveCandidateTypes, mongo.Int)
				})
			})

			Convey("Including another doc with same fields but another type in the subdoc", func() {
				err = collection.IncludeSample(bsonutil.NewD(
					bsonutil.NewDocElem("a", int32(1)),
					bsonutil.NewDocElem("b", bsonutil.NewD(
						bsonutil.NewDocElem("c", "string"),
						bsonutil.NewDocElem("d", int64(1)),
					)),
				))
				So(err, ShouldBeNil)

				Convey("Should result in 2 properties", func() {
					So(collection.Properties, ShouldHaveLength, 2)

					So(collection.Properties["a"], shouldHaveSampleCount, 2)
					So(collection.Properties["a"], shouldHaveCandidateTypes, mongo.Int)

					So(collection.Properties["b"], shouldHaveSampleCount, 2)
					So(collection.Properties["b"], shouldHaveCandidateTypes, mongo.Object)

					Convey("And the nested document should have 2 properties", func() {
						doc := mapping.PolymorphicMajorityCountHeuristic(collection.Properties["b"])
						So(doc[0], shouldHaveType, mongo.Object)

						So(doc[0].Properties, ShouldHaveLength, 2)

						So(doc[0].Properties["c"], shouldHaveSampleCount, 2)
						So(doc[0].Properties["c"], shouldHaveCandidateTypes, mongo.Int,
							mongo.String)

						So(doc[0].Properties["d"], shouldHaveSampleCount, 2)
						So(doc[0].Properties["d"], shouldHaveCandidateTypes, mongo.Int, mongo.Long)
					})
				})
			})

			Convey("And including an additional document with a different structure", func() {
				err = collection.IncludeSample(bsonutil.NewD(
					bsonutil.NewDocElem("c", int32(1)),
					bsonutil.NewDocElem("b", bsonutil.NewD(
						bsonutil.NewDocElem("c", "string"),
						bsonutil.NewDocElem("e", int32(1)),
					)),
				))
				So(err, ShouldBeNil)

				Convey("Should result in 3 properties", func() {
					So(collection.Properties, ShouldHaveLength, 3)

					So(collection.Properties["a"], shouldHaveSampleCount, 1)
					So(collection.Properties["a"], shouldHaveCandidateTypes, mongo.Int)

					So(collection.Properties["b"], shouldHaveSampleCount, 2)
					So(collection.Properties["b"], shouldHaveCandidateTypes, mongo.Object)

					So(collection.Properties["c"], shouldHaveSampleCount, 1)
					So(collection.Properties["c"], shouldHaveCandidateTypes, mongo.Int)

					Convey("And the nested document should have 3 properties", func() {
						doc := mapping.PolymorphicMajorityCountHeuristic(collection.Properties["b"])
						So(doc[0].BSONType, ShouldEqual, mongo.Object)

						So(doc[0].Properties, ShouldHaveLength, 3)

						So(doc[0].Properties["c"], shouldHaveSampleCount, 2)
						So(doc[0].Properties["c"], shouldHaveCandidateTypes, mongo.String,
							mongo.Int)

						So(doc[0].Properties["d"], shouldHaveSampleCount, 1)
						So(doc[0].Properties["d"], shouldHaveCandidateTypes, mongo.Int)

						So(doc[0].Properties["e"], shouldHaveSampleCount, 1)
						So(doc[0].Properties["e"], shouldHaveCandidateTypes, mongo.Int)
					})
				})
			})
		})

		Convey("Including a document with a homogenous array", func() {
			err := collection.IncludeSample(bsonutil.NewD(
				bsonutil.NewDocElem("a", int32(1)),
				bsonutil.NewDocElem("b", bsonutil.NewArray(
					"a",
					"b",
					"c",
				)),
			))
			So(err, ShouldBeNil)

			Convey("Should result in 2 properties", func() {
				So(collection.Properties, ShouldHaveLength, 2)

				So(collection.Properties["a"], shouldHaveSampleCount, 1)
				So(collection.Properties["a"], shouldHaveCandidateTypes, mongo.Int)

				So(collection.Properties["b"], shouldHaveSampleCount, 1)
				So(collection.Properties["b"], shouldHaveCandidateTypes, mongo.Array)

				Convey("The array should have 1 candidate type with the right count", func() {
					array := mapping.PolymorphicMajorityCountHeuristic(collection.Properties["b"])
					So(array[0], shouldHaveType, mongo.Array)

					So(array[0].Items, shouldHaveSampleCount, 3)
					So(array[0].Items, shouldHaveCandidateTypes, mongo.String)
				})
			})

			Convey("And including an additional document with the same structure", func() {
				err = collection.IncludeSample(bsonutil.NewD(
					bsonutil.NewDocElem("a", int32(1)),
					bsonutil.NewDocElem("b", bsonutil.NewArray(
						"b",
						"c",
						"d",
					)),
				))
				So(err, ShouldBeNil)

				Convey("Should result in 2 properties", func() {

					So(collection.Properties, ShouldHaveLength, 2)

					So(collection.Properties["a"], shouldHaveSampleCount, 2)
					So(collection.Properties["a"], shouldHaveCandidateTypes, mongo.Int)

					So(collection.Properties["b"], shouldHaveSampleCount, 2)
					So(collection.Properties["b"], shouldHaveCandidateTypes, mongo.Array)

					Convey("Array should have 1 candidate type with right sample count", func() {
						array := mapping.PolymorphicMajorityCountHeuristic(
							collection.Properties["b"])
						So(array[0], shouldHaveType, mongo.Array)

						So(array[0].Items, shouldHaveSampleCount, 6)
						So(array[0].Items, shouldHaveCandidateTypes, mongo.String)
					})
				})
			})

			Convey("And including an additional document with a different structure", func() {
				err = collection.IncludeSample(bsonutil.NewD(
					bsonutil.NewDocElem("c", int32(1)),
					bsonutil.NewDocElem("b", bsonutil.NewArray(
						time.Now(),
					)),
				))
				So(err, ShouldBeNil)

				Convey("Should result in 3 properties", func() {
					So(collection.Properties, ShouldHaveLength, 3)

					So(collection.Properties["a"], shouldHaveSampleCount, 1)
					So(collection.Properties["a"], shouldHaveCandidateTypes, mongo.Int)

					So(collection.Properties["b"], shouldHaveSampleCount, 2)
					So(collection.Properties["b"], shouldHaveCandidateTypes, mongo.Array)

					So(collection.Properties["c"], shouldHaveSampleCount, 1)
					So(collection.Properties["c"], shouldHaveCandidateTypes, mongo.Int)

					Convey("And the array should have two candidate types", func() {
						array := mapping.PolymorphicMajorityCountHeuristic(
							collection.Properties["b"])
						So(array[0], shouldHaveType, mongo.Array)

						So(array[0].Items, shouldHaveSampleCount, 4)
						So(array[0].Items, shouldHaveCandidateTypes, mongo.Date, mongo.String)

						Convey("Dominant type for array should be chosen by sample freq", func() {
							So(array[0].Items, shouldHaveDominantType, mongo.String)
						})
					})
				})
			})
		})
	})
}

func TestValidateSchema(t *testing.T) {
	Convey("Given an empty schema", t, func() {
		schema := &mongo.Schema{}
		Convey("It should pass validation", func() {
			So(schema, shouldBeValidSchema)
		})

		scalars := []mongo.BSONType{
			mongo.Int, mongo.Long, mongo.Double, mongo.Decimal,
			mongo.Boolean, mongo.String, mongo.Date, mongo.ObjectID,
		}
		for _, scalar := range scalars {

			desc := fmt.Sprintf("Changing the BSONType to %s", scalar)
			Convey(desc, func() {
				schema.BSONType = scalar

				Convey("Should yield a valid schema", func() {
					So(schema, shouldBeValidSchema)
				})

				Convey("And making Items a non-nil Schemata", func() {
					schema.Items = &mongo.Schemata{}

					Convey("Should yield an invalid schema", func() {
						So(schema, shouldBeInvalidSchema)
					})
				})

				Convey("And making Properties a non-nil map", func() {
					schema.Properties = make(map[string]*mongo.Schemata)

					Convey("Should yield an invalid schema", func() {
						So(schema, shouldBeInvalidSchema)
					})
				})
			})
		}

		Convey("Changing the BSONType to BinData", func() {
			schema.BSONType = mongo.BinData

			Convey("Should yield a valid schema", func() {
				So(schema, shouldBeValidSchema)
			})

			validSpecialTypes := []mongo.SpecialType{
				mongo.UUID3, mongo.UUID4,
			}

			invalidSpecialTypes := []mongo.SpecialType{
				mongo.GeoPoint, "randomstring",
			}

			for _, st := range validSpecialTypes {
				desc := fmt.Sprintf("And setting the SpecialType to %s", st)
				Convey(desc, func() {
					schema.SpecialType = st

					Convey("Should yield a valid schema", func() {
						So(schema, shouldBeValidSchema)
					})
				})
			}

			for _, st := range invalidSpecialTypes {
				desc := fmt.Sprintf("And setting the SpecialType to %s", st)
				Convey(desc, func() {
					schema.SpecialType = st

					Convey("Should yield an invalid schema", func() {
						So(schema, shouldBeInvalidSchema)
					})
				})
			}
		})

		specialTypes := []mongo.SpecialType{
			mongo.GeoPoint, mongo.UUID3, mongo.UUID4, "aldkjfalksdfj",
		}
		for _, st := range specialTypes {
			desc := fmt.Sprintf("Changing the specialType to %s", st)

			Convey(desc, func() {
				schema.SpecialType = st
				Convey("Should yield an invalid schema", func() {
					So(schema, shouldBeInvalidSchema)
				})
			})
		}
	})

	Convey("Given a fresh object schema", t, func() {
		schema := mongo.NewCollectionSchema()
		Convey("It should pass validation", func() {
			So(schema, shouldBeValidSchema)
		})

		invalidSpecialTypes := []mongo.SpecialType{
			mongo.UUID3, mongo.UUID4, "randomstring",
		}
		for _, st := range invalidSpecialTypes {
			desc := fmt.Sprintf("Setting the SpecialType to %s", st)
			Convey(desc, func() {
				schema.SpecialType = st

				Convey("Should yield an invalid schema", func() {
					So(schema, shouldBeInvalidSchema)
				})
			})
		}

		validSpecialTypes := []mongo.SpecialType{
			mongo.GeoPoint,
		}
		for _, st := range validSpecialTypes {
			desc := fmt.Sprintf("Setting the SpecialType to %s", st)
			Convey(desc, func() {
				schema.SpecialType = st

				Convey("Should yield a valid schema", func() {
					So(schema, shouldBeValidSchema)
				})
			})
		}

		Convey("Setting a valid schemata as a property", func() {
			s := mongo.NewCollectionSchema()
			So(s, shouldBeValidSchema)

			prop := mongo.NewSchemata(s)
			So(prop, shouldBeValidSchemata)

			schema.Properties["validProp"] = prop

			Convey("Should yield a valid schema", func() {
				So(schema, shouldBeValidSchema)
			})

			Convey("And setting a valid schemata as another property", func() {
				s := mongo.NewCollectionSchema()
				So(s, shouldBeValidSchema)

				prop := mongo.NewSchemata(s)
				So(prop, shouldBeValidSchemata)

				schema.Properties["anotherValidProp"] = prop

				Convey("Should yield a valid schema", func() {
					So(schema, shouldBeValidSchema)
				})
			})
		})

		Convey("Setting an invalid schemata as a property", func() {
			s := mongo.NewCollectionSchema()
			s.SpecialType = mongo.UUID3
			So(s, shouldBeInvalidSchema)

			prop := mongo.NewSchemata(s)
			So(prop, shouldBeInvalidSchemata)

			schema.Properties["invalidProp"] = prop

			Convey("Should yield an invalid schema", func() {
				So(schema, shouldBeInvalidSchema)
			})

			Convey("And setting a valid schemata as another property", func() {
				s := mongo.NewCollectionSchema()
				So(s, shouldBeValidSchema)

				prop := mongo.NewSchemata(s)
				So(prop, shouldBeValidSchemata)

				schema.Properties["validProp"] = prop

				Convey("Should yield an invalid schema", func() {
					So(schema, shouldBeInvalidSchema)
				})
			})
		})
	})

	Convey("Given a fresh array schema", t, func() {
		schema, err := mongo.NewArraySchema([]interface{}{})
		So(err, ShouldBeNil)

		Convey("It should pass validation", func() {
			So(schema, shouldBeValidSchema)
		})

		invalidSpecialTypes := []mongo.SpecialType{
			mongo.UUID3, mongo.UUID4, "randomstring",
		}
		for _, st := range invalidSpecialTypes {
			desc := fmt.Sprintf("Setting the SpecialType to %s", st)
			Convey(desc, func() {
				schema.SpecialType = st

				Convey("Should yield an invalid schema", func() {
					So(schema, shouldBeInvalidSchema)
				})
			})
		}

		validSpecialTypes := []mongo.SpecialType{
			mongo.GeoPoint,
		}
		for _, st := range validSpecialTypes {
			desc := fmt.Sprintf("Setting the SpecialType to %s", st)
			Convey(desc, func() {
				schema.SpecialType = st

				Convey("Should yield a valid schema", func() {
					So(schema, shouldBeValidSchema)
				})
			})
		}

		Convey("Setting a valid schemata as Items", func() {
			s := mongo.NewCollectionSchema()
			So(s, shouldBeValidSchema)

			prop := mongo.NewSchemata(s)
			So(prop, shouldBeValidSchemata)

			schema.Items = prop

			Convey("Should yield a valid schema", func() {
				So(schema, shouldBeValidSchema)
			})
		})

		Convey("Setting an invalid schemata as Items", func() {
			s := mongo.NewCollectionSchema()
			s.SpecialType = mongo.UUID3
			So(s, shouldBeInvalidSchema)

			prop := mongo.NewSchemata(s)
			So(prop, shouldBeInvalidSchemata)

			schema.Items = prop

			Convey("Should yield an invalid schema", func() {
				So(schema, shouldBeInvalidSchema)
			})
		})
	})

	Convey("Given a fresh schemata", t, func() {
		schemata := mongo.NewSchemata(nil)

		Convey("It should pass validation", func() {
			So(schemata, shouldBeValidSchemata)
		})

		Convey("Including an invalid schema", func() {
			s, err := mongo.NewArraySchema([]interface{}{})
			So(err, ShouldBeNil)

			s.SpecialType = mongo.UUID3
			So(s, shouldBeInvalidSchema)

			err = schemata.IncludeSchema(s, 1)
			So(err, ShouldBeNil)

			Convey("Should yield an invalid schema", func() {
				So(schemata, shouldBeInvalidSchemata)
			})
		})

		Convey("Including a valid schema", func() {
			s := mongo.NewCollectionSchema()
			So(s, shouldBeValidSchema)

			err := schemata.IncludeSchema(s, 1)
			So(err, ShouldBeNil)

			Convey("Should yield a valid schema", func() {
				So(schemata, shouldBeValidSchemata)
			})
		})
	})
}

var shouldHaveType = func(actual interface{}, expected ...interface{}) string {
	schema, ok := actual.(*mongo.Schema)
	if !ok {
		return fmt.Sprintf("Expected arg of type *mongo.Schema, got a %T", actual)
	}
	actualType := schema.BSONType

	expectedType, ok := expected[0].(mongo.BSONType)
	if !ok {
		return fmt.Sprintf("Expected arg of type mongo.BSONType, got a %T", expected[0])
	}

	if actualType != expectedType {
		return fmt.Sprintf(
			"Expected schema's bsonType to be %s, but it was %s",
			expectedType, actualType,
		)
	}

	return ""
}

var shouldHaveSampleCount = func(actual interface{}, expected ...interface{}) string {
	schemata, ok := actual.(*mongo.Schemata)
	if !ok {
		return fmt.Sprintf("Expected arg of type *mongo.Schemata, got a %T", actual)
	}

	var actualSampleCount int
	for _, count := range schemata.Counts {
		actualSampleCount += count
	}

	expectedSampleCount, ok := expected[0].(int)
	if !ok {
		return fmt.Sprintf("Expected arg of type int got a %T", expected[0])
	}

	if actualSampleCount != expectedSampleCount {
		return fmt.Sprintf(
			"Expected schemata's sample count to be %d, but it was %d",
			expectedSampleCount, actualSampleCount,
		)
	}

	return ""
}

var shouldHaveDominantType = func(actual interface{}, expected ...interface{}) string {
	schemata, ok := actual.(*mongo.Schemata)
	if !ok {
		return fmt.Sprintf("Expected arg of type *mongo.Schemata, got a %T", actual)
	}
	actualDominantType := mapping.PolymorphicMajorityCountHeuristic(schemata)[0].BSONType

	expectedDominantType, ok := expected[0].(mongo.BSONType)
	if !ok {
		return fmt.Sprintf("Expected arg of type mongo.BSONType, got a %T", expected[0])
	}

	if actualDominantType != expectedDominantType {
		return fmt.Sprintf(
			"Expected schemata's dominant type to be %s, but it was %s",
			expectedDominantType, actualDominantType,
		)
	}

	return ""
}

var shouldHaveCandidateTypes = func(actual interface{}, expected ...interface{}) string {
	schemata, ok := actual.(*mongo.Schemata)
	if !ok {
		return fmt.Sprintf("Expected arg of type *mongo.Schemata, got a %T", actual)
	}
	actualCandidateTypes := make([]mongo.BSONType, 0, len(schemata.Schemas))
	for _, schema := range schemata.Schemas {
		actualCandidateTypes = append(actualCandidateTypes, schema.BSONType)
	}

	expectedCandidateTypes := make([]mongo.BSONType, 0, len(expected))
	for _, exp := range expected {
		typ, ok := exp.(mongo.BSONType)
		if !ok {
			return fmt.Sprintf("Expected arg of type mongo.BSONType, got a %T", exp)
		}
		expectedCandidateTypes = append(expectedCandidateTypes, typ)
	}

	success := true

	if len(expectedCandidateTypes) != len(actualCandidateTypes) {
		success = false
	}

	for _, exp := range expectedCandidateTypes {
		err := ShouldContain(actualCandidateTypes, exp)
		if err != "" {
			success = false
			break
		}
	}

	if !success {
		return fmt.Sprintf(
			"Expected schemata to have candidate types %v, but got %v",
			expectedCandidateTypes, actualCandidateTypes,
		)
	}
	return ""
}

var shouldBeValidSchema = func(actual interface{}, expected ...interface{}) string {
	schema, ok := actual.(*mongo.Schema)
	if !ok {
		return fmt.Sprintf("Expected arg of type *mongo.Schema, got a %T", actual)
	}

	err := schema.Validate()
	if err != nil {
		return fmt.Sprintf("Got error while validating schema: %s", err.Error())
	}
	return ""
}

var shouldBeInvalidSchema = func(actual interface{}, expected ...interface{}) string {
	schema, ok := actual.(*mongo.Schema)
	if !ok {
		return fmt.Sprintf("Expected arg of type *mongo.Schema, got a %T", actual)
	}

	err := schema.Validate()
	if err == nil {
		return fmt.Sprintf("Expected schema to fail validation, but it passed")
	}
	return ""
}

var shouldBeValidSchemata = func(actual interface{}, expected ...interface{}) string {
	schemata, ok := actual.(*mongo.Schemata)
	if !ok {
		return fmt.Sprintf("Expected arg of type *mongo.Schemata, got a %T", actual)
	}

	err := schemata.Validate()
	if err != nil {
		return fmt.Sprintf("Got error while validating schemata: %s", err.Error())
	}
	return ""
}

var shouldBeInvalidSchemata = func(actual interface{}, expected ...interface{}) string {

	schemata, ok := actual.(*mongo.Schemata)
	if !ok {
		return fmt.Sprintf("Expected arg of type *mongo.Schemata, got a %T", actual)
	}

	err := schemata.Validate()
	if err == nil {
		return fmt.Sprintf("Expected schemata to fail validation, but it passed")
	}
	return ""
}
