package ecc

import (
	"math/big"
	"testing"
)

func TestNafDecomposition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string // large number in decimal form
		expected []int8 // expected NAF representation
	}{
		{"13", []int8{1, 0, -1, 0, 1}},    // existing test case
		{"0", []int8{}},                   // edge case - zero
		{"1", []int8{1}},                  // edge case - one
		{"7", []int8{-1, 0, 0, 1}},        // 7 = 2³ - 2⁰ (8 - 1)
		{"15", []int8{-1, 0, 0, 0, 1}},    // 15 = 2⁴ - 2⁰
		{"31", []int8{-1, 0, 0, 0, 0, 1}}, // 31 = 2⁵ - 2⁰
	}

	for i, test := range tests {
		input, success := new(big.Int).SetString(test.input, 10)
		if !success {
			t.Errorf("Failed to parse input number %s", test.input)
			continue
		}

		var result [400]int8
		length := NafDecomposition(input, result[:])
		naf := result[:length]

		// Length check
		if len(naf) != len(test.expected) {
			t.Errorf("Test %d: Incorrect length for input %s. Got %d, want %d",
				i, test.input, len(naf), len(test.expected))
			continue
		}

		// Value check
		for j := range naf {
			if naf[j] != test.expected[j] {
				t.Errorf("Test %d: Mismatch at position %d for input %s. Got %d, want %d",
					i, j, test.input, naf[j], test.expected[j])
			}
		}

		// Checking NAF properties:
		// 1. All digits must be -1, 0, or 1
		// 2. No two non-zero digits should be adjacent
		for j := range naf {
			if naf[j] < -1 || naf[j] > 1 {
				t.Errorf("Test %d: Invalid NAF digit at position %d: %d", i, j, naf[j])
			}
			if j > 0 && naf[j] != 0 && naf[j-1] != 0 {
				t.Errorf("Test %d: Adjacent non-zero digits at positions %d and %d", i, j-1, j)
			}
		}

		// Verify that the NAF representation equals the original number
		reconstructed := new(big.Int)
		power := new(big.Int).SetInt64(1)
		for j := range naf {
			if naf[j] != 0 {
				term := new(big.Int).Mul(power, big.NewInt(int64(naf[j])))
				reconstructed.Add(reconstructed, term)
			}
			power.Mul(power, big.NewInt(2))
		}
		if reconstructed.Cmp(input) != 0 {
			t.Errorf("Test %d: NAF reconstruction failed for input %s. Got %s",
				i, test.input, reconstructed.String())
		}
	}
}

func TestSplitting(t *testing.T) {
	t.Parallel()

	var lambda, r, s, _s, zero big.Int
	var l Lattice

	r.SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
	lambda.SetString("4407920970296243842393367215006156084916469457145843978461", 10)

	PrecomputeLattice(&r, &lambda, &l)

	s.SetString("183927522224640574525727508854836440041603434369820418657580", 10)

	v := SplitScalar(&s, &l)
	_s.Mul(&v[1], &lambda).Add(&_s, &v[0]).Sub(&_s, &s)
	_s.Mod(&_s, &r)
	if _s.Cmp(&zero) != 0 {
		t.Fatal("Error split scalar")
	}
}

func BenchmarkSplitting256(b *testing.B) {
	var lambda, r, s big.Int
	var l Lattice

	r.SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
	lambda.SetString("4407920970296243842393367215006156084916469457145843978461", 10)
	PrecomputeLattice(&r, &lambda, &l)
	s.SetString("183927522224640574525727508854836440041603434369820418657580", 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SplitScalar(&s, &l)
	}
}
