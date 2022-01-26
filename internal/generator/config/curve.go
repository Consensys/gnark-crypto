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

	Z int // z (or zeta) is a quadratic non-residue with //TODO: some extra nice properties, refer to WB19

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
	CoordExtDegree   uint8
	PointName        string
	GLV              bool  // scalar mulitplication using GLV
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

func NewHashSuiteInfo(fieldModulus *big.Int, g *Point, name string, suite *HashSuite) HashSuiteInfo {

	var fieldSize big.Int
	fieldSize.Exp(fieldModulus, big.NewInt(int64(g.CoordExtDegree)), nil)
	fieldSizeMod256 := uint8(fieldSize.Bits()[0])

	Z := int64(suite.Z)
	var c []big.Int

	//TODO: Works only for fp
	if fieldSizeMod256%4 == 3 {
		c = make([]big.Int, 2)
		c[0].Rsh(&fieldSize, 2)

		c[1].SetInt64(-Z)
		c[1].ModSqrt(&c[1], fieldModulus)
		field.IntToMont(&c[1], fieldModulus)

	} else if fieldSizeMod256%8 == 5 {
		c = make([]big.Int, 3)
		c[0].Rsh(&fieldSize, 3)

		c[1].SetInt64(-1)
		c[1].ModSqrt(&c[1], fieldModulus)

		c[2].ModInverse(&c[1], fieldModulus)
		c[2].Mul(&c[2], big.NewInt(Z))

		c[2].ModSqrt(&c[2], fieldModulus)

		field.IntToMont(&c[1], fieldModulus)
		field.IntToMont(&c[2], fieldModulus)
	} else if fieldSizeMod256%8 == 1 {
		ONE := big.NewInt(1)
		c = make([]big.Int, 7)

		c[0].Sub(&fieldSize, big.NewInt(1))
		c1 := c[0].TrailingZeroBits()
		c[0].SetUint64(uint64(c1))

		var twoPowC1 big.Int
		twoPowC1.Lsh(ONE, c1)
		c[1].Rsh(&fieldSize, c1)
		c[2].Rsh(&c[1], 1)
		c[3].Sub(&twoPowC1, ONE)
		c[4].Rsh(&twoPowC1, 1)
		powMod(&c[5], big.NewInt(Z), &c[1], fieldModulus)
		var c7Pow big.Int
		c7Pow.Add(&c[1], ONE)
		c7Pow.Rsh(&c7Pow, 1)
		powMod(&c[6], big.NewInt(Z), &c7Pow, fieldModulus)

		field.IntToMont(&c[5], fieldModulus)
		field.IntToMont(&c[6], fieldModulus)

	} else {
		panic("this is logically impossible")
	}

	return HashSuiteInfo{
		A:                field.HexToMont(suite.A, fieldModulus),
		B:                field.HexToMont(suite.B, fieldModulus),
		Z:                suite.Z,
		CoordType:        g.CoordType,
		CofactorCleaning: g.CofactorCleaning,
		Name:             name,
		Isogeny:          newIsogenousCurveInfoOptional(fieldModulus, suite.Isogeny),
		FieldSizeMod256:  fieldSizeMod256,
		SqrtRatioParams:  c,
	}
}

func powMod(res *big.Int, x *big.Int, pow *big.Int, mod *big.Int) *big.Int {
	res.SetInt64(1)

	for i := pow.BitLen() - 1; ; {

		if pow.Bit(i) == 1 {
			res.Mul(res, x)
		}

		if i == 0 {
			break
		}
		i--

		res.Mul(res, res).Mod(res, mod)
	}

	res.Mod(res, mod)
	return res
}

func newIsogenousCurveInfoOptional(fieldModulus *big.Int, isogenousCurve *Isogeny) *IsogenyInfo {
	if isogenousCurve == nil {
		return nil
	}
	return &IsogenyInfo{
		XMap: RationalPolynomialInfo{
			hexSliceToIntSlice(isogenousCurve.XMap.Num, fieldModulus),
			hexSliceToIntSlice(isogenousCurve.XMap.Den, fieldModulus),
		},
		YMap: RationalPolynomialInfo{
			hexSliceToIntSlice(isogenousCurve.YMap.Num, fieldModulus),
			hexSliceToIntSlice(isogenousCurve.YMap.Den, fieldModulus),
		},
	}
}

func hexSliceToIntSlice(hexSlice []string, fieldModulus *big.Int) []big.Int {
	res := make([]big.Int, len(hexSlice))

	for i, hex := range hexSlice {
		res[i] = field.HexToMont(hex, fieldModulus)
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
	Isogeny *IsogenyInfo //pointer so it's nullable. TODO: Bad practice or ok?

	A big.Int
	B big.Int

	CoordType        string
	Name             string
	FieldSizeMod256  uint8
	SqrtRatioParams  []big.Int
	Z                int // z (or zeta) is a quadratic non-residue with //TODO: some extra nice properties, refer to WB19
	CofactorCleaning bool
}
