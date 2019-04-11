package decimalutil_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/10gen/mongoast/util/decimalutil"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNegate(t *testing.T) {
	testCases := []struct {
		x        string
		expected string
	}{
		{
			x:        "1",
			expected: "-1",
		},
		{
			x:        "-1",
			expected: "1",
		},
		{
			x:        "0",
			expected: "0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.x, func(t *testing.T) {
			xPrimitive, err := primitive.ParseDecimal128(tc.x)
			if err != nil {
				t.Fatalf("failed to parse decimal %s", tc.x)
			}
			x := decimalutil.FromPrimitive(xPrimitive)
			expectedPrimitive, err := primitive.ParseDecimal128(tc.expected)
			if err != nil {
				t.Fatalf("failed to parse decimal %s", tc.expected)
			}
			expected := decimalutil.FromPrimitive(expectedPrimitive)
			actual := decimalutil.Negate(x)
			if decimalutil.Compare(actual, expected) != 0 {
				t.Fatalf("expected %v, but got %v", expected, actual)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	testCases := []struct {
		x        string
		y        string
		expected int
	}{
		{
			x:        "1",
			y:        "2",
			expected: -1,
		},
		{
			x:        "2",
			y:        "1",
			expected: 1,
		},
		{
			x:        "1",
			y:        "1",
			expected: 0,
		},
		{
			x:        "0",
			y:        "0",
			expected: 0,
		},
		{
			x:        "1e3",
			y:        "10e2",
			expected: 0,
		},
		{
			x:        "2e3",
			y:        "15e2",
			expected: 1,
		},
		{
			x:        "-1e3",
			y:        "-10e2",
			expected: 0,
		},
		{
			x:        "-2e3",
			y:        "-15e2",
			expected: -1,
		},
		{
			x:        "1e100",
			y:        "1e5",
			expected: 1,
		},
		{
			x:        "-1",
			y:        "0",
			expected: -1,
		},
		{
			x:        "0",
			y:        "1",
			expected: -1,
		},
		{
			x:        "-1",
			y:        "1",
			expected: -1,
		},
		{
			x:        "1",
			y:        "-1",
			expected: 1,
		},
		{
			x:        "1",
			y:        "0",
			expected: 1,
		},
		{
			x:        "0",
			y:        "-1",
			expected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s + %s", tc.x, tc.y), func(t *testing.T) {
			xPrimitive, err := primitive.ParseDecimal128(tc.x)
			if err != nil {
				t.Fatalf("failed to parse decimal %s", tc.x)
			}
			x := decimalutil.FromPrimitive(xPrimitive)
			yPrimitive, err := primitive.ParseDecimal128(tc.y)
			if err != nil {
				t.Fatalf("failed to parse decimal %s", tc.y)
			}
			y := decimalutil.FromPrimitive(yPrimitive)
			actual := decimalutil.Compare(x, y)
			if actual != tc.expected {
				t.Fatalf("expected %v, but got %v", tc.expected, actual)
			}
		})
	}
}

func TestFromInt32(t *testing.T) {
	testCases := []int32{1, 0, -1, math.MaxInt32, math.MinInt32}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tc), func(t *testing.T) {
			expectedPrimitive, err := primitive.ParseDecimal128(fmt.Sprintf("%d", tc))
			if err != nil {
				t.Fatalf("failed to parse decimal %d", tc)
			}
			expected := decimalutil.FromPrimitive(expectedPrimitive)
			actual := decimalutil.FromInt32(tc)
			if actual != expected {
				t.Fatalf("expected %v, but got %v", expected, actual)
			}
		})
	}
}

func TestFromInt64(t *testing.T) {
	testCases := []int64{1, 0, -1, math.MaxInt64, math.MinInt64}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tc), func(t *testing.T) {
			expectedPrimitive, err := primitive.ParseDecimal128(fmt.Sprintf("%d", tc))
			if err != nil {
				t.Fatalf("failed to parse decimal %d", tc)
			}
			expected := decimalutil.FromPrimitive(expectedPrimitive)
			actual := decimalutil.FromInt64(tc)
			if actual != expected {
				t.Fatalf("expected %v, but got %v", expected, actual)
			}
		})
	}
}
