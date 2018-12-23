package bsonutil_test

import (
	"fmt"
	"testing"

	. "github.com/10gen/sqlproxy/internal/bsonutil"

	"github.com/stretchr/testify/require"
)

func TestConvertBSONToMap(t *testing.T) {
	req := require.New(t)

	type test struct {
		name     string
		document interface{}
		mapped   map[string]interface{}
	}

	runTests := func(tests []test) {
		for _, test := range tests {
			t.Run(fmt.Sprintf("convert_bson_to_map_%s", test.name), func(t *testing.T) {
				req.Equal(test.mapped, ConvertBSONToMap(test.document))
			})
		}
	}

	tests := []test{
		{"empty", NewD(), map[string]interface{}{}},
		{"bsonD_simple", NewD(NewDocElem("a", "b")), map[string]interface{}{"a": "b"}},
		{"bsonM_simple", NewM(NewDocElem("a", "b")), map[string]interface{}{"a": "b"}},
		{
			"mixed", NewD(
				NewDocElem("a", NewM(NewDocElem("z", NewD(NewDocElem("b", "c"))))),
			), map[string]interface{}{
				"a": map[string]interface{}{
					"z": map[string]interface{}{"b": "c"},
				},
			},
		},
		{
			"bsonD_nested", NewD(NewDocElem("a", NewD(NewDocElem("b", "c")))),
			map[string]interface{}{"a": map[string]interface{}{"b": "c"}},
		},
		{
			"bsonM_nested", NewM(NewDocElem("a", NewM(NewDocElem("b", "c")))),
			map[string]interface{}{"a": map[string]interface{}{"b": "c"}},
		},
	}

	runTests(tests)
}
