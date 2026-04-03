package secp256r1

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/secp256r1/fp"
)

func TestCardanoRoots(t *testing.T) {
	// Test with c = b - y² for small y values on P-256 (y² = x³ - 3x + b)
	var b fp.Element
	// P-256 b = 0x5ac635d8aa3a93e7b3ebbd55769886bc651d06b0cc53b0f63bce3c3e27d2604b
	b.SetString("41058363725152142129326129780047268409114441015993725554835256314039467401291")

	for y := int64(0); y < 256; y++ {
		var yFp, y2, c fp.Element
		yFp.SetInt64(y)
		y2.Square(&yFp)
		c.Sub(&b, &y2)

		roots := CardanoRoots(c)
		// verify each root satisfies x³ - 3x + c = 0
		for _, r := range roots {
			var r3, three, threex, lhs fp.Element
			r3.Square(&r).Mul(&r3, &r)
			three.SetInt64(3)
			threex.Mul(&three, &r)
			lhs.Sub(&r3, &threex).Add(&lhs, &c)
			if !lhs.IsZero() {
				t.Fatalf("y=%d: root %v does not satisfy x³ - 3x + c = 0", y, r)
			}
		}
	}
}
