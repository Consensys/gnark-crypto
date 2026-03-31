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

	// E2 Cbrt precomputes (for Fp² cube root)
	E2CbrtP2Mod9  uint64 // p² mod 9
	E2CbrtP2Mod27 uint64 // p² mod 27
	// Torus-based E2 Cbrt
	E2CbrtTorusEnabled       bool     // whether torus cbrt is available
	E2CbrtTorusBeta          int64    // beta from Fp2=Fp[u]/(u²-beta), e.g. -1 or -5
	E2CbrtTorusBetaAbs       int64    // |beta|
	E2CbrtTorusBetaInvNeg    []uint64 // 1/|beta| in Montgomery form (when beta != -1)
	E2CbrtTorusHelperSquared bool     // m = norm·t² (true) or norm·t (false)
	E2CbrtTorusUniqueRoot    bool     // q ≡ 7 mod 9: no ζ-adjustment needed
	E2CbrtTorusNormInvM      int      // power of m in normInv = m^a · t^b
	E2CbrtTorusNormInvT      int      // power of t in normInv = m^a · t^b
	E2CbrtTorusLucasExponent []uint64 // 3⁻¹ mod (p+1) little-endian
	E2CbrtTorusLucasNLimbs   int
	E2CbrtTorusLucasTopBit   int // bit length - 1
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
	GLV              bool   // scalar multiplication using GLV
	CofactorCleaning bool   // flag telling if the Cofactor cleaning is available
	CRange           []int  // multiexp bucket method: generate inner methods (with const arrays) for each c
	Projective       bool   // generate projective coordinates
	A                string // A linear coefficient in Weierstrass form y²=x³+ax+b
	B                string // B constant term in Weierstrass form y²=x³+ax+b
}

func (p Point) IsZeroA() bool {
	return p.A == "" || p.A == "0"
}

func (p Point) IsNeg3A() bool {
	return p.A == "-3"
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

	// Torus-based E2 Cbrt parameters
	// Check q mod 3 == 1 (needed for torus cbrt helper)
	pMod3 := new(big.Int).Mod(p, big.NewInt(3)).Uint64()
	if pMod3 == 1 {
		qMod9Torus := new(big.Int).Mod(p, big.NewInt(9)).Uint64()
		qMod27Torus := new(big.Int).Mod(p, big.NewInt(27)).Uint64()

		// Check if one of the supported cases
		var hasTorus bool
		switch {
		case qMod9Torus == 7, qMod9Torus == 4:
			hasTorus = true
		case qMod27Torus == 10, qMod27Torus == 19:
			hasTorus = true
		}

		if hasTorus {
			c.E2CbrtTorusEnabled = true

			// beta from Fp2 = Fp[u]/(u² - beta)
			c.E2CbrtTorusBeta = c.G2.CoordExtRoot
			if c.E2CbrtTorusBeta == 0 {
				c.E2CbrtTorusBeta = -1 // BN254 default
			}
			c.E2CbrtTorusBetaAbs = -c.E2CbrtTorusBeta

			c.E2CbrtTorusHelperSquared = (qMod27Torus == 10)
			c.E2CbrtTorusUniqueRoot = (qMod9Torus == 7)

			// normInv = m^a · t^b where a = r-2, b = d - (r-2)*s
			var d, r uint64
			var s int
			switch {
			case qMod9Torus == 7:
				d, r, s = 9, 7, 1
			case qMod9Torus == 4:
				d, r, s = 9, 4, 1
			case qMod27Torus == 10:
				d, r, s = 27, 10, 2
			case qMod27Torus == 19:
				d, r, s = 27, 19, 1
			}
			c.E2CbrtTorusNormInvM = int(r - 2)
			c.E2CbrtTorusNormInvT = int(d) - c.E2CbrtTorusNormInvM*s

			// Lucas exponent: 3⁻¹ mod (p+1)
			pPlus1 := new(big.Int).Add(p, big.NewInt(1))
			three := big.NewInt(3)
			lucasExp := new(big.Int).ModInverse(three, pPlus1)
			nLucasLimbs := (lucasExp.BitLen() + 63) / 64
			c.E2CbrtTorusLucasExponent = bigIntToLimbs(lucasExp, nLucasLimbs)
			c.E2CbrtTorusLucasNLimbs = nLucasLimbs
			c.E2CbrtTorusLucasTopBit = lucasExp.BitLen() - 1
		}
	}

	Curves = append(Curves, *c)
}

func addTwistedEdwardCurve(c *TwistedEdwardsCurve) {
	TwistedEdwardsCurves = append(TwistedEdwardsCurves, *c)
}

// bigIntToLimbs converts a big.Int to a little-endian slice of uint64 limbs.
func bigIntToLimbs(n *big.Int, nLimbs int) []uint64 {
	limbs := make([]uint64, nLimbs)
	mask := new(big.Int).SetUint64(^uint64(0))
	tmp := new(big.Int).Set(n)
	for i := 0; i < nLimbs; i++ {
		limbs[i] = new(big.Int).And(tmp, mask).Uint64()
		tmp.Rsh(tmp, 64)
	}
	return limbs
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
