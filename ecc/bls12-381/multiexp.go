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

package bls12381

import (
	"errors"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/gnark-crypto/internal/parallel"
	"math"
	"runtime"
)

// MultiExp implements section 4 of https://eprint.iacr.org/2012/549.pdf
//
// This call return an error if len(scalars) != len(points) or if provided config is invalid.
func (p *G1Affine) MultiExp(points []G1Affine, scalars []fr.Element, config ecc.MultiExpConfig) (*G1Affine, error) {
	var _p G1Jac
	if _, err := _p.MultiExp(points, scalars, config); err != nil {
		return nil, err
	}
	p.FromJacobian(&_p)
	return p, nil
}

// MultiExp implements section 4 of https://eprint.iacr.org/2012/549.pdf
//
// This call return an error if len(scalars) != len(points) or if provided config is invalid.
func (p *G1Jac) MultiExp(points []G1Affine, scalars []fr.Element, config ecc.MultiExpConfig) (*G1Jac, error) {
	// TODO @gbotrel replace the ecc.MultiExpConfig by a Option pattern for maintainability.
	// note:
	// each of the msmCX method is the same, except for the c constant it declares
	// duplicating (through template generation) these methods allows to declare the buckets on the stack
	// the choice of c needs to be improved:
	// there is a theoretical value that gives optimal asymptotics
	// but in practice, other factors come into play, including:
	// * if c doesn't divide 64, the word size, then we're bound to select bits over 2 words of our scalars, instead of 1
	// * number of CPUs
	// * cache friendliness (which depends on the host, G1 or G2... )
	//	--> for example, on BN254, a G1 point fits into one cache line of 64bytes, but a G2 point don't.

	// for each msmCX
	// step 1
	// we compute, for each scalars over c-bit wide windows, nbChunk digits
	// if the digit is larger than 2^{c-1}, then, we borrow 2^c from the next window and subtract
	// 2^{c} to the current digit, making it negative.
	// negative digits will be processed in the next step as adding -G into the bucket instead of G
	// (computing -G is cheap, and this saves us half of the buckets)
	// step 2
	// buckets are declared on the stack
	// notice that we have 2^{c-1} buckets instead of 2^{c} (see step1)
	// we use jacobian extended formulas here as they are faster than mixed addition
	// msmProcessChunk places points into buckets base on their selector and return the weighted bucket sum in given channel
	// step 3
	// reduce the buckets weighed sums into our result (msmReduceChunk)

	// ensure len(points) == len(scalars)
	nbPoints := len(points)
	if nbPoints != len(scalars) {
		return nil, errors.New("len(points) != len(scalars)")
	}

	// if nbTasks is not set, use all available CPUs
	if config.NbTasks <= 0 {
		config.NbTasks = runtime.NumCPU() * 3
	} else if config.NbTasks > 1024 {
		return nil, errors.New("invalid config: config.NbTasks > 1024")
	}

	// here, we compute the best C for nbPoints
	// we split recursively until nbChunks(c) >= nbTasks,
	bestC := func(nbPoints int) uint64 {
		// implemented msmC methods (the c we use must be in this slice)
		implementedCs := []uint64{4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
		var C uint64
		// approximate cost (in group operations)
		// cost = bits/c * (nbPoints + 2^{c})
		// this needs to be verified empirically.
		// for example, on a MBP 2016, for G2 MultiExp > 8M points, hand picking c gives better results
		min := math.MaxFloat64
		for _, c := range implementedCs {
			cc := (fr.Bits + 1) * (nbPoints + (1 << c))
			cost := float64(cc) / float64(c)
			if cost < min {
				min = cost
				C = c
			}
		}
		return C
	}

	C := bestC(nbPoints)
	nbChunks := int(computeNbChunks(C))

	// should we recursively split the msm in half? (see below)
	// we want to minimize the execution time of the algorithm;
	// splitting the msm will **add** operations, but if it allows to use more CPU, it might be worth it.

	// costFunction returns a metric that represent the "wall time" of the algorithm
	costFunction := func(nbTasks, nbCpus, costPerTask int) int {
		// cost for the reduction of all tasks (msmReduceChunk)
		totalCost := nbTasks

		// cost for the computation of each task (msmProcessChunk)
		for nbTasks >= nbCpus {
			nbTasks -= nbCpus
			totalCost += costPerTask
		}
		if nbTasks > 0 {
			totalCost += costPerTask
		}
		return totalCost
	}

	// costPerTask is the approximate number of group ops per task
	costPerTask := func(c uint64, nbPoints int) int { return (nbPoints + int((1 << c))) }

	costPreSplit := costFunction(nbChunks, config.NbTasks, costPerTask(C, nbPoints))

	cPostSplit := bestC(nbPoints / 2)
	nbChunksPostSplit := int(computeNbChunks(cPostSplit))
	costPostSplit := costFunction(nbChunksPostSplit*2, config.NbTasks, costPerTask(cPostSplit, nbPoints/2))

	// if the cost of the split msm is lower than the cost of the non split msm, we split
	if costPostSplit < costPreSplit {
		config.NbTasks = int(math.Ceil(float64(config.NbTasks) / 2.0))
		var _p G1Jac
		chDone := make(chan struct{}, 1)
		go func() {
			_p.MultiExp(points[:nbPoints/2], scalars[:nbPoints/2], config)
			close(chDone)
		}()
		p.MultiExp(points[nbPoints/2:], scalars[nbPoints/2:], config)
		<-chDone
		p.AddAssign(&_p)
		return p, nil
	}

	// if we don't split, we use the best C we found
	_innerMsmG1(p, C, points, scalars, config)

	return p, nil
}

func _innerMsmG1(p *G1Jac, c uint64, points []G1Affine, scalars []fr.Element, config ecc.MultiExpConfig) *G1Jac {
	// partition the scalars
	digits, chunkStats := partitionScalars(scalars, c, config.NbTasks)

	nbChunks := computeNbChunks(c)

	// for each chunk, spawn one go routine that'll loop through all the scalars in the
	// corresponding bit-window
	// note that buckets is an array allocated on the stack and this is critical for performance

	// each go routine sends its result in chChunks[i] channel
	chChunks := make([]chan g1JacExtended, nbChunks)
	for i := 0; i < len(chChunks); i++ {
		chChunks[i] = make(chan g1JacExtended, 1)
	}

	// we use a semaphore to limit the number of go routines running concurrently
	// (only if nbTasks < nbCPU)
	var sem chan struct{}
	if config.NbTasks < runtime.NumCPU() {
		sem = make(chan struct{}, config.NbTasks*2) // *2 because if chunk is overweight we split it in two
		for i := 0; i < config.NbTasks; i++ {
			sem <- struct{}{}
		}
	}

	// the last chunk may be processed with a different method than the rest, as it could be smaller.
	n := len(points)
	for j := int(nbChunks - 1); j >= 0; j-- {
		processChunk := getChunkProcessorG1(c, chunkStats[j])
		if j == int(nbChunks-1) {
			processChunk = getChunkProcessorG1(lastC(c), chunkStats[j])
		}
		if chunkStats[j].weight >= 115 {
			// we split this in more go routines since this chunk has more work to do than the others.
			// else what would happen is this go routine would finish much later than the others.
			chSplit := make(chan g1JacExtended, 2)
			split := n / 2

			if sem != nil {
				sem <- struct{}{} // add another token to the semaphore, since we split in two.
			}
			go processChunk(uint64(j), chSplit, c, points[:split], digits[j*n:(j*n)+split], sem)
			go processChunk(uint64(j), chSplit, c, points[split:], digits[(j*n)+split:(j+1)*n], sem)
			go func(chunkID int) {
				s1 := <-chSplit
				s2 := <-chSplit
				close(chSplit)
				s1.add(&s2)
				chChunks[chunkID] <- s1
			}(j)
			continue
		}
		go processChunk(uint64(j), chChunks[j], c, points, digits[j*n:(j+1)*n], sem)
	}

	return msmReduceChunkG1Affine(p, int(c), chChunks[:])
}

// getChunkProcessorG1 decides, depending on c window size and statistics for the chunk
// to return the best algorithm to process the chunk.
func getChunkProcessorG1(c uint64, stat chunkStat) func(chunkID uint64, chRes chan<- g1JacExtended, c uint64, points []G1Affine, digits []uint16, sem chan struct{}) {
	switch c {

	case 3:
		return processChunkG1Jacobian[bucketg1JacExtendedC3]
	case 4:
		return processChunkG1Jacobian[bucketg1JacExtendedC4]
	case 5:
		return processChunkG1Jacobian[bucketg1JacExtendedC5]
	case 6:
		return processChunkG1Jacobian[bucketg1JacExtendedC6]
	case 7:
		return processChunkG1Jacobian[bucketg1JacExtendedC7]
	case 8:
		return processChunkG1Jacobian[bucketg1JacExtendedC8]
	case 9:
		return processChunkG1Jacobian[bucketg1JacExtendedC9]
	case 10:
		const batchSize = 80
		// here we could check some chunk statistic (deviation, ...) to determine if calling
		// the batch affine version is worth it.
		if stat.nbBucketFilled < batchSize {
			// clear indicator that batch affine method is not appropriate here.
			return processChunkG1Jacobian[bucketg1JacExtendedC10]
		}
		return processChunkG1BatchAffine[bucketg1JacExtendedC10, bucketG1AffineC10, bitSetC10, pG1AffineC10, ppG1AffineC10, qG1AffineC10, cG1AffineC10]
	case 11:
		const batchSize = 150
		// here we could check some chunk statistic (deviation, ...) to determine if calling
		// the batch affine version is worth it.
		if stat.nbBucketFilled < batchSize {
			// clear indicator that batch affine method is not appropriate here.
			return processChunkG1Jacobian[bucketg1JacExtendedC11]
		}
		return processChunkG1BatchAffine[bucketg1JacExtendedC11, bucketG1AffineC11, bitSetC11, pG1AffineC11, ppG1AffineC11, qG1AffineC11, cG1AffineC11]
	case 12:
		const batchSize = 200
		// here we could check some chunk statistic (deviation, ...) to determine if calling
		// the batch affine version is worth it.
		if stat.nbBucketFilled < batchSize {
			// clear indicator that batch affine method is not appropriate here.
			return processChunkG1Jacobian[bucketg1JacExtendedC12]
		}
		return processChunkG1BatchAffine[bucketg1JacExtendedC12, bucketG1AffineC12, bitSetC12, pG1AffineC12, ppG1AffineC12, qG1AffineC12, cG1AffineC12]
	case 13:
		const batchSize = 350
		// here we could check some chunk statistic (deviation, ...) to determine if calling
		// the batch affine version is worth it.
		if stat.nbBucketFilled < batchSize {
			// clear indicator that batch affine method is not appropriate here.
			return processChunkG1Jacobian[bucketg1JacExtendedC13]
		}
		return processChunkG1BatchAffine[bucketg1JacExtendedC13, bucketG1AffineC13, bitSetC13, pG1AffineC13, ppG1AffineC13, qG1AffineC13, cG1AffineC13]
	case 14:
		const batchSize = 400
		// here we could check some chunk statistic (deviation, ...) to determine if calling
		// the batch affine version is worth it.
		if stat.nbBucketFilled < batchSize {
			// clear indicator that batch affine method is not appropriate here.
			return processChunkG1Jacobian[bucketg1JacExtendedC14]
		}
		return processChunkG1BatchAffine[bucketg1JacExtendedC14, bucketG1AffineC14, bitSetC14, pG1AffineC14, ppG1AffineC14, qG1AffineC14, cG1AffineC14]
	case 15:
		const batchSize = 500
		// here we could check some chunk statistic (deviation, ...) to determine if calling
		// the batch affine version is worth it.
		if stat.nbBucketFilled < batchSize {
			// clear indicator that batch affine method is not appropriate here.
			return processChunkG1Jacobian[bucketg1JacExtendedC15]
		}
		return processChunkG1BatchAffine[bucketg1JacExtendedC15, bucketG1AffineC15, bitSetC15, pG1AffineC15, ppG1AffineC15, qG1AffineC15, cG1AffineC15]
	case 16:
		const batchSize = 640
		// here we could check some chunk statistic (deviation, ...) to determine if calling
		// the batch affine version is worth it.
		if stat.nbBucketFilled < batchSize {
			// clear indicator that batch affine method is not appropriate here.
			return processChunkG1Jacobian[bucketg1JacExtendedC16]
		}
		return processChunkG1BatchAffine[bucketg1JacExtendedC16, bucketG1AffineC16, bitSetC16, pG1AffineC16, ppG1AffineC16, qG1AffineC16, cG1AffineC16]
	default:
		// panic("will not happen c != previous values is not generated by templates")
		return processChunkG1Jacobian[bucketg1JacExtendedC16]
	}
}

// msmReduceChunkG1Affine reduces the weighted sum of the buckets into the result of the multiExp
func msmReduceChunkG1Affine(p *G1Jac, c int, chChunks []chan g1JacExtended) *G1Jac {
	var _p g1JacExtended
	totalj := <-chChunks[len(chChunks)-1]
	_p.Set(&totalj)
	for j := len(chChunks) - 2; j >= 0; j-- {
		for l := 0; l < c; l++ {
			_p.double(&_p)
		}
		totalj := <-chChunks[j]
		_p.add(&totalj)
	}

	return p.unsafeFromJacExtended(&_p)
}

// MultiExp implements section 4 of https://eprint.iacr.org/2012/549.pdf
//
// This call return an error if len(scalars) != len(points) or if provided config is invalid.
func (p *G2Affine) MultiExp(points []G2Affine, scalars []fr.Element, config ecc.MultiExpConfig) (*G2Affine, error) {
	var _p G2Jac
	if _, err := _p.MultiExp(points, scalars, config); err != nil {
		return nil, err
	}
	p.FromJacobian(&_p)
	return p, nil
}

// MultiExp implements section 4 of https://eprint.iacr.org/2012/549.pdf
//
// This call return an error if len(scalars) != len(points) or if provided config is invalid.
func (p *G2Jac) MultiExp(points []G2Affine, scalars []fr.Element, config ecc.MultiExpConfig) (*G2Jac, error) {
	// TODO @gbotrel replace the ecc.MultiExpConfig by a Option pattern for maintainability.
	// note:
	// each of the msmCX method is the same, except for the c constant it declares
	// duplicating (through template generation) these methods allows to declare the buckets on the stack
	// the choice of c needs to be improved:
	// there is a theoretical value that gives optimal asymptotics
	// but in practice, other factors come into play, including:
	// * if c doesn't divide 64, the word size, then we're bound to select bits over 2 words of our scalars, instead of 1
	// * number of CPUs
	// * cache friendliness (which depends on the host, G1 or G2... )
	//	--> for example, on BN254, a G1 point fits into one cache line of 64bytes, but a G2 point don't.

	// for each msmCX
	// step 1
	// we compute, for each scalars over c-bit wide windows, nbChunk digits
	// if the digit is larger than 2^{c-1}, then, we borrow 2^c from the next window and subtract
	// 2^{c} to the current digit, making it negative.
	// negative digits will be processed in the next step as adding -G into the bucket instead of G
	// (computing -G is cheap, and this saves us half of the buckets)
	// step 2
	// buckets are declared on the stack
	// notice that we have 2^{c-1} buckets instead of 2^{c} (see step1)
	// we use jacobian extended formulas here as they are faster than mixed addition
	// msmProcessChunk places points into buckets base on their selector and return the weighted bucket sum in given channel
	// step 3
	// reduce the buckets weighed sums into our result (msmReduceChunk)

	// ensure len(points) == len(scalars)
	nbPoints := len(points)
	if nbPoints != len(scalars) {
		return nil, errors.New("len(points) != len(scalars)")
	}

	// if nbTasks is not set, use all available CPUs
	if config.NbTasks <= 0 {
		config.NbTasks = runtime.NumCPU() * 3
	} else if config.NbTasks > 1024 {
		return nil, errors.New("invalid config: config.NbTasks > 1024")
	}

	// here, we compute the best C for nbPoints
	// we split recursively until nbChunks(c) >= nbTasks,
	bestC := func(nbPoints int) uint64 {
		// implemented msmC methods (the c we use must be in this slice)
		implementedCs := []uint64{4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
		var C uint64
		// approximate cost (in group operations)
		// cost = bits/c * (nbPoints + 2^{c})
		// this needs to be verified empirically.
		// for example, on a MBP 2016, for G2 MultiExp > 8M points, hand picking c gives better results
		min := math.MaxFloat64
		for _, c := range implementedCs {
			cc := (fr.Bits + 1) * (nbPoints + (1 << c))
			cost := float64(cc) / float64(c)
			if cost < min {
				min = cost
				C = c
			}
		}
		return C
	}

	C := bestC(nbPoints)
	nbChunks := int(computeNbChunks(C))

	// should we recursively split the msm in half? (see below)
	// we want to minimize the execution time of the algorithm;
	// splitting the msm will **add** operations, but if it allows to use more CPU, it might be worth it.

	// costFunction returns a metric that represent the "wall time" of the algorithm
	costFunction := func(nbTasks, nbCpus, costPerTask int) int {
		// cost for the reduction of all tasks (msmReduceChunk)
		totalCost := nbTasks

		// cost for the computation of each task (msmProcessChunk)
		for nbTasks >= nbCpus {
			nbTasks -= nbCpus
			totalCost += costPerTask
		}
		if nbTasks > 0 {
			totalCost += costPerTask
		}
		return totalCost
	}

	// costPerTask is the approximate number of group ops per task
	costPerTask := func(c uint64, nbPoints int) int { return (nbPoints + int((1 << c))) }

	costPreSplit := costFunction(nbChunks, config.NbTasks, costPerTask(C, nbPoints))

	cPostSplit := bestC(nbPoints / 2)
	nbChunksPostSplit := int(computeNbChunks(cPostSplit))
	costPostSplit := costFunction(nbChunksPostSplit*2, config.NbTasks, costPerTask(cPostSplit, nbPoints/2))

	// if the cost of the split msm is lower than the cost of the non split msm, we split
	if costPostSplit < costPreSplit {
		config.NbTasks = int(math.Ceil(float64(config.NbTasks) / 2.0))
		var _p G2Jac
		chDone := make(chan struct{}, 1)
		go func() {
			_p.MultiExp(points[:nbPoints/2], scalars[:nbPoints/2], config)
			close(chDone)
		}()
		p.MultiExp(points[nbPoints/2:], scalars[nbPoints/2:], config)
		<-chDone
		p.AddAssign(&_p)
		return p, nil
	}

	// if we don't split, we use the best C we found
	_innerMsmG2(p, C, points, scalars, config)

	return p, nil
}

func _innerMsmG2(p *G2Jac, c uint64, points []G2Affine, scalars []fr.Element, config ecc.MultiExpConfig) *G2Jac {
	// partition the scalars
	digits, chunkStats := partitionScalars(scalars, c, config.NbTasks)

	nbChunks := computeNbChunks(c)

	// for each chunk, spawn one go routine that'll loop through all the scalars in the
	// corresponding bit-window
	// note that buckets is an array allocated on the stack and this is critical for performance

	// each go routine sends its result in chChunks[i] channel
	chChunks := make([]chan g2JacExtended, nbChunks)
	for i := 0; i < len(chChunks); i++ {
		chChunks[i] = make(chan g2JacExtended, 1)
	}

	// we use a semaphore to limit the number of go routines running concurrently
	// (only if nbTasks < nbCPU)
	var sem chan struct{}
	if config.NbTasks < runtime.NumCPU() {
		sem = make(chan struct{}, config.NbTasks*2) // *2 because if chunk is overweight we split it in two
		for i := 0; i < config.NbTasks; i++ {
			sem <- struct{}{}
		}
	}

	// the last chunk may be processed with a different method than the rest, as it could be smaller.
	n := len(points)
	for j := int(nbChunks - 1); j >= 0; j-- {
		processChunk := getChunkProcessorG2(c, chunkStats[j])
		if j == int(nbChunks-1) {
			processChunk = getChunkProcessorG2(lastC(c), chunkStats[j])
		}
		if chunkStats[j].weight >= 115 {
			// we split this in more go routines since this chunk has more work to do than the others.
			// else what would happen is this go routine would finish much later than the others.
			chSplit := make(chan g2JacExtended, 2)
			split := n / 2

			if sem != nil {
				sem <- struct{}{} // add another token to the semaphore, since we split in two.
			}
			go processChunk(uint64(j), chSplit, c, points[:split], digits[j*n:(j*n)+split], sem)
			go processChunk(uint64(j), chSplit, c, points[split:], digits[(j*n)+split:(j+1)*n], sem)
			go func(chunkID int) {
				s1 := <-chSplit
				s2 := <-chSplit
				close(chSplit)
				s1.add(&s2)
				chChunks[chunkID] <- s1
			}(j)
			continue
		}
		go processChunk(uint64(j), chChunks[j], c, points, digits[j*n:(j+1)*n], sem)
	}

	return msmReduceChunkG2Affine(p, int(c), chChunks[:])
}

// getChunkProcessorG2 decides, depending on c window size and statistics for the chunk
// to return the best algorithm to process the chunk.
func getChunkProcessorG2(c uint64, stat chunkStat) func(chunkID uint64, chRes chan<- g2JacExtended, c uint64, points []G2Affine, digits []uint16, sem chan struct{}) {
	switch c {

	case 3:
		return processChunkG2Jacobian[bucketg2JacExtendedC3]
	case 4:
		return processChunkG2Jacobian[bucketg2JacExtendedC4]
	case 5:
		return processChunkG2Jacobian[bucketg2JacExtendedC5]
	case 6:
		return processChunkG2Jacobian[bucketg2JacExtendedC6]
	case 7:
		return processChunkG2Jacobian[bucketg2JacExtendedC7]
	case 8:
		return processChunkG2Jacobian[bucketg2JacExtendedC8]
	case 9:
		return processChunkG2Jacobian[bucketg2JacExtendedC9]
	case 10:
		const batchSize = 80
		// here we could check some chunk statistic (deviation, ...) to determine if calling
		// the batch affine version is worth it.
		if stat.nbBucketFilled < batchSize {
			// clear indicator that batch affine method is not appropriate here.
			return processChunkG2Jacobian[bucketg2JacExtendedC10]
		}
		return processChunkG2BatchAffine[bucketg2JacExtendedC10, bucketG2AffineC10, bitSetC10, pG2AffineC10, ppG2AffineC10, qG2AffineC10, cG2AffineC10]
	case 11:
		const batchSize = 150
		// here we could check some chunk statistic (deviation, ...) to determine if calling
		// the batch affine version is worth it.
		if stat.nbBucketFilled < batchSize {
			// clear indicator that batch affine method is not appropriate here.
			return processChunkG2Jacobian[bucketg2JacExtendedC11]
		}
		return processChunkG2BatchAffine[bucketg2JacExtendedC11, bucketG2AffineC11, bitSetC11, pG2AffineC11, ppG2AffineC11, qG2AffineC11, cG2AffineC11]
	case 12:
		const batchSize = 200
		// here we could check some chunk statistic (deviation, ...) to determine if calling
		// the batch affine version is worth it.
		if stat.nbBucketFilled < batchSize {
			// clear indicator that batch affine method is not appropriate here.
			return processChunkG2Jacobian[bucketg2JacExtendedC12]
		}
		return processChunkG2BatchAffine[bucketg2JacExtendedC12, bucketG2AffineC12, bitSetC12, pG2AffineC12, ppG2AffineC12, qG2AffineC12, cG2AffineC12]
	case 13:
		const batchSize = 350
		// here we could check some chunk statistic (deviation, ...) to determine if calling
		// the batch affine version is worth it.
		if stat.nbBucketFilled < batchSize {
			// clear indicator that batch affine method is not appropriate here.
			return processChunkG2Jacobian[bucketg2JacExtendedC13]
		}
		return processChunkG2BatchAffine[bucketg2JacExtendedC13, bucketG2AffineC13, bitSetC13, pG2AffineC13, ppG2AffineC13, qG2AffineC13, cG2AffineC13]
	case 14:
		const batchSize = 400
		// here we could check some chunk statistic (deviation, ...) to determine if calling
		// the batch affine version is worth it.
		if stat.nbBucketFilled < batchSize {
			// clear indicator that batch affine method is not appropriate here.
			return processChunkG2Jacobian[bucketg2JacExtendedC14]
		}
		return processChunkG2BatchAffine[bucketg2JacExtendedC14, bucketG2AffineC14, bitSetC14, pG2AffineC14, ppG2AffineC14, qG2AffineC14, cG2AffineC14]
	case 15:
		const batchSize = 500
		// here we could check some chunk statistic (deviation, ...) to determine if calling
		// the batch affine version is worth it.
		if stat.nbBucketFilled < batchSize {
			// clear indicator that batch affine method is not appropriate here.
			return processChunkG2Jacobian[bucketg2JacExtendedC15]
		}
		return processChunkG2BatchAffine[bucketg2JacExtendedC15, bucketG2AffineC15, bitSetC15, pG2AffineC15, ppG2AffineC15, qG2AffineC15, cG2AffineC15]
	case 16:
		const batchSize = 640
		// here we could check some chunk statistic (deviation, ...) to determine if calling
		// the batch affine version is worth it.
		if stat.nbBucketFilled < batchSize {
			// clear indicator that batch affine method is not appropriate here.
			return processChunkG2Jacobian[bucketg2JacExtendedC16]
		}
		return processChunkG2BatchAffine[bucketg2JacExtendedC16, bucketG2AffineC16, bitSetC16, pG2AffineC16, ppG2AffineC16, qG2AffineC16, cG2AffineC16]
	default:
		// panic("will not happen c != previous values is not generated by templates")
		return processChunkG2Jacobian[bucketg2JacExtendedC16]
	}
}

// msmReduceChunkG2Affine reduces the weighted sum of the buckets into the result of the multiExp
func msmReduceChunkG2Affine(p *G2Jac, c int, chChunks []chan g2JacExtended) *G2Jac {
	var _p g2JacExtended
	totalj := <-chChunks[len(chChunks)-1]
	_p.Set(&totalj)
	for j := len(chChunks) - 2; j >= 0; j-- {
		for l := 0; l < c; l++ {
			_p.double(&_p)
		}
		totalj := <-chChunks[j]
		_p.add(&totalj)
	}

	return p.unsafeFromJacExtended(&_p)
}

// selector stores the index, mask and shifts needed to select bits from a scalar
// it is used during the multiExp algorithm or the batch scalar multiplication
type selector struct {
	index uint64 // index in the multi-word scalar to select bits from
	mask  uint64 // mask (c-bit wide)
	shift uint64 // shift needed to get our bits on low positions

	multiWordSelect bool   // set to true if we need to select bits from 2 words (case where c doesn't divide 64)
	maskHigh        uint64 // same than mask, for index+1
	shiftHigh       uint64 // same than shift, for index+1
}

// return number of chunks for a given window size c
// the last chunk may be bigger to accommodate a potential carry from the NAF decomposition
func computeNbChunks(c uint64) uint64 {
	return (fr.Bits + c - 1) / c
}

// return the last window size for a scalar;
// this last window should accommodate a carry (from the NAF decomposition)
// it can be == c if we have 1 available bit
// it can be > c if we have 0 available bit
// it can be < c if we have 2+ available bits
func lastC(c uint64) uint64 {
	nbAvailableBits := (computeNbChunks(c) * c) - fr.Bits
	return c + 1 - nbAvailableBits
}

type chunkStat struct {
	// relative weight of work compared to other chunks. 100.0 -> nominal weight.
	weight float32

	// percentage of bucket filled in the window;
	ppBucketFilled float32
	nbBucketFilled int
}

// partitionScalars  compute, for each scalars over c-bit wide windows, nbChunk digits
// if the digit is larger than 2^{c-1}, then, we borrow 2^c from the next window and subtract
// 2^{c} to the current digit, making it negative.
// negative digits can be processed in a later step as adding -G into the bucket instead of G
// (computing -G is cheap, and this saves us half of the buckets in the MultiExp or BatchScalarMultiplication)
func partitionScalars(scalars []fr.Element, c uint64, nbTasks int) ([]uint16, []chunkStat) {
	// no benefit here to have more tasks than CPUs
	if nbTasks > runtime.NumCPU() {
		nbTasks = runtime.NumCPU()
	}

	// number of c-bit radixes in a scalar
	nbChunks := computeNbChunks(c)

	digits := make([]uint16, len(scalars)*int(nbChunks))

	mask := uint64((1 << c) - 1) // low c bits are 1
	max := int(1<<(c-1)) - 1     // max value (inclusive) we want for our digits
	cDivides64 := (64 % c) == 0  // if c doesn't divide 64, we may need to select over multiple words

	// compute offset and word selector / shift to select the right bits of our windows
	selectors := make([]selector, nbChunks)
	for chunk := uint64(0); chunk < nbChunks; chunk++ {
		jc := uint64(chunk * c)
		d := selector{}
		d.index = jc / 64
		d.shift = jc - (d.index * 64)
		d.mask = mask << d.shift
		d.multiWordSelect = !cDivides64 && d.shift > (64-c) && d.index < (fr.Limbs-1)
		if d.multiWordSelect {
			nbBitsHigh := d.shift - uint64(64-c)
			d.maskHigh = (1 << nbBitsHigh) - 1
			d.shiftHigh = (c - nbBitsHigh)
		}
		selectors[chunk] = d
	}

	parallel.Execute(len(scalars), func(start, end int) {
		for i := start; i < end; i++ {
			if scalars[i].IsZero() {
				// everything is 0, no need to process this scalar
				continue
			}
			scalar := scalars[i].Bits()

			var carry int

			// for each chunk in the scalar, compute the current digit, and an eventual carry
			for chunk := uint64(0); chunk < nbChunks-1; chunk++ {
				s := selectors[chunk]

				// init with carry if any
				digit := carry
				carry = 0

				// digit = value of the c-bit window
				digit += int((scalar[s.index] & s.mask) >> s.shift)

				if s.multiWordSelect {
					// we are selecting bits over 2 words
					digit += int(scalar[s.index+1]&s.maskHigh) << s.shiftHigh
				}

				// if the digit is larger than 2^{c-1}, then, we borrow 2^c from the next window and subtract
				// 2^{c} to the current digit, making it negative.
				if digit > max {
					digit -= (1 << c)
					carry = 1
				}

				// if digit is zero, no impact on result
				if digit == 0 {
					continue
				}

				var bits uint16
				if digit > 0 {
					bits = uint16(digit) << 1
				} else {
					bits = (uint16(-digit-1) << 1) + 1
				}
				digits[int(chunk)*len(scalars)+i] = bits
			}

			// for the last chunk, we don't want to borrow from a next window
			// (but may have a larger max value)
			chunk := nbChunks - 1
			s := selectors[chunk]
			// init with carry if any
			digit := carry
			// digit = value of the c-bit window
			digit += int((scalar[s.index] & s.mask) >> s.shift)
			if s.multiWordSelect {
				// we are selecting bits over 2 words
				digit += int(scalar[s.index+1]&s.maskHigh) << s.shiftHigh
			}
			digits[int(chunk)*len(scalars)+i] = uint16(digit) << 1
		}

	}, nbTasks)

	// aggregate  chunk stats
	chunkStats := make([]chunkStat, nbChunks)
	if c <= 9 {
		// no need to compute stats for small window sizes
		return digits, chunkStats
	}
	parallel.Execute(len(chunkStats), func(start, end int) {
		// for each chunk compute the statistics
		for chunkID := start; chunkID < end; chunkID++ {
			// indicates if a bucket is hit.
			var b bitSetC16

			// digits for the chunk
			chunkDigits := digits[chunkID*len(scalars) : (chunkID+1)*len(scalars)]

			totalOps := 0
			nz := 0 // non zero buckets count
			for _, digit := range chunkDigits {
				if digit == 0 {
					continue
				}
				totalOps++
				bucketID := digit >> 1
				if digit&1 == 0 {
					bucketID -= 1
				}
				if !b[bucketID] {
					nz++
					b[bucketID] = true
				}
			}
			chunkStats[chunkID].weight = float32(totalOps) // count number of ops for now, we will compute the weight after
			chunkStats[chunkID].ppBucketFilled = (float32(nz) * 100.0) / float32(int(1<<(c-1)))
			chunkStats[chunkID].nbBucketFilled = nz
		}
	}, nbTasks)

	totalOps := float32(0.0)
	for _, stat := range chunkStats {
		totalOps += stat.weight
	}

	target := totalOps / float32(nbChunks)
	if target != 0.0 {
		// if target == 0, it means all the scalars are 0 everywhere, there is no work to be done.
		for i := 0; i < len(chunkStats); i++ {
			chunkStats[i].weight = (chunkStats[i].weight * 100.0) / target
		}
	}

	return digits, chunkStats
}
