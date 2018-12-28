package strutil

import (
	"fmt"
	"reflect"
)

// SliceContains is a generic function that returns true if elt is in slice.
// It panics if slice is not of Kind reflect.Slice.
func SliceContains(slice, elt interface{}) bool {
	if slice == nil {
		return false
	}
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		panic(fmt.Sprintf("Cannot call SliceContains on a non-slice %#v of "+
			"kind %#v", slice, v.Kind().String()))
	}
	for i := 0; i < v.Len(); i++ {
		if reflect.DeepEqual(v.Index(i).Interface(), elt) {
			return true
		}
	}
	return false
}

// IntSliceContains reports whether i is in the slice.
func IntSliceContains(slice []int, i int) bool {
	return IntSliceIndex(slice, i) != -1
}

// IntSliceIndex returns the first index at which the given int
// can be found in the slice, or -1 if it is not present.
func IntSliceIndex(slice []int, i int) int {
	idx := -1
	for j, v := range slice {
		if v == i {
			idx = j
			break
		}
	}
	return idx
}

// StringSliceContains reports whether str is in the slice.
func StringSliceContains(slice []string, str string) bool {
	return StringSliceIndex(slice, str) != -1
}

// StringSliceIndex returns the first index at which the given string
// can be found in the slice, or -1 if it is not present.
func StringSliceIndex(slice []string, str string) int {
	i := -1
	for j, v := range slice {
		if v == str {
			i = j
			break
		}
	}
	return i
}

// SliceCount is a generic function that returns number of instances of 'elt' in 'slice'.
// It panics if slice is not of Kind reflect.Slice.
func SliceCount(slice, elt interface{}) int {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		panic(fmt.Sprintf("Cannot call SliceCount on a non-slice %#v of kind "+
			"%#v", slice, v.Kind().String()))
	}
	counter := 0
	for i := 0; i < v.Len(); i++ {
		if reflect.DeepEqual(v.Index(i).Interface(), elt) {
			counter++
		}
	}
	return counter
}
