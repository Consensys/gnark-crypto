package config

import (
	"math/big"

	"github.com/consensys/gnark-crypto/field"
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

	Hash     HashSuite
	HashInfo HashSuiteInfo
}

type IsogenousCurve struct {
	AHex string
	BHex string

	Z int // z (or zeta) is a quadratic non-residue with //TODO: some extra nice properties, refer to WB19

	//Isogeny to original curve
	XMap RationalPolynomial
	YMap RationalPolynomial // The y map is also evaluated on x. The result is multiplied by y.
}

type RationalPolynomial struct {
	NumHex []string
	DenHex []string //Denominator is monic. The leading coefficient (1) is omitted.
}

type HashSuite struct {
	E1Prime *IsogenousCurve //pointer so it's nullable. TODO: Bad practice or ok?
	//E2Prime *IsogenousCurve
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
	c.HashInfo = newHashSuiteInfo(c.FpInfo.Modulus(), &c.Hash)
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

func newHashSuiteInfo(fieldModulus *big.Int, suite *HashSuite) HashSuiteInfo {
	return HashSuiteInfo{
		newIsogenousCurveInfoOptional(fieldModulus, suite.E1Prime),
	}
}

func newIsogenousCurveInfoOptional(fieldModulus *big.Int, isogenousCurve *IsogenousCurve) *IsogenousCurveInfo {
	if isogenousCurve == nil {
		return nil
	}
	return &IsogenousCurveInfo{
		A: field.HexToMontSlice(fieldModulus, isogenousCurve.AHex),
		B: field.HexToMontSlice(fieldModulus, isogenousCurve.BHex),
		Z: isogenousCurve.Z,
		XMap: RationalPolynomialInfo{
			hexSliceToMontSliceSlice(fieldModulus, isogenousCurve.XMap.NumHex),
			hexSliceToMontSliceSlice(fieldModulus, isogenousCurve.XMap.DenHex),
		},
		YMap: RationalPolynomialInfo{
			hexSliceToMontSliceSlice(fieldModulus, isogenousCurve.YMap.NumHex),
			hexSliceToMontSliceSlice(fieldModulus, isogenousCurve.YMap.DenHex),
		},
	}
}

func hexSliceToMontSliceSlice(fieldModulus *big.Int, hexSlice []string) [][]uint64 {
	res := make([][]uint64, len(hexSlice))

	for i, hex := range hexSlice {
		res[i] = field.HexToMontSlice(fieldModulus, hex)
	}

	return res
}

type IsogenousCurveInfo struct {
	A []uint64
	B []uint64

	Z int // z (or zeta) is a quadratic non-residue with //TODO: some extra nice properties, refer to WB19

	//Isogeny to original curve
	XMap RationalPolynomialInfo
	YMap RationalPolynomialInfo // The y map is also evaluated on x. The result is multiplied by y.
}

type RationalPolynomialInfo struct {
	Num [][]uint64
	Den [][]uint64 //Denominator is monic. The leading coefficient (1) is omitted.
}

type HashSuiteInfo struct {
	E1Prime *IsogenousCurveInfo //pointer so it's nullable. TODO: Bad practice or ok?
	//E2Prime *IsogenousCurve
}
