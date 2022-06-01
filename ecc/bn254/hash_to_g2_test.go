package bn254

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/internal/fptower"
	"strings"
	"testing"
)

func TestHashToFpG2Vectors(t *testing.T) {
	for _, c := range encodeToG2Vector.cases {
		elems, err := hashToFp([]byte(c.msg), encodeToG2Vector.dst, 2)
		if err != nil {
			t.Error(err)
		}
		u := fptower.E2{A0: elems[0], A1: elems[1]}
		g2TestMatchCoord(t, "u", c.msg, c.u, &u)
	}

	for _, c := range hashToG2Vector.cases {
		elems, err := hashToFp([]byte(c.msg), hashToG2Vector.dst, 4)
		if err != nil {
			t.Error(err)
		}
		u0 := fptower.E2{A0: elems[0], A1: elems[1]}
		u1 := fptower.E2{A0: elems[2], A1: elems[3]}

		g2TestMatchCoord(t, "u0", c.msg, c.u0, &u0)
		g2TestMatchCoord(t, "u1", c.msg, c.u1, &u1)
	}
}

func TestMapToG2Vectors(t *testing.T) {
	for _, c := range encodeToG2Vector.cases {
		var u fptower.E2
		g2CoordSetString(&u, c.u)
		q := MapToG2(u)
		g2TestMatchPoint(t, "Q", c.msg, c.Q, &q)
	}

	for _, c := range hashToG2Vector.cases {
		var u fptower.E2
		g2CoordSetString(&u, c.u0)
		q := MapToG2(u)
		g2TestMatchPoint(t, "Q0", c.msg, c.Q0, &q)

		g2CoordSetString(&u, c.u1)
		q = MapToG2(u)
		g2TestMatchPoint(t, "Q1", c.msg, c.Q1, &q)
	}
}

func TestEncodeToG2Vectors(t *testing.T) {
	for _, c := range encodeToG2Vector.cases {
		p, err := EncodeToG2([]byte(c.msg), encodeToG2Vector.dst)
		if err != nil {
			t.Fatal(err)
		}
		g2TestMatchPoint(t, "P", c.msg, c.P, &p)
	}
}

func TestHashToG2Vectors(t *testing.T) {
	for _, c := range hashToG2Vector.cases {
		p, err := HashToG2([]byte(c.msg), hashToG2Vector.dst)
		if err != nil {
			t.Fatal(err)
		}
		g2TestMatchPoint(t, "P", c.msg, c.P, &p)
	}
}

func g2TestMatchCoord(t *testing.T, coordName string, msg string, expectedStr string, seen *fptower.E2) {
	var expected fptower.E2

	g2CoordSetString(&expected, expectedStr)

	if !expected.Equal(seen) {
		t.Errorf("mismatch on \"%s\", %s:\n\texpected %s\n\tsaw      %s", msg, coordName, expected.String(), seen)
	}
}

func g2TestMatchPoint(t *testing.T, pointName string, msg string, expected point, seen *G2Affine) {
	g2TestMatchCoord(t, pointName+".x", msg, expected.x, &seen.X)
	g2TestMatchCoord(t, pointName+".y", msg, expected.y, &seen.Y)
}

//Only works on simple extensions (two-story towers)
func g2CoordSetString(z *fptower.E2, s string) {
	ssplit := strings.Split(s, ",")
	if len(ssplit) != 2 {
		panic("not equal to tower size")
	}
	z.SetString(

		ssplit[0],
		ssplit[1],
	)
}

var hashToG2Vector hashTestVector
var encodeToG2Vector encodeTestVector
