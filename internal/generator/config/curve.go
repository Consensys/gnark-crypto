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

	// E2 Cbrt Frobenius decomposition: e = e₀ + e₁·p
	E2CbrtFrobeniusE0      []uint64 // e₀ limbs (little-endian)
	E2CbrtFrobeniusE1      []uint64 // e₁ limbs (little-endian)
	E2CbrtFrobeniusNLimbs  int      // number of limbs
	E2CbrtFrobeniusMaxBit  int      // bit length of max(e₀, e₁)
	E2CbrtFrobeniusEven    bool     // whether MaxBit is even (determines loop alignment)
	E2CbrtFrobeniusTopLimb int      // limb index of the MSB: (MaxBit-1)/64
	E2CbrtFrobeniusTopBit  int      // bit position within the top limb: (MaxBit-1)%64
	E2CbrtFrobeniusStart   int      // starting bit for the 2-bit windowed loop

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

	// Frobenius decomposition: e = e₀ + e₁·p
	e1 := new(big.Int).Div(&exp, p)
	e0 := new(big.Int).Mod(&exp, p)
	maxBits := e0.BitLen()
	if e1.BitLen() > maxBits {
		maxBits = e1.BitLen()
	}
	// Number of uint64 limbs needed
	nLimbs := (maxBits + 63) / 64
	c.E2CbrtFrobeniusE0 = bigIntToLimbs(e0, nLimbs)
	c.E2CbrtFrobeniusE1 = bigIntToLimbs(e1, nLimbs)
	c.E2CbrtFrobeniusNLimbs = nLimbs
	c.E2CbrtFrobeniusMaxBit = maxBits
	c.E2CbrtFrobeniusEven = maxBits%2 == 0
	c.E2CbrtFrobeniusTopLimb = (maxBits - 1) / 64
	c.E2CbrtFrobeniusTopBit = (maxBits - 1) % 64
	if maxBits%2 == 0 {
		c.E2CbrtFrobeniusStart = maxBits - 2
	} else {
		c.E2CbrtFrobeniusStart = maxBits - 3
	}

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
