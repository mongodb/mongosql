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
	"unsafe"

	"go.mongodb.org/mongo-driver/bson"
)

// version returns the version of the underlying c translation
// library. The consumer of this library should ensure that the
// version of the go library matches that of the c library.
func version() string {
	cVersion := C.version()
	defer C.delete_string(cVersion)
	version := C.GoString(cVersion)
	return version
}

// callTranslate is a thin wrapper around the translate FFI call. It
// passes the provided TranslationArgs to the c translation library,
// and returns the string returned by the c library (a base64-encoded
// bson document representing the result of the translation).
func callTranslate(args TranslationArgs) (string, error) {
	cSQL := C.CString(args.SQL)
	defer C.free(unsafe.Pointer(cSQL))

	cDB := C.CString(args.DB)
	defer C.free(unsafe.Pointer(cDB))

	// Convert the catalog schema into a base64-encoded bson document
	catalogSchemaBson, err := bson.Marshal(args.CatalogSchema)
	if err != nil {
		return "", NewInternalError(fmt.Errorf("failed to marshal catalog schema to BSON: %w", err))
	}

	cCatalogSchema := C.CString(base64.StdEncoding.EncodeToString(catalogSchemaBson))
	defer C.free(unsafe.Pointer(cCatalogSchema))

	cRelaxSchemaChecking := C.int(0)
	if args.relaxSchemaChecking {
		cRelaxSchemaChecking = C.int(1)
	}

	cExcludeNamespaces := C.int(0)
	if args.ExcludeNamespaces {
		cExcludeNamespaces = C.int(1)
	}

	cTranslationBase64 := C.translate(cDB, cSQL, cCatalogSchema, cRelaxSchemaChecking, cExcludeNamespaces)
	defer C.delete_string(cTranslationBase64)

	translationBase64 := C.GoString(cTranslationBase64)

	return translationBase64, nil
}

// callGetNamespaces is a thin wrapper around the translate FFI call. It
// passes the provided arguments to the c translation library,
// and returns the string returned by the c library (a base64-encoded
// bson document representing the result of the get_namespaces call).
func callGetNamespaces(currentDB, sql string) string {
	cSQL := C.CString(sql)
	defer C.free(unsafe.Pointer(cSQL))

	cDB := C.CString(currentDB)
	defer C.free(unsafe.Pointer(cDB))

	cResultBase64 := C.get_namespaces(cDB, cSQL)
	defer C.delete_string(cResultBase64)

	resultBase64 := C.GoString(cResultBase64)
	return resultBase64
}
