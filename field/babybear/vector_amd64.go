//go:build !purego

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

package babybear

//go:noescape
func addVec(res, a, b *Element, n uint64)

//go:noescape
func subVec(res, a, b *Element, n uint64)

//go:noescape
func sumVec(t *uint64, a *Element, n uint64)

//go:noescape
func mulVec(res, a, b *Element, n uint64)

//go:noescape
func scalarMulVec(res, a, b *Element, n uint64)

//go:noescape
func innerProdVec(t *uint64, a, b *Element, n uint64)

// Add adds two vectors element-wise and stores the result in self.
// It panics if the vectors don't have the same length.
func (vector *Vector) Add(a, b Vector) {
	if len(a) != len(b) || len(a) != len(*vector) {
		panic("vector.Add: vectors don't have the same length")
	}
	n := uint64(len(a))
	if n == 0 {
		return
	}
	if !supportAvx512 {
		// call addVecGeneric
		addVecGeneric(*vector, a, b)
		return
	}

	const blockSize = 16
	addVec(&(*vector)[0], &a[0], &b[0], n/blockSize)
	if n%blockSize != 0 {
		// call addVecGeneric on the rest
		start := n - n%blockSize
		addVecGeneric((*vector)[start:], a[start:], b[start:])
	}
}

// Sub subtracts two vectors element-wise and stores the result in self.
// It panics if the vectors don't have the same length.
func (vector *Vector) Sub(a, b Vector) {
	if len(a) != len(b) || len(a) != len(*vector) {
		panic("vector.Sub: vectors don't have the same length")
	}
	n := uint64(len(a))
	if n == 0 {
		return
	}
	if !supportAvx512 {
		// call subVecGeneric
		subVecGeneric(*vector, a, b)
		return
	}

	const blockSize = 16
	subVec(&(*vector)[0], &a[0], &b[0], n/blockSize)
	if n%blockSize != 0 {
		// call subVecGeneric on the rest
		start := n - n%blockSize
		subVecGeneric((*vector)[start:], a[start:], b[start:])
	}
}

// ScalarMul multiplies a vector by a scalar element-wise and stores the result in self.
// It panics if the vectors don't have the same length.
func (vector *Vector) ScalarMul(a Vector, b *Element) {
	if len(a) != len(*vector) {
		panic("vector.ScalarMul: vectors don't have the same length")
	}
	n := uint64(len(a))
	if n == 0 {
		return
	}
	if !supportAvx512 {
		// call scalarMulVecGeneric
		scalarMulVecGeneric(*vector, a, b)
		return
	}

	const blockSize = 8
	scalarMulVec(&(*vector)[0], &a[0], b, n/blockSize)
	if n%blockSize != 0 {
		// call scalarMulVecGeneric on the rest
		start := n - n%blockSize
		scalarMulVecGeneric((*vector)[start:], a[start:], b)
	}
}

// Sum computes the sum of all elements in the vector.
func (vector *Vector) Sum() (res Element) {
	n := uint64(len(*vector))
	if n == 0 {
		return
	}
	if !supportAvx512 {
		// call sumVecGeneric
		sumVecGeneric(&res, *vector)
		return
	}

	const blockSize = 16
	var t [8]uint64 // stores the accumulators (not reduced mod q)
	sumVec(&t[0], &(*vector)[0], n/blockSize)
	// we reduce the accumulators mod q and add to res
	var v Element
	for i := 0; i < 8; i++ {
		t[i] %= q
		v[0] = uint32(t[i])
		res.Add(&res, &v)
	}
	if n%blockSize != 0 {
		// call sumVecGeneric on the rest
		start := n - n%blockSize
		sumVecGeneric(&res, (*vector)[start:])
	}

	return
}

// InnerProduct computes the inner product of two vectors.
// It panics if the vectors don't have the same length.
func (vector *Vector) InnerProduct(other Vector) (res Element) {
	n := uint64(len(*vector))
	if n == 0 {
		return
	}
	if n != uint64(len(other)) {
		panic("vector.InnerProduct: vectors don't have the same length")
	}
	if !supportAvx512 {
		// call innerProductVecGeneric
		innerProductVecGeneric(&res, *vector, other)
		return
	}

	const blockSize = 8
	var t [8]uint64 // stores the accumulators (not reduced mod q)
	innerProdVec(&t[0], &(*vector)[0], &other[0], n/blockSize)
	// we reduce the accumulators mod q and add to res
	var v Element
	for i := 0; i < 8; i++ {
		t[i] %= q
		v[0] = uint32(t[i])
		res.Add(&res, &v)
	}
	if n%blockSize != 0 {
		// call innerProductVecGeneric on the rest
		start := n - n%blockSize
		innerProductVecGeneric(&res, (*vector)[start:], other[start:])
	}

	return
}

// Mul multiplies two vectors element-wise and stores the result in self.
// It panics if the vectors don't have the same length.
func (vector *Vector) Mul(a, b Vector) {
	if len(a) != len(b) || len(a) != len(*vector) {
		panic("vector.Mul: vectors don't have the same length")
	}
	n := uint64(len(a))
	if n == 0 {
		return
	}
	if !supportAvx512 {
		// call mulVecGeneric
		mulVecGeneric(*vector, a, b)
		return
	}

	const blockSize = 8
	mulVec(&(*vector)[0], &a[0], &b[0], n/blockSize)
	if n%blockSize != 0 {
		// call mulVecGeneric on the rest
		start := n - n%blockSize
		mulVecGeneric((*vector)[start:], a[start:], b[start:])
	}
}
