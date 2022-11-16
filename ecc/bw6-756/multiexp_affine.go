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

package bw6756

import (
	"github.com/consensys/gnark-crypto/ecc/bw6-756/fp"
)

type batchOpG1Affine struct {
	bucketID uint16
	point    G1Affine
}

// processChunkG1BatchAffine process a chunk of the scalars during the msm
// using affine coordinates for the buckets. To amortize the cost of the inverse in the affine addition
// we use a batch affine addition.
//
// this is derived from a PR by 0x0ece : https://github.com/ConsenSys/gnark-crypto/pull/249
// See Section 5.3: ia.cr/2022/1396
func processChunkG1BatchAffine[BJE ibg1JacExtended, B ibG1Affine, BS bitSet, TP pG1Affine, TPP ppG1Affine, TQ qOpsG1Affine, TC cG1Affine](
	chunk uint64,
	chRes chan<- g1JacExtended,
	c uint64,
	points []G1Affine,
	digits []uint16) {

	// the batch affine addition needs independent points; in other words, for a window of batchSize
	// we want to hit independent bucketIDs when processing the digit. if there is a conflict (we're trying
	// to add 2 different points to the same bucket), then we push the conflicted point to a queue.
	// each time the batch is full, we execute it, and tentatively put the points (if not conflict)
	// from the top of the queue into the next batch.
	// if the queue is full, we "flush it"; we sequentially add the points to the buckets in
	// g1JacExtended coordinates.
	// The reasoning behind this is the following; batchSize is chosen such as, for a uniformly random
	// input, the number of conflicts is going to be low, and the element added to the queue should be immediatly
	// processed in the next batch. If it's not the case, then our inputs are not random; and we fallback to
	// non-batch-affine version.

	// note that we have 2 sets of buckets
	// 1 in G1Affine used with the batch affine additions
	// 1 in g1JacExtended used in case the queue of conflicting points
	var buckets B
	var bucketsJE BJE
	for i := 0; i < len(buckets); i++ {
		buckets[i].setInfinity()
		bucketsJE[i].setInfinity()
	}

	// setup for the batch affine;
	var (
		bucketIds BS  // bitSet to signify presence of a bucket in current batch
		cptAdd    int // count the number of bucket + point added to current batch
		R         TPP // bucket references
		P         TP  // points to be added to R (buckets); it is beneficial to store them on the stack (ie copy)
		queue     TQ  // queue of points that conflict the current batch
		qID       int // current position in queue
	)

	batchSize := len(P)

	isFull := func() bool { return cptAdd == batchSize }

	executeAndReset := func() {
		batchAddG1Affine[TP, TPP, TC](&R, &P, cptAdd)
		var tmp BS
		bucketIds = tmp
		cptAdd = 0
	}

	addFromQueue := func(op batchOpG1Affine) {
		// @precondition: must ensures bucket is not "used" in current batch
		// note that there is a bit of duplicate logic between add and addFromQueue
		// the reason is that as of Go 1.19.3, if we pass a pointer to the queue item (see add signature)
		// the compiler will put the queue on the heap.
		BK := &buckets[op.bucketID]

		// handle special cases with inf or -P / P
		if BK.IsInfinity() {
			BK.Set(&op.point)
			return
		}
		if BK.X.Equal(&op.point.X) {
			if BK.Y.Equal(&op.point.Y) {
				// P + P: doubling, which should be quite rare --
				// TODO FIXME @gbotrel / @yelhousni this path is not taken by our tests.
				// need doubling in affine implemented ?
				BK.Add(BK, BK)
				return
			}
			BK.setInfinity()
			return
		}

		bucketIds[op.bucketID] = true
		R[cptAdd] = BK
		P[cptAdd] = op.point
		cptAdd++
	}

	add := func(bucketID uint16, PP *G1Affine, isAdd bool) {
		// @precondition: ensures bucket is not "used" in current batch
		BK := &buckets[bucketID]
		// handle special cases with inf or -P / P
		if BK.IsInfinity() {
			if isAdd {
				BK.Set(PP)
			} else {
				BK.Neg(PP)
			}
			return
		}
		if BK.X.Equal(&PP.X) {
			if BK.Y.Equal(&PP.Y) {
				// P + P: doubling, which should be quite rare --
				// TODO FIXME @gbotrel / @yelhousni this path is not taken by our tests.
				// need doubling in affine implemented ?
				if isAdd {
					BK.Add(BK, BK)
				} else {
					BK.setInfinity()
				}
				return
			}
			if isAdd {
				BK.setInfinity()
			} else {
				BK.Add(BK, BK)
			}
			return
		}

		bucketIds[bucketID] = true
		R[cptAdd] = BK
		if isAdd {
			P[cptAdd].Set(PP)
		} else {
			P[cptAdd].Neg(PP)
		}
		cptAdd++
	}

	flushQueue := func() {
		for i := 0; i < qID; i++ {
			bucketsJE[queue[i].bucketID].addMixed(&queue[i].point)
		}
		qID = 0
	}

	processTopQueue := func() {
		for i := qID - 1; i >= 0; i-- {
			if bucketIds[queue[i].bucketID] {
				return
			}
			addFromQueue(queue[i])
			// len(queue) < batchSize so no need to check for full batch.
			qID--
		}
	}

	for i, digit := range digits {

		if digit == 0 || points[i].IsInfinity() {
			continue
		}

		bucketID := uint16((digit >> 1))
		isAdd := digit&1 == 0
		if isAdd {
			// add
			bucketID -= 1
		}

		if bucketIds[bucketID] {
			// put it in queue
			queue[qID].bucketID = bucketID
			if isAdd {
				queue[qID].point.Set(&points[i])
			} else {
				queue[qID].point.Neg(&points[i])
			}
			qID++

			// queue is full, flush it.
			if qID == len(queue)-1 {
				flushQueue()
			}
			continue
		}

		// we add the point to the batch.
		add(bucketID, &points[i], isAdd)
		if isFull() {
			executeAndReset()
			processTopQueue()
		}
	}

	// flush items in batch.
	executeAndReset()

	// empty the queue
	flushQueue()

	// reduce buckets into total
	// total =  bucket[0] + 2*bucket[1] + 3*bucket[2] ... + n*bucket[n-1]
	var runningSum, total g1JacExtended
	runningSum.setInfinity()
	total.setInfinity()
	for k := len(buckets) - 1; k >= 0; k-- {
		runningSum.addMixed(&buckets[k])
		if !bucketsJE[k].ZZ.IsZero() {
			runningSum.add(&bucketsJE[k])
		}
		total.add(&runningSum)
	}

	chRes <- total

}

// we declare the buckets as fixed-size array types
// this allow us to allocate the buckets on the stack
type bucketG1AffineC11 [1 << (11 - 1)]G1Affine
type bucketG1AffineC16 [1 << (16 - 1)]G1Affine

// buckets: array of G1Affine points of size 1 << (c-1)
type ibG1Affine interface {
	bucketG1AffineC11 |
		bucketG1AffineC16
}

// array of coordinates fp.Element
type cG1Affine interface {
	cG1AffineC11 |
		cG1AffineC16
}

// buckets: array of G1Affine points (for the batch addition)
type pG1Affine interface {
	pG1AffineC11 |
		pG1AffineC16
}

// buckets: array of *G1Affine points (for the batch addition)
type ppG1Affine interface {
	ppG1AffineC11 |
		ppG1AffineC16
}

// buckets: array of G1Affine queue operations (for the batch addition)
type qOpsG1Affine interface {
	qG1AffineC11 |
		qG1AffineC16
}

// batch size 150 when c = 11
type cG1AffineC11 [150]fp.Element
type pG1AffineC11 [150]G1Affine
type ppG1AffineC11 [150]*G1Affine
type qG1AffineC11 [150]batchOpG1Affine

// batch size 640 when c = 16
type cG1AffineC16 [640]fp.Element
type pG1AffineC16 [640]G1Affine
type ppG1AffineC16 [640]*G1Affine
type qG1AffineC16 [640]batchOpG1Affine

type batchOpG2Affine struct {
	bucketID uint16
	point    G2Affine
}

// processChunkG2BatchAffine process a chunk of the scalars during the msm
// using affine coordinates for the buckets. To amortize the cost of the inverse in the affine addition
// we use a batch affine addition.
//
// this is derived from a PR by 0x0ece : https://github.com/ConsenSys/gnark-crypto/pull/249
// See Section 5.3: ia.cr/2022/1396
func processChunkG2BatchAffine[BJE ibg2JacExtended, B ibG2Affine, BS bitSet, TP pG2Affine, TPP ppG2Affine, TQ qOpsG2Affine, TC cG2Affine](
	chunk uint64,
	chRes chan<- g2JacExtended,
	c uint64,
	points []G2Affine,
	digits []uint16) {

	// the batch affine addition needs independent points; in other words, for a window of batchSize
	// we want to hit independent bucketIDs when processing the digit. if there is a conflict (we're trying
	// to add 2 different points to the same bucket), then we push the conflicted point to a queue.
	// each time the batch is full, we execute it, and tentatively put the points (if not conflict)
	// from the top of the queue into the next batch.
	// if the queue is full, we "flush it"; we sequentially add the points to the buckets in
	// g2JacExtended coordinates.
	// The reasoning behind this is the following; batchSize is chosen such as, for a uniformly random
	// input, the number of conflicts is going to be low, and the element added to the queue should be immediatly
	// processed in the next batch. If it's not the case, then our inputs are not random; and we fallback to
	// non-batch-affine version.

	// note that we have 2 sets of buckets
	// 1 in G2Affine used with the batch affine additions
	// 1 in g2JacExtended used in case the queue of conflicting points
	var buckets B
	var bucketsJE BJE
	for i := 0; i < len(buckets); i++ {
		buckets[i].setInfinity()
		bucketsJE[i].setInfinity()
	}

	// setup for the batch affine;
	var (
		bucketIds BS  // bitSet to signify presence of a bucket in current batch
		cptAdd    int // count the number of bucket + point added to current batch
		R         TPP // bucket references
		P         TP  // points to be added to R (buckets); it is beneficial to store them on the stack (ie copy)
		queue     TQ  // queue of points that conflict the current batch
		qID       int // current position in queue
	)

	batchSize := len(P)

	isFull := func() bool { return cptAdd == batchSize }

	executeAndReset := func() {
		batchAddG2Affine[TP, TPP, TC](&R, &P, cptAdd)
		var tmp BS
		bucketIds = tmp
		cptAdd = 0
	}

	addFromQueue := func(op batchOpG2Affine) {
		// @precondition: must ensures bucket is not "used" in current batch
		// note that there is a bit of duplicate logic between add and addFromQueue
		// the reason is that as of Go 1.19.3, if we pass a pointer to the queue item (see add signature)
		// the compiler will put the queue on the heap.
		BK := &buckets[op.bucketID]

		// handle special cases with inf or -P / P
		if BK.IsInfinity() {
			BK.Set(&op.point)
			return
		}
		if BK.X.Equal(&op.point.X) {
			if BK.Y.Equal(&op.point.Y) {
				// P + P: doubling, which should be quite rare --
				// TODO FIXME @gbotrel / @yelhousni this path is not taken by our tests.
				// need doubling in affine implemented ?
				BK.Add(BK, BK)
				return
			}
			BK.setInfinity()
			return
		}

		bucketIds[op.bucketID] = true
		R[cptAdd] = BK
		P[cptAdd] = op.point
		cptAdd++
	}

	add := func(bucketID uint16, PP *G2Affine, isAdd bool) {
		// @precondition: ensures bucket is not "used" in current batch
		BK := &buckets[bucketID]
		// handle special cases with inf or -P / P
		if BK.IsInfinity() {
			if isAdd {
				BK.Set(PP)
			} else {
				BK.Neg(PP)
			}
			return
		}
		if BK.X.Equal(&PP.X) {
			if BK.Y.Equal(&PP.Y) {
				// P + P: doubling, which should be quite rare --
				// TODO FIXME @gbotrel / @yelhousni this path is not taken by our tests.
				// need doubling in affine implemented ?
				if isAdd {
					BK.Add(BK, BK)
				} else {
					BK.setInfinity()
				}
				return
			}
			if isAdd {
				BK.setInfinity()
			} else {
				BK.Add(BK, BK)
			}
			return
		}

		bucketIds[bucketID] = true
		R[cptAdd] = BK
		if isAdd {
			P[cptAdd].Set(PP)
		} else {
			P[cptAdd].Neg(PP)
		}
		cptAdd++
	}

	flushQueue := func() {
		for i := 0; i < qID; i++ {
			bucketsJE[queue[i].bucketID].addMixed(&queue[i].point)
		}
		qID = 0
	}

	processTopQueue := func() {
		for i := qID - 1; i >= 0; i-- {
			if bucketIds[queue[i].bucketID] {
				return
			}
			addFromQueue(queue[i])
			// len(queue) < batchSize so no need to check for full batch.
			qID--
		}
	}

	for i, digit := range digits {

		if digit == 0 || points[i].IsInfinity() {
			continue
		}

		bucketID := uint16((digit >> 1))
		isAdd := digit&1 == 0
		if isAdd {
			// add
			bucketID -= 1
		}

		if bucketIds[bucketID] {
			// put it in queue
			queue[qID].bucketID = bucketID
			if isAdd {
				queue[qID].point.Set(&points[i])
			} else {
				queue[qID].point.Neg(&points[i])
			}
			qID++

			// queue is full, flush it.
			if qID == len(queue)-1 {
				flushQueue()
			}
			continue
		}

		// we add the point to the batch.
		add(bucketID, &points[i], isAdd)
		if isFull() {
			executeAndReset()
			processTopQueue()
		}
	}

	// flush items in batch.
	executeAndReset()

	// empty the queue
	flushQueue()

	// reduce buckets into total
	// total =  bucket[0] + 2*bucket[1] + 3*bucket[2] ... + n*bucket[n-1]
	var runningSum, total g2JacExtended
	runningSum.setInfinity()
	total.setInfinity()
	for k := len(buckets) - 1; k >= 0; k-- {
		runningSum.addMixed(&buckets[k])
		if !bucketsJE[k].ZZ.IsZero() {
			runningSum.add(&bucketsJE[k])
		}
		total.add(&runningSum)
	}

	chRes <- total

}

// we declare the buckets as fixed-size array types
// this allow us to allocate the buckets on the stack
type bucketG2AffineC11 [1 << (11 - 1)]G2Affine
type bucketG2AffineC16 [1 << (16 - 1)]G2Affine

// buckets: array of G2Affine points of size 1 << (c-1)
type ibG2Affine interface {
	bucketG2AffineC11 |
		bucketG2AffineC16
}

// array of coordinates fp.Element
type cG2Affine interface {
	cG2AffineC11 |
		cG2AffineC16
}

// buckets: array of G2Affine points (for the batch addition)
type pG2Affine interface {
	pG2AffineC11 |
		pG2AffineC16
}

// buckets: array of *G2Affine points (for the batch addition)
type ppG2Affine interface {
	ppG2AffineC11 |
		ppG2AffineC16
}

// buckets: array of G2Affine queue operations (for the batch addition)
type qOpsG2Affine interface {
	qG2AffineC11 |
		qG2AffineC16
}

// batch size 150 when c = 11
type cG2AffineC11 [150]fp.Element
type pG2AffineC11 [150]G2Affine
type ppG2AffineC11 [150]*G2Affine
type qG2AffineC11 [150]batchOpG2Affine

// batch size 640 when c = 16
type cG2AffineC16 [640]fp.Element
type pG2AffineC16 [640]G2Affine
type ppG2AffineC16 [640]*G2Affine
type qG2AffineC16 [640]batchOpG2Affine

type bitSetC3 [1 << (3 - 1)]bool
type bitSetC4 [1 << (4 - 1)]bool
type bitSetC5 [1 << (5 - 1)]bool
type bitSetC8 [1 << (8 - 1)]bool
type bitSetC11 [1 << (11 - 1)]bool
type bitSetC16 [1 << (16 - 1)]bool

type bitSet interface {
	bitSetC3 |
		bitSetC4 |
		bitSetC5 |
		bitSetC8 |
		bitSetC11 |
		bitSetC16
}
