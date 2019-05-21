package stringutil

import "sort"

// StringSet implements a set of strings.
type StringSet struct {
	set map[string]struct{}
}

// NewStringSet creates a new string set.
func NewStringSet() *StringSet {
	return &StringSet{
		set: make(map[string]struct{}),
	}
}

// Add adds a new element to the set.
func (s *StringSet) Add(elem string) {
	s.set[elem] = struct{}{}
}

// Remove removes an element from the set.
func (s *StringSet) Remove(elem string) {
	delete(s.set, elem)
}

// RemoveAll removes all elements from the set.
func (s *StringSet) RemoveAll() {
	s.set = make(map[string]struct{})
}

// Contains checks if an element is present in the set.
func (s *StringSet) Contains(elem string) bool {
	_, ok := s.set[elem]
	return ok
}

// Len gets the length of the set.
func (s *StringSet) Len() int {
	return len(s.set)
}

// AddSlice adds all elements in a slice to the set.
func (s *StringSet) AddSlice(list []string) {
	for _, elem := range list {
		s.set[elem] = struct{}{}
	}
}

// AddSet adds all elements in another set to the set.
func (s *StringSet) AddSet(other *StringSet) {
	for elem := range other.set {
		s.set[elem] = struct{}{}
	}
}

// Slice converts the set to a slice.
func (s *StringSet) Slice() []string {
	list := make([]string, 0, len(s.set))
	for elem := range s.set {
		list = append(list, elem)
	}
	return list
}

// SortedSlice converts the set to a sorted slice.
func (s *StringSet) SortedSlice() []string {
	slice := s.Slice()
	sort.Strings(slice)
	return slice
}

// Intersection computes the intersection of two StringSets.
func (s *StringSet) Intersection(other *StringSet) *StringSet {
	out := NewStringSet()
	for value := range s.set {
		if other.Contains(value) {
			out.Add(value)
		}
	}
	return out
}

// HasIntersection returns true if there is a non-empty intersection between two sets.
func (s *StringSet) HasIntersection(other *StringSet) bool {
	for value := range s.set {
		if other.Contains(value) {
			return true
		}
	}
	return false
}

// Equals returns true if these sets are equal.
func (s *StringSet) Equals(other *StringSet) bool {
	if s.Len() != other.Len() {
		return false
	}
	for value := range s.set {
		if !other.Contains(value) {
			return false
		}
	}
	return true
}
