package mongosql

import (
	"syscall"
	"unsafe"
)


func uintptr2string(u uintptr) string {
	b := byte(56)
	p := unsafe.Pointer(u)
	bs := []byte{}
	for ;b != byte(0); {
		b = *(*byte)(p)
		bs = append(bs, b)
		p = unsafe.Pointer(uintptr(p) +  1)
	}
	return string(bs)
}

func string2uintptr(s string) uintptr {
	bs := []byte(s)
	bs = append(bs, 0)
	p := unsafe.Pointer(&bs[0])
	return uintptr(p)
}

// Version returns the version of the underlying translation library. This version should match the version of the
func Version() string {
	dll, err := syscall.LoadDLL("mongosql.dll")
	if err != nil {
		panic("could not find mongosql.dll")
	}
	proc, err := dll.FindProc("version")
	if err != nil {
		panic("could not find version in mongosql.dll")
	}
	ret1, _, _ := proc.Call()
	return uintptr2string(ret1)
}

// Translate takes a SQL string and returns an extJSON string
// representation of its agg-pipeline translation.
func Translate(sql string) string {
	dll, err := syscall.LoadDLL("mongosql.dll")
	if err != nil {
		panic("could not find mongosql.dll")
	}
	proc, err := dll.FindProc("translate")
	if err != nil {
		panic("could not find translate in mongosql.dll")
	}
	arg := string2uintptr(sql)
	ret1, _, _ := proc.Call(arg)
	return uintptr2string(ret1)
}

// ResultSetSchema takes a SQL string and returns a string
// representation of the MongoDB schema of the result set.
func ResultSetSchema(sql string) string {
	panic("unimplemented")
}
