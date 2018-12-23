package bsonutil_test

import (
	"math"
	"testing"

	. "github.com/10gen/sqlproxy/internal/bsonutil"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/stretchr/testify/require"
)

func TestJSTruthyValues(t *testing.T) {

	req := require.New(t)

	req.True(IsTruthy(true))

	var myMap map[string]interface{}
	req.True(IsTruthy(myMap))
	myMap = map[string]interface{}{"a": 1}
	req.True(IsTruthy(myMap))

	var mySlice []byte
	req.True(IsTruthy(mySlice))
	mySlice = []byte{21, 12}
	req.True(IsTruthy(mySlice))

	req.True(IsTruthy(""))

	req.False(IsTruthy(false))

	req.False(IsTruthy(0))

	req.False(IsTruthy(float64(0)))

	req.False(IsTruthy(nil))

	req.False(IsTruthy(bson.Undefined))

	req.True(IsTruthy([]int{1, 2, 3}))
	req.True(IsTruthy("true"))
	req.True(IsTruthy("false"))
	req.True(IsTruthy(25))
	req.True(IsTruthy(math.NaN()))
	req.True(IsTruthy(25.1))
	req.True(IsTruthy(struct{ A int }{A: 12}))

}
