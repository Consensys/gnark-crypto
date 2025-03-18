package vortex

import (
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/field/koalabear/sis"
)

// Params collects the public parameters of the commitment scheme. The object
// should not be constructed directly (use [NewParamsSis] or [NewParamsNoSis])
// instead nor be modified after having been constructed.
type Params struct {
	// RSis stores the public parameters of the ring-SIS instance in use to
	// hash the columns.
	Key *sis.RSis
	// ReedSolomonInvRate corresponds to the inverse-rate of the Reed-Solomon code
	// in use to encode the rows of the committed matrices. This is a power of
	// two and can't be one.
	ReedSolomonInvRate int
	// Domain[0]: domain to perform the FFT^-1, of size NbColumns is meant to
	// be run over the non-encoded rows when RS encoding.
	// Domain[1]: domain to perform FFT, of size BlowUp * NbColumns is meant
	// to be obtain the codeword when RS encoding.
	Domains [2]*fft.Domain
	// NbColumns number of columns of the matrix storing the polynomials. The
	// total size of the polynomials which are committed is NbColumns x NbRows.
	// The Number of columns is a power of 2, it corresponds to the original
	// size of the codewords of the Reed Solomon code.
	NbColumns int
	// MaxNbRows number of rows of the matrix storing the polynomials. If a
	// polynomial p is appended whose size if not 0 mod MaxNbRows, it is padded
	// as p' so that len(p')=0 mod MaxNbRows.
	MaxNbRows int
	// NumSelectedColumns indicates the number of columns to open in the
	// column opening phase.
	NumSelectedColumns int
}

// NewParams constructs a new set of public parameters.
func NewParams(
	numColumns int,
	maxNumRow int,
	sisParams *sis.RSis,
	reedSolomonInvRate int,
	numSelectedColumns int,
) *Params {

	return &Params{
		Key: sisParams,
		Domains: [2]*fft.Domain{
			fft.NewDomain(uint64(numColumns)),
			fft.NewDomain(uint64(numColumns * reedSolomonInvRate)),
		},
		ReedSolomonInvRate: reedSolomonInvRate,
		NbColumns:          numColumns,
		MaxNbRows:          maxNumRow,
		NumSelectedColumns: numSelectedColumns,
	}

}

// NbEncodedColumns returns the number of columns of the matrix *after* the encoding
// has been performed.
func (p *Params) NbEncodedColumns() int {
	return p.NbColumns * p.ReedSolomonInvRate
}
