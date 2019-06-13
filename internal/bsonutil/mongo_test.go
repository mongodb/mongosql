package bsonutil_test

import (
	"testing"

	. "github.com/10gen/sqlproxy/internal/bsonutil"

	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson"
)

func TestPipelineToMapSlice(t *testing.T) {
	req := require.New(t)

	type test struct {
		name     string
		pipeline []bson.D
		mapped   []map[string]interface{}
	}

	tests := []test{
		{"empty", []bson.D{}, []map[string]interface{}{}},
		{"empty bsonD", []bson.D{NewD()}, []map[string]interface{}{{}}},
		{"simple bsonD", []bson.D{NewD(NewDocElem("a", "b"))}, []map[string]interface{}{{"a": "b"}}},
		{
			"nested bsonD",
			[]bson.D{NewD(NewDocElem("a", NewD(NewDocElem("b", "c"))))},
			[]map[string]interface{}{{"a": map[string]interface{}{"b": "c"}}},
		},
		{
			"mixed",
			[]bson.D{
				NewD(NewDocElem("a",
					NewM(NewDocElem("b",
						NewD(NewDocElem("c",
							NewArray(
								"d",
								NewD(NewDocElem("e", "f")),
								NewM(NewDocElem("g", "h")),
							),
						)),
					)),
				)),
			},
			[]map[string]interface{}{
				{
					"a": map[string]interface{}{
						"b": map[string]interface{}{
							"c": []interface{}{
								"d",
								map[string]interface{}{"e": "f"},
								map[string]interface{}{"g": "h"},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req.Equal(test.mapped, PipelineToMapSlice(test.pipeline))
		})
	}
}
