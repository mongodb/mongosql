package mongosql

/*
#cgo LDFLAGS: -L../../target/debug/ -lmongosql -lm -lgcc -lws2_32 -luser32 -lkernel32 -luserenv
#include <stdlib.h>
#include "./mongosql.h"
*/
import "C"
import (
	"unsafe"
)

// Version returns the version of the underlying translation library. This version should match the version of the
func Version() string {
	cVersion := C.version()
	version := C.GoString(cVersion)
	C.free(unsafe.Pointer(cVersion))
	return version
}

// Translate takes a SQL string and returns an extJSON string
// representation of its agg-pipeline translation.
func Translate(sql string) string {
	cSQL := C.CString(sql)

	cTranslation := C.translate(cSQL)
	translation := C.GoString(cTranslation)

	C.free(unsafe.Pointer(cSQL))
	C.free(unsafe.Pointer(cTranslation))

	return translation
}

// ResultSetSchema takes a SQL string and returns a string
// representation of the MongoDB schema of the result set.
func ResultSetSchema(sql string) string {
	panic("unimplemented")
}
