package config

import (
	"fmt"
	"math/big"
	"math/bits"

	"github.com/consensys/gnark-crypto/field"
)

type FieldElementToCurvePoint string

const (
	SSWU FieldElementToCurvePoint = "SSWU"
	//SVDW FieldElementToCurvePoint = "SVDW"
)

type Isogeny struct {

	//Isogeny to original curve
	XMap RationalPolynomial
	YMap RationalPolynomial // The y map is also evaluated on x. The result is multiplied by y.
}

type RationalPolynomial struct {
	Num [][]string //Num is stored
	Den [][]string //Den is stored. It is also monic. The leading coefficient (1) is omitted.
}

type HashSuite struct {
	A []string // A is the Weierstrass curve coefficient of x in the isogenous curve over which the SSWU map is evaluated.
	B []string // B is the Weierstrass curve constant term in the isogenous curve over which the SSWU map is evaluated.

	Z []int // z (or zeta) is a quadratic non-residue with //TODO: some extra nice properties, refer to WB19

	Isogeny *Isogeny
}

type WeierstrassCoefficients struct {
	f *field.Extension
	A field.Element
	B field.Element
}

func (c *WeierstrassCoefficients) g(x field.Element) field.Element {
	gx := c.f.Mul(x, x)
	gx = c.f.Add(gx, c.A)
	gx = c.f.Mul(gx, x)
	gx = c.f.Add(gx, c.B)
	return gx
}

func NewWeierstrassCoefficients(f *field.Extension, A []string, B []string) WeierstrassCoefficients {
	return WeierstrassCoefficients{
		f: f,
		A: field.NewElement(A),
		B: field.NewElement(B),
	}
}

func toBigIntSlice(z []int) []big.Int {
	res := make([]big.Int, len(z))
	for i := 0; i < len(z); i++ {
		res[i].SetInt64(int64(z[i]))
	}
	return res
}

func findSvdwZ(f *field.Extension, curve WeierstrassCoefficients) field.Element {

	fourA := f.Mul(f.FromInt64(4), curve.A)
	negHalf := f.Div(f.FromInt64(-1), f.FromInt64(2))
	checkCandidate := func(z field.Element) bool {
		gZ := curve.g(z)
		if f.IsZero(gZ) {
			return false
		}

		h := f.Div(
			f.Add(
				f.Mul(f.FromInt64(3), f.Mul(gZ, gZ)),
				fourA,
			),
			f.Mul(
				f.FromInt64(-4),
				gZ,
			),
		)

		if f.IsZero(h) {
			return false
		}
		if f.Sqrt(h) == nil {
			return false
		}

		if f.Sqrt(gZ) != nil {
			return true
		}

		return f.Sqrt(curve.g(f.Mul(z, negHalf))) != nil
	}

	for _z := int64(1); _z < 1024; _z++ {
		z := f.FromInt64(_z)
		if checkCandidate(z) {
			return z
		}
		z = f.FromInt64(-_z)
		if checkCandidate(z) {
			return z
		}
	}

}

func NewHashSuiteInfoSvdW(baseField *field.Field, g *Point, curve *Curve, name string) HashSuiteInfo {

	f := field.NewTower(baseField, g.CoordExtDegree, g.CoordExtRoot)
	c := make([]big.Int, 4)

}

//TODO: Find Z automatically
func NewHashSuiteInfoSSWU(baseField *field.Field, g *Point, name string, suite *HashSuite) HashSuiteInfo {

	f := field.NewTower(baseField, g.CoordExtDegree, g.CoordExtRoot)
	fieldSizeMod256 := uint8(f.Size.Bits()[0])

	Z := toBigIntSlice(suite.Z)
	var c [][]big.Int

	if fieldSizeMod256%4 == 3 {
		c = make([][]big.Int, 2)
		c[0] = make([]big.Int, 1)
		c[0][0].Rsh(&f.Size, 2)

		c[1] = f.Neg(Z)
		c[1] = f.Sqrt(c[1])

	} else if fieldSizeMod256%8 == 5 {
		c = make([][]big.Int, 3)
		c[0] = make([]big.Int, 1)
		c[0][0].Rsh(&f.Size, 3)

		c[1] = make([]big.Int, f.Degree)
		c[1][0].SetInt64(-1)
		c[1] = f.Sqrt(c[1])

		c[2] = f.Inverse(c[1])
		c[2] = f.Mul(Z, c[2])
		c[2] = f.Sqrt(c[2])

	} else if fieldSizeMod256%8 == 1 {
		ONE := big.NewInt(1)
		c = make([][]big.Int, 3)

		c[0] = make([]big.Int, 5)
		// c1 .. c5 stored as c[0][0] .. c[0][4]
		c[0][0].Sub(&f.Size, big.NewInt(1))
		c1 := c[0][0].TrailingZeroBits()
		c[0][0].SetUint64(uint64(c1))

		var twoPowC1 big.Int
		twoPowC1.Lsh(ONE, c1)
		c[0][1].Rsh(&f.Size, c1)
		c[0][2].Rsh(&c[0][1], 1)
		c[0][3].Sub(&twoPowC1, ONE)
		c[0][4].Rsh(&twoPowC1, 1)

		// c6, c7 stored as c[1], c[2] respectively
		c[1] = f.Exp(Z, &c[0][1])

		var c7Pow big.Int
		c7Pow.Add(&c[0][1], ONE)
		c7Pow.Rsh(&c7Pow, 1)
		c[2] = f.Exp(Z, &c7Pow)

	} else {
		panic("this is logically impossible")
	}

	return HashSuiteInfo{
		A:                 field.NewElement(suite.A),
		B:                 field.NewElement(suite.B),
		Z:                 Z,
		Point:             g,
		CofactorClearing:  g.CofactorCleaning,
		Name:              name,
		Isogeny:           newIsogenousCurveInfoOptional(suite.Isogeny),
		FieldSizeMod256:   fieldSizeMod256,
		PrecomputedParams: c,
		Field:             &f,
		FieldCoordName:    coordNameForExtensionDegree(g.CoordExtDegree),
	}
}

func coordNameForExtensionDegree(degree uint8) string {
	switch degree {
	case 1:
		return ""
	case 2:
		return "A"
	case 6:
		return "B"
	case 12:
		return "C"
	}
	panic(fmt.Sprint("unknown extension degree", degree))
}

func stringMatrixToIntMatrix(s [][]string) []field.Element {
	res := make([]field.Element, len(s))
	for i, S := range s {
		res[i] = field.NewElement(S)
	}
	return res
}

func newIsogenousCurveInfoOptional(isogenousCurve *Isogeny) *IsogenyInfo {
	if isogenousCurve == nil {
		return nil
	}
	return &IsogenyInfo{
		XMap: RationalPolynomialInfo{
			stringMatrixToIntMatrix(isogenousCurve.XMap.Num),
			stringMatrixToIntMatrix(isogenousCurve.XMap.Den),
		},
		YMap: RationalPolynomialInfo{
			stringMatrixToIntMatrix(isogenousCurve.YMap.Num),
			stringMatrixToIntMatrix(isogenousCurve.YMap.Den),
		},
	}
}

type IsogenyInfo struct {
	XMap RationalPolynomialInfo
	YMap RationalPolynomialInfo // The y map is also evaluated on x. The result is multiplied by y.
}

type RationalPolynomialInfo struct {
	Num []field.Element
	Den []field.Element //Denominator is monic. The leading coefficient (1) is omitted.
}

type HashSuiteInfo struct {
	//Isogeny to original curve
	Isogeny *IsogenyInfo //pointer so it's nullable.

	A []big.Int
	B []big.Int

	Point             *Point
	Field             *field.Extension
	FieldCoordName    string
	Name              string
	FieldSizeMod256   uint8
	PrecomputedParams [][]big.Int // PrecomputedParams[0][n] correspond to integer cₙ₋₁ in std doc
	// PrecomputedParams[n≥1] correspond to field element c_( len(PrecomputedParams[0]) + n - 1 ) in std doc
	Z                []big.Int // z (or zeta) is a quadratic non-residue with //TODO: some extra nice properties, refer to WB19
	CofactorClearing bool
}
