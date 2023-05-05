// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fr

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"math/big"
	"math/bits"
	"reflect"
	"strconv"
	"strings"

	"github.com/bits-and-blooms/bitset"
	"github.com/consensys/gnark-crypto/field/hash"
	"github.com/consensys/gnark-crypto/field/pool"
)

// Element represents a field element stored on 4 words (uint64)
//
// Element are assumed to be in Montgomery form in all methods.
//
// Modulus q =
//
//	q[base10] = 115792089237316195423570985008687907852837564279074904382605163141518161494337
//	q[base16] = 0xfffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141
//
// # Warning
//
// This code has not been audited and is provided as-is. In particular, there is no security guarantees such as constant time implementation or side-channel attack resistance.
type Element [4]uint64

const (
	Limbs = 4   // number of 64 bits words needed to represent a Element
	Bits  = 256 // number of bits needed to represent a Element
	Bytes = 32  // number of bytes needed to represent a Element
)

// Field modulus q
const (
	q0 uint64 = 13822214165235122497
	q1 uint64 = 13451932020343611451
	q2 uint64 = 18446744073709551614
	q3 uint64 = 18446744073709551615
)

var qElement = Element{
	q0,
	q1,
	q2,
	q3,
}

var _modulus big.Int // q stored as big.Int

// Modulus returns q as a big.Int
//
//	q[base10] = 115792089237316195423570985008687907852837564279074904382605163141518161494337
//	q[base16] = 0xfffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141
func Modulus() *big.Int {
	return new(big.Int).Set(&_modulus)
}

// q + r'.r = 1, i.e., qInvNeg = - q⁻¹ mod r
// used for Montgomery reduction
const qInvNeg uint64 = 5408259542528602431

func init() {
	_modulus.SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
}

// NewElement returns a new Element from a uint64 value
//
// it is equivalent to
//
//	var v Element
//	v.SetUint64(...)
func NewElement(v uint64) Element {
	z := Element{v}
	z.Mul(&z, &rSquare)
	return z
}

// SetUint64 sets z to v and returns z
func (z *Element) SetUint64(v uint64) *Element {
	//  sets z LSB to v (non-Montgomery form) and convert z to Montgomery form
	*z = Element{v}
	return z.Mul(z, &rSquare) // z.toMont()
}

// SetInt64 sets z to v and returns z
func (z *Element) SetInt64(v int64) *Element {

	// absolute value of v
	m := v >> 63
	z.SetUint64(uint64((v ^ m) - m))

	if m != 0 {
		// v is negative
		z.Neg(z)
	}

	return z
}

// Set z = x and returns z
func (z *Element) Set(x *Element) *Element {
	z[0] = x[0]
	z[1] = x[1]
	z[2] = x[2]
	z[3] = x[3]
	return z
}

// SetInterface converts provided interface into Element
// returns an error if provided type is not supported
// supported types:
//
//	Element
//	*Element
//	uint64
//	int
//	string (see SetString for valid formats)
//	*big.Int
//	big.Int
//	[]byte
func (z *Element) SetInterface(i1 interface{}) (*Element, error) {
	if i1 == nil {
		return nil, errors.New("can't set fr.Element with <nil>")
	}

	switch c1 := i1.(type) {
	case Element:
		return z.Set(&c1), nil
	case *Element:
		if c1 == nil {
			return nil, errors.New("can't set fr.Element with <nil>")
		}
		return z.Set(c1), nil
	case uint8:
		return z.SetUint64(uint64(c1)), nil
	case uint16:
		return z.SetUint64(uint64(c1)), nil
	case uint32:
		return z.SetUint64(uint64(c1)), nil
	case uint:
		return z.SetUint64(uint64(c1)), nil
	case uint64:
		return z.SetUint64(c1), nil
	case int8:
		return z.SetInt64(int64(c1)), nil
	case int16:
		return z.SetInt64(int64(c1)), nil
	case int32:
		return z.SetInt64(int64(c1)), nil
	case int64:
		return z.SetInt64(c1), nil
	case int:
		return z.SetInt64(int64(c1)), nil
	case string:
		return z.SetString(c1)
	case *big.Int:
		if c1 == nil {
			return nil, errors.New("can't set fr.Element with <nil>")
		}
		return z.SetBigInt(c1), nil
	case big.Int:
		return z.SetBigInt(&c1), nil
	case []byte:
		return z.SetBytes(c1), nil
	default:
		return nil, errors.New("can't set fr.Element from type " + reflect.TypeOf(i1).String())
	}
}

// SetZero z = 0
func (z *Element) SetZero() *Element {
	z[0] = 0
	z[1] = 0
	z[2] = 0
	z[3] = 0
	return z
}

// SetOne z = 1 (in Montgomery form)
func (z *Element) SetOne() *Element {
	z[0] = 4624529908474429119
	z[1] = 4994812053365940164
	z[2] = 1
	z[3] = 0
	return z
}

// Div z = x*y⁻¹ (mod q)
func (z *Element) Div(x, y *Element) *Element {
	var yInv Element
	yInv.Inverse(y)
	z.Mul(x, &yInv)
	return z
}

// Equal returns z == x; constant-time
func (z *Element) Equal(x *Element) bool {
	return z.NotEqual(x) == 0
}

// NotEqual returns 0 if and only if z == x; constant-time
func (z *Element) NotEqual(x *Element) uint64 {
	return (z[3] ^ x[3]) | (z[2] ^ x[2]) | (z[1] ^ x[1]) | (z[0] ^ x[0])
}

// IsZero returns z == 0
func (z *Element) IsZero() bool {
	return (z[3] | z[2] | z[1] | z[0]) == 0
}

// IsOne returns z == 1
func (z *Element) IsOne() bool {
	return (z[3] | z[2] ^ 1 | z[1] ^ 4994812053365940164 | z[0] ^ 4624529908474429119) == 0
}

// IsUint64 reports whether z can be represented as an uint64.
func (z *Element) IsUint64() bool {
	zz := *z
	zz.fromMont()
	return zz.FitsOnOneWord()
}

// Uint64 returns the uint64 representation of x. If x cannot be represented in a uint64, the result is undefined.
func (z *Element) Uint64() uint64 {
	return z.Bits()[0]
}

// FitsOnOneWord reports whether z words (except the least significant word) are 0
//
// It is the responsibility of the caller to convert from Montgomery to Regular form if needed.
func (z *Element) FitsOnOneWord() bool {
	return (z[3] | z[2] | z[1]) == 0
}

// Cmp compares (lexicographic order) z and x and returns:
//
//	-1 if z <  x
//	 0 if z == x
//	+1 if z >  x
func (z *Element) Cmp(x *Element) int {
	_z := z.Bits()
	_x := x.Bits()
	if _z[3] > _x[3] {
		return 1
	} else if _z[3] < _x[3] {
		return -1
	}
	if _z[2] > _x[2] {
		return 1
	} else if _z[2] < _x[2] {
		return -1
	}
	if _z[1] > _x[1] {
		return 1
	} else if _z[1] < _x[1] {
		return -1
	}
	if _z[0] > _x[0] {
		return 1
	} else if _z[0] < _x[0] {
		return -1
	}
	return 0
}

// LexicographicallyLargest returns true if this element is strictly lexicographically
// larger than its negation, false otherwise
func (z *Element) LexicographicallyLargest() bool {
	// adapted from github.com/zkcrypto/bls12_381
	// we check if the element is larger than (q-1) / 2
	// if z - (((q -1) / 2) + 1) have no underflow, then z > (q-1) / 2

	_z := z.Bits()

	var b uint64
	_, b = bits.Sub64(_z[0], 16134479119472337057, 0)
	_, b = bits.Sub64(_z[1], 6725966010171805725, b)
	_, b = bits.Sub64(_z[2], 18446744073709551615, b)
	_, b = bits.Sub64(_z[3], 9223372036854775807, b)

	return b == 0
}

// SetRandom sets z to a uniform random value in [0, q).
//
// This might error only if reading from crypto/rand.Reader errors,
// in which case, value of z is undefined.
func (z *Element) SetRandom() (*Element, error) {
	// this code is generated for all modulus
	// and derived from go/src/crypto/rand/util.go

	// l is number of limbs * 8; the number of bytes needed to reconstruct 4 uint64
	const l = 32

	// bitLen is the maximum bit length needed to encode a value < q.
	const bitLen = 256

	// k is the maximum byte length needed to encode a value < q.
	const k = (bitLen + 7) / 8

	// b is the number of bits in the most significant byte of q-1.
	b := uint(bitLen % 8)
	if b == 0 {
		b = 8
	}

	var bytes [l]byte

	for {
		// note that bytes[k:l] is always 0
		if _, err := io.ReadFull(rand.Reader, bytes[:k]); err != nil {
			return nil, err
		}

		// Clear unused bits in in the most significant byte to increase probability
		// that the candidate is < q.
		bytes[k-1] &= uint8(int(1<<b) - 1)
		z[0] = binary.LittleEndian.Uint64(bytes[0:8])
		z[1] = binary.LittleEndian.Uint64(bytes[8:16])
		z[2] = binary.LittleEndian.Uint64(bytes[16:24])
		z[3] = binary.LittleEndian.Uint64(bytes[24:32])

		if !z.smallerThanModulus() {
			continue // ignore the candidate and re-sample
		}

		return z, nil
	}
}

// smallerThanModulus returns true if z < q
// This is not constant time
func (z *Element) smallerThanModulus() bool {
	return (z[3] < q3 || (z[3] == q3 && (z[2] < q2 || (z[2] == q2 && (z[1] < q1 || (z[1] == q1 && (z[0] < q0)))))))
}

// One returns 1
func One() Element {
	var one Element
	one.SetOne()
	return one
}

// Halve sets z to z / 2 (mod q)
func (z *Element) Halve() {
	var carry uint64

	if z[0]&1 == 1 {
		// z = z + q
		z[0], carry = bits.Add64(z[0], q0, 0)
		z[1], carry = bits.Add64(z[1], q1, carry)
		z[2], carry = bits.Add64(z[2], q2, carry)
		z[3], carry = bits.Add64(z[3], q3, carry)

	}
	// z = z >> 1
	z[0] = z[0]>>1 | z[1]<<63
	z[1] = z[1]>>1 | z[2]<<63
	z[2] = z[2]>>1 | z[3]<<63
	z[3] >>= 1

	if carry != 0 {
		// when we added q, the result was larger than our available limbs
		// when we shift right, we need to set the highest bit
		z[3] |= (1 << 63)
	}

}

// fromMont converts z in place (i.e. mutates) from Montgomery to regular representation
// sets and returns z = z * 1
func (z *Element) fromMont() *Element {
	fromMont(z)
	return z
}

// Add z = x + y (mod q)
func (z *Element) Add(x, y *Element) *Element {

	var carry uint64
	z[0], carry = bits.Add64(x[0], y[0], 0)
	z[1], carry = bits.Add64(x[1], y[1], carry)
	z[2], carry = bits.Add64(x[2], y[2], carry)
	z[3], carry = bits.Add64(x[3], y[3], carry)
	// if we overflowed the last addition, z >= q
	// if z >= q, z = z - q
	if carry != 0 {
		var b uint64
		// we overflowed, so z >= q
		z[0], b = bits.Sub64(z[0], q0, 0)
		z[1], b = bits.Sub64(z[1], q1, b)
		z[2], b = bits.Sub64(z[2], q2, b)
		z[3], _ = bits.Sub64(z[3], q3, b)
		return z
	}

	// if z ⩾ q → z -= q
	if !z.smallerThanModulus() {
		var b uint64
		z[0], b = bits.Sub64(z[0], q0, 0)
		z[1], b = bits.Sub64(z[1], q1, b)
		z[2], b = bits.Sub64(z[2], q2, b)
		z[3], _ = bits.Sub64(z[3], q3, b)
	}
	return z
}

// Double z = x + x (mod q), aka Lsh 1
func (z *Element) Double(x *Element) *Element {

	var carry uint64
	z[0], carry = bits.Add64(x[0], x[0], 0)
	z[1], carry = bits.Add64(x[1], x[1], carry)
	z[2], carry = bits.Add64(x[2], x[2], carry)
	z[3], carry = bits.Add64(x[3], x[3], carry)
	// if we overflowed the last addition, z >= q
	// if z >= q, z = z - q
	if carry != 0 {
		var b uint64
		// we overflowed, so z >= q
		z[0], b = bits.Sub64(z[0], q0, 0)
		z[1], b = bits.Sub64(z[1], q1, b)
		z[2], b = bits.Sub64(z[2], q2, b)
		z[3], _ = bits.Sub64(z[3], q3, b)
		return z
	}

	// if z ⩾ q → z -= q
	if !z.smallerThanModulus() {
		var b uint64
		z[0], b = bits.Sub64(z[0], q0, 0)
		z[1], b = bits.Sub64(z[1], q1, b)
		z[2], b = bits.Sub64(z[2], q2, b)
		z[3], _ = bits.Sub64(z[3], q3, b)
	}
	return z
}

// Sub z = x - y (mod q)
func (z *Element) Sub(x, y *Element) *Element {
	var b uint64
	z[0], b = bits.Sub64(x[0], y[0], 0)
	z[1], b = bits.Sub64(x[1], y[1], b)
	z[2], b = bits.Sub64(x[2], y[2], b)
	z[3], b = bits.Sub64(x[3], y[3], b)
	if b != 0 {
		var c uint64
		z[0], c = bits.Add64(z[0], q0, 0)
		z[1], c = bits.Add64(z[1], q1, c)
		z[2], c = bits.Add64(z[2], q2, c)
		z[3], _ = bits.Add64(z[3], q3, c)
	}
	return z
}

// Neg z = q - x
func (z *Element) Neg(x *Element) *Element {
	if x.IsZero() {
		z.SetZero()
		return z
	}
	var borrow uint64
	z[0], borrow = bits.Sub64(q0, x[0], 0)
	z[1], borrow = bits.Sub64(q1, x[1], borrow)
	z[2], borrow = bits.Sub64(q2, x[2], borrow)
	z[3], _ = bits.Sub64(q3, x[3], borrow)
	return z
}

// Select is a constant-time conditional move.
// If c=0, z = x0. Else z = x1
func (z *Element) Select(c int, x0 *Element, x1 *Element) *Element {
	cC := uint64((int64(c) | -int64(c)) >> 63) // "canonicized" into: 0 if c=0, -1 otherwise
	z[0] = x0[0] ^ cC&(x0[0]^x1[0])
	z[1] = x0[1] ^ cC&(x0[1]^x1[1])
	z[2] = x0[2] ^ cC&(x0[2]^x1[2])
	z[3] = x0[3] ^ cC&(x0[3]^x1[3])
	return z
}

// _mulGeneric is unoptimized textbook CIOS
// it is a fallback solution on x86 when ADX instruction set is not available
// and is used for testing purposes.
func _mulGeneric(z, x, y *Element) {

	// Implements CIOS multiplication -- section 2.3.2 of Tolga Acar's thesis
	// https://www.microsoft.com/en-us/research/wp-content/uploads/1998/06/97Acar.pdf
	//
	// The algorithm:
	//
	// for i=0 to N-1
	// 		C := 0
	// 		for j=0 to N-1
	// 			(C,t[j]) := t[j] + x[j]*y[i] + C
	// 		(t[N+1],t[N]) := t[N] + C
	//
	// 		C := 0
	// 		m := t[0]*q'[0] mod D
	// 		(C,_) := t[0] + m*q[0]
	// 		for j=1 to N-1
	// 			(C,t[j-1]) := t[j] + m*q[j] + C
	//
	// 		(C,t[N-1]) := t[N] + C
	// 		t[N] := t[N+1] + C
	//
	// → N is the number of machine words needed to store the modulus q
	// → D is the word size. For example, on a 64-bit architecture D is 2	64
	// → x[i], y[i], q[i] is the ith word of the numbers x,y,q
	// → q'[0] is the lowest word of the number -q⁻¹ mod r. This quantity is pre-computed, as it does not depend on the inputs.
	// → t is a temporary array of size N+2
	// → C, S are machine words. A pair (C,S) refers to (hi-bits, lo-bits) of a two-word number

	var t [5]uint64
	var D uint64
	var m, C uint64
	// -----------------------------------
	// First loop

	C, t[0] = bits.Mul64(y[0], x[0])
	C, t[1] = madd1(y[0], x[1], C)
	C, t[2] = madd1(y[0], x[2], C)
	C, t[3] = madd1(y[0], x[3], C)

	t[4], D = bits.Add64(t[4], C, 0)

	// m = t[0]n'[0] mod W
	m = t[0] * qInvNeg

	// -----------------------------------
	// Second loop
	C = madd0(m, q0, t[0])
	C, t[0] = madd2(m, q1, t[1], C)
	C, t[1] = madd2(m, q2, t[2], C)
	C, t[2] = madd2(m, q3, t[3], C)

	t[3], C = bits.Add64(t[4], C, 0)
	t[4], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(y[1], x[0], t[0])
	C, t[1] = madd2(y[1], x[1], t[1], C)
	C, t[2] = madd2(y[1], x[2], t[2], C)
	C, t[3] = madd2(y[1], x[3], t[3], C)

	t[4], D = bits.Add64(t[4], C, 0)

	// m = t[0]n'[0] mod W
	m = t[0] * qInvNeg

	// -----------------------------------
	// Second loop
	C = madd0(m, q0, t[0])
	C, t[0] = madd2(m, q1, t[1], C)
	C, t[1] = madd2(m, q2, t[2], C)
	C, t[2] = madd2(m, q3, t[3], C)

	t[3], C = bits.Add64(t[4], C, 0)
	t[4], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(y[2], x[0], t[0])
	C, t[1] = madd2(y[2], x[1], t[1], C)
	C, t[2] = madd2(y[2], x[2], t[2], C)
	C, t[3] = madd2(y[2], x[3], t[3], C)

	t[4], D = bits.Add64(t[4], C, 0)

	// m = t[0]n'[0] mod W
	m = t[0] * qInvNeg

	// -----------------------------------
	// Second loop
	C = madd0(m, q0, t[0])
	C, t[0] = madd2(m, q1, t[1], C)
	C, t[1] = madd2(m, q2, t[2], C)
	C, t[2] = madd2(m, q3, t[3], C)

	t[3], C = bits.Add64(t[4], C, 0)
	t[4], _ = bits.Add64(0, D, C)
	// -----------------------------------
	// First loop

	C, t[0] = madd1(y[3], x[0], t[0])
	C, t[1] = madd2(y[3], x[1], t[1], C)
	C, t[2] = madd2(y[3], x[2], t[2], C)
	C, t[3] = madd2(y[3], x[3], t[3], C)

	t[4], D = bits.Add64(t[4], C, 0)

	// m = t[0]n'[0] mod W
	m = t[0] * qInvNeg

	// -----------------------------------
	// Second loop
	C = madd0(m, q0, t[0])
	C, t[0] = madd2(m, q1, t[1], C)
	C, t[1] = madd2(m, q2, t[2], C)
	C, t[2] = madd2(m, q3, t[3], C)

	t[3], C = bits.Add64(t[4], C, 0)
	t[4], _ = bits.Add64(0, D, C)

	if t[4] != 0 {
		// we need to reduce, we have a result on 5 words
		var b uint64
		z[0], b = bits.Sub64(t[0], q0, 0)
		z[1], b = bits.Sub64(t[1], q1, b)
		z[2], b = bits.Sub64(t[2], q2, b)
		z[3], _ = bits.Sub64(t[3], q3, b)
		return
	}

	// copy t into z
	z[0] = t[0]
	z[1] = t[1]
	z[2] = t[2]
	z[3] = t[3]

	// if z ⩾ q → z -= q
	if !z.smallerThanModulus() {
		var b uint64
		z[0], b = bits.Sub64(z[0], q0, 0)
		z[1], b = bits.Sub64(z[1], q1, b)
		z[2], b = bits.Sub64(z[2], q2, b)
		z[3], _ = bits.Sub64(z[3], q3, b)
	}
}

func _fromMontGeneric(z *Element) {
	// the following lines implement z = z * 1
	// with a modified CIOS montgomery multiplication
	// see Mul for algorithm documentation
	{
		// m = z[0]n'[0] mod W
		m := z[0] * qInvNeg
		C := madd0(m, q0, z[0])
		C, z[0] = madd2(m, q1, z[1], C)
		C, z[1] = madd2(m, q2, z[2], C)
		C, z[2] = madd2(m, q3, z[3], C)
		z[3] = C
	}
	{
		// m = z[0]n'[0] mod W
		m := z[0] * qInvNeg
		C := madd0(m, q0, z[0])
		C, z[0] = madd2(m, q1, z[1], C)
		C, z[1] = madd2(m, q2, z[2], C)
		C, z[2] = madd2(m, q3, z[3], C)
		z[3] = C
	}
	{
		// m = z[0]n'[0] mod W
		m := z[0] * qInvNeg
		C := madd0(m, q0, z[0])
		C, z[0] = madd2(m, q1, z[1], C)
		C, z[1] = madd2(m, q2, z[2], C)
		C, z[2] = madd2(m, q3, z[3], C)
		z[3] = C
	}
	{
		// m = z[0]n'[0] mod W
		m := z[0] * qInvNeg
		C := madd0(m, q0, z[0])
		C, z[0] = madd2(m, q1, z[1], C)
		C, z[1] = madd2(m, q2, z[2], C)
		C, z[2] = madd2(m, q3, z[3], C)
		z[3] = C
	}

	// if z ⩾ q → z -= q
	if !z.smallerThanModulus() {
		var b uint64
		z[0], b = bits.Sub64(z[0], q0, 0)
		z[1], b = bits.Sub64(z[1], q1, b)
		z[2], b = bits.Sub64(z[2], q2, b)
		z[3], _ = bits.Sub64(z[3], q3, b)
	}
}

func _reduceGeneric(z *Element) {

	// if z ⩾ q → z -= q
	if !z.smallerThanModulus() {
		var b uint64
		z[0], b = bits.Sub64(z[0], q0, 0)
		z[1], b = bits.Sub64(z[1], q1, b)
		z[2], b = bits.Sub64(z[2], q2, b)
		z[3], _ = bits.Sub64(z[3], q3, b)
	}
}

// BatchInvert returns a new slice with every element inverted.
// Uses Montgomery batch inversion trick
func BatchInvert(a []Element) []Element {
	res := make([]Element, len(a))
	if len(a) == 0 {
		return res
	}

	zeroes := bitset.New(uint(len(a)))
	accumulator := One()

	for i := 0; i < len(a); i++ {
		if a[i].IsZero() {
			zeroes.Set(uint(i))
			continue
		}
		res[i] = accumulator
		accumulator.Mul(&accumulator, &a[i])
	}

	accumulator.Inverse(&accumulator)

	for i := len(a) - 1; i >= 0; i-- {
		if zeroes.Test(uint(i)) {
			continue
		}
		res[i].Mul(&res[i], &accumulator)
		accumulator.Mul(&accumulator, &a[i])
	}

	return res
}

func _butterflyGeneric(a, b *Element) {
	t := *a
	a.Add(a, b)
	b.Sub(&t, b)
}

// BitLen returns the minimum number of bits needed to represent z
// returns 0 if z == 0
func (z *Element) BitLen() int {
	if z[3] != 0 {
		return 192 + bits.Len64(z[3])
	}
	if z[2] != 0 {
		return 128 + bits.Len64(z[2])
	}
	if z[1] != 0 {
		return 64 + bits.Len64(z[1])
	}
	return bits.Len64(z[0])
}

// Hash msg to count prime field elements.
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-5.2
func Hash(msg, dst []byte, count int) ([]Element, error) {
	// 128 bits of security
	// L = ceil((ceil(log2(p)) + k) / 8), where k is the security parameter = 128
	const Bytes = 1 + (Bits-1)/8
	const L = 16 + Bytes

	lenInBytes := count * L
	pseudoRandomBytes, err := hash.ExpandMsgXmd(msg, dst, lenInBytes)
	if err != nil {
		return nil, err
	}

	// get temporary big int from the pool
	vv := pool.BigInt.Get()

	res := make([]Element, count)
	for i := 0; i < count; i++ {
		vv.SetBytes(pseudoRandomBytes[i*L : (i+1)*L])
		res[i].SetBigInt(vv)
	}

	// release object into pool
	pool.BigInt.Put(vv)

	return res, nil
}

// Exp z = xᵏ (mod q)
func (z *Element) Exp(x Element, k *big.Int) *Element {
	if k.IsUint64() && k.Uint64() == 0 {
		return z.SetOne()
	}

	e := k
	if k.Sign() == -1 {
		// negative k, we invert
		// if k < 0: xᵏ (mod q) == (x⁻¹)ᵏ (mod q)
		x.Inverse(&x)

		// we negate k in a temp big.Int since
		// Int.Bit(_) of k and -k is different
		e = pool.BigInt.Get()
		defer pool.BigInt.Put(e)
		e.Neg(k)
	}

	z.Set(&x)

	for i := e.BitLen() - 2; i >= 0; i-- {
		z.Square(z)
		if e.Bit(i) == 1 {
			z.Mul(z, &x)
		}
	}

	return z
}

// rSquare where r is the Montgommery constant
// see section 2.3.2 of Tolga Acar's thesis
// https://www.microsoft.com/en-us/research/wp-content/uploads/1998/06/97Acar.pdf
var rSquare = Element{
	9902555850136342848,
	8364476168144746616,
	16616019711348246470,
	11342065889886772165,
}

// toMont converts z to Montgomery form
// sets and returns z = z * r²
func (z *Element) toMont() *Element {
	return z.Mul(z, &rSquare)
}

// String returns the decimal representation of z as generated by
// z.Text(10).
func (z *Element) String() string {
	return z.Text(10)
}

// toBigInt returns z as a big.Int in Montgomery form
func (z *Element) toBigInt(res *big.Int) *big.Int {
	var b [Bytes]byte
	binary.BigEndian.PutUint64(b[24:32], z[0])
	binary.BigEndian.PutUint64(b[16:24], z[1])
	binary.BigEndian.PutUint64(b[8:16], z[2])
	binary.BigEndian.PutUint64(b[0:8], z[3])

	return res.SetBytes(b[:])
}

// Text returns the string representation of z in the given base.
// Base must be between 2 and 36, inclusive. The result uses the
// lower-case letters 'a' to 'z' for digit values 10 to 35.
// No prefix (such as "0x") is added to the string. If z is a nil
// pointer it returns "<nil>".
// If base == 10 and -z fits in a uint16 prefix "-" is added to the string.
func (z *Element) Text(base int) string {
	if base < 2 || base > 36 {
		panic("invalid base")
	}
	if z == nil {
		return "<nil>"
	}

	const maxUint16 = 65535
	if base == 10 {
		var zzNeg Element
		zzNeg.Neg(z)
		zzNeg.fromMont()
		if zzNeg.FitsOnOneWord() && zzNeg[0] <= maxUint16 && zzNeg[0] != 0 {
			return "-" + strconv.FormatUint(zzNeg[0], base)
		}
	}
	zz := *z
	zz.fromMont()
	if zz.FitsOnOneWord() {
		return strconv.FormatUint(zz[0], base)
	}
	vv := pool.BigInt.Get()
	r := zz.toBigInt(vv).Text(base)
	pool.BigInt.Put(vv)
	return r
}

// BigInt sets and return z as a *big.Int
func (z *Element) BigInt(res *big.Int) *big.Int {
	_z := *z
	_z.fromMont()
	return _z.toBigInt(res)
}

// ToBigIntRegular returns z as a big.Int in regular form
//
// Deprecated: use BigInt(*big.Int) instead
func (z Element) ToBigIntRegular(res *big.Int) *big.Int {
	z.fromMont()
	return z.toBigInt(res)
}

// Bits provides access to z by returning its value as a little-endian [4]uint64 array.
// Bits is intended to support implementation of missing low-level Element
// functionality outside this package; it should be avoided otherwise.
func (z *Element) Bits() [4]uint64 {
	_z := *z
	fromMont(&_z)
	return _z
}

// Bytes returns the value of z as a big-endian byte array
func (z *Element) Bytes() (res [Bytes]byte) {
	BigEndian.PutElement(&res, *z)
	return
}

// Marshal returns the value of z as a big-endian byte slice
func (z *Element) Marshal() []byte {
	b := z.Bytes()
	return b[:]
}

// SetBytes interprets e as the bytes of a big-endian unsigned integer,
// sets z to that value, and returns z.
func (z *Element) SetBytes(e []byte) *Element {
	if len(e) == Bytes {
		// fast path
		v, err := BigEndian.Element((*[Bytes]byte)(e))
		if err == nil {
			*z = v
			return z
		}
	}

	// slow path.
	// get a big int from our pool
	vv := pool.BigInt.Get()
	vv.SetBytes(e)

	// set big int
	z.SetBigInt(vv)

	// put temporary object back in pool
	pool.BigInt.Put(vv)

	return z
}

// SetBytesCanonical interprets e as the bytes of a big-endian 32-byte integer.
// If e is not a 32-byte slice or encodes a value higher than q,
// SetBytesCanonical returns an error.
func (z *Element) SetBytesCanonical(e []byte) error {
	if len(e) != Bytes {
		return errors.New("invalid fr.Element encoding")
	}
	v, err := BigEndian.Element((*[Bytes]byte)(e))
	if err != nil {
		return err
	}
	*z = v
	return nil
}

// SetBigInt sets z to v and returns z
func (z *Element) SetBigInt(v *big.Int) *Element {
	z.SetZero()

	var zero big.Int

	// fast path
	c := v.Cmp(&_modulus)
	if c == 0 {
		// v == 0
		return z
	} else if c != 1 && v.Cmp(&zero) != -1 {
		// 0 < v < q
		return z.setBigInt(v)
	}

	// get temporary big int from the pool
	vv := pool.BigInt.Get()

	// copy input + modular reduction
	vv.Mod(v, &_modulus)

	// set big int byte value
	z.setBigInt(vv)

	// release object into pool
	pool.BigInt.Put(vv)
	return z
}

// setBigInt assumes 0 ⩽ v < q
func (z *Element) setBigInt(v *big.Int) *Element {
	vBits := v.Bits()

	if bits.UintSize == 64 {
		for i := 0; i < len(vBits); i++ {
			z[i] = uint64(vBits[i])
		}
	} else {
		for i := 0; i < len(vBits); i++ {
			if i%2 == 0 {
				z[i/2] = uint64(vBits[i])
			} else {
				z[i/2] |= uint64(vBits[i]) << 32
			}
		}
	}

	return z.toMont()
}

// SetString creates a big.Int with number and calls SetBigInt on z
//
// The number prefix determines the actual base: A prefix of
// ”0b” or ”0B” selects base 2, ”0”, ”0o” or ”0O” selects base 8,
// and ”0x” or ”0X” selects base 16. Otherwise, the selected base is 10
// and no prefix is accepted.
//
// For base 16, lower and upper case letters are considered the same:
// The letters 'a' to 'f' and 'A' to 'F' represent digit values 10 to 15.
//
// An underscore character ”_” may appear between a base
// prefix and an adjacent digit, and between successive digits; such
// underscores do not change the value of the number.
// Incorrect placement of underscores is reported as a panic if there
// are no other errors.
//
// If the number is invalid this method leaves z unchanged and returns nil, error.
func (z *Element) SetString(number string) (*Element, error) {
	// get temporary big int from the pool
	vv := pool.BigInt.Get()

	if _, ok := vv.SetString(number, 0); !ok {
		return nil, errors.New("Element.SetString failed -> can't parse number into a big.Int " + number)
	}

	z.SetBigInt(vv)

	// release object into pool
	pool.BigInt.Put(vv)

	return z, nil
}

// MarshalJSON returns json encoding of z (z.Text(10))
// If z == nil, returns null
func (z *Element) MarshalJSON() ([]byte, error) {
	if z == nil {
		return []byte("null"), nil
	}
	const maxSafeBound = 15 // we encode it as number if it's small
	s := z.Text(10)
	if len(s) <= maxSafeBound {
		return []byte(s), nil
	}
	var sbb strings.Builder
	sbb.WriteByte('"')
	sbb.WriteString(s)
	sbb.WriteByte('"')
	return []byte(sbb.String()), nil
}

// UnmarshalJSON accepts numbers and strings as input
// See Element.SetString for valid prefixes (0x, 0b, ...)
func (z *Element) UnmarshalJSON(data []byte) error {
	s := string(data)
	if len(s) > Bits*3 {
		return errors.New("value too large (max = Element.Bits * 3)")
	}

	// we accept numbers and strings, remove leading and trailing quotes if any
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}

	// get temporary big int from the pool
	vv := pool.BigInt.Get()

	if _, ok := vv.SetString(s, 0); !ok {
		return errors.New("can't parse into a big.Int: " + s)
	}

	z.SetBigInt(vv)

	// release object into pool
	pool.BigInt.Put(vv)
	return nil
}

// A ByteOrder specifies how to convert byte slices into a Element
type ByteOrder interface {
	Element(*[Bytes]byte) (Element, error)
	PutElement(*[Bytes]byte, Element)
	String() string
}

// BigEndian is the big-endian implementation of ByteOrder and AppendByteOrder.
var BigEndian bigEndian

type bigEndian struct{}

// Element interpret b is a big-endian 32-byte slice.
// If b encodes a value higher than q, Element returns error.
func (bigEndian) Element(b *[Bytes]byte) (Element, error) {
	var z Element
	z[0] = binary.BigEndian.Uint64((*b)[24:32])
	z[1] = binary.BigEndian.Uint64((*b)[16:24])
	z[2] = binary.BigEndian.Uint64((*b)[8:16])
	z[3] = binary.BigEndian.Uint64((*b)[0:8])

	if !z.smallerThanModulus() {
		return Element{}, errors.New("invalid fr.Element encoding")
	}

	z.toMont()
	return z, nil
}

func (bigEndian) PutElement(b *[Bytes]byte, e Element) {
	e.fromMont()
	binary.BigEndian.PutUint64((*b)[24:32], e[0])
	binary.BigEndian.PutUint64((*b)[16:24], e[1])
	binary.BigEndian.PutUint64((*b)[8:16], e[2])
	binary.BigEndian.PutUint64((*b)[0:8], e[3])
}

func (bigEndian) String() string { return "BigEndian" }

// LittleEndian is the little-endian implementation of ByteOrder and AppendByteOrder.
var LittleEndian littleEndian

type littleEndian struct{}

func (littleEndian) Element(b *[Bytes]byte) (Element, error) {
	var z Element
	z[0] = binary.LittleEndian.Uint64((*b)[0:8])
	z[1] = binary.LittleEndian.Uint64((*b)[8:16])
	z[2] = binary.LittleEndian.Uint64((*b)[16:24])
	z[3] = binary.LittleEndian.Uint64((*b)[24:32])

	if !z.smallerThanModulus() {
		return Element{}, errors.New("invalid fr.Element encoding")
	}

	z.toMont()
	return z, nil
}

func (littleEndian) PutElement(b *[Bytes]byte, e Element) {
	e.fromMont()
	binary.LittleEndian.PutUint64((*b)[0:8], e[0])
	binary.LittleEndian.PutUint64((*b)[8:16], e[1])
	binary.LittleEndian.PutUint64((*b)[16:24], e[2])
	binary.LittleEndian.PutUint64((*b)[24:32], e[3])
}

func (littleEndian) String() string { return "LittleEndian" }

// Legendre returns the Legendre symbol of z (either +1, -1, or 0.)
func (z *Element) Legendre() int {
	var l Element
	// z^((q-1)/2)
	l.expByLegendreExp(*z)

	if l.IsZero() {
		return 0
	}

	// if l == 1
	if l.IsOne() {
		return 1
	}
	return -1
}

// Sqrt z = √x (mod q)
// if the square root doesn't exist (x is not a square mod q)
// Sqrt leaves z unchanged and returns nil
func (z *Element) Sqrt(x *Element) *Element {
	// q ≡ 1 (mod 4)
	// see modSqrtTonelliShanks in math/big/int.go
	// using https://www.maa.org/sites/default/files/pdf/upload_library/22/Polya/07468342.di020786.02p0470a.pdf

	var y, b, t, w Element
	// w = x^((s-1)/2))
	w.expBySqrtExp(*x)

	// y = x^((s+1)/2)) = w * x
	y.Mul(x, &w)

	// b = xˢ = w * w * x = y * x
	b.Mul(&w, &y)

	// g = nonResidue ^ s
	var g = Element{
		16727483617216526287,
		14607548025256143850,
		15265302390528700431,
		15433920720005950142,
	}
	r := uint64(6)

	// compute legendre symbol
	// t = x^((q-1)/2) = r-1 squaring of xˢ
	t = b
	for i := uint64(0); i < r-1; i++ {
		t.Square(&t)
	}
	if t.IsZero() {
		return z.SetZero()
	}
	if !t.IsOne() {
		// t != 1, we don't have a square root
		return nil
	}
	for {
		var m uint64
		t = b

		// for t != 1
		for !t.IsOne() {
			t.Square(&t)
			m++
		}

		if m == 0 {
			return z.Set(&y)
		}
		// t = g^(2^(r-m-1)) (mod q)
		ge := int(r - m - 1)
		t = g
		for ge > 0 {
			t.Square(&t)
			ge--
		}

		g.Square(&t)
		y.Mul(&y, &t)
		b.Mul(&b, &g)
		r = m
	}
}

// Inverse z = x⁻¹ (mod q)
//
// note: allocates a big.Int (math/big)
func (z *Element) Inverse(x *Element) *Element {
	var _xNonMont big.Int
	x.BigInt(&_xNonMont)
	_xNonMont.ModInverse(&_xNonMont, Modulus())
	z.SetBigInt(&_xNonMont)
	return z
}
