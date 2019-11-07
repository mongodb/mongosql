//+build !go1.13

package memdebug

const (
	// KB is the number of bytes in a KB.
	KB = 1024.0
	// MB is th number of bytes in a MB.
	MB = 1048576.0
)

// Sizeof returns the estimated memory usage of object(s) not just the size of the type.
// On 64bit Sizeof("test") == 12 (8 = sizeof(StringHeader) + 4 bytes).
func Sizeof(objs ...interface{}) (sz uint64) {
	panic("memdebug functions not available before go 1.13")
}

// SizeofMB returns object size in MegaBytes.
func SizeofMB(objs ...interface{}) float64 {
	panic("memdebug functions not available before go 1.13")
}

// SizeofKB returns object size in KiloBytes.
func SizeofKB(objs ...interface{}) float64 {
	panic("memdebug functions not available before go 1.13")
}
