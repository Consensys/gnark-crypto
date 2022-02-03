package config

import (
	"github.com/consensys/gnark-crypto/field"
	"math/big"
)

// Curve describes parameters of the curve useful for the template
type Curve struct {
	Name         string
	CurvePackage string
	Package      string // current package being generated
	EnumID       string
	FpModulus    string
	FrModulus    string

	Fp           *field.Field
	Fr           *field.Field
	FpUnusedBits int

	FpInfo, FrInfo Field
	G1             Point
	G2             Point

	HashE1 *HashSuite
}

type Isogeny struct {

	//Isogeny to original curve
	XMap RationalPolynomial
	YMap RationalPolynomial // The y map is also evaluated on x. The result is multiplied by y.
}

type RationalPolynomial struct {
	Num []string //Num is stored as a hex string
	Den []string //Den is stored as a hex string. It is also monic. The leading coefficient (1) is omitted.
}

type HashSuite struct {
	A string // A is the hex-encoded Weierstrass curve coefficient of x in the isogenous curve over which the SSWU map is evaluated.
	B string // B is the hex-encoded Weierstrass curve constant term in the isogenous curve over which the SSWU map is evaluated.

	Z []int // z (or zeta) is a quadratic non-residue with //TODO: some extra nice properties, refer to WB19

	Isogeny *Isogeny
}

type Field struct {
	Bits    int
	Bytes   int
	Modulus func() *big.Int
}

func (c Curve) Equal(other Curve) bool {
	return c.Name == other.Name
}

type Point struct {
	CoordType        string
	CoordExtDegree   uint8 // value n, such that q = pⁿ
	CoordExtRoot     int64 // value a, such that the field is Fp[X]/(Xⁿ - a)
	PointName        string
	GLV              bool  // scalar multiplication using GLV
	CofactorCleaning bool  // flag telling if the Cofactor cleaning is available
	CRange           []int // multiexp bucket method: generate inner methods (with const arrays) for each c
	Projective       bool  // generate projective coordinates
}

var Curves []Curve

func defaultCRange() []int {
	// default range for C values in the multiExp
	return []int{4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 20, 21, 22}
}

func addCurve(c *Curve) {
	// init FpInfo and FrInfo
	c.FpInfo = newFieldInfo(c.FpModulus)
	c.FrInfo = newFieldInfo(c.FrModulus)
	Curves = append(Curves, *c)
}

func newFieldInfo(modulus string) Field {
	var F Field
	var bModulus big.Int
	if _, ok := bModulus.SetString(modulus, 10); !ok {
		panic("invalid modulus " + modulus)
	}

	F.Bits = bModulus.BitLen()
	F.Bytes = len(bModulus.Bits()) * 8
	F.Modulus = func() *big.Int { return new(big.Int).Set(&bModulus) }
	return F
}

func toInt64Slice(z []int) []int64 {
	ret := make([]int64, len(z))

	for i := 0; i < len(z); i++ {
		ret[i] = int64(z[i])
	}

	return ret
}

func toBigIntSlice(z []int) []big.Int {
	res := make([]big.Int, len(z))
	for i := 0; i < len(z); i++ {
		res[i].SetInt64(int64(z[i]))
	}
	return res
}

func NewHashSuiteInfo(baseField *field.Field, g *Point, name string, suite *HashSuite) HashSuiteInfo {

	f := field.NewTower(baseField, g.CoordExtDegree, g.CoordExtRoot)
	fieldSizeMod256 := uint8(f.Size.Bits()[0])

	Z := toBigIntSlice(suite.Z)
	var c [][]big.Int

	if fieldSizeMod256%4 == 3 {
		c = make([][]big.Int, 2)
		c[0] = make([]big.Int, 1)
		c[0][0].Rsh(&f.Size, 2)

		f.Neg(&c[1], Z).Sqrt(&c[1], c[1])
		f.ToMont(&c[1], c[1])

	} else if fieldSizeMod256%8 == 5 {
		c = make([][]big.Int, 3)
		c[0] = make([]big.Int, 1)
		c[0][0].Rsh(&f.Size, 3)

		c[1] = make([]big.Int, f.Degree)
		c[1][0].SetInt64(-1)
		f.Sqrt(&c[1], c[1])

		f.Inverse(&c[2], c[1])
		f.Mul(&c[2], Z, c[2]).Sqrt(&c[2], c[2])

		f.ToMont(&c[1], c[1])
		f.ToMont(&c[2], c[2])
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
		c[0][2].Rsh(&c[1][0], 1)
		c[0][3].Sub(&twoPowC1, ONE)
		c[0][4].Rsh(&twoPowC1, 1)

		// c6, c7 stored as c[1], c[2] respectively
		f.Exp(&c[1], Z, &c[0][1])
		var c7Pow big.Int
		c7Pow.Add(&c[0][1], ONE)
		c7Pow.Rsh(&c7Pow, 1)
		f.Exp(&c[2], Z, &c7Pow)

		f.ToMont(&c[1], c[1])
		f.ToMont(&c[2], c[2])

	} else {
		panic("this is logically impossible")
	}

	return HashSuiteInfo{
		A:                f.HexToMont(suite.A),
		B:                f.HexToMont(suite.B),
		Z:                suite.Z,
		CoordType:        g.CoordType,
		CoordExtDegree:   g.CoordExtDegree,
		CofactorCleaning: g.CofactorCleaning,
		Name:             name,
		Isogeny:          newIsogenousCurveInfoOptional(f, suite.Isogeny),
		FieldSizeMod256:  fieldSizeMod256,
		SqrtRatioParams:  c,
		Field:            f,
	}
}

func newIsogenousCurveInfoOptional(f *field.Field, isogenousCurve *Isogeny) *IsogenyInfo {
	if isogenousCurve == nil {
		return nil
	}
	return &IsogenyInfo{
		XMap: RationalPolynomialInfo{
			hexSliceToIntSlice(isogenousCurve.XMap.Num, f),
			hexSliceToIntSlice(isogenousCurve.XMap.Den, f),
		},
		YMap: RationalPolynomialInfo{
			hexSliceToIntSlice(isogenousCurve.YMap.Num, f),
			hexSliceToIntSlice(isogenousCurve.YMap.Den, f),
		},
	}
}

func hexSliceToIntSlice(hexSlice []string, f *field.Field) []big.Int {
	res := make([]big.Int, len(hexSlice))

	for i, hex := range hexSlice {
		res[i] = f.HexToMont(hex)
	}

	return res
}

type IsogenyInfo struct {
	XMap RationalPolynomialInfo
	YMap RationalPolynomialInfo // The y map is also evaluated on x. The result is multiplied by y.
}

type RationalPolynomialInfo struct {
	Num []big.Int
	Den []big.Int //Denominator is monic. The leading coefficient (1) is omitted.
}

type HashSuiteInfo struct {
	//Isogeny to original curve
	Isogeny *IsogenyInfo //pointer so it's nullable.

	A big.Int
	B big.Int

	Field            *field.Field
	CoordType        string
	CoordExtDegree   uint8
	Name             string
	FieldSizeMod256  uint8
	SqrtRatioParams  []big.Int
	Z                int // z (or zeta) is a quadratic non-residue with //TODO: some extra nice properties, refer to WB19
	CofactorCleaning bool
}
