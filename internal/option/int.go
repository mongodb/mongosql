package option

import "fmt"

// Int represents an optional int value.
type Int struct {
	isSome bool
	i      int
}

// SomeInt constructs a Some variant of option.Int with the provided int.
func SomeInt(i int) Int {
	return Int{
		isSome: true,
		i:      i,
	}
}

// NoneInt constructs the None variant of option.Int.
func NoneInt() Int {
	return Int{
		isSome: false,
		i:      0,
	}
}

// Copy returns a copy of this Int.
func (v Int) Copy() Int {
	return v
}

// Set makes this Int a Some variant with the provided value.
func (v *Int) Set(i int) {
	v.i = i
	v.isSome = true
}

// IsSome returns true if this option is a Some, and false if it is a None.
func (v Int) IsSome() bool {
	return v.isSome
}

// IsNone returns true if this option is a None, and false if it is a Some.
func (v Int) IsNone() bool {
	return !v.IsSome()
}

// Unwrap returns the value of the Int if it is a Some, and panics
// with a default error message otherwise.
func (v Int) Unwrap() int {
	return v.Expect("tried to unwrap a None variant")
}

// Expect returns the value of the Int if it is a Some, and panics
// with the provided error message otherwise.
func (v Int) Expect(msg string) int {
	if !v.isSome {
		panic(msg)
	}
	return v.i
}

// Else returns the value of the Int if it is a Some, and returns the
// provided alternative value if it is a None.
func (v Int) Else(alt int) int {
	if v.isSome {
		return v.i
	}
	return alt
}

// IntMapFunc is a function that maps a int to another int.
type IntMapFunc func(int) int

// Map returns a new Int with its value transformed by the provided
// IntMapFunc if a Some, and None otherwise.
func (v Int) Map(f IntMapFunc) Int {
	if v.IsSome() {
		v.i = f(v.i)
	}
	return v
}

// IntMapStringFunc is a function that maps a int to a string.
type IntMapStringFunc func(int) string

// MapString returns a new String with its value transformed by the provided
// IntMapStringFunc if a Some, and None otherwise.
func (v Int) MapString(f IntMapStringFunc) String {
	if v.IsSome() {
		return SomeString(f(v.i))
	}
	return NoneString()
}

// Int returns a string representation of this Int.
func (v Int) String() string {
	return v.
		MapString(func(i int) string { return fmt.Sprintf("Some(%d)", i) }).
		Else("None")
}
