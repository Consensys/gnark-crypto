package config

import (
	"math/big"

	"github.com/consensys/gnark-crypto/internal/generator/addchain"
	"github.com/consensys/gnark-crypto/internal/generator/field/config"
)

// Curve describes parameters of the curve useful for the template
type Curve struct {
	Name         string
	CurvePackage string
	Package      string // current package being generated
	EnumID       string
	FpModulus    string
	FrModulus    string

	Fp           *config.Field
	Fr           *config.Field
	FpUnusedBits int

	FpInfo, FrInfo FieldInfo
	G1             Point
	G2             Point

	HashE1 HashSuite
	HashE2 HashSuite

	// E2 Cbrt precomputes (for Fp² cube root)
	E2CbrtP2Mod9       uint64                 // p² mod 9
	E2CbrtP2Mod27      uint64                 // p² mod 27
	E2CbrtExponentHex  string                 // precomputed exponent as hex string
	E2CbrtExponentData *addchain.AddChainData // addition chain data for E2 Cbrt exponentiation
}

type TwistedEdwardsCurve struct {
	Name    string
	Package string
	EnumID  string

	A, D, Cofactor, Order, BaseX, BaseY string

	// set if endomorphism
	HasEndomorphism bool
	Endo0, Endo1    string
	Lambda          string
}

type FieldInfo struct {
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
	GLV              bool     // scalar multiplication using GLV
	CofactorCleaning bool     // flag telling if the Cofactor cleaning is available
	CRange           []int    // multiexp bucket method: generate inner methods (with const arrays) for each c
	Projective       bool     // generate projective coordinates
	A                []string //A linear coefficient in Weierstrass form
	B                []string //B constant term in Weierstrass form
}

var Curves []Curve
var TwistedEdwardsCurves []TwistedEdwardsCurve

func defaultCRange() []int {
	// default range for C values in the multiExp
	return []int{4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
}

func addCurve(c *Curve) {
	// init FpInfo and FrInfo
	c.FpInfo = newFieldInfo(c.FpModulus)
	c.FrInfo = newFieldInfo(c.FrModulus)

	// Compute E2 Cbrt precomputes
	p := c.FpInfo.Modulus()
	p2 := new(big.Int).Mul(p, p)
	c.E2CbrtP2Mod9 = new(big.Int).Mod(p2, big.NewInt(9)).Uint64()
	c.E2CbrtP2Mod27 = new(big.Int).Mod(p2, big.NewInt(27)).Uint64()

	// Compute the E2 Cbrt exponent based on p² mod 9/27
	var exp big.Int
	switch c.E2CbrtP2Mod9 {
	case 7:
		// p² ≡ 7 (mod 9): exponent = (p²+2)/9
		exp.Add(p2, big.NewInt(2))
		exp.Div(&exp, big.NewInt(9))
	case 4:
		// p² ≡ 4 (mod 9): exponent = (2p²+1)/9
		exp.Lsh(p2, 1)
		exp.Add(&exp, big.NewInt(1))
		exp.Div(&exp, big.NewInt(9))
	default:
		// p² ≡ 1 (mod 9): need p² mod 27
		switch c.E2CbrtP2Mod27 {
		case 10:
			// p² ≡ 10 (mod 27): exponent = (2p²+7)/27
			exp.Lsh(p2, 1)
			exp.Add(&exp, big.NewInt(7))
			exp.Div(&exp, big.NewInt(27))
		case 19:
			// p² ≡ 19 (mod 27): exponent = (p²+8)/27
			exp.Add(p2, big.NewInt(8))
			exp.Div(&exp, big.NewInt(27))
		default:
			// Generic fallback: exponent = (2(p²-1)+1)/3
			exp.Sub(p2, big.NewInt(1))
			exp.Lsh(&exp, 1)
			exp.Add(&exp, big.NewInt(1))
			exp.Div(&exp, big.NewInt(3))
		}
	}
	c.E2CbrtExponentHex = exp.Text(16)
	c.E2CbrtExponentData = addchain.GetAddChain(&exp)

	Curves = append(Curves, *c)
}

func addTwistedEdwardCurve(c *TwistedEdwardsCurve) {
	TwistedEdwardsCurves = append(TwistedEdwardsCurves, *c)
}

func newFieldInfo(modulus string) FieldInfo {
	var F FieldInfo
	var bModulus big.Int
	if _, ok := bModulus.SetString(modulus, 10); !ok {
		panic("invalid modulus " + modulus)
	}

	F.Bits = bModulus.BitLen()
	F.Bytes = (F.Bits + 7) / 8
	F.Modulus = func() *big.Int { return new(big.Int).Set(&bModulus) }
	return F
}
