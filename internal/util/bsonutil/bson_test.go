package bsonutil

import (
	"fmt"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/stretchr/testify/require"
)

func TestEmptyBsonStructs(t *testing.T) {
	req := require.New(t)

	structs := []interface{}{bson.M{}, []bson.M{}, bson.D{}, []bson.D{}, []interface{}{}}
	funcCalls := []interface{}{NewM(), NewMArray(), NewD(), NewDArray(), NewArray()}
	names := []string{"NewM()", "NewMArray()", "NewD()", "NewDArray()", "NewArray()"}

	for i, s := range structs {
		req.Equal(s, funcCalls[i], fmt.Sprintf("%v does not correctly create corresponding empty struct", names[i]))
	}
}
