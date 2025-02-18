package vortex

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
)

// EncodeReedSolomon encodes a vector of field elements into a reed-solomon codewords.
// The function checks that:
//   - the input argument has the right size
func (p *Params) EncodeReedSolomon(input []koalabear.Element) ([]koalabear.Element, error) {

	codeword := make([]koalabear.Element, p.NbEncodedColumns())

	if len(input) != p.NbColumns {
		return nil, fmt.Errorf("expected %d input values, got %d", p.NbColumns, len(input))
	}

	if len(codeword) != p.NbEncodedColumns() {
		return nil, fmt.Errorf("expected %d codeword values, got %v", p.NbColumns*p.ReedSolomonInvRate, len(codeword))
	}

	copy(codeword, input)

	for i := p.NbColumns; i < p.NbEncodedColumns(); i++ {
		codeword[i].SetZero()
	}

	p.Domains[0].FFTInverse(codeword[:p.NbColumns], fft.DIT)
	p.Domains[1].FFT(codeword, fft.DIT)

	return codeword, nil
}

// IsCodeword returns nil iff the argument `v` is a correct codeword and an
// error is returned otherwise.
func (p *Params) IsReedSolomonCodewords(codeword []koalabear.Element) error {
	coeffs := append([]koalabear.Element(nil), codeword...)
	p.Domains[1].FFTInverse(coeffs, fft.DIT)
	for i := p.NbColumns; i < p.NbEncodedColumns(); i++ {
		if !coeffs[i].IsZero() {
			return fmt.Errorf("not a reed-solomon codeword")
		}
	}
	return nil
}
