// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package sis

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash"
	"math/bits"

	"github.com/bits-and-blooms/bitset"
	"github.com/consensys/gnark-crypto/field/goldilocks"
	"github.com/consensys/gnark-crypto/field/goldilocks/fft"
	"github.com/consensys/gnark-crypto/internal/parallel"
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
	// Ag the evaluation form of the polynomials in A on the coset √(g) * <g>
	A  [][]goldilocks.Element
	Ag [][]goldilocks.Element

	// LogTwoBound (Infinity norm) of the vector to hash. It means that each component in m
	// is < 2^B, where m is the vector to hash (the hash being A*m).
	// cf https://hackmd.io/7OODKWQZRRW9RxM5BaXtIw , B >= 3.
	LogTwoBound int

	// domain for the polynomial multiplication
	Domain        *fft.Domain
	twiddleCosets []goldilocks.Element // see FFT64 and precomputeTwiddlesCoset

	// d, the degree of X^{d}+1
	Degree int

	// in bytes, represents the maximum number of bytes the .Write(...) will handle;
	// ( maximum number of bytes to sum )
	capacity            int
	maxNbElementsToHash int

	// allocate memory once per instance (used in Sum())
	bufM, bufRes goldilocks.Vector
	bufMValues   *bitset.BitSet
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
	capacity := maxNbElementsToHash * goldilocks.Bytes

	// n: number of polynomials in A
	// len(m) == degree * n
	// with each element in m being logTwoBounds bits from the instance buffer.
	// that is, to fill m, we need [degree * n * logTwoBound] bits of data
	// capacity == [degree * n * logTwoBound] / 8
	// n == (capacity*8)/(degree*logTwoBound)

	// First n <- #limbs to represent a single field element
	n := (goldilocks.Bytes * 8) / logTwoBound
	if n*logTwoBound < goldilocks.Bytes*8 {
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
	shift, err := goldilocks.Generator(uint64(2 * degree))
	if err != nil {
		return nil, err
	}

	r := &RSis{
		LogTwoBound:         logTwoBound,
		capacity:            capacity,
		Degree:              degree,
		Domain:              fft.NewDomain(uint64(degree), fft.WithShift(shift)),
		A:                   make([][]goldilocks.Element, n),
		Ag:                  make([][]goldilocks.Element, n),
		bufM:                make(goldilocks.Vector, degree*n),
		bufRes:              make(goldilocks.Vector, degree),
		bufMValues:          bitset.New(uint(n)),
		maxNbElementsToHash: maxNbElementsToHash,
	}
	if r.LogTwoBound == 8 && r.Degree == 64 {
		// TODO @gbotrel fixme, that's dirty.
		r.twiddleCosets = PrecomputeTwiddlesCoset(r.Domain.Generator, r.Domain.FrMultiplicativeGen)
	}

	// filling A
	a := make([]goldilocks.Element, n*r.Degree)
	ag := make([]goldilocks.Element, n*r.Degree)

	parallel.Execute(n, func(start, end int) {
		var buf bytes.Buffer
		for i := start; i < end; i++ {
			rstart, rend := i*r.Degree, (i+1)*r.Degree
			r.A[i] = a[rstart:rend:rend]
			r.Ag[i] = ag[rstart:rend:rend]
			for j := 0; j < r.Degree; j++ {
				r.A[i][j] = genRandom(seed, int64(i), int64(j), &buf)
			}

			// fill Ag the evaluation form of the polynomials in A on the coset √(g) * <g>
			copy(r.Ag[i], r.A[i])
			r.Domain.FFT(r.Ag[i], fft.DIF, fft.OnCoset())
		}
	})

	return r, nil
}

func (r *RSis) Write(p []byte) (n int, err error) {
	r.buffer.Write(p)
	return len(p), nil
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
// The instance buffer is interpreted as a sequence of coefficients of size r.Bound bits long.
// The function returns the hash of the polynomial as a a sequence []goldilocks.Elements, interpreted as []bytes,
// corresponding to sum_i A[i]*m Mod X^{d}+1
func (r *RSis) Sum(b []byte) []byte {
	buf := r.buffer.Bytes()
	if len(buf) > r.capacity {
		panic("buffer too large")
	}

	fastPath := r.LogTwoBound == 8 && r.Degree == 64

	// clear the buffers of the instance.
	defer r.cleanupBuffers()

	m := r.bufM
	mValues := r.bufMValues

	if fastPath {
		// fast path.
		limbDecomposeBytes8_64(buf, m, mValues)
	} else {
		limbDecomposeBytes(buf, m, r.LogTwoBound, r.Degree, mValues)
	}

	// we can hash now.
	res := r.bufRes

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

	resBytes, err := res.MarshalBinary()
	if err != nil {
		panic(err)
	}

	return append(b, resBytes[4:]...) // first 4 bytes are uint32(len(res))
}

// Reset resets the Hash to its initial state.
func (r *RSis) Reset() {
	r.buffer.Reset()
}

// Size returns the number of bytes Sum will return.
func (r *RSis) Size() int {

	// The size in bits is the size in bits of a polynomial in A.
	degree := len(r.A[0])
	totalSize := degree * goldilocks.Modulus().BitLen() / 8

	return totalSize
}

// BlockSize returns the hash's underlying block size.
// The Write method must be able to accept any amount
// of data, but it may operate more efficiently if all writes
// are a multiple of the block size.
func (r *RSis) BlockSize() int {
	return 0
}

// Construct a hasher generator. It takes as input the same parameters
// as `NewRingSIS` and outputs a function which returns fresh hasher
// everytime it is called
func NewRingSISMaker(seed int64, logTwoDegree, logTwoBound, maxNbElementsToHash int) (func() hash.Hash, error) {
	return func() hash.Hash {
		h, err := NewRSis(seed, logTwoDegree, logTwoBound, maxNbElementsToHash)
		if err != nil {
			panic(err)
		}
		return h
	}, nil

}

func genRandom(seed, i, j int64, buf *bytes.Buffer) goldilocks.Element {

	buf.Reset()
	buf.WriteString("SIS")
	binary.Write(buf, binary.BigEndian, seed)
	binary.Write(buf, binary.BigEndian, i)
	binary.Write(buf, binary.BigEndian, j)

	digest := blake2b.Sum256(buf.Bytes())

	var res goldilocks.Element
	res.SetBytes(digest[:])

	return res
}

// mulMod computes p * q in ℤ_{p}[X]/Xᵈ+1.
// Is assumed that pLagrangeShifted and qLagrangeShifted are of the correct sizes
// and that they are in evaluation form on √(g) * <g>
// The result is not FFTinversed. The fft inverse is done once every
// multiplications are done.
func mulMod(pLagrangeCosetBitReversed, qLagrangeCosetBitReversed []goldilocks.Element) []goldilocks.Element {

	res := make([]goldilocks.Element, len(pLagrangeCosetBitReversed))
	for i := 0; i < len(pLagrangeCosetBitReversed); i++ {
		res[i].Mul(&pLagrangeCosetBitReversed[i], &qLagrangeCosetBitReversed[i])
	}

	// NOT fft inv for now, wait until every part of the keys have been multiplied
	// r.Domain.FFTInverse(res, fft.DIT, true)

	return res

}

// mulMod + accumulate in res.
func mulModAcc(res []goldilocks.Element, pLagrangeCosetBitReversed, qLagrangeCosetBitReversed []goldilocks.Element) {
	var t goldilocks.Element
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
	res.bufM = make(goldilocks.Vector, len(r.bufM))
	res.bufMValues = bitset.New(r.bufMValues.Len())
	res.bufRes = make(goldilocks.Vector, len(r.bufRes))
	return res
}

// Cleanup the buffers of the RSis instance
func (r *RSis) cleanupBuffers() {
	r.bufMValues.ClearAll()
	for i := 0; i < len(r.bufM); i++ {
		r.bufM[i].SetZero()
	}
	for i := 0; i < len(r.bufRes); i++ {
		r.bufRes[i].SetZero()
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
func LimbDecomposeBytes(buf []byte, m goldilocks.Vector, logTwoBound int) {
	limbDecomposeBytes(buf, m, logTwoBound, 0, nil)
}

// From a slice of field elements m:=[a_0, a_1, ...]
// Doing h.Sum(h.Write([Marshal[a_i] for i in len(m)])) is the same than
// writing the a_i in little endian, and then taking logTwoBound bits at a time.
//
// ex: m := [0x1, 0x3]
// in the hash buffer, it is interpreted like that as a stream of bits:
// [100...0 110...0] corresponding to [0x1, 0x3] in little endian, so first bit = LSbit
// then the stream of bits is splitted in chunks of logTwoBound bits.
//
// This function is called when logTwoBound divides the number of bits used to represent a
// goldilocks element.
func limbDecomposeBytesFast_2(buf []byte, m goldilocks.Vector, logTwoBound, degree int, mValues *bitset.BitSet) {
	mask := byte(0x3)
	nbChunksPerBytes := 8 / logTwoBound
	nbFieldsElmts := len(buf) / goldilocks.Bytes
	for i := 0; i < nbFieldsElmts; i++ {
		for j := goldilocks.Bytes - 1; j >= 0; j-- {
			curByte := buf[i*goldilocks.Bytes+j]
			curPos := i*goldilocks.Bytes*nbChunksPerBytes + (goldilocks.Bytes-1-j)*nbChunksPerBytes
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
func limbDecomposeBytes(buf []byte, m goldilocks.Vector, logTwoBound, degree int, mValues *bitset.BitSet) {

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
		for bitInField := 0; bitInField < goldilocks.Bytes*8; {

			j := bitInField % logTwoBound

			// r.LogTwoBound < 64; we just use the first word of our element here,
			// and set the bits from LSB to MSB.
			at := fieldStart + goldilocks.Bytes*8 - bitInField - 1

			m[mPos][0] |= uint64(getIthBit(at) << j)

			bitInField++

			// Check if mPos is zero and mark as non-zero in the bitset if not
			if m[mPos][0] != 0 && mValues != nil {
				mValues.Set(uint(mPos / degree))
			}

			if j == logTwoBound-1 || bitInField == goldilocks.Bytes*8 {
				mPos++
			}
		}
		fieldStart += goldilocks.Bytes * 8
	}
}

// see limbDecomposeBytes; this function is optimized for the case where
// logTwoBound == 8 and degree == 64
func limbDecomposeBytes8_64(buf []byte, m goldilocks.Vector, mValues *bitset.BitSet) {
	// with logTwoBound == 8, we can actually advance byte per byte.
	const degree = 64
	j := 0

	for startPos := goldilocks.Bytes - 1; startPos < len(buf); startPos += goldilocks.Bytes {
		for i := startPos; i >= startPos-goldilocks.Bytes+1; i-- {

			m[j][0] = uint64(buf[i])

			if m[j][0] != 0 {
				mValues.Set(uint(j / degree))
			}
			j++
		}
	}
}
