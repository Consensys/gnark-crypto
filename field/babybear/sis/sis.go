// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package sis

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"

	"github.com/bits-and-blooms/bitset"
	"github.com/consensys/gnark-crypto/field/babybear"
	"github.com/consensys/gnark-crypto/field/babybear/fft"
	"github.com/consensys/gnark-crypto/internal/parallel"
	"golang.org/x/crypto/blake2b"
)

// Ring-SIS instance
type RSis struct {

	// buffer storing the data to hash
	buffer bytes.Buffer

	// Vectors in ℤ_{p}/Xⁿ+1
	// A[i] is the i-th polynomial.
	// Ag the evaluation form of the polynomials in A on the coset √(g) * <g>
	A  [][]babybear.Element
	Ag [][]babybear.Element

	// LogTwoBound (Infinity norm) of the vector to hash. It means that each component in m
	// is < 2^B, where m is the vector to hash (the hash being A*m).
	// cf https://hackmd.io/7OODKWQZRRW9RxM5BaXtIw , B >= 3.
	LogTwoBound int

	// domain for the polynomial multiplication
	Domain        *fft.Domain
	twiddleCosets []babybear.Element // see FFT64 and precomputeTwiddlesCoset

	// d, the degree of X^{d}+1
	Degree int

	// in bytes, represents the maximum number of bytes the .Write(...) will handle;
	// ( maximum number of bytes to sum )
	capacity            int
	maxNbElementsToHash int

	// allocate memory once per instance (used in Sum())
	bufM       babybear.Vector
	bufMValues *bitset.BitSet
}

// NewRSis creates an instance of RSis.
// seed: seed for the randomness for generating A.
// logTwoDegree: if d := logTwoDegree, the ring will be ℤ_{p}[X]/Xᵈ-1, where X^{2ᵈ} is the 2ᵈ⁺¹-th cyclotomic polynomial
// logTwoBound: the bound of the vector to hash (using the infinity norm).
// maxNbElementsToHash: maximum number of field elements the instance handles
// used to derived n, the number of polynomials in A, and max size of instance's internal buffer.
func NewRSis(seed int64, logTwoDegree, logTwoBound, maxNbElementsToHash int) (*RSis, error) {

	if logTwoBound > 64 {
		return nil, errors.New("logTwoBound too large")
	}
	if bits.UintSize == 32 {
		return nil, errors.New("unsupported architecture; need 64bit target")
	}

	degree := 1 << logTwoDegree
	capacity := maxNbElementsToHash * babybear.Bytes

	// n: number of polynomials in A
	// len(m) == degree * n
	// with each element in m being logTwoBounds bits from the instance buffer.
	// that is, to fill m, we need [degree * n * logTwoBound] bits of data
	// capacity == [degree * n * logTwoBound] / 8
	// n == (capacity*8)/(degree*logTwoBound)

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
		capacity:            capacity,
		Degree:              degree,
		Domain:              fft.NewDomain(uint64(degree), fft.WithShift(shift)),
		A:                   make([][]babybear.Element, n),
		Ag:                  make([][]babybear.Element, n),
		bufM:                make(babybear.Vector, degree*n),
		bufMValues:          bitset.New(uint(n)),
		maxNbElementsToHash: maxNbElementsToHash,
	}
	if r.LogTwoBound == 8 && r.Degree == 64 {
		// TODO @gbotrel fixme, that's dirty.
		r.twiddleCosets = PrecomputeTwiddlesCoset(r.Domain.Generator, r.Domain.FrMultiplicativeGen)
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
//
// It is equivalent to calling r.Write(element.Marshal()); outBytes = r.Sum(nil);
// ! note @gbotrel: this is a place holder, may not make sense
func (r *RSis) Hash(v, res []babybear.Element) error {
	if len(res) != r.Degree {
		return fmt.Errorf("output vector must have length %d", r.Degree)
	}
	// TODO @gbotrel check that this is needed.
	for i := 0; i < len(res); i++ {
		res[i].SetZero()
	}
	if len(v) > r.maxNbElementsToHash {
		return fmt.Errorf("can't hash more than %d elements with params provided in constructor", r.maxNbElementsToHash)
	}

	// reset the buffer
	r.buffer.Reset()

	// write the elements to the buffer
	// TODO @gbotrel for now we use a buffer, we will kill it later in the refactoring.
	for _, e := range v {
		r.buffer.Write(e.Marshal())
	}

	{
		// previous Sum()

		buf := r.buffer.Bytes()
		if len(buf) > r.capacity {
			panic("buffer too large")
		}

		fastPath := r.LogTwoBound == 8 && r.Degree == 64

		// clear the buffers of the instance.
		defer r.cleanupBuffers()

		m := r.bufM
		mValues := r.bufMValues

		if r.LogTwoBound < 8 && (8%r.LogTwoBound == 0) {
			limbDecomposeBytesSmallBound(buf, m, r.LogTwoBound, r.Degree, mValues)
		} else if r.LogTwoBound >= 8 && (babybear.Bytes*8)%r.LogTwoBound == 0 {
			limbDecomposeBytesMiddleBound(buf, m, r.LogTwoBound, r.Degree, mValues)
		} else {
			limbDecomposeBytes(buf, m, r.LogTwoBound, r.Degree, mValues)
		}

		// method 1: fft
		for i := 0; i < len(r.Ag); i++ {
			if !mValues.Test(uint(i)) {
				// means m[i*r.Degree : (i+1)*r.Degree] == [0...0]
				// we can skip this, FFT(0) = 0
				continue
			}
			k := m[i*r.Degree : (i+1)*r.Degree]
			if fastPath {
				// fast path.
				FFT64(k, r.twiddleCosets)
			} else {
				r.Domain.FFT(k, fft.DIF, fft.OnCoset(), fft.WithNbTasks(1))
			}
			mulModAcc(res, r.Ag[i], k)
		}
		r.Domain.FFTInverse(res, fft.DIT, fft.OnCoset(), fft.WithNbTasks(1)) // -> reduces mod Xᵈ+1

		return nil
	}
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

// Returns a clone of the RSis parameters with a fresh and empty buffer. Does not
// mutate the current instance. The keys and the public parameters of the SIS
// instance are not deep-copied. It is useful when we want to hash in parallel.
// Otherwise, we would have to generate an entire RSis for each thread.
func (r *RSis) CopyWithFreshBuffer() RSis {
	res := *r
	res.buffer = bytes.Buffer{}
	res.bufM = make(babybear.Vector, len(r.bufM))
	res.bufMValues = bitset.New(r.bufMValues.Len())
	return res
}

// Cleanup the buffers of the RSis instance
func (r *RSis) cleanupBuffers() {
	r.bufMValues.ClearAll()
	for i := 0; i < len(r.bufM); i++ {
		r.bufM[i].SetZero()
	}
}

// Split an slice of bytes representing an array of serialized field element in
// big-endian form into an array of limbs representing the same field elements
// in little-endian form. Namely, if our field is represented with 64 bits and we
// have the following field element 0x0123456789abcdef (0 being the most significant
// character and and f being the least significant one) and our log norm bound is
// 16 (so 1 hex character = 1 limb). The function assigns the values of m to [f, e,
// d, c, b, a, ..., 3, 2, 1, 0]. m should be preallocated and zeroized. Additionally,
// we have the guarantee that 2 bits contributing to different field elements cannot
// be part of the same limb.
func LimbDecomposeBytes(buf []byte, m babybear.Vector, logTwoBound int) {
	limbDecomposeBytes(buf, m, logTwoBound, 0, nil)
}

// decomposes m as by taking chunks of logTwoBound bits at a time. The buffer is interpreted like this:
// [0xa,        ..            , 0x1 | 0xa ... ]
//
//	<- #bytes in a field element ->
//	<-0xa is the MSB, 0x1 the LSB->
//	<-we read this chunk from right
//	   			to left 	 	 ->
//
// This function is called when logTwoBound divides the number of bits used to represent a
// babybear element.
// From a slice of field elements m:=[a_0, a_1, ...]
// Doing h.Sum(h.Write([Marshal[a_i] for i in len(m)])) is the same than
// writing the a_i in little endian, and then taking logTwoBound bits at a time.
//
// ex: m := [0x1, 0x3]
// in the hash buffer, it is interpreted like that as a stream of bits:
// [100...0 110...0] corresponding to [0x1, 0x3] in little endian, so first bit = LSbit
// then the stream of bits is splitted in chunks of logTwoBound bits.
//
// This function is called when logTwoBound divides 8.
func limbDecomposeBytesSmallBound(buf []byte, m babybear.Vector, logTwoBound, degree int, mValues *bitset.BitSet) {
	mask := byte((1 << logTwoBound) - 1)
	nbChunksPerBytes := 8 / logTwoBound
	nbFieldsElmts := len(buf) / babybear.Bytes
	for i := 0; i < nbFieldsElmts; i++ {
		for j := babybear.Bytes - 1; j >= 0; j-- {
			curByte := buf[i*babybear.Bytes+j]
			curPos := i*babybear.Bytes*nbChunksPerBytes + (babybear.Bytes-1-j)*nbChunksPerBytes
			for k := 0; k < nbChunksPerBytes; k++ {

				m[curPos+k][0] = uint32((curByte >> (k * logTwoBound)) & mask)

				// Check if mPos is zero and mark as non-zero in the bitset if not
				if m[curPos+k][0] != 0 && mValues != nil {
					mValues.Set(uint((curPos + k) / degree))
				}
			}
		}
	}
}

// limbDecomposeBytesMiddleBound same function than limbDecomposeBytesSmallBound, but logTwoBound is
// a multiple of 8, and divides the number of bits of the fields.
func limbDecomposeBytesMiddleBound(buf []byte, m babybear.Vector, logTwoBound, degree int, mValues *bitset.BitSet) {
	nbFieldsElmts := len(buf) / babybear.Bytes
	nbChunksPerElements := babybear.Bytes * 8 / logTwoBound
	nbBytesInChunk := logTwoBound / 8
	curElmt := 0
	for i := 0; i < nbFieldsElmts; i++ {
		for j := nbChunksPerElements; j > 0; j-- {
			curPos := i*babybear.Bytes + j*nbBytesInChunk
			for k := 1; k <= nbBytesInChunk; k++ {

				m[curElmt][0] |= (uint32(buf[curPos-k]) << ((k - 1) * 8))

			}
			// Check if mPos is zero and mark as non-zero in the bitset if not
			if m[curElmt][0] != 0 && mValues != nil {
				mValues.Set(uint((curElmt) / degree))
			}
			curElmt += 1
		}
	}
}

// Split an slice of bytes representing an array of serialized field element in
// big-endian form into an array of limbs representing the same field elements
// in little-endian form. Namely, if our field is represented with 64 bits and we
// have the following field element 0x0123456789abcdef (0 being the most significant
// character and and f being the least significant one) and our log norm bound is
// 16 (so 1 hex character = 1 limb). The function assigns the values of m to [f, e,
// d, c, b, a, ..., 3, 2, 1, 0]. m should be preallocated and zeroized. mValues is
// an optional bitSet. If provided, it must be empty. The function will set bit "i"
// to indicate the that i-th SIS input polynomial should be non-zero. Recall, that a
// SIS polynomial corresponds to a chunk of limbs of size `degree`. Additionally,
// we have the guarantee that 2 bits contributing to different field elements cannot
// be part of the same limb.
func limbDecomposeBytes(buf []byte, m babybear.Vector, logTwoBound, degree int, mValues *bitset.BitSet) {

	// bitwise decomposition of the buffer, in order to build m (the vector to hash)
	// as a list of polynomials, whose coefficients are less than r.B bits long.
	// Say buf=[0xbe,0x0f]. As a stream of bits it is interpreted like this:
	// 10111110 00001111. getIthBit(0)=1 (=leftmost bit), getIthBit(1)=0 (=second leftmost bit), etc.
	nbBits := len(buf) * 8
	getIthBit := func(i int) uint8 {
		k := i / 8
		if k >= len(buf) {
			return 0
		}
		b := buf[k]
		j := i % 8
		return b >> (7 - j) & 1
	}

	// we process the input buffer by blocks of r.LogTwoBound bits
	// each of these block (<< 64bits) are interpreted as a coefficient
	mPos := 0
	for fieldStart := 0; fieldStart < nbBits; {
		for bitInField := 0; bitInField < babybear.Bytes*8; {

			j := bitInField % logTwoBound

			// r.LogTwoBound < 64; we just use the first word of our element here,
			// and set the bits from LSB to MSB.
			at := fieldStart + babybear.Bytes*8 - bitInField - 1

			m[mPos][0] |= uint32(getIthBit(at) << j)

			bitInField++

			// Check if mPos is zero and mark as non-zero in the bitset if not
			if m[mPos][0] != 0 && mValues != nil {
				mValues.Set(uint(mPos / degree))
			}

			if j == logTwoBound-1 || bitInField == babybear.Bytes*8 {
				mPos++
			}
		}
		fieldStart += babybear.Bytes * 8
	}
}

// see limbDecomposeBytes; this function is optimized for the case where
// logTwoBound == 8 and degree == 64
func limbDecomposeBytes8_64(buf []byte, m babybear.Vector, mValues *bitset.BitSet) {
	// with logTwoBound == 8, we can actually advance byte per byte.
	const degree = 64
	j := 0

	for startPos := babybear.Bytes - 1; startPos < len(buf); startPos += babybear.Bytes {
		for i := startPos; i >= startPos-babybear.Bytes+1; i-- {

			m[j][0] = uint32(buf[i])

			if m[j][0] != 0 {
				mValues.Set(uint(j / degree))
			}
			j++
		}
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
