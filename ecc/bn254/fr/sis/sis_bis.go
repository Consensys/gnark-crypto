package sis

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	"golang.org/x/crypto/blake2b"
)

var (
	ErrNotAPowerOfTwo = errors.New("d must be a power of 2")
)

// Ring-SIS instance
type RSis struct {

	// buffer storing the data to hash
	buffer bytes.Buffer

	// Vectors in ℤ_{p}/Xⁿ+1
	// A[i] is the i-th polynomial.
	// AFftBitreversed the evaluation form of the polynomials in A on the coset √(g) * <g>
	A                    [][]fr.Element
	AfftCosetBitreversed [][]fr.Element

	// LogTwoBound (Inifinty norm) of the vector to hash. It means that each component in m
	// is < 2^B, where m is the vector to hash (the hash being A*m).
	// cd https://hackmd.io/7OODKWQZRRW9RxM5BaXtIw , B >= 3.
	LogTwoBound int

	// maximal number of bytes to sum
	maxSizeByte int

	// domain for the polynomial multiplication
	Domain *fft.Domain

	// d, the degree of X^{d}+1
	d int
}

func genRandom(seed, i, j int64) fr.Element {

	var buf bytes.Buffer
	buf.WriteString("SIS")
	binary.Write(&buf, binary.BigEndian, seed)
	binary.Write(&buf, binary.BigEndian, i)
	binary.Write(&buf, binary.BigEndian, j)

	slice := buf.Bytes()
	digest := blake2b.Sum256(slice)

	var res fr.Element
	res.SetBytes(digest[:])

	return res
}

// NewRSis creates an instance of RSis.
// seed: seed for the randomness for generating A.
// logTwoDegree: if d := logTwoDegree, the ring will be ℤ_{p}[X]/Xᵈ-1, where X^{2ᵈ} is the 2ᵈ⁺¹-th cyclotomic polynomial
// b: the bound of the vector to hash (using the infinity norm).
// keySize: number of polynomials in A.
func NewRSis(seed int64, logTwoDegree, logTwoBound, keySize int) (RSis, error) {

	var res RSis

	// domains (shift is √{gen} )
	var shift fr.Element
	shift.SetString("19103219067921713944291392827692070036145651957329286315305642004821462161904") // -> 2²⁸-th root of unity of bn254
	e := int64(1 << (28 - (logTwoDegree + 1)))
	shift.Exp(shift, big.NewInt(e))
	res.Domain = fft.NewDomain(uint64(1<<logTwoDegree), shift)

	// bound
	res.LogTwoBound = logTwoBound

	// filling A
	degree := 1 << logTwoDegree
	res.A = make([][]fr.Element, keySize)
	for i := 0; i < keySize; i++ {
		res.A[i] = make([]fr.Element, degree)
		for j := 0; j < degree; j++ {
			res.A[i][j] = genRandom(seed, int64(i), int64(j))
		}
	}

	// filling AfftCosetBitreversed
	res.AfftCosetBitreversed = make([][]fr.Element, keySize)
	for i := 0; i < keySize; i++ {
		res.AfftCosetBitreversed[i] = make([]fr.Element, degree)
		for j := 0; j < degree; j++ {
			copy(res.AfftCosetBitreversed[i], res.A[i])
			res.Domain.FFT(res.AfftCosetBitreversed[i], fft.DIF, true)
		}
	}

	// computing the maximal size in bytes of a vector to hash
	res.maxSizeByte = res.LogTwoBound * degree * len(res.A) / 8

	// degree
	res.d = degree

	return res, nil
}

func printPoly(p []fr.Element) {
	for i := 0; i < len(p); i++ {
		fmt.Printf("%s*x**%d + ", p[i].String(), i)
	}
	fmt.Println("")
}

// mulMod computes p * q in ℤ_{p}[X]/Xᵈ+1.
// Is assumed that pLagrangeShifted and qLagrangeShifted are of the corret sizes
// and that they are in evaluation form on √(g) * <g>
func (r RSis) mulMod(pLagrangeCosetBitReversed, qLagrangeCosetBitReversed []fr.Element) []fr.Element {

	res := make([]fr.Element, len(pLagrangeCosetBitReversed))
	for i := 0; i < len(pLagrangeCosetBitReversed); i++ {
		res[i].Mul(&pLagrangeCosetBitReversed[i], &qLagrangeCosetBitReversed[i])
	}

	// FFTinv on the coset, it automagically reduces mod Xᵈ+1
	r.Domain.FFTInverse(res, fft.DIT, true)

	return res

}

func (r *RSis) Write(p []byte) (n int, err error) {
	return 0, nil
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
// b is interpreted as a sequence of coefficients of size r.Bound bits long.
// Each coefficient is interpreted in big endian.
// Ex: b = [0xa4, ...] and r.Bound = 4, means that b is decomposed as [10, 4, ...]
func (r *RSis) Sum(b []byte) []byte {

	// if maxSizeBytes is not reached, the buffer is padded with zeroes.
	sizeBuffer := r.buffer.Len()
	if sizeBuffer < r.maxSizeByte {
		toPadd := make([]byte, r.maxSizeByte-sizeBuffer)
		_, err := r.buffer.Write(toPadd)
		if err != nil {
			panic(err)
		}
	}

	// bitwise decomposition of the buffer, in order to build m (the vector to hash)
	// as a list of polynomials, whose coefficients are less than r.B bits long.
	mBits := make([]byte, r.maxSizeByte*8)
	var tmp [1]byte
	for i := 0; i < r.maxSizeByte; i++ {
		_, err := r.buffer.Read(tmp[:])
		if err != nil {
			panic(err)
		}
		for j := 0; j < 8; j++ {
			mBits[i*8+j] = (tmp[0] >> (7 - j)) & 1
		}
	}

	// now we an construct m
	nbBytesPerCoefficients := (r.LogTwoBound - (r.LogTwoBound % 8)) / 8
	nbBitsPerCoefficients := r.LogTwoBound
	offset := nbBitsPerCoefficients % 8
	sizeM := r.d * len(r.A)
	buf := make([]byte, nbBytesPerCoefficients)
	m := make([]fr.Element, sizeM)
	for i := 0; i < sizeM; i++ {
		for j := 0; j < offset; j++ {
			buf[0] += (mBits[i*nbBitsPerCoefficients+j]) << (offset - 1 - j)
		}
		for j := 0; j < nbBytesPerCoefficients; j++ {
			for k := 0; k < 8; k++ {
				buf[j+1] += (mBits[i*nbBitsPerCoefficients+offset+8*j+k]) << (7 - k)
			}
		}
		m[i].SetBytes(buf)
	}

	return nil
}

// Reset resets the Hash to its initial state.
func (r *RSis) Reset() {

	r.buffer.Reset()

	return
}

// Size returns the number of bytes Sum will return.
func (r *RSis) Size() int {

	// The size in bits is the size in bits of a polynomial in A.
	degree := len(r.A[0])
	totalSize := degree * fr.Modulus().BitLen() / 8

	return totalSize
}

// BlockSize returns the hash's underlying block size.
// The Write method must be able to accept any amount
// of data, but it may operate more efficiently if all writes
// are a multiple of the block size.
func (r *RSis) BlockSize() int {
	return 0
}
