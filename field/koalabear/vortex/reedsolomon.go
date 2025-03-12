package vortex

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
)

// EncodeReedSolomon encodes a vector of field elements into a reed-solomon codewords.
// The function checks that:
//   - the input argument has the right size
func (p *Params) EncodeReedSolomon(input []koalabear.Element) ([]koalabear.Element, error) {
	if len(input) != p.NbColumns {
		return nil, fmt.Errorf("expected %d input values, got %d", p.NbColumns, len(input))
	}

	codeword := make([]koalabear.Element, p.NbEncodedColumns())
	copy(codeword, input)

	p.Domains[0].FFTInverse(codeword[:p.NbColumns], fft.DIF)
	fft.BitReverse(codeword[:p.NbColumns])
	p.Domains[1].FFT(codeword, fft.DIF)
	fft.BitReverse(codeword)

	return codeword, nil
}

// IsCodeword returns nil iff the argument `v` is a correct codeword and an
// error is returned otherwise.
func (p *Params) IsReedSolomonCodewords(codeword []fext.E4) error {

	// As we don't have a dedicated FFT for field extensions, we apply
	// the FFT algorithm coordinates-by-coordinates. This might be
	// improvable by a direct AVX implementation but this only matters
	// for the verifier and not for the prover.

	coeffs := make([]koalabear.Element, p.NbEncodedColumns())

	for i := range coeffs {
		coeffs[i] = codeword[i].B0.A0
	}

	p.Domains[1].FFTInverse(coeffs, fft.DIF)
	fft.BitReverse(coeffs)
	for i := p.NbColumns; i < p.NbEncodedColumns(); i++ {
		if !coeffs[i].IsZero() {
			return fmt.Errorf("not a reed-solomon codeword")
		}
	}

	for i := range coeffs {
		coeffs[i] = codeword[i].B0.A1
	}

	p.Domains[1].FFTInverse(coeffs, fft.DIF)
	fft.BitReverse(coeffs)
	for i := p.NbColumns; i < p.NbEncodedColumns(); i++ {
		if !coeffs[i].IsZero() {
			return fmt.Errorf("not a reed-solomon codeword")
		}
	}

	for i := range coeffs {
		coeffs[i] = codeword[i].B1.A0
	}

	p.Domains[1].FFTInverse(coeffs, fft.DIF)
	fft.BitReverse(coeffs)
	for i := p.NbColumns; i < p.NbEncodedColumns(); i++ {
		if !coeffs[i].IsZero() {
			return fmt.Errorf("not a reed-solomon codeword")
		}
	}

	for i := range coeffs {
		coeffs[i] = codeword[i].B1.A1
	}

	p.Domains[1].FFTInverse(coeffs, fft.DIF)
	fft.BitReverse(coeffs)
	for i := p.NbColumns; i < p.NbEncodedColumns(); i++ {
		if !coeffs[i].IsZero() {
			return fmt.Errorf("not a reed-solomon codeword")
		}
	}

	return nil
}
