package config

import (
	"github.com/consensys/gnark-crypto/field"
	"math/big"
)

//TODO: Same data structure used in runtime and for code generation. HashInfo not needed for runtime

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

	HashE1     *HashSuite
	HashInfoE1 *HashSuiteInfo
}

type Isogeny struct {

	//Isogeny to original curve
	XMap RationalPolynomial
	YMap RationalPolynomial // The y map is also evaluated on x. The result is multiplied by y.
}

type RationalPolynomial struct {
	NumHex []string
	DenHex []string //Denominator is monic. The leading coefficient (1) is omitted.
}

type HashSuite struct {
	AHex string
	BHex string

	Z int // z (or zeta) is a quadratic non-residue with //TODO: some extra nice properties, refer to WB19

	Isogeny *Isogeny //pointer so it's nullable. TODO: Bad practice or ok?
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
	// c.Fp is nil here. TODO: Why? Fix if no good reason
	c.HashInfoE1 = newHashSuiteInfo(c.FpInfo.Modulus(), &c.G1, c.HashE1)
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

// stdSqrt is an in-place standardized square root, returning the even root. It assumes quadratic residuosity.
func stdSqrt(a *big.Int, modulus *big.Int) {
	a.ModSqrt(a, modulus)
	//take standard value with sgn == 0
	if a.Bit(0) == 1 {
		a.Sub(modulus, a)
	}
}

// twoFactor returns the greatest m such that 2ᵐ divides x-1
func twoFactor(x *big.Int) int {
	if x.BitLen() <= 1 {
		return -1 //substitute for ∞
	}
	if x.Bit(0) == 0 {
		return 0 //rightmost bit of x-1 is 1
	}
	var m int
	for m = 1; x.Bit(m) == 0; m++ {
	}
	return m
}

func newHashSuiteInfo(fieldModulus *big.Int, G *Point, suite *HashSuite) *HashSuiteInfo {

	if suite == nil {
		//return nil
		// Make up some Z value to test sqrtRatio with
		Z := int64(0)

		for {
			if legendreFp(big.NewInt(Z), fieldModulus) == -1 {
				break
			}
			if legendreFp(big.NewInt(-Z), fieldModulus) == -1 {
				Z = -Z
				break
			}
			Z += 1
		}

		suite = &HashSuite{Z: int(Z)}
	}

	fieldSize := pow(fieldModulus, G.CoordExtDegree)
	fieldSizeMod256 := uint8(fieldSize.Bits()[0])

	Z := int64(suite.Z)
	var c []big.Int

	//TODO: Works only for fp
	if fieldSizeMod256%4 == 3 {
		c = make([]big.Int, 2)
		c[0].Rsh(fieldSize, 2)

		c[1].SetInt64(-Z)
		c[1].ModSqrt(&c[1], fieldModulus)
		field.IntToMont(&c[1], fieldModulus)

	} else if fieldSizeMod256%8 == 5 {
		c = make([]big.Int, 3)
		c[0].Rsh(fieldSize, 3)

		c[1].SetInt64(-1)
		c[1].ModSqrt(&c[1], fieldModulus)

		c[2].DivMod(big.NewInt(Z), &c[1], fieldModulus)
	} else if fieldSizeMod256%8 == 1 {
		ONE := big.NewInt(1)
		c = make([]big.Int, 7)
		c1 := twoFactor(fieldSize)
		c[0].SetInt64(int64(c1))
		var twoPowC1 big.Int
		twoPowC1.Lsh(ONE, uint(c1))
		c[1].Rsh(fieldSize, uint(c1))
		c[2].Rsh(&c[1], 1)
		c[3].Sub(&twoPowC1, ONE)
		c[4].Rsh(&twoPowC1, 1)
		powMod(&c[5], big.NewInt(Z), &c[1], fieldModulus)
		var c7Pow big.Int
		c7Pow.Add(&c[1], ONE)
		c7Pow.Rsh(&c7Pow, 1)

	} else {
		panic("this is logically impossible")
	}

	return &HashSuiteInfo{
		A:               field.HexToMont(suite.AHex, fieldModulus),
		B:               field.HexToMont(suite.BHex, fieldModulus),
		Z:               suite.Z,
		Isogeny:         newIsogenousCurveInfoOptional(fieldModulus, suite.Isogeny),
		FieldSizeMod256: fieldSizeMod256,
		SqrtRatioParams: c,
	}
}

//TODO: We probably have lots of duplicated ad-hoc utility funcs
func legendreFp(x *big.Int, mod *big.Int) int {
	var pow big.Int
	pow.Rsh(mod, 1)
	var resBig big.Int
	powMod(&resBig, x, &pow, mod)
	if !resBig.IsInt64() {
		panic("legendre fail")
	}
	res := resBig.Int64()
	if res > 1 || res < -1 {
		panic("abs legendre too large")
	}
	return int(res)
}

func powMod(res *big.Int, x *big.Int, pow *big.Int, mod *big.Int) *big.Int {
	res.SetInt64(1)

	for i := pow.BitLen(); ; {

		if pow.Bit(i) == 1 {
			res.Add(res, x)
		}

		if i == 0 {
			break
		}
		i--

		res.Lsh(res, 1)
		res.Mod(res, mod)
	}

	res.Mod(res, mod)
	return res
}

func pow(p *big.Int, pow uint8) *big.Int {

	res := big.NewInt(1)

	for ; pow != 0 && pow&128 == 0; pow *= 2 {
	}

	for {
		if pow&128 != 0 {
			res.Mul(res, p)
		}
		pow *= 2
		if pow != 0 {
			res.Lsh(res, 1)
		} else {
			break
		}
	}
	return res
}

func newIsogenousCurveInfoOptional(fieldModulus *big.Int, isogenousCurve *Isogeny) *IsogenyInfo {
	if isogenousCurve == nil {
		return nil
	}
	return &IsogenyInfo{
		XMap: RationalPolynomialInfo{
			hexSliceToIntSlice(isogenousCurve.XMap.NumHex, fieldModulus),
			hexSliceToIntSlice(isogenousCurve.XMap.DenHex, fieldModulus),
		},
		YMap: RationalPolynomialInfo{
			hexSliceToIntSlice(isogenousCurve.YMap.NumHex, fieldModulus),
			hexSliceToIntSlice(isogenousCurve.YMap.DenHex, fieldModulus),
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

	FieldSizeMod256 uint8
	SqrtRatioParams []big.Int
	Z               int // z (or zeta) is a quadratic non-residue with //TODO: some extra nice properties, refer to WB19
}
