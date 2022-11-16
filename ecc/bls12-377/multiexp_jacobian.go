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

package bls12377

func processChunkG1Jacobian[B ibg1JacExtended](chunk uint64,
	chRes chan<- g1JacExtended,
	c uint64,
	points []G1Affine,
	digits []uint16) {

	var buckets B
	for i := 0; i < len(buckets); i++ {
		buckets[i].setInfinity()
	}

	// for each scalars, get the digit corresponding to the chunk we're processing.
	for i, digit := range digits {
		if digit == 0 {
			continue
		}

		// if msbWindow bit is set, we need to substract
		if digit&1 == 0 {
			// add
			buckets[(digit>>1)-1].addMixed(&points[i])
		} else {
			// sub
			buckets[(digit >> 1)].subMixed(&points[i])
		}
	}

	// reduce buckets into total
	// total =  bucket[0] + 2*bucket[1] + 3*bucket[2] ... + n*bucket[n-1]

	var runningSum, total g1JacExtended
	runningSum.setInfinity()
	total.setInfinity()
	for k := len(buckets) - 1; k >= 0; k-- {
		if !buckets[k].ZZ.IsZero() {
			runningSum.add(&buckets[k])
		}
		total.add(&runningSum)
	}

	chRes <- total
}

// we declare the buckets as fixed-size array types
// this allow us to allocate the buckets on the stack
type bucketg1JacExtendedC1 [1 << (1 - 1)]g1JacExtended
type bucketg1JacExtendedC2 [1 << (2 - 1)]g1JacExtended
type bucketg1JacExtendedC4 [1 << (4 - 1)]g1JacExtended
type bucketg1JacExtendedC5 [1 << (5 - 1)]g1JacExtended
type bucketg1JacExtendedC6 [1 << (6 - 1)]g1JacExtended
type bucketg1JacExtendedC7 [1 << (7 - 1)]g1JacExtended
type bucketg1JacExtendedC8 [1 << (8 - 1)]g1JacExtended
type bucketg1JacExtendedC9 [1 << (9 - 1)]g1JacExtended
type bucketg1JacExtendedC10 [1 << (10 - 1)]g1JacExtended
type bucketg1JacExtendedC11 [1 << (11 - 1)]g1JacExtended
type bucketg1JacExtendedC12 [1 << (12 - 1)]g1JacExtended
type bucketg1JacExtendedC13 [1 << (13 - 1)]g1JacExtended
type bucketg1JacExtendedC14 [1 << (14 - 1)]g1JacExtended
type bucketg1JacExtendedC15 [1 << (15 - 1)]g1JacExtended
type bucketg1JacExtendedC16 [1 << (16 - 1)]g1JacExtended

type ibg1JacExtended interface {
	bucketg1JacExtendedC1 |
		bucketg1JacExtendedC2 |
		bucketg1JacExtendedC4 |
		bucketg1JacExtendedC5 |
		bucketg1JacExtendedC6 |
		bucketg1JacExtendedC7 |
		bucketg1JacExtendedC8 |
		bucketg1JacExtendedC9 |
		bucketg1JacExtendedC10 |
		bucketg1JacExtendedC11 |
		bucketg1JacExtendedC12 |
		bucketg1JacExtendedC13 |
		bucketg1JacExtendedC14 |
		bucketg1JacExtendedC15 |
		bucketg1JacExtendedC16
}

func processChunkG2Jacobian[B ibg2JacExtended](chunk uint64,
	chRes chan<- g2JacExtended,
	c uint64,
	points []G2Affine,
	digits []uint16) {

	var buckets B
	for i := 0; i < len(buckets); i++ {
		buckets[i].setInfinity()
	}

	// for each scalars, get the digit corresponding to the chunk we're processing.
	for i, digit := range digits {
		if digit == 0 {
			continue
		}

		// if msbWindow bit is set, we need to substract
		if digit&1 == 0 {
			// add
			buckets[(digit>>1)-1].addMixed(&points[i])
		} else {
			// sub
			buckets[(digit >> 1)].subMixed(&points[i])
		}
	}

	// reduce buckets into total
	// total =  bucket[0] + 2*bucket[1] + 3*bucket[2] ... + n*bucket[n-1]

	var runningSum, total g2JacExtended
	runningSum.setInfinity()
	total.setInfinity()
	for k := len(buckets) - 1; k >= 0; k-- {
		if !buckets[k].ZZ.IsZero() {
			runningSum.add(&buckets[k])
		}
		total.add(&runningSum)
	}

	chRes <- total
}

// we declare the buckets as fixed-size array types
// this allow us to allocate the buckets on the stack
type bucketg2JacExtendedC1 [1 << (1 - 1)]g2JacExtended
type bucketg2JacExtendedC2 [1 << (2 - 1)]g2JacExtended
type bucketg2JacExtendedC4 [1 << (4 - 1)]g2JacExtended
type bucketg2JacExtendedC5 [1 << (5 - 1)]g2JacExtended
type bucketg2JacExtendedC6 [1 << (6 - 1)]g2JacExtended
type bucketg2JacExtendedC7 [1 << (7 - 1)]g2JacExtended
type bucketg2JacExtendedC8 [1 << (8 - 1)]g2JacExtended
type bucketg2JacExtendedC9 [1 << (9 - 1)]g2JacExtended
type bucketg2JacExtendedC10 [1 << (10 - 1)]g2JacExtended
type bucketg2JacExtendedC11 [1 << (11 - 1)]g2JacExtended
type bucketg2JacExtendedC12 [1 << (12 - 1)]g2JacExtended
type bucketg2JacExtendedC13 [1 << (13 - 1)]g2JacExtended
type bucketg2JacExtendedC14 [1 << (14 - 1)]g2JacExtended
type bucketg2JacExtendedC15 [1 << (15 - 1)]g2JacExtended
type bucketg2JacExtendedC16 [1 << (16 - 1)]g2JacExtended

type ibg2JacExtended interface {
	bucketg2JacExtendedC1 |
		bucketg2JacExtendedC2 |
		bucketg2JacExtendedC4 |
		bucketg2JacExtendedC5 |
		bucketg2JacExtendedC6 |
		bucketg2JacExtendedC7 |
		bucketg2JacExtendedC8 |
		bucketg2JacExtendedC9 |
		bucketg2JacExtendedC10 |
		bucketg2JacExtendedC11 |
		bucketg2JacExtendedC12 |
		bucketg2JacExtendedC13 |
		bucketg2JacExtendedC14 |
		bucketg2JacExtendedC15 |
		bucketg2JacExtendedC16
}
