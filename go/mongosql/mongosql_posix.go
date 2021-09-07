//+build darwin linux

package mongosql

/*
#cgo LDFLAGS: -lmongosql -lpthread -ldl -lm
#include <stdlib.h>
#include "./mongosql.h"
*/
import "C"
import (
	"encoding/base64"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"unsafe"
)

// version returns the version of the underlying c translation
// library. The consumer of this library should ensure that the
// version of the the go library matches that of the c library.
func version() string {
	cVersion := C.version()
	version := C.GoString(cVersion)
	C.free(unsafe.Pointer(cVersion))
	return version
}

// callTranslate is a thin wrapper around the translate FFI call. It
// passes the provided TranslationArgs to the c translation library,
// and returns the string returned by the c library (a base64-encoded
// bson document representing the result of the translation).
func callTranslate(args TranslationArgs) (string, error) {
	cSQL := C.CString(args.SQL)
	cDB := C.CString(args.DB)

	// Convert the catalog schema into a base64-encoded bson document
	catalogSchemaBson, err := bson.Marshal(args.CatalogSchema)
	if err != nil {
		return "", fmt.Errorf("failed to marshal catalog schema to BSON: %w", err)
	}
	cCatalogSchema := C.CString(base64.StdEncoding.EncodeToString(catalogSchemaBson))

	cTranslationBase64 := C.translate(cDB, cSQL, cCatalogSchema)
	translationBase64 := C.GoString(cTranslationBase64)

	C.free(unsafe.Pointer(cSQL))
	C.free(unsafe.Pointer(cDB))
	C.free(unsafe.Pointer(cCatalogSchema))
	C.free(unsafe.Pointer(cTranslationBase64))

	return translationBase64, nil
}
