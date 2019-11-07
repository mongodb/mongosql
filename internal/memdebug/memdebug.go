//+build go1.13

package memdebug

import (
	"fmt"
	"reflect"
)

const (
	// KB is the number of bytes in a KB.
	KB = 1024.0
	// MB is th number of bytes in a MB.
	MB = 1048576.0
)

var (
	sliceSize  = uint64(reflect.TypeOf(reflect.SliceHeader{}).Size())
	stringSize = uint64(reflect.TypeOf(reflect.StringHeader{}).Size())
)

func isNativeType(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Bool:
		return true
	}
	return false
}

func sizeofInternal(seen map[reflect.Value]struct{}, val reflect.Value, fromStruct bool, depth int) (sz uint64) {
	if depth++; depth > 1000 {
		panic("sizeOf recursed more than 1000 times.")
	}

	if val.IsZero() {
		sz += uint64(val.Type().Size())
		return
	}

	typ := val.Type()

	if !fromStruct {
		sz = uint64(typ.Size())
	}

	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			break
		}
		// Don't multi-count memory! This code could not handle cycles before.
		if _, ok := seen[val]; ok {
			break
		}
		sz += sizeofInternal(seen, val.Elem(), false, depth)
		// only pointers need to go in seen, since only pointers can cause
		// cycles
		seen[val] = struct{}{}

	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			sz += sizeofInternal(seen, val.Field(i), true, depth)
		}

	case reflect.Array:
		el := typ.Elem()
		if isNativeType(el.Kind()) {
			sz += uint64(val.Len()) * uint64(el.Size())
			break
		}
		sz = 0
		for i := 0; i < val.Len(); i++ {
			sz += sizeofInternal(seen, val.Index(i), false, depth)
		}
	case reflect.Slice:
		if !fromStruct {
			sz = sliceSize
		}
		el := typ.Elem()
		if isNativeType(el.Kind()) {
			sz += uint64(val.Len()) * uint64(el.Size())
			break
		}
		for i := 0; i < val.Len(); i++ {
			sz += sizeofInternal(seen, val.Index(i), false, depth)
		}
	case reflect.Map:
		if val.IsNil() {
			break
		}
		kel, vel := typ.Key(), typ.Elem()
		if isNativeType(kel.Kind()) && isNativeType(vel.Kind()) {
			sz += uint64(kel.Size()+vel.Size()) * uint64(val.Len())
			break
		}
		keys := val.MapKeys()
		for i := 0; i < len(keys); i++ {
			sz += sizeofInternal(seen, keys[i], false, depth) + sizeofInternal(seen, val.MapIndex(keys[i]), false, depth)
		}
	case reflect.String:
		if !fromStruct {
			sz = stringSize
		}
		sz += uint64(val.Len())
	case reflect.Interface:
		sz += sizeofInternal(seen, val.Elem(), false, depth)
	case reflect.Func:
		sz += 8
	default:
		if isNativeType(typ.Kind()) {
			sz += uint64(typ.Size())
			break
		}
		fmt.Println("Unknown value type: ", val, ":", typ.Kind())
	}
	return
}

// Sizeof returns the estimated memory usage of object(s) not just the size of the type.
// On 64bit Sizeof("test") == 12 (8 = sizeof(StringHeader) + 4 bytes).
func Sizeof(objs ...interface{}) (sz uint64) {
	seen := make(map[reflect.Value]struct{})
	for i := range objs {
		sz += sizeofInternal(seen, reflect.ValueOf(objs[i]), false, 0)
	}
	return
}

// SizeofMB returns object size in MegaBytes.
func SizeofMB(objs ...interface{}) float64 {
	return float64(Sizeof(objs...)) / MB
}

// SizeofKB returns object size in KiloBytes.
func SizeofKB(objs ...interface{}) float64 {
	return float64(Sizeof(objs...)) / KB
}
