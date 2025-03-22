package vortex

import (
	"errors"
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
)

// func TestLocal(t *testing.T) {

// 	params := NewParams(4, 4, nil, 2, 2)

// 	size := 4
// 	toEncode := make([]koalabear.Element, size)
// 	for i := 0; i < size; i++ {
// 		toEncode[i].SetUint64(uint64(i + 2))
// 	}

// 	encodedRS, _ := params.EncodeReedSolomon(toEncode, false)
// 	for i := 0; i < len(encodedRS); i++ {
// 		fmt.Println(encodedRS[i].String())
// 	}
// 	fmt.Println("--")

// 	rho := 2
// 	ds := fft.NewDomain(uint64(size))
// 	db := fft.NewDomain(uint64(size * rho))
// 	encodedClassic := make([]koalabear.Element, rho*size)
// 	copy(encodedClassic, toEncode)
// 	ds.FFTInverse(encodedClassic[:size], fft.DIF)
// 	fft.BitReverse(encodedClassic[:size])
// 	db.FFT(encodedClassic, fft.DIF)
// 	fft.BitReverse(encodedClassic)
// 	for i := 0; i < len(encodedRS); i++ {
// 		fmt.Println(encodedClassic[i].String())
// 	}

// }

func TestLagrangeSimple(t *testing.T) {

	params := NewParams(4, 4, nil, 2, 2)

	t.Run("0-1-2-3", func(t *testing.T) {

		v := []koalabear.Element{
			koalabear.NewElement(0),
			koalabear.NewElement(1),
			koalabear.NewElement(2),
			koalabear.NewElement(3),
		}

		codeword := make([]koalabear.Element, params.SizeCodeWord())
		params.EncodeReedSolomon(v, codeword)

		for i := 0; i < len(codeword); i += 2 {
			if codeword[i] != v[i/2] {
				t.Errorf("failure at position (%v %v)", i, i/2)
			}
		}
	})

	t.Run("shifting", func(t *testing.T) {

		v := []koalabear.Element{
			koalabear.NewElement(0),
			koalabear.NewElement(1),
			koalabear.NewElement(2),
			koalabear.NewElement(3),
		}

		vShifted := []koalabear.Element{
			koalabear.NewElement(1),
			koalabear.NewElement(2),
			koalabear.NewElement(3),
			koalabear.NewElement(0),
		}

		codeword := make([]koalabear.Element, params.SizeCodeWord())
		params.EncodeReedSolomon(v, codeword)

		codewordShifted := make([]koalabear.Element, params.SizeCodeWord())
		params.EncodeReedSolomon(vShifted, codewordShifted)

		for i := 0; i < len(codeword); i++ {

			iShifted := i - 2
			if iShifted < 0 {
				iShifted += 8
			}

			if codeword[i] != codewordShifted[iShifted] {
				t.Errorf("mismatch between codeword and shifted codeword")
			}
		}

	})
}

func TestReedSolomonProperty(t *testing.T) {

	var (
		size         = 16
		invRate      = 2
		v            = make([]koalabear.Element, size)
		encodedVFext = make([]fext.E4, size*invRate)
		params       = NewParams(size, 4, nil, 2, 2)

		// #nosec G404 -- test case generation does not require a cryptographic PRNG
		rng   = rand.New(rand.NewChaCha8([32]byte{}))
		randX = randFext(rng)
	)

	for i := range v {
		v[i] = randElement(rng)
	}

	encodedV := make([]koalabear.Element, params.SizeCodeWord())
	params.EncodeReedSolomon(v, encodedV)

	for i := range encodedVFext {
		encodedVFext[i].B0.A0.Set(&encodedV[i])
	}

	if err := params.IsReedSolomonCodewords(encodedVFext); err != nil {
		t.Fatalf("codeword does not pass rs check")
	}

	var (
		y0, err0 = EvalBasePolyLagrange(v, randX)
		y1, err1 = EvalBasePolyLagrange(encodedV, randX)
		y2, err2 = EvalFextPolyLagrange(encodedVFext, randX)
	)

	if err := errors.Join(err0, err1, err2); err != nil {
		t.Fatal(err)
	}

	if y0 != y1 || y1 != y2 {
		t.Fatalf("rs inconsistent with lagrange basis evaluation, %v %v %v", y0.String(), y1.String(), y2.String())
	}

}
