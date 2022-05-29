package bn254

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"testing"
)

var DST = []byte("QUUX-V01-CS02-with-BN254G1_XMD:SHA-256_SW_NU_")

func TestHashToFpG1Vectors(t *testing.T) {
	for _, c := range encodeToG1Vector.cases {
		elems, err := hashToFp([]byte(c.msg), encodeToG1Vector.dst, 1)
		if err != nil {
			t.Error(err)
		}
		g1TestMatchCoord(t, "u", c.msg, c.u, &elems[0])
	}

	for _, c := range hashToG1Vector.cases {
		elems, err := hashToFp([]byte(c.msg), hashToG1Vector.dst, 2)
		if err != nil {
			t.Error(err)
		}
		g1TestMatchCoord(t, "u0", c.msg, c.u, &elems[0])
		g1TestMatchCoord(t, "u1", c.msg, c.u1, &elems[1])
	}
}

func TestMapToG1Vectors(t *testing.T) {
	for _, c := range encodeToG1Vector.cases {
		var u fp.Element
		u.SetString(c.u)
		q := MapToG1(u)
		g1TestMatchPoint(t, "Q", c.msg, c.Q, &q)
	}

	for _, c := range hashToG1Vector.cases {
		var u fp.Element
		u.SetString(c.u)
		q := MapToG1(u)
		g1TestMatchPoint(t, "Q0", c.msg, c.Q, &q)

		u.SetString(c.u1)
		q = MapToG1(u)
		g1TestMatchPoint(t, "Q1", c.msg, c.Q1, &q)
	}
}

func g1TestMatchCoord(t *testing.T, coordName string, msg string, expectedStr string, seen *fp.Element) {
	var expected fp.Element

	expected.SetString(expectedStr)

	if !expected.Equal(seen) {
		t.Errorf("mismatch on \"%s\", %s:\n\texpected %s\n\tsaw      %s", msg, coordName, expectedStr, seen.Text(16))
	}
}

func g1TestMatchPoint(t *testing.T, pointName string, msg string, expected point, seen *G1Affine) {
	g1TestMatchCoord(t, pointName+".x", msg, expected.x, &seen.X)
	g1TestMatchCoord(t, pointName+".y", msg, expected.y, &seen.Y)
}

type hashTestVector struct {
	dst   []byte
	cases []hashTestCase
}

type point struct {
	x string
	y string
}

type hashTestCase struct {
	msg string
	P   point  //pY a coordinate of P, the final output
	u   string //u hashed onto the field
	u1  string //u1 extra hashed onto the field for HashTo
	Q   point  //Q map to curve output
	Q1  point  //Q extra map to curve output for HashTo
}

var hashToG1Vector hashTestVector
var encodeToG1Vector hashTestVector
