// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package sis

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/internal/parallel"
	"golang.org/x/crypto/blake2b"
)

// RSis is the Ring-SIS instance
type RSis struct {
	// Vectors in ℤ_{p}/Xⁿ+1
	// A[i] is the i-th polynomial.
	// Ag the evaluation form of the polynomials in A on the coset √(g) * <g>
	A  [][]koalabear.Element
	Ag [][]koalabear.Element

	// LogTwoBound (Infinity norm) of the vector to hash. It means that each component in m
	// is < 2^B, where m is the vector to hash (the hash being A*m).
	// cf https://hackmd.io/7OODKWQZRRW9RxM5BaXtIw , B >= 3.
	LogTwoBound int

	// d, the degree of X^{d}+1
	Degree int

	// domain for the polynomial multiplication
	Domain *fft.Domain

	maxNbElementsToHash int

	smallFFT      func([]koalabear.Element)
	twiddlesCoset []koalabear.Element // used in conjunction with the smallFFT;
}

// NewRSis creates an instance of RSis.
// seed: seed for the randomness for generating A.
// logTwoDegree: if d := logTwoDegree, the ring will be ℤ_{p}[X]/Xᵈ-1, where X^{2ᵈ} is the 2ᵈ⁺¹-th cyclotomic polynomial
// logTwoBound: the bound of the vector to hash (using the infinity norm).
// maxNbElementsToHash: maximum number of field elements the instance handles
// used to derived n, the number of polynomials in A, and max size of instance's internal buffer.
func NewRSis(seed int64, logTwoDegree, logTwoBound, maxNbElementsToHash int) (*RSis, error) {

	if logTwoBound > 64 || logTwoBound > koalabear.Bits {
		return nil, errors.New("logTwoBound too large")
	}
	if logTwoBound%8 != 0 {
		return nil, errors.New("logTwoBound must be a multiple of 8")
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
	nbBytesPerLimb := logTwoBound / 8
	if koalabear.Bytes%nbBytesPerLimb != 0 {
		return nil, errors.New("nbBytesPerLimb must divide field size")
	}
	n := koalabear.Bytes / nbBytesPerLimb

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
	shift, err := koalabear.Generator(uint64(2 * degree))
	if err != nil {
		return nil, err
	}

	r := &RSis{
		LogTwoBound:         logTwoBound,
		Degree:              degree,
		Domain:              fft.NewDomain(uint64(degree), fft.WithShift(shift)),
		A:                   make([][]koalabear.Element, n),
		Ag:                  make([][]koalabear.Element, n),
		maxNbElementsToHash: maxNbElementsToHash,
	}

	r.smallFFT = func(p []koalabear.Element) {
		r.Domain.FFT(p, fft.DIF, fft.OnCoset(), fft.WithNbTasks(1))
	}

	// if we have a FFT kernel of the size of the domain cardinality, we use it.
	if r.Domain.Cardinality == 64 {
		r.twiddlesCoset = PrecomputeTwiddlesCoset(r.Domain.Generator, shift)
		r.smallFFT = func(a []koalabear.Element) {
			FFT64(a, r.twiddlesCoset)
		}
	}

	// filling A
	a := make([]koalabear.Element, n*r.Degree)
	ag := make([]koalabear.Element, n*r.Degree)

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
			r.Domain.FFT(r.Ag[i], fft.DIF, fft.OnCoset(), fft.WithNbTasks(1))
		}
	})

	return r, nil
}

// Hash interprets the input vector as a sequence of coefficients of size r.LogTwoBound bits long,
// and return the hash of the polynomial corresponding to the sum sum_i A[i]*m Mod X^{d}+1
func (r *RSis) Hash(v, res []koalabear.Element) error {
	if len(res) != r.Degree {
		return fmt.Errorf("output vector must have length %d", r.Degree)
	}

	if len(v) > r.maxNbElementsToHash {
		return fmt.Errorf("can't hash more than %d elements with params provided in constructor", r.maxNbElementsToHash)
	}

	// zeroing res
	for i := 0; i < len(res); i++ {
		res[i].SetZero()
	}

	k := make([]koalabear.Element, r.Degree)

	// inner hash
	it := NewLimbIterator(&VectorIterator{v: v}, r.LogTwoBound/8)
	for i := 0; i < len(r.Ag); i++ {
		r.InnerHash(it, res, k, i)
	}

	// reduces mod Xᵈ+1
	r.Domain.FFTInverse(res, fft.DIT, fft.OnCoset(), fft.WithNbTasks(1))

	return nil
}

func (r *RSis) InnerHash(it *LimbIterator, res, k koalabear.Vector, polId int) {
	zero := uint32(0)
	for j := 0; j < r.Degree; j++ {
		l, ok := it.NextLimb()
		if !ok {
			// we need to pad; note that we should use a deterministic padding
			// other than 0, but it is not an issue for the current use cases.
			for m := j; m < r.Degree; m++ {
				k[m].SetZero()
			}
			break
		}
		zero |= l
		k[j].SetZero()
		k[j][0] = l
	}
	if zero == 0 {
		// means m[i*r.Degree : (i+1)*r.Degree] == [0...0]
		// we can skip this, FFT(0) = 0
		return
	}

	// r.Domain.FFT(k, fft.DIF, fft.OnCoset(), fft.WithNbTasks(1))
	// for perf, we use directly what's exposed;
	r.smallFFT(k)
	// k.Mul(k, fr.Vector(r.cosetTable))
	// if r.Domain.KernelDIF != nil {
	// 	r.Domain.KernelDIF(k)
	// } else {
	// 	r.Domain.FFT(k, fft.DIF, fft.WithNbTasks(1))
	// }

	mulModAcc(res, r.Ag[polId], k)
}

// mulModAcc computes p * q in ℤ_{p}[X]/Xᵈ+1.
// Is assumed that pLagrangeShifted and qLagrangeShifted are of the correct sizes
// and that they are in evaluation form on √(g) * <g>
// The result is not FFTinversed. The fft inverse is done once every
// multiplications are done.
// then accumulates the mulMod result in res.
// qLagrangeCosetBitReversed and res are mutated.
// pLagrangeCosetBitReversed is not mutated.
func mulModAcc(res, pLagrangeCosetBitReversed, qLagrangeCosetBitReversed koalabear.Vector) {
	qLagrangeCosetBitReversed.Mul(qLagrangeCosetBitReversed, pLagrangeCosetBitReversed)
	res.Add(res, qLagrangeCosetBitReversed)
}

func deriveRandomElementFromSeed(seed, i, j int64) koalabear.Element {
	var buf [3 + 3*8]byte
	copy(buf[:3], "SIS")
	binary.BigEndian.PutUint64(buf[3:], uint64(seed))
	binary.BigEndian.PutUint64(buf[11:], uint64(i))
	binary.BigEndian.PutUint64(buf[19:], uint64(j))

	digest := blake2b.Sum256(buf[:])

	var res koalabear.Element
	res.SetBytes(digest[:])

	return res
}

// TODO @gbotrel explore generic perf impact + go 1.23 iterators

// ElementIterator is an iterator over a stream of field elements.
type ElementIterator interface {
	Next() (koalabear.Element, bool)
}

type VectorIterator struct {
	v koalabear.Vector
	i int
}

func NewVectorIterator(v koalabear.Vector) *VectorIterator {
	return &VectorIterator{v: v}
}

func (vi *VectorIterator) Next() (koalabear.Element, bool) {
	if vi.i == len(vi.v) {
		return koalabear.Element{}, false
	}
	vi.i++
	return vi.v[vi.i-1], true
}

// LimbIterator iterates over a vector of field element, limb by limb.
type LimbIterator struct {
	it  ElementIterator
	buf [koalabear.Bytes]byte

	j int // position in buf

	next func(buf []byte, pos *int) uint32
}

// NewLimbIterator creates a new LimbIterator
// v: the vector to read
// limbSize: the size of the limb in bytes (1, 2, 4 or 8)
// The elements are interpreted in little endian.
// The limb is also in little endian.
func NewLimbIterator(it ElementIterator, limbSize int) *LimbIterator {
	var next func(buf []byte, pos *int) uint32
	switch limbSize {
	case 1:
		next = nextUint8
	case 2:
		next = nextUint16

	default:
		panic("unsupported limb size")
	}
	return &LimbIterator{
		it:   it,
		j:    koalabear.Bytes,
		next: next,
	}
}

// NextLimb returns the next limb of the vector.
// This does not perform any bound check, may trigger an out of bound panic.
// If underlying vector is "out of limb"
func (vr *LimbIterator) NextLimb() (uint32, bool) {
	if vr.j == koalabear.Bytes {
		next, ok := vr.it.Next()
		if !ok {
			return 0, false
		}
		vr.j = 0
		koalabear.LittleEndian.PutElement(&vr.buf, next)
	}
	return vr.next(vr.buf[:], &vr.j), true
}

func (vr *LimbIterator) Reset(it ElementIterator) {
	vr.it = it
	vr.j = koalabear.Bytes
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
