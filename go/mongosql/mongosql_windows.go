package mongosql

import (
	"encoding/base64"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"syscall"
	"unsafe"
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

// version returns the version of the underlying c translation
// library. The consumer of this library should ensure that the
// version of the the go library matches that of the c library.
func version() string {
	ret1, _, _ := versionProc.Call()
	goRetVal := uintptrToString(ret1)

	// delete the returned uintptr
	deleteStringProc.Call(ret1)

	return goRetVal
}

// callTranslate is a thin wrapper around the translate FFI call. It
// passes the provided TranslationArgs to the c translation library,
// and returns the string returned by the c library (a base64-encoded
// bson document representing the result of the translation).
func callTranslate(args TranslationArgs) (string, error) {

	// Convert the catalog schema into a base64-encoded bson document
	catalogSchemaBson, err := bson.Marshal(args.CatalogSchema)
	if err != nil {
		return "", fmt.Errorf("failed to marshal catalog schema to BSON: %w", err)
	}
	catalogSchemaBase64 := base64.StdEncoding.EncodeToString(catalogSchemaBson)
	dbArg, sqlArg, catalogArg := stringToUnsafePointer(args.DB), stringToUnsafePointer(args.SQL), stringToUnsafePointer(catalogSchemaBase64)
	ret1, _, _ := translateProc.Call(uintptr(dbArg), uintptr(sqlArg), uintptr(catalogArg))
	translationBase64 := uintptrToString(ret1)

	// delete the returned uintptr
	deleteStringProc.Call(ret1)

	return translationBase64, nil
}
