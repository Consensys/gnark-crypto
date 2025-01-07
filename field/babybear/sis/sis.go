// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package sis

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"

	"github.com/consensys/gnark-crypto/field/babybear"
	"github.com/consensys/gnark-crypto/field/babybear/fft"
	"github.com/consensys/gnark-crypto/internal/parallel"
	"golang.org/x/crypto/blake2b"
)

// RSis is the Ring-SIS instance
type RSis struct {
	// Vectors in ℤ_{p}/Xⁿ+1
	// A[i] is the i-th polynomial.
	// Ag the evaluation form of the polynomials in A on the coset √(g) * <g>
	A  [][]babybear.Element
	Ag [][]babybear.Element

	// LogTwoBound (Infinity norm) of the vector to hash. It means that each component in m
	// is < 2^B, where m is the vector to hash (the hash being A*m).
	// cf https://hackmd.io/7OODKWQZRRW9RxM5BaXtIw , B >= 3.
	LogTwoBound int

	// d, the degree of X^{d}+1
	Degree int

	// domain for the polynomial multiplication
	Domain *fft.Domain

	maxNbElementsToHash int
}

// NewRSis creates an instance of RSis.
// seed: seed for the randomness for generating A.
// logTwoDegree: if d := logTwoDegree, the ring will be ℤ_{p}[X]/Xᵈ-1, where X^{2ᵈ} is the 2ᵈ⁺¹-th cyclotomic polynomial
// logTwoBound: the bound of the vector to hash (using the infinity norm).
// maxNbElementsToHash: maximum number of field elements the instance handles
// used to derived n, the number of polynomials in A, and max size of instance's internal buffer.
func NewRSis(seed int64, logTwoDegree, logTwoBound, maxNbElementsToHash int) (*RSis, error) {

	if logTwoBound > 64 || logTwoBound > babybear.Bits {
		return nil, errors.New("logTwoBound too large")
	}
	if logTwoBound%8 != 0 {
		panic("logTwoBound must be a multiple of 8")
	}
	if bits.UintSize == 32 {
		return nil, errors.New("unsupported architecture; need 64bit target")
	}

	degree := 1 << logTwoDegree

	// n: number of polynomials in A
	// len(m) == degree * n
	// with each element in m being logTwoBounds bits from the instance buffer.
	// that is, to fill m, we need [degree * n * logTwoBound] bits of data

	// First n <- #limbs to represent a single field element
	n := (babybear.Bytes * 8) / logTwoBound
	if n*logTwoBound < babybear.Bytes*8 {
		n++
	}

	// Then multiply by the number of field elements
	n *= maxNbElementsToHash

	// And divide (+ ceil) to get the number of polynomials
	if n%degree == 0 {
		n /= degree
	} else {
		n /= degree // number of polynomials
		n++
	}

	// domains (shift is √{gen} )
	shift, err := babybear.Generator(uint64(2 * degree))
	if err != nil {
		return nil, err
	}

	r := &RSis{
		LogTwoBound:         logTwoBound,
		Degree:              degree,
		Domain:              fft.NewDomain(uint64(degree), fft.WithShift(shift)),
		A:                   make([][]babybear.Element, n),
		Ag:                  make([][]babybear.Element, n),
		maxNbElementsToHash: maxNbElementsToHash,
	}

	// filling A
	a := make([]babybear.Element, n*r.Degree)
	ag := make([]babybear.Element, n*r.Degree)

	parallel.Execute(n, func(start, end int) {
		for i := start; i < end; i++ {
			rstart, rend := i*r.Degree, (i+1)*r.Degree
			r.A[i] = a[rstart:rend:rend]
			r.Ag[i] = ag[rstart:rend:rend]
			for j := 0; j < r.Degree; j++ {
				r.A[i][j] = deriveRandomElementFromSeed(seed, int64(i), int64(j))
			}

			// fill Ag the evaluation form of the polynomials in A on the coset √(g) * <g>
			copy(r.Ag[i], r.A[i])
			r.Domain.FFT(r.Ag[i], fft.DIF, fft.OnCoset())
		}
	})

	return r, nil
}

// Hash interprets the input vector as a sequence of coefficients of size r.LogTwoBound bits long,
// and return the hash of the polynomial corresponding to the sum sum_i A[i]*m Mod X^{d}+1
func (r *RSis) Hash(v, res []babybear.Element) error {
	if len(res) != r.Degree {
		return fmt.Errorf("output vector must have length %d", r.Degree)
	}

	for i := 0; i < len(res); i++ {
		// TODO @gbotrel ensure that this is needed.
		res[i].SetZero()
	}
	if len(v) > r.maxNbElementsToHash {
		return fmt.Errorf("can't hash more than %d elements with params provided in constructor", r.maxNbElementsToHash)
	}

	reader := NewVectorLimbReader(v, r.LogTwoBound/8)

	kz := make([]babybear.Element, r.Degree)
	k := make([]babybear.Element, r.Degree)
	for i := 0; i < len(r.Ag); i++ {
		copy(k, kz)

		zero := uint32(0)
		for j := 0; j < r.Degree; j++ {
			l := reader.NextLimb()
			zero |= l
			k[j][0] = l
		}
		if zero == 0 {
			// means m[i*r.Degree : (i+1)*r.Degree] == [0...0]
			// we can skip this, FFT(0) = 0
			continue
		}

		r.Domain.FFT(k, fft.DIF, fft.OnCoset(), fft.WithNbTasks(1))

		mulModAcc(res, r.Ag[i], k)
	}
	r.Domain.FFTInverse(res, fft.DIT, fft.OnCoset(), fft.WithNbTasks(1)) // -> reduces mod Xᵈ+1

	return nil
}

// mulModAcc computes p * q in ℤ_{p}[X]/Xᵈ+1.
// Is assumed that pLagrangeShifted and qLagrangeShifted are of the correct sizes
// and that they are in evaluation form on √(g) * <g>
// The result is not FFTinversed. The fft inverse is done once every
// multiplications are done.
// then accumulates the mulMod result in res.
func mulModAcc(res []babybear.Element, pLagrangeCosetBitReversed, qLagrangeCosetBitReversed []babybear.Element) {
	var t babybear.Element
	for i := 0; i < len(pLagrangeCosetBitReversed); i++ {
		t.Mul(&pLagrangeCosetBitReversed[i], &qLagrangeCosetBitReversed[i])
		res[i].Add(&res[i], &t)
	}
}

func deriveRandomElementFromSeed(seed, i, j int64) babybear.Element {
	var buf [3 + 3*8]byte
	copy(buf[:3], "SIS")
	binary.BigEndian.PutUint64(buf[3:], uint64(seed))
	binary.BigEndian.PutUint64(buf[11:], uint64(i))
	binary.BigEndian.PutUint64(buf[19:], uint64(j))

	digest := blake2b.Sum256(buf[:])

	var res babybear.Element
	res.SetBytes(digest[:])

	return res
}

// VectorLimbReader reads a vector of field element, limb by limb.
// The elements are interpreted in little endian.
// The limb is also interpreted in little endian.
type VectorLimbReader struct {
	v   babybear.Vector
	buf [babybear.Bytes]byte

	i int // position in vector
	j int // position in buf

	next func(buf []byte, pos *int) uint32
}

// NewVectorLimbReader creates a new VectorLimbReader
// v: the vector to read
// limbSize: the size of the limb in bytes (1, 2, 4 or 8)
func NewVectorLimbReader(v babybear.Vector, limbSize int) *VectorLimbReader {
	var next func(buf []byte, pos *int) uint32
	switch limbSize {
	case 1:
		next = nextUint8
	case 2:
		next = nextUint16

	default:
		panic("unsupported limb size")
	}
	return &VectorLimbReader{
		v:    v,
		j:    babybear.Bytes,
		next: next,
	}
}

// NextLimb returns the next limb of the vector.
// This does not perform any bound check, may trigger an out of bound panic.
// If underlying vector is "out of limb"
func (vr *VectorLimbReader) NextLimb() uint32 {
	if vr.j == babybear.Bytes {
		vr.j = 0
		// TODO @gbotrel we could return 0 in the case vr.i == len(vr.v)
		babybear.LittleEndian.PutElement(&vr.buf, vr.v[vr.i])
		vr.i++
	}
	return vr.next(vr.buf[:], &vr.j)
}

func nextUint8(buf []byte, pos *int) uint32 {
	r := uint32(buf[*pos])
	*pos++
	return r
}

func nextUint16(buf []byte, pos *int) uint32 {
	r := uint32(binary.LittleEndian.Uint16(buf[*pos:]))
	*pos += 2
	return r
}
