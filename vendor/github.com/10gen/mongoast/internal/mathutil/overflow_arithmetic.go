package mathutil

// Add32 performs addition on 32 bit ints and returns whether or not there is overflow
func Add32(a, b int32) (int32, bool) {
	c := a + b
	return c, (c > a) == (b > 0)
}

// Add64 performs addition on 64 bit ints and returns whether or not there is overflow
func Add64(a, b int64) (int64, bool) {
	c := a + b
	return c, (c > a) == (b > 0)
}

// Sub32 performs subtraction on 32 bit ints and returns whether or not there is overflow
func Sub32(a, b int32) (int32, bool) {
	c := a - b
	return c, (c < a) == (b > 0)
}

// Sub64 performs subtraction on 64 bit ints and returns whether or not there is overflow
func Sub64(a, b int64) (int64, bool) {
	c := a - b
	return c, (c < a) == (b > 0)
}

// Mul32 performs subtraction on 32 bit ints and returns whether or not there is overflow
func Mul32(a, b int32) (int32, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	c := a * b
	return c, (c < 0) == ((a < 0) != (b < 0)) && c/b == a
}

// Mul64 performs subtraction on 64 bit ints and returns whether or not there is overflow
func Mul64(a, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	c := a * b
	return c, (c < 0) == ((a < 0) != (b < 0)) && c/b == a
}
