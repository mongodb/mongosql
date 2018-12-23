package option

// String represents an optional string value.
type String struct {
	isSome bool
	str    string
}

// SomeString constructs a Some variant of option.String with the provided string.
func SomeString(s string) String {
	return String{
		isSome: true,
		str:    s,
	}
}

// NoneString constructs the None variant of option.String.
func NoneString() String {
	return String{
		isSome: false,
		str:    "",
	}
}

// Copy returns a copy of this String.
func (s String) Copy() String {
	return s
}

// Set makes this String a Some variant with the provided value.
func (s *String) Set(str string) {
	s.str = str
	s.isSome = true
}

// IsSome returns true if this option is a Some, and false if it is a None.
func (s String) IsSome() bool {
	return s.isSome
}

// IsNone returns true if this option is a None, and false if it is a Some.
func (s String) IsNone() bool {
	return !s.IsSome()
}

// Unwrap returns the value of the String if it is a Some, and panics
// with a default error message otherwise.
func (s String) Unwrap() string {
	return s.Expect("tried to unwrap a None variant")
}

// Expect returns the value of the String if it is a Some, and panics
// with the provided error message otherwise.
func (s String) Expect(msg string) string {
	if !s.isSome {
		panic(msg)
	}
	return s.str
}

// Else returns the value of the String if it is a Some, and returns the
// provided alternative value if it is a None.
func (s String) Else(alt string) string {
	if s.isSome {
		return s.str
	}
	return alt
}

// StringMapFunc is a function that maps a string to another string.
type StringMapFunc func(string) string

// Map returns a new String with its value transformed by the provided
// StringMapFunc if a Some, and None otherwise.
func (s String) Map(f StringMapFunc) String {
	if s.IsSome() {
		s.str = f(s.str)
	}
	return s
}
