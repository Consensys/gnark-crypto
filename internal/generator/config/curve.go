package config

import (
	"math/big"

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

	// NoFieldSuite disables the field-based auxiliary packages generated from Fr/Fp
	// in internal/generator/main.go: MiMC, polynomial, Poseidon2, and hash_to_field.
	// Curves generate that suite by default unless it is explicitly disabled here.
	NoFieldSuite bool

	// NoECC disables the ECC package generation (G1, G2, multiexp, marshal).
	// Used for curves that only need field arithmetic and ECDSA (e.g., stark-curve, secp256r1).
	NoECC bool

	// ECDSAKeyRecovery enables ECDSA public key recovery (SignForRecover, RecoverPublicKey).
	ECDSAKeyRecovery bool
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

func (c Curve) HasG1() bool {
	return c.G1.PointName != ""
}

func (c Curve) HasG2() bool {
	return c.G2.PointName != ""
}

func (c Curve) MarshalMetadataBits() int {
	return c.FpInfo.Bytes*8 - c.FpInfo.Bits
}

func (c Curve) SupportsPointCompression() bool {
	return c.MarshalMetadataBits() > 0
}

func (c Curve) GenerateMarshal() bool {
	return c.HasG1()
}

func (c Curve) GenerateECC() bool {
	return c.HasG1() && !c.NoECC
}

func (c Curve) GenerateFieldSuite() bool {
	return !c.NoFieldSuite
}

func (c Curve) GenerateFFT() bool {
	return c.GenerateFieldSuite() && c.GeneratePairingPackages()
}

func (c Curve) GenerateHashToCurve1() bool {
	return c.HashE1 != nil
}

func (c Curve) GenerateHashToCurve2() bool {
	return c.HashE2 != nil
}

func (c Curve) GenerateHashToCurve() bool {
	return c.GenerateHashToCurve1() || c.GenerateHashToCurve2()
}

func (c Curve) GeneratePairingPackages() bool {
	return c.HasG2()
}

func (p Point) CMax() int {
	if len(p.CRange) == 0 {
		return 0
	}
	return p.CRange[len(p.CRange)-1]
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
