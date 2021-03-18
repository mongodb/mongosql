//+build darwin linux

package mongosql

/*
#cgo LDFLAGS: -lmongosql -ldl -lm
#include <stdlib.h>
#include "./mongosql.h"
*/
import "C"
import (
	"encoding/base64"
	"unsafe"

	"go.mongodb.org/mongo-driver/bson"
)

// Version returns the version of the underlying c translation
// library. The consumer of this library should ensure that the
// version of the the go library matches that of the c library.
func Version() string {
	cVersion := C.version()
	version := C.GoString(cVersion)
	C.free(unsafe.Pointer(cVersion))
	return version
}

// Translate takes a SQL query string and the database in which that
// query should be executed, returning a target database, target
// collection, and MongoDB aggregation pipeline as a BSON array.
func Translate(db, sql string) (queryDB string, queryCollection string, pipeline []byte) {
	cSQL := C.CString(sql)
	cDB := C.CString(db)

	cTranslationBase64 := C.translate(cDB, cSQL)
	translationBase64 := C.GoString(cTranslationBase64)

	C.free(unsafe.Pointer(cSQL))
	C.free(unsafe.Pointer(cDB))
	C.free(unsafe.Pointer(cTranslationBase64))

	translation := struct {
		Db         string   `bson:"target_db"`
		Collection string   `bson:"target_collection"`
		Pipeline   []bson.M `bson:"pipeline"`
	}{}

	translationBytes, err := base64.StdEncoding.DecodeString(translationBase64)
	if err != nil {
		panic(err)
	}

	err = bson.Unmarshal(translationBytes, &translation)
	if err != nil {
		panic(err)
	}

	typ, pipelineBytes, err := bson.MarshalValue(translation.Pipeline)
	if err != nil {
		panic(err)
	}
	if typ.String() != "array" {
		panic("didn't marshal to array")
	}

	return translation.Db, translation.Collection, pipelineBytes
}
