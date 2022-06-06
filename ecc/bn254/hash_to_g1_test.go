package bn254

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"testing"
)

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
		g1TestMatchCoord(t, "u0", c.msg, c.u0, &elems[0])
		g1TestMatchCoord(t, "u1", c.msg, c.u1, &elems[1])
	}
}

func TestMapToG1Vectors(t *testing.T) {
	for _, c := range encodeToG1Vector.cases {
		var u fp.Element
		u.SetString(c.u)
		q := mapToCurve1(u)
		g1TestMatchPoint(t, "Q", c.msg, c.Q, &q)
	}

	for _, c := range hashToG1Vector.cases {
		var u fp.Element
		u.SetString(c.u0)
		q := mapToCurve1(u)
		g1TestMatchPoint(t, "Q0", c.msg, c.Q0, &q)

		u.SetString(c.u1)
		q = mapToCurve1(u)
		g1TestMatchPoint(t, "Q1", c.msg, c.Q1, &q)
	}
}

func TestEncodeToG1Vectors(t *testing.T) {
	for _, c := range encodeToG1Vector.cases {
		p, err := EncodeToG1([]byte(c.msg), encodeToG1Vector.dst)
		if err != nil {
			t.Fatal(err)
		}
		g1TestMatchPoint(t, "P", c.msg, c.P, &p)
	}
}

func TestHashToG1Vectors(t *testing.T) {
	for _, c := range hashToG1Vector.cases {
		p, err := HashToG1([]byte(c.msg), hashToG1Vector.dst)
		if err != nil {
			t.Fatal(err)
		}
		g1TestMatchPoint(t, "P", c.msg, c.P, &p)
	}
}

func g1TestMatchCoord(t *testing.T, coordName string, msg string, expectedStr string, seen *fp.Element) {
	var expected fp.Element

	expected.SetString(expectedStr)

	if !expected.Equal(seen) {
		t.Errorf("mismatch on \"%s\", %s:\n\texpected %s\n\tsaw      0x%s", msg, coordName, expectedStr, seen.Text(16))
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

type encodeTestVector struct {
	dst   []byte
	cases []encodeTestCase
}

type point struct {
	x string
	y string
}

type encodeTestCase struct {
	msg string
	P   point  //pY a coordinate of P, the final output
	u   string //u hashed onto the field
	Q   point  //Q map to curve output
}

type hashTestCase struct {
	msg string
	P   point  //pY a coordinate of P, the final output
	u0  string //u0 hashed onto the field
	u1  string //u1 extra hashed onto the field
	Q0  point  //Q0 map to curve output
	Q1  point  //Q1 extra map to curve output
}

var hashToG1Vector hashTestVector
var encodeToG1Vector encodeTestVector
