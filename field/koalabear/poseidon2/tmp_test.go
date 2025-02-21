package poseidon2

import (
	"fmt"
	"testing"

	fr "github.com/consensys/gnark-crypto/field/koalabear"
)

//go:noescape
func validation(input []fr.Element)

func TestValidation(t *testing.T) {
	input := make([]fr.Element, 24)
	sum := fr.Element{}
	for i := range input {
		input[i].SetUint64(uint64(i + 10))
		sum.Add(&sum, &input[i])
	}
	sbox0 := input[0]
	sbox0.Square(&sbox0)
	sbox0.Mul(&sbox0, &input[0])

	fmt.Println("before")
	for i := range input {
		fmt.Printf("%s, ", input[i].String())
	}
	fmt.Println()

	validation(input)

	for i := range input {
		fmt.Printf("%s, ", input[i].String())
	}

	if !sbox0.Equal(&input[0]) {
		t.Fatalf("mismatch error sbox0, expected %s, got %s", sbox0.String(), input[0].String())
	}

	// expected := []uint32{2, 3, 0, 1, 6, 7, 4, 5}
	// for i := range input {
	// 	if input[i] != expected[i] {
	// 		t.Fatal("mismatch error")
	// 	}
	// }
}

func TestAVX512(t *testing.T) {
	// generate 1 random vector of 24 elements
	for j := 0; j < 10; j++ {
		var input, expected [24]fr.Element
		for i := 0; i < 24; i++ {
			input[i].SetRandom()
		}

		expected = input

		h := NewPermutation(24, 6, 21)
		h.params.hasFast24_6_21 = false
		h.Permutation(input[:])
		h.params.hasFast24_6_21 = true
		h.Permutation(expected[:])

		for i := 0; i < 24; i++ {
			if !input[i].Equal(&expected[i]) {
				t.Fatal("mismatch error avx512")
			}
		}
	}
}
