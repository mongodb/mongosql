package bsonutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestEmptyBsonStructs(t *testing.T) {
	req := require.New(t)

	structs := []interface{}{bson.M{}, []bson.M{}, bson.D{}, []bson.D{}, bson.A{}}
	funcCalls := []interface{}{NewM(), NewMArray(), NewD(), NewDArray(), NewArray()}
	names := []string{"NewM()", "NewMArray()", "NewD()", "NewDArray()", "NewArray()"}

	for i, s := range structs {
		req.Equal(s, funcCalls[i], fmt.Sprintf("%v does not correctly create corresponding empty struct", names[i]))
	}
}
