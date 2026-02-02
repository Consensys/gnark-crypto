package lattice

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

const (
	nbFuzzShort = 2
	nbFuzz      = 20
)

// BN254 curve parameters for testing
var (
	bn254r, _      = new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
	bn254Lambda, _ = new(big.Int).SetString("4407920970296243842393367215006156084916469457145843978461", 10)
)

func TestRationalReconstruct(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	r := bn254r

	properties.Property("RationalReconstruct: k = x/z mod r", prop.ForAll(
		func(kRaw *big.Int) bool {
			k := new(big.Int).Mod(kRaw, r)
			if k.Sign() == 0 {
				k.SetInt64(1)
			}

			result := RationalReconstruct(k, r)
			x, z := result[0], result[1]

			// Verify: x ≡ k*z (mod r)
			expected := new(big.Int).Mul(k, z)
			expected.Mod(expected, r)

			xMod := new(big.Int).Mod(x, r)
			if xMod.Sign() < 0 {
				xMod.Add(xMod, r)
			}

			return xMod.Cmp(expected) == 0
		},
		GenNumber(256),
	))

	properties.Property("RationalReconstruct: outputs are small (< ~1.5*r^(1/2))", prop.ForAll(
		func(kRaw *big.Int) bool {
			k := new(big.Int).Mod(kRaw, r)
			if k.Sign() == 0 {
				k.SetInt64(1)
			}

			result := RationalReconstruct(k, r)

			// Expected bound: ~1.5 * r^(1/2) (with some margin)
			// r^(1/2) for BN254 ≈ 2^127
			bound := new(big.Int).Exp(big.NewInt(2), big.NewInt(135), nil) // 2^135 with margin

			for i := 0; i < 2; i++ {
				absVal := new(big.Int).Abs(result[i])
				if absVal.Cmp(bound) > 0 {
					return false
				}
			}
			return true
		},
		GenNumber(256),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestMultiRationalReconstruct(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	r := bn254r

	properties.Property("MultiRationalReconstruct: k1 = x1/z and k2 = x2/z mod r", prop.ForAll(
		func(k1Raw, k2Raw *big.Int) bool {
			k1 := new(big.Int).Mod(k1Raw, r)
			k2 := new(big.Int).Mod(k2Raw, r)
			if k1.Sign() == 0 {
				k1.SetInt64(1)
			}
			if k2.Sign() == 0 {
				k2.SetInt64(2)
			}

			result := MultiRationalReconstruct(k1, k2, r)
			x1, x2, z := result[0], result[1], result[2]

			// Verify k1: x1 ≡ k1*z (mod r)
			expected1 := new(big.Int).Mul(k1, z)
			expected1.Mod(expected1, r)
			x1Mod := new(big.Int).Mod(x1, r)
			if x1Mod.Sign() < 0 {
				x1Mod.Add(x1Mod, r)
			}

			// Verify k2: x2 ≡ k2*z (mod r)
			expected2 := new(big.Int).Mul(k2, z)
			expected2.Mod(expected2, r)
			x2Mod := new(big.Int).Mod(x2, r)
			if x2Mod.Sign() < 0 {
				x2Mod.Add(x2Mod, r)
			}

			return x1Mod.Cmp(expected1) == 0 && x2Mod.Cmp(expected2) == 0
		},
		GenNumber(256),
		GenNumber(256),
	))

	properties.Property("MultiRationalReconstruct: outputs are small (< ~1.22*r^(2/3))", prop.ForAll(
		func(k1Raw, k2Raw *big.Int) bool {
			k1 := new(big.Int).Mod(k1Raw, r)
			k2 := new(big.Int).Mod(k2Raw, r)
			if k1.Sign() == 0 {
				k1.SetInt64(1)
			}
			if k2.Sign() == 0 {
				k2.SetInt64(2)
			}

			result := MultiRationalReconstruct(k1, k2, r)

			// Expected bound: ~1.22 * r^(2/3) (per paper, with δ=0.99)
			// Lattice has det = r², so Minkowski bound gives ~r^(2/3)
			// r^(2/3) for BN254 ≈ 2^169, so 1.22*r^(2/3) ≈ 2^170
			bound := new(big.Int).Exp(big.NewInt(2), big.NewInt(177), nil) // 2^177 with margin

			for i := 0; i < 3; i++ {
				absVal := new(big.Int).Abs(result[i])
				if absVal.Cmp(bound) > 0 {
					return false
				}
			}
			return true
		},
		GenNumber(256),
		GenNumber(256),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestRationalReconstructExt(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	// Use BN254 curve parameters
	r := bn254r
	lambda := bn254Lambda

	properties.Property("RationalReconstructExt: k = (x + λy)/(z + λt) mod r", prop.ForAll(
		func(kRaw *big.Int) bool {
			// Reduce k mod r
			k := new(big.Int).Mod(kRaw, r)
			if k.Sign() == 0 {
				k.SetInt64(1) // Avoid zero scalar
			}

			result := RationalReconstructExt(k, r, lambda)
			x, y, z, tt := result[0], result[1], result[2], result[3]

			// Verify: (x + λy) ≡ k*(z + λt) (mod r)
			// which is: x + λy - k*z - k*λ*t ≡ 0 (mod r)

			// Compute numerator: x + λy
			num := new(big.Int).Mul(lambda, y)
			num.Add(num, x)
			num.Mod(num, r)

			// Compute denominator: z + λt
			den := new(big.Int).Mul(lambda, tt)
			den.Add(den, z)
			den.Mod(den, r)

			// Verify: num ≡ k * den (mod r)
			expected := new(big.Int).Mul(k, den)
			expected.Mod(expected, r)

			return num.Cmp(expected) == 0
		},
		GenNumber(256),
	))

	properties.Property("RationalReconstructExt: outputs are small (< ~1.5*r^(1/4))", prop.ForAll(
		func(kRaw *big.Int) bool {
			k := new(big.Int).Mod(kRaw, r)
			if k.Sign() == 0 {
				k.SetInt64(1)
			}

			result := RationalReconstructExt(k, r, lambda)

			// Expected bound: ~1.5 * r^(1/4) (with some margin)
			// r^(1/4) for BN254 ≈ 2^64
			bound := new(big.Int).Exp(big.NewInt(2), big.NewInt(72), nil) // 2^72 with margin

			for i := 0; i < 4; i++ {
				absVal := new(big.Int).Abs(result[i])
				if absVal.Cmp(bound) > 0 {
					return false
				}
			}
			return true
		},
		GenNumber(256),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestMultiRationalReconstructExt(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	r := bn254r
	lambda := bn254Lambda

	properties.Property("MultiRationalReconstructExt: k1 = (x1 + λy1)/(z + λt) and k2 = (x2 + λy2)/(z + λt) mod r", prop.ForAll(
		func(k1Raw, k2Raw *big.Int) bool {
			k1 := new(big.Int).Mod(k1Raw, r)
			k2 := new(big.Int).Mod(k2Raw, r)
			if k1.Sign() == 0 {
				k1.SetInt64(1)
			}
			if k2.Sign() == 0 {
				k2.SetInt64(2)
			}

			result := MultiRationalReconstructExt(k1, k2, r, lambda)
			x1, y1, x2, y2, z, tt := result[0], result[1], result[2], result[3], result[4], result[5]

			// Compute denominator: z + λt
			den := new(big.Int).Mul(lambda, tt)
			den.Add(den, z)
			den.Mod(den, r)

			// Verify k1: (x1 + λy1) ≡ k1 * (z + λt) (mod r)
			num1 := new(big.Int).Mul(lambda, y1)
			num1.Add(num1, x1)
			num1.Mod(num1, r)
			expected1 := new(big.Int).Mul(k1, den)
			expected1.Mod(expected1, r)

			// Verify k2: (x2 + λy2) ≡ k2 * (z + λt) (mod r)
			num2 := new(big.Int).Mul(lambda, y2)
			num2.Add(num2, x2)
			num2.Mod(num2, r)
			expected2 := new(big.Int).Mul(k2, den)
			expected2.Mod(expected2, r)

			return num1.Cmp(expected1) == 0 && num2.Cmp(expected2) == 0
		},
		GenNumber(256),
		GenNumber(256),
	))

	properties.Property("MultiRationalReconstructExt: outputs are small (< ~1.28*r^(1/3))", prop.ForAll(
		func(k1Raw, k2Raw *big.Int) bool {
			k1 := new(big.Int).Mod(k1Raw, r)
			k2 := new(big.Int).Mod(k2Raw, r)
			if k1.Sign() == 0 {
				k1.SetInt64(1)
			}
			if k2.Sign() == 0 {
				k2.SetInt64(2)
			}

			result := MultiRationalReconstructExt(k1, k2, r, lambda)

			// Expected bound: ~1.28 * r^(1/3) (per paper, with δ=0.99)
			// r^(1/3) for BN254 ≈ 2^85, so 1.28*r^(1/3) ≈ 2^85
			bound := new(big.Int).Exp(big.NewInt(2), big.NewInt(92), nil) // 2^92 with margin

			for i := 0; i < 6; i++ {
				absVal := new(big.Int).Abs(result[i])
				if absVal.Cmp(bound) > 0 {
					return false
				}
			}
			return true
		},
		GenNumber(256),
		GenNumber(256),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

var benchRationalReconstructRes [2]*big.Int

func BenchmarkRationalReconstruct(b *testing.B) {
	k, _ := new(big.Int).SetString("12345678901234567890123456789012345678901234567890", 10)
	k.Mod(k, bn254r)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchRationalReconstructRes = RationalReconstruct(k, bn254r)
	}
}

var benchMultiRationalReconstructRes [3]*big.Int

func BenchmarkMultiRationalReconstruct(b *testing.B) {
	k1, _ := new(big.Int).SetString("12345678901234567890123456789012345678901234567890", 10)
	k2, _ := new(big.Int).SetString("98765432109876543210987654321098765432109876543210", 10)
	k1.Mod(k1, bn254r)
	k2.Mod(k2, bn254r)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchMultiRationalReconstructRes = MultiRationalReconstruct(k1, k2, bn254r)
	}
}

var benchRationalReconstructExtRes [4]*big.Int

func BenchmarkRationalReconstructExt(b *testing.B) {
	k, _ := new(big.Int).SetString("12345678901234567890123456789012345678901234567890", 10)
	k.Mod(k, bn254r)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchRationalReconstructExtRes = RationalReconstructExt(k, bn254r, bn254Lambda)
	}
}

var benchMultiRationalReconstructExtRes [6]*big.Int

func BenchmarkMultiRationalReconstructExt(b *testing.B) {
	k1, _ := new(big.Int).SetString("12345678901234567890123456789012345678901234567890", 10)
	k2, _ := new(big.Int).SetString("98765432109876543210987654321098765432109876543210", 10)
	k1.Mod(k1, bn254r)
	k2.Mod(k2, bn254r)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchMultiRationalReconstructExtRes = MultiRationalReconstructExt(k1, k2, bn254r, bn254Lambda)
	}
}

// GenNumber generates a random integer
func GenNumber(boundSize int64) gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var bound big.Int
		bound.Exp(big.NewInt(2), big.NewInt(boundSize), nil)
		elmt, _ := rand.Int(genParams.Rng, &bound)
		genResult := gopter.NewGenResult(elmt, gopter.NoShrinker)
		return genResult
	}
}
