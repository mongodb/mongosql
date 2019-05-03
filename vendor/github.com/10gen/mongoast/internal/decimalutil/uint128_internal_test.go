package decimalutil

import (
	"fmt"
	"testing"
)

func TestAdd(t *testing.T) {
	testCases := []struct {
		x        uint128
		y        uint128
		expected uint128
	}{
		{
			x:        uint128{High: 1, Low: 2},
			y:        uint128{High: 3, Low: 4},
			expected: uint128{High: 4, Low: 6},
		},
		{
			x:        uint128{High: 1, Low: 0xFFFFFFFFFFFFFFFF},
			y:        uint128{High: 2, Low: 2},
			expected: uint128{High: 4, Low: 1},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v + %v", tc.x, tc.y), func(t *testing.T) {
			actual := tc.x.Add(tc.y)
			if actual != tc.expected {
				t.Fatalf("expected %v, but got %v", tc.expected, actual)
			}
		})
	}
}

func TestCompareTo(t *testing.T) {
	testCases := []struct {
		x        uint128
		y        uint128
		expected int
	}{
		{
			x:        uint128{High: 1, Low: 2},
			y:        uint128{High: 2, Low: 1},
			expected: -1,
		},
		{
			x:        uint128{High: 2, Low: 1},
			y:        uint128{High: 1, Low: 2},
			expected: 1,
		},
		{
			x:        uint128{High: 1, Low: 1},
			y:        uint128{High: 1, Low: 2},
			expected: -1,
		},
		{
			x:        uint128{High: 1, Low: 2},
			y:        uint128{High: 1, Low: 1},
			expected: 1,
		},
		{
			x:        uint128{High: 1, Low: 1},
			y:        uint128{High: 1, Low: 1},
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v <=> %v", tc.x, tc.y), func(t *testing.T) {
			actual := tc.x.CompareTo(tc.y)
			if actual != tc.expected {
				t.Fatalf("expected %v, but got %v", tc.expected, actual)
			}
		})
	}
}

func TestDivide(t *testing.T) {
	testCases := []struct {
		dividend          uint128
		divisor           uint32
		expectedQuotient  uint128
		expectedRemainder uint32
	}{
		{
			dividend:          uint128{High: 0, Low: 3},
			divisor:           2,
			expectedQuotient:  uint128{High: 0, Low: 1},
			expectedRemainder: 1,
		},
		{
			dividend:         uint128{High: 0, Low: 1 << 32},
			divisor:          2,
			expectedQuotient: uint128{High: 0, Low: 1 << 31},
		},
		{
			dividend:         uint128{High: 1, Low: 0},
			divisor:          2,
			expectedQuotient: uint128{High: 0, Low: 1 << 63},
		},
		{
			dividend:         uint128{High: 1 << 32, Low: 0},
			divisor:          2,
			expectedQuotient: uint128{High: 1 << 31, Low: 0},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v / %v", tc.dividend, tc.divisor), func(t *testing.T) {
			actualQuotient, actualRemainder := tc.dividend.Divide(tc.divisor)
			if actualQuotient != tc.expectedQuotient {
				t.Fatalf("expected quotient %v, but got %v", tc.expectedQuotient, actualQuotient)
			}
			if actualRemainder != tc.expectedRemainder {
				t.Fatalf("expected remainder %v, but got %v", tc.expectedRemainder, actualRemainder)
			}
		})
	}
}

func TestMultiply(t *testing.T) {
	testCases := []struct {
		x        uint128
		y        uint32
		expected uint128
	}{
		{
			x:        uint128{High: 0, Low: 1},
			y:        2,
			expected: uint128{High: 0, Low: 2},
		},
		{
			x:        uint128{High: 0, Low: 1 << 31},
			y:        2,
			expected: uint128{High: 0, Low: 1 << 32},
		},
		{
			x:        uint128{High: 0, Low: 1 << 63},
			y:        2,
			expected: uint128{High: 1, Low: 0},
		},
		{
			x:        uint128{High: 1 << 31, Low: 0},
			y:        2,
			expected: uint128{High: 1 << 32, Low: 0},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v * %v", tc.x, tc.y), func(t *testing.T) {
			actual := tc.x.Multiply(tc.y)
			if actual != tc.expected {
				t.Fatalf("expected %v, but got %v", tc.expected, actual)
			}
		})
	}
}

func TestParseUInt128(t *testing.T) {
	testCases := []struct {
		s        string
		expected uint128
	}{
		{
			s:        "0",
			expected: uint128{High: 0, Low: 0},
		},
		{
			s:        "000000",
			expected: uint128{High: 0, Low: 0},
		},
		{
			s:        "1",
			expected: uint128{High: 0, Low: 1},
		},
		{
			s:        "000001",
			expected: uint128{High: 0, Low: 1},
		},
		{
			s:        "1234567890",
			expected: uint128{High: 0, Low: 1234567890},
		},
		{
			s:        "18446744073709551616",
			expected: uint128{High: 1, Low: 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.s, func(t *testing.T) {
			actual := parseUint128(tc.s)
			if actual != tc.expected {
				t.Fatalf("expected %v, but got %v", tc.expected, actual)
			}
		})
	}
}
