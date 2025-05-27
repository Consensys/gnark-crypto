package vortex

import (
	"errors"
	"hash"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/field/koalabear/sis"
)

var (
	ErrWrongSizeHash = errors.New("the hash size should be 32 bytes")
)

// NewHash a functions returning a hash. Hash functions are stored this way, to allocate
// them when needed and parallelise the execution when possible.
type NewHash = func() hash.Hash

// Configuration options of the vortex prover
type Config struct {
	// hash function used to build the Merkle tree. By default, this hash is poseidon2.
	merkleHashFunc NewHash
	// hash function used to hash the stacked codewords. By default, this hash function is SIS.
	otherThanSis NewHash
}

// Option provides options for altering the default behavior of the vortex prover.
// See the descriptions of the functions returning instances of this
// type for available options.
type Option func(opt *Config) error

// WithMerkleHash specifies the hash function used to build the Merkle tree of the hashed
// columns of the stacked codewords.
func WithMerkleHash(h hash.Hash) Option {
	return func(opt *Config) error {
		bs := h.Size()
		if bs != 32 {
			return ErrWrongSizeHash
		}
		opt.merkleHashFunc = func() hash.Hash { return h }
		return nil
	}
}

// WithNoSis specifies the hash function used to hash the columns of the stacked codewords.
func WithNoSis(h hash.Hash) Option {
	return func(opt *Config) error {
		bs := h.Size()
		if bs != 32 {
			return ErrWrongSizeHash
		}
		opt.otherThanSis = func() hash.Hash { return h }
		return nil
	}
}

func defaultConfig() Config {
	return Config{merkleHashFunc: nil, otherThanSis: nil}
}

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

	// Coset table of the small domain, bit reversed
	CosetTableBitReverse koalabear.Vector

	// Conf is used to provide some customisation and to alter the default behavior
	// of the vortex prover.
	Conf Config
}

// NewParams constructs a new set of public parameters.
func NewParams(
	numColumns int,
	maxNumRow int,
	sisParams *sis.RSis,
	reedSolomonInvRate int,
	numSelectedColumns int,
	opts ...Option,
) (*Params, error) {
	if numColumns < 1 || !isPowerOfTwo(numColumns) {
		return nil, errors.New("number of columns must be a power of two")
	}

	if reedSolomonInvRate != 2 && reedSolomonInvRate != 4 && reedSolomonInvRate != 8 {
		// note: tested only with these.
		return nil, errors.New("reed solomon rate must be 2, 4 or 8")
	}

	conf := defaultConfig()
	if len(opts) != 0 {
		for _, opt := range opts {
			err := opt(&conf)
			if err != nil {
				return nil, err
			}
		}
	}

	shift, err := koalabear.Generator(uint64(numColumns * reedSolomonInvRate))
	if err != nil {
		return nil, err
	}

	smallDomain := fft.NewDomain(uint64(numColumns), fft.WithShift(shift))
	cosetTable, err := smallDomain.CosetTable()
	if err != nil {
		return nil, err
	}
	cosetTableBitReverse := make(koalabear.Vector, len(cosetTable))
	copy(cosetTableBitReverse, cosetTable)
	fft.BitReverse(cosetTableBitReverse)
	bigDomain := fft.NewDomain(uint64(numColumns * reedSolomonInvRate))

	return &Params{
		Key: sisParams,
		Domains: [2]*fft.Domain{
			smallDomain,
			bigDomain,
		},
		ReedSolomonInvRate:   reedSolomonInvRate,
		NbColumns:            numColumns,
		MaxNbRows:            maxNumRow,
		NumSelectedColumns:   numSelectedColumns,
		CosetTableBitReverse: cosetTableBitReverse,
	}, nil

}

// SizeCodeWord returns the number of columns of the matrix *after* the encoding
// has been performed.
func (p *Params) SizeCodeWord() int {
	return p.NbColumns * p.ReedSolomonInvRate
}
