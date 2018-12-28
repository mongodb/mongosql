package mathutil_test

import (
	"testing"

	. "github.com/10gen/sqlproxy/internal/mathutil"
	"github.com/stretchr/testify/require"
)

// TestUint128 tests the functionality of Uint128 code.
func TestUint128(t *testing.T) {
	req := require.New(t)
	x := Uint128{H: 0x0, L: 0xf000000000000000}
	y := Uint128{H: 0x0, L: 0x000000000000000f}
	res := Uint128{H: 0xe, L: 0x1000000000000000}

	x.Mult(y)
	req.Equal(res, x)

	x.Plus(0x19)
	res = Uint128{H: 0xe, L: 0x1000000000000019}
	req.Equal(res, x)

	x.Plus(0xf000000000000000)
	res = Uint128{H: 0xf, L: 0x0000000000000019}
	req.Equal(res, x)

	x.Xor(0xff)
	res = Uint128{H: 0xf, L: 0x00000000000000e6}
	req.Equal(res, x)
}
