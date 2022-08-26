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

	// Vectors in ℤ_{p}/Xⁿ+1
	// A[i] is the i-th polynomial.
	// AFftBitreversed the evaluation form of the polynomials in A on the coset √(g) * <g>
	A                    [][]fr.Element
	AfftCosetBitreversed [][]fr.Element

	// Bound (Inifinty norm) of the vector to hash, in binary.
	B int

	// domain for the polynomial multiplication
	Domain *fft.Domain
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
// d: the ring will be ℤ_{p}[X]/Xᵈ-1, where X^{2ᵈ} is the 2ᵈ⁺¹-th cyclotomic polynomial
// b: the bound of the vector to hash (using the infinity norm).
// keySize: number of polynomials in A.
func NewRSis(seed int64, degreeBin, bound, keySize int) (RSis, error) {

	var res RSis

	// domains (shift is √{gen} )
	var shift fr.Element
	shift.SetString("19103219067921713944291392827692070036145651957329286315305642004821462161904") // -> 2²⁸-th root of unity of bn254
	e := int64(1 << (28 - (degreeBin + 1)))
	shift.Exp(shift, big.NewInt(e))
	res.Domain = fft.NewDomain(uint64(1<<degreeBin), shift)

	// bound
	res.B = bound

	// filling A
	degree := 1 << degreeBin
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
func (r RSis) mulMod(pLagrangeShifted, qLagrangeShifted []fr.Element) []fr.Element {

	res := make([]fr.Element, len(pLagrangeShifted))
	for i := 0; i < len(pLagrangeShifted); i++ {
		res[i].Mul(&pLagrangeShifted[i], &qLagrangeShifted[i])
	}

	// FFTinv on the coset, it automagically reduces mod Xᵈ+1
	r.Domain.FFTInverse(res, fft.DIT, true)

	return res

}
