package mongosql

import (
	"encoding/base64"
	"syscall"
	"unsafe"

	"github.com/10gen/mongosql-rs/go/internal/desugarer"
	"go.mongodb.org/mongo-driver/bson"
)

var versionProc *syscall.LazyProc
var deleteStringProc *syscall.LazyProc
var translateProc *syscall.LazyProc

func init() {
	dll := syscall.NewLazyDLL("mongosql.dll")
	versionProc = dll.NewProc("version")
	deleteStringProc = dll.NewProc("delete_string")
	translateProc = dll.NewProc("translate")
}

// uintptrToString converts a uintptr return value from
// a LazyProc.Call into a go string. It does this by
// copying the entire memory out of the DLL space into
// go, making it safe to manually delete the DLL space
// memory.
func uintptrToString(u uintptr) string {
	var b byte
	p := unsafe.Pointer(u)
	bs := []byte{}
	for {
		b = *(*byte)(p)
		if b == 0 {
			break
		}
		bs = append(bs, b)
		p = unsafe.Pointer(uintptr(p) + 1)
	}
	return string(bs)
}

// stringToUnsafePointer gets an unsafe.Pointer to a string to pass to a LazyProc.Call
// DLL call. It has to nul-terminate the string, which may result in copying
// the string value, but may not. Go will automatically delete this memory
// once the dll call has completed during a gc cycle. In order to pass this
// value to a syscall it must be converted to uintptr in the syscall expression.
func stringToUnsafePointer(s string) unsafe.Pointer {
	bs := []byte(s)
	bs = append(bs, 0)
	return unsafe.Pointer(&bs[0])
}

// Version returns the version of the underlying c translation
// library. The consumer of this library should ensure that the
// version of the the go library matches that of the c library.
func Version() string {
	ret1, _, _ := versionProc.Call()
	goRetVal := uintptrToString(ret1)

	// delete the returned uintptr
	deleteStringProc.Call(ret1)

	return goRetVal
}

// Translate takes a SQL query string and the database in which that
// query should be executed, returning a target database, target
// collection, and MongoDB aggregation pipeline as a BSON array.
func Translate(db, sql string) (queryDB string, queryCollection string, pipeline []byte) {

	dbArg, sqlArg := stringToUnsafePointer(db), stringToUnsafePointer(sql)
	ret1, _, _ := translateProc.Call(uintptr(dbArg), uintptr(sqlArg))
	translationBase64 := uintptrToString(ret1)

	// delete the returned uintptr
	deleteStringProc.Call(ret1)

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

	pipelineBytes, err = desugarer.Desugar(pipelineBytes)
	if err != nil {
		panic(err)
	}

	return translation.Db, translation.Collection, pipelineBytes
}
