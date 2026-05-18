// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package bn254

import (
	"errors"
	"math/bits"
	"runtime"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/parallel"
)

const (
	glvMSMG1Window = 16
	glvMSMG1Chunks = 8

	glvMSMG1WideWindow      = 19
	glvMSMG1WideChunks      = 7
	glvMSMG1WideStripeBits  = 12
	glvMSMG1WideStripeSize  = 1 << glvMSMG1WideStripeBits
	glvMSMG1WideStripeCount = 1 << (glvMSMG1WideWindow - 1 - glvMSMG1WideStripeBits)
	glvMSMG1WideBatchSize   = 640
)

type glvMSMG1Scalar struct {
	k0     [glvMSMG1Chunks]int16
	k1     [glvMSMG1Chunks]int16
	extra0 int8
	extra1 int8
}

type glvMSMG1WideScalar struct {
	k0     [glvMSMG1WideChunks]int32
	k1     [glvMSMG1WideChunks]int32
	extra0 int8
	extra1 int8
}

type glvMSMG1WideOp struct {
	index  uint32
	bucket uint32
	neg    bool
}

type batchOpG1AffineWide struct {
	bucketID uint32
	point    G1Affine
}

var (
	glvMSMG1B1 = [6]uint64{
		0x96ce4aece61f0339,
		0x2e3ff027efccd68a,
		0x8fa7d32d2fafba64,
		0x6eb9c714773a6ef2,
		0xd91d232ec7e0b3d7,
		0x0000000000000002,
	}
	glvMSMG1B2Abs = [7]uint64{
		0xd073ced5f11aeea9,
		0x7abf2e6fc85f00fa,
		0x869375169b9bdffa,
		0xa5e38cfb5eaa26d9,
		0x7a7bd9d4391eb18d,
		0x4ccef014a773d2cf,
		0x0000000000000002,
	}
	glvMSMG1V11Abs = [2]uint64{0x8211bbeb7d4f1128, 0x6f4d8248eeb859fc}
	glvMSMG1V20    = [2]uint64{0x0be4e1541221250b, 0x6f4d8248eeb859fd}
)

const glvMSMG1V10 = uint64(0x89d3256894d213e3)

// MultiExpGLV computes sum(scalars[i] * points[i]) using the BN254 G1 GLV
// endomorphism. It is an experimental opt-in MSM variant for large MSMs.
func (p *G1Affine) MultiExpGLV(points []G1Affine, scalars []fr.Element, config ecc.MultiExpConfig) (*G1Affine, error) {
	var _p G1Jac
	if _, err := _p.MultiExpGLV(points, scalars, config); err != nil {
		return nil, err
	}
	p.FromJacobian(&_p)
	return p, nil
}

// MultiExpGLV computes sum(scalars[i] * points[i]) using a deferred-phi GLV
// Pippenger variant.
//
// Each scalar is decomposed as k0 + lambda*k1. For each window we build two
// bucket tables over the original input points and apply phi only after the k1
// bucket table has been reduced:
//
//	sum k0_i*P_i + phi(sum k1_i*P_i)
//
// This avoids materializing phi(P_i) for every input point.
func (p *G1Jac) MultiExpGLV(points []G1Affine, scalars []fr.Element, config ecc.MultiExpConfig) (*G1Jac, error) {
	nbPoints := len(points)
	if nbPoints != len(scalars) {
		return nil, errors.New("len(points) != len(scalars)")
	}
	if config.NbTasks <= 0 {
		config.NbTasks = runtime.NumCPU() * 2
	} else if config.NbTasks > 1024 {
		return nil, errors.New("invalid config: config.NbTasks > 1024")
	}

	digits := decomposeScalarsGLVG1(scalars, config.NbTasks)
	return innerMSMGLVG1(p, points, digits), nil
}

// MultiExpGLVWide computes sum(scalars[i] * points[i]) using the deferred-phi
// GLV MSM with 19-bit windows and a striped bucket backend.
func (p *G1Affine) MultiExpGLVWide(points []G1Affine, scalars []fr.Element, config ecc.MultiExpConfig) (*G1Affine, error) {
	var _p G1Jac
	if _, err := _p.MultiExpGLVWide(points, scalars, config); err != nil {
		return nil, err
	}
	p.FromJacobian(&_p)
	return p, nil
}

// MultiExpGLVWide computes sum(scalars[i] * points[i]) using 19-bit GLV
// component windows. The larger window reduces bucket insertions to 14N for
// BN254-sized scalars, while stripes avoid full 2^18 bucket-table allocation.
func (p *G1Jac) MultiExpGLVWide(points []G1Affine, scalars []fr.Element, config ecc.MultiExpConfig) (*G1Jac, error) {
	nbPoints := len(points)
	if nbPoints != len(scalars) {
		return nil, errors.New("len(points) != len(scalars)")
	}
	if nbPoints > 1<<32-1 {
		return nil, errors.New("invalid input: len(points) > math.MaxUint32")
	}
	if config.NbTasks <= 0 {
		config.NbTasks = runtime.NumCPU() * 2
	} else if config.NbTasks > 1024 {
		return nil, errors.New("invalid config: config.NbTasks > 1024")
	}

	digits := decomposeScalarsGLVG1Wide(scalars, config.NbTasks)
	return innerMSMGLVG1Wide(p, points, digits), nil
}

// MultiExpGLVWideDense computes sum(scalars[i] * points[i]) with the 19-bit
// deferred-phi GLV MSM and full dynamic bucket tables. It is intended for
// benchmarking the algorithmic tradeoff against the lower-memory striped path.
func (p *G1Affine) MultiExpGLVWideDense(points []G1Affine, scalars []fr.Element, config ecc.MultiExpConfig) (*G1Affine, error) {
	var _p G1Jac
	if _, err := _p.MultiExpGLVWideDense(points, scalars, config); err != nil {
		return nil, err
	}
	p.FromJacobian(&_p)
	return p, nil
}

// MultiExpGLVWideDense computes sum(scalars[i] * points[i]) with the 19-bit
// deferred-phi GLV MSM and full dynamic bucket tables.
func (p *G1Jac) MultiExpGLVWideDense(points []G1Affine, scalars []fr.Element, config ecc.MultiExpConfig) (*G1Jac, error) {
	nbPoints := len(points)
	if nbPoints != len(scalars) {
		return nil, errors.New("len(points) != len(scalars)")
	}
	if config.NbTasks <= 0 {
		config.NbTasks = runtime.NumCPU() * 2
	} else if config.NbTasks > 1024 {
		return nil, errors.New("invalid config: config.NbTasks > 1024")
	}

	digits := decomposeScalarsGLVG1Wide(scalars, config.NbTasks)
	return innerMSMGLVG1WideDense(p, points, digits), nil
}

func decomposeScalarsGLVG1(scalars []fr.Element, nbTasks int) []glvMSMG1Scalar {
	digits := make([]glvMSMG1Scalar, len(scalars))

	parallel.Execute(len(scalars), func(start, end int) {
		for i := start; i < end; i++ {
			if scalars[i].IsZero() {
				continue
			}

			splitScalarGLVG1Fixed(scalars[i].Bits(), &digits[i])
		}
	}, nbTasks)

	return digits
}

func decomposeScalarsGLVG1Wide(scalars []fr.Element, nbTasks int) []glvMSMG1WideScalar {
	digits := make([]glvMSMG1WideScalar, len(scalars))

	parallel.Execute(len(scalars), func(start, end int) {
		for i := start; i < end; i++ {
			if scalars[i].IsZero() {
				continue
			}

			splitScalarGLVG1FixedWide(scalars[i].Bits(), &digits[i])
		}
	}, nbTasks)

	return digits
}

func splitScalarGLVG1Fixed(s [4]uint64, out *glvMSMG1Scalar) {
	sign0, mag0, sign1, mag1 := splitScalarGLVG1FixedComponents(s)
	recodeSignedGLVComponentG1(sign0, mag0, &out.k0, &out.extra0)
	recodeSignedGLVComponentG1(sign1, mag1, &out.k1, &out.extra1)
}

func splitScalarGLVG1FixedWide(s [4]uint64, out *glvMSMG1WideScalar) {
	sign0, mag0, sign1, mag1 := splitScalarGLVG1FixedComponents(s)
	recodeSignedGLVComponentG1Wide(sign0, mag0, &out.k0, &out.extra0)
	recodeSignedGLVComponentG1Wide(sign1, mag1, &out.k1, &out.extra1)
}

func splitScalarGLVG1FixedComponents(s [4]uint64) (int, [4]uint64, int, [4]uint64) {
	c1 := mul4x6High512(s, glvMSMG1B1)
	c2 := mul4x7High512(s, glvMSMG1B2Abs)

	var v0 [4]uint64
	add4(&v0, mul1x1(c1, glvMSMG1V10))
	add4(&v0, mul2x2(c2, glvMSMG1V20))

	sign0, mag0 := subSigned4(s, v0)

	a := mul1x2(c1, glvMSMG1V11Abs)
	b := mul2x1(c2, glvMSMG1V10)
	sign1, mag1 := subSigned4(a, b)

	return sign0, mag0, sign1, mag1
}

func recodeSignedGLVComponentG1(sign int, mag [4]uint64, digits *[glvMSMG1Chunks]int16, extra *int8) {
	const (
		base     = int64(1 << glvMSMG1Window)
		minDigit = -int64(1 << (glvMSMG1Window - 1))
		maxDigit = int64(1<<(glvMSMG1Window-1)) - 1
	)

	if sign == 0 {
		return
	}

	var carry int64
	for i := range glvMSMG1Chunks {
		di := int64(window16U256(mag, i))
		if sign < 0 {
			di = -di
		}
		di += carry
		carry = 0

		if di > maxDigit {
			di -= base
			carry = 1
		} else if di < minDigit {
			di += base
			carry = -1
		}

		digits[i] = int16(di)
	}

	*extra = int8(carry)
}

func recodeSignedGLVComponentG1Wide(sign int, mag [4]uint64, digits *[glvMSMG1WideChunks]int32, extra *int8) {
	const (
		base     = int64(1 << glvMSMG1WideWindow)
		minDigit = -int64(1 << (glvMSMG1WideWindow - 1))
		maxDigit = int64(1<<(glvMSMG1WideWindow-1)) - 1
	)

	if sign == 0 {
		return
	}

	var carry int64
	for i := range glvMSMG1WideChunks {
		di := int64(windowU256(mag, i*glvMSMG1WideWindow, glvMSMG1WideWindow))
		if sign < 0 {
			di = -di
		}
		di += carry
		carry = 0

		if di > maxDigit {
			di -= base
			carry = 1
		} else if di < minDigit {
			di += base
			carry = -1
		}

		digits[i] = int32(di)
	}

	*extra = int8(carry)
}

func window16U256(words [4]uint64, chunk int) uint64 {
	const mask = uint64(1<<glvMSMG1Window) - 1
	return (words[chunk/4] >> (uint(chunk%4) * glvMSMG1Window)) & mask
}

func windowU256(words [4]uint64, bitOffset, width int) uint64 {
	mask := uint64(1<<width) - 1
	wordIndex := bitOffset / 64
	shift := uint(bitOffset % 64)
	value := words[wordIndex] >> shift
	if shift != 0 && wordIndex+1 < len(words) && int(shift)+width > 64 {
		value |= words[wordIndex+1] << (64 - shift)
	}
	return value & mask
}

func mul4x6High512(a [4]uint64, b [6]uint64) uint64 {
	var product [10]uint64
	for i := range a {
		for j := range b {
			addMul64(product[:], i+j, a[i], b[j])
		}
	}
	return product[8]
}

func mul4x7High512(a [4]uint64, b [7]uint64) [2]uint64 {
	var product [11]uint64
	for i := range a {
		for j := range b {
			addMul64(product[:], i+j, a[i], b[j])
		}
	}
	return [2]uint64{product[8], product[9]}
}

func mul1x1(a, b uint64) (out [4]uint64) {
	out[1], out[0] = bits.Mul64(a, b)
	return out
}

func mul1x2(a uint64, b [2]uint64) (out [4]uint64) {
	hi, lo := bits.Mul64(a, b[0])
	out[0] = lo
	carry := hi
	hi, lo = bits.Mul64(a, b[1])
	out[1], carry = bits.Add64(lo, carry, 0)
	out[2], _ = bits.Add64(hi, 0, carry)
	return out
}

func mul2x1(a [2]uint64, b uint64) (out [4]uint64) {
	hi, lo := bits.Mul64(a[0], b)
	out[0] = lo
	carry := hi
	hi, lo = bits.Mul64(a[1], b)
	out[1], carry = bits.Add64(lo, carry, 0)
	out[2], _ = bits.Add64(hi, 0, carry)
	return out
}

func mul2x2(a, b [2]uint64) (out [4]uint64) {
	for i := range a {
		for j := range b {
			addMul64(out[:], i+j, a[i], b[j])
		}
	}
	return out
}

func addMul64(out []uint64, index int, a, b uint64) {
	hi, lo := bits.Mul64(a, b)

	var carry uint64
	out[index], carry = bits.Add64(out[index], lo, 0)
	hi, hiCarry := bits.Add64(hi, 0, carry)
	out[index+1], carry = bits.Add64(out[index+1], hi, 0)
	carry += hiCarry

	for i := index + 2; carry != 0 && i < len(out); i++ {
		out[i], carry = bits.Add64(out[i], 0, carry)
	}
}

func add4(z *[4]uint64, x [4]uint64) {
	var carry uint64
	for i := range 4 {
		z[i], carry = bits.Add64(z[i], x[i], carry)
	}
}

func subSigned4(a, b [4]uint64) (int, [4]uint64) {
	switch cmp4(a, b) {
	case 1:
		return 1, sub4(a, b)
	case -1:
		return -1, sub4(b, a)
	default:
		return 0, [4]uint64{}
	}
}

func cmp4(a, b [4]uint64) int {
	for i := 3; i >= 0; i-- {
		if a[i] > b[i] {
			return 1
		}
		if a[i] < b[i] {
			return -1
		}
	}
	return 0
}

func sub4(a, b [4]uint64) (out [4]uint64) {
	var borrow uint64
	for i := range 4 {
		out[i], borrow = bits.Sub64(a[i], b[i], borrow)
	}
	return out
}

func innerMSMGLVG1(p *G1Jac, points []G1Affine, digits []glvMSMG1Scalar) *G1Jac {
	chunks := make([]chan g1JacExtended, glvMSMG1Chunks+1)
	for i := range chunks {
		chunks[i] = make(chan g1JacExtended, 1)
	}

	for chunk := range glvMSMG1Chunks {
		go processChunkG1GLVBatchAffineC16(chunk, chunks[chunk], points, digits)
	}
	go processExtraChunkG1GLV(chunks[glvMSMG1Chunks], points, digits)

	return msmReduceChunkG1Affine(p, glvMSMG1Window, chunks)
}

func innerMSMGLVG1Wide(p *G1Jac, points []G1Affine, digits []glvMSMG1WideScalar) *G1Jac {
	chunks := make([]chan g1JacExtended, glvMSMG1WideChunks+1)
	for i := range chunks {
		chunks[i] = make(chan g1JacExtended, 1)
	}

	for chunk := range glvMSMG1WideChunks {
		go processChunkG1GLVWideStriped(chunk, chunks[chunk], points, digits)
	}
	go processExtraChunkG1GLVWide(chunks[glvMSMG1WideChunks], points, digits)

	return msmReduceChunkG1Affine(p, glvMSMG1WideWindow, chunks)
}

func innerMSMGLVG1WideDense(p *G1Jac, points []G1Affine, digits []glvMSMG1WideScalar) *G1Jac {
	chunks := make([]chan g1JacExtended, glvMSMG1WideChunks+1)
	for i := range chunks {
		chunks[i] = make(chan g1JacExtended, 1)
	}

	for chunk := range glvMSMG1WideChunks {
		go processChunkG1GLVWideDense(chunk, chunks[chunk], points, digits)
	}
	go processExtraChunkG1GLVWide(chunks[glvMSMG1WideChunks], points, digits)

	return msmReduceChunkG1Affine(p, glvMSMG1WideWindow, chunks)
}

type g1GLVBatchAffineC16 struct {
	buckets   bucketG1AffineC16
	bucketsJE bucketg1JacExtendedC16
	bucketIds bitSetC16
	cptAdd    int
	R         ppG1AffineC16
	P         pG1AffineC16
	queue     qG1AffineC16
	qID       int
}

func (ctx *g1GLVBatchAffineC16) init() {
	for i := range len(ctx.bucketsJE) {
		ctx.bucketsJE[i].SetInfinity()
	}
}

func (ctx *g1GLVBatchAffineC16) executeAndReset() {
	if ctx.cptAdd == 0 {
		return
	}
	batchAddG1Affine[pG1AffineC16, ppG1AffineC16, cG1AffineC16](&ctx.R, &ctx.P, ctx.cptAdd)
	ctx.bucketIds = bitSetC16{}
	ctx.cptAdd = 0
}

func (ctx *g1GLVBatchAffineC16) addFromQueue(op batchOpG1Affine) {
	BK := &ctx.buckets[op.bucketID]
	if BK.IsInfinity() {
		BK.Set(&op.point)
		return
	}
	if BK.X.Equal(&op.point.X) {
		if BK.Y.Equal(&op.point.Y) {
			ctx.bucketsJE[op.bucketID].addMixed(&op.point)
			return
		}
		BK.SetInfinity()
		return
	}

	ctx.bucketIds[op.bucketID] = true
	ctx.R[ctx.cptAdd] = BK
	ctx.P[ctx.cptAdd] = op.point
	ctx.cptAdd++
}

func (ctx *g1GLVBatchAffineC16) add(bucketID uint16, point *G1Affine, isAdd bool) {
	BK := &ctx.buckets[bucketID]
	if BK.IsInfinity() {
		if isAdd {
			BK.Set(point)
		} else {
			BK.Neg(point)
		}
		return
	}
	if BK.X.Equal(&point.X) {
		if BK.Y.Equal(&point.Y) {
			if isAdd {
				ctx.bucketsJE[bucketID].addMixed(point)
			} else {
				BK.SetInfinity()
			}
			return
		}
		if isAdd {
			BK.SetInfinity()
		} else {
			ctx.bucketsJE[bucketID].subMixed(point)
		}
		return
	}

	ctx.bucketIds[bucketID] = true
	ctx.R[ctx.cptAdd] = BK
	if isAdd {
		ctx.P[ctx.cptAdd].Set(point)
	} else {
		ctx.P[ctx.cptAdd].Neg(point)
	}
	ctx.cptAdd++
}

func (ctx *g1GLVBatchAffineC16) flushQueue() {
	for i := range ctx.qID {
		ctx.bucketsJE[ctx.queue[i].bucketID].addMixed(&ctx.queue[i].point)
	}
	ctx.qID = 0
}

func (ctx *g1GLVBatchAffineC16) processTopQueue() {
	for i := ctx.qID - 1; i >= 0; i-- {
		if ctx.bucketIds[ctx.queue[i].bucketID] {
			return
		}
		ctx.addFromQueue(ctx.queue[i])
		ctx.qID--
	}
}

func (ctx *g1GLVBatchAffineC16) addDigit(digit int16, point *G1Affine) {
	if digit == 0 || point.IsInfinity() {
		return
	}

	d := int32(digit)
	isAdd := d > 0
	if !isAdd {
		d = -d
	}
	bucketID := uint16(d - 1)

	if ctx.bucketIds[bucketID] {
		ctx.queue[ctx.qID].bucketID = bucketID
		if isAdd {
			ctx.queue[ctx.qID].point.Set(point)
		} else {
			ctx.queue[ctx.qID].point.Neg(point)
		}
		ctx.qID++
		if ctx.qID == len(ctx.queue)-1 {
			ctx.flushQueue()
		}
		return
	}

	ctx.add(bucketID, point, isAdd)
	if ctx.cptAdd == len(ctx.P) {
		ctx.executeAndReset()
		ctx.processTopQueue()
	}
}

func (ctx *g1GLVBatchAffineC16) reduce(total *g1JacExtended) {
	ctx.executeAndReset()
	ctx.flushQueue()

	var runningSum g1JacExtended
	runningSum.SetInfinity()
	total.SetInfinity()
	for k := len(ctx.buckets) - 1; k >= 0; k-- {
		runningSum.addMixed(&ctx.buckets[k])
		if !ctx.bucketsJE[k].IsInfinity() {
			runningSum.add(&ctx.bucketsJE[k])
		}
		total.add(&runningSum)
	}
}

func processChunkG1GLVBatchAffineC16(chunk int, chRes chan<- g1JacExtended, points []G1Affine, digits []glvMSMG1Scalar) {
	var bucket0, bucket1 g1GLVBatchAffineC16
	bucket0.init()
	bucket1.init()

	for i := range points {
		bucket0.addDigit(digits[i].k0[chunk], &points[i])
		bucket1.addDigit(digits[i].k1[chunk], &points[i])
	}

	var total0, total1 g1JacExtended
	bucket0.reduce(&total0)
	bucket1.reduce(&total1)
	phiG1JacExtended(&total1)
	total0.add(&total1)
	chRes <- total0
}

func processExtraChunkG1GLV(chRes chan<- g1JacExtended, points []G1Affine, digits []glvMSMG1Scalar) {
	var total0, total1 g1JacExtended
	total0.SetInfinity()
	total1.SetInfinity()

	for i := range points {
		switch digits[i].extra0 {
		case 1:
			total0.addMixed(&points[i])
		case -1:
			total0.subMixed(&points[i])
		}
		switch digits[i].extra1 {
		case 1:
			total1.addMixed(&points[i])
		case -1:
			total1.subMixed(&points[i])
		}
	}

	phiG1JacExtended(&total1)
	total0.add(&total1)
	chRes <- total0
}

func phiG1JacExtended(p *g1JacExtended) {
	if p.IsInfinity() {
		return
	}
	p.X.Mul(&p.X, &thirdRootOneG1)
}

type g1GLVStripeBatchAffine struct {
	buckets      []G1Affine
	bucketsJE    []g1JacExtended
	bucketStamps []uint32
	stamp        uint32
	touched      []uint32
	touchedMarks []bool
	cptAdd       int
	R            []*G1Affine
	P            []G1Affine
	queue        []batchOpG1AffineWide
	qID          int
	lambda       []fp.Element
	lambdain     []fp.Element
}

func newG1GLVBatchAffineDynamic(size int) g1GLVStripeBatchAffine {
	return g1GLVStripeBatchAffine{
		buckets:      make([]G1Affine, size),
		bucketsJE:    make([]g1JacExtended, size),
		bucketStamps: make([]uint32, size),
		stamp:        1,
		touched:      make([]uint32, 0, min(size, glvMSMG1WideStripeSize)),
		touchedMarks: make([]bool, size),
		R:            make([]*G1Affine, glvMSMG1WideBatchSize),
		P:            make([]G1Affine, glvMSMG1WideBatchSize),
		queue:        make([]batchOpG1AffineWide, glvMSMG1WideBatchSize),
		lambda:       make([]fp.Element, glvMSMG1WideBatchSize),
		lambdain:     make([]fp.Element, glvMSMG1WideBatchSize),
	}
}

func newG1GLVStripeBatchAffine() g1GLVStripeBatchAffine {
	return newG1GLVBatchAffineDynamic(glvMSMG1WideStripeSize)
}

func (ctx *g1GLVStripeBatchAffine) resetBatch() {
	ctx.stamp++
	if ctx.stamp == 0 {
		clear(ctx.bucketStamps)
		ctx.stamp = 1
	}
	ctx.cptAdd = 0
}

func (ctx *g1GLVStripeBatchAffine) touch(bucketID uint32) {
	if !ctx.touchedMarks[bucketID] {
		ctx.touchedMarks[bucketID] = true
		ctx.touched = append(ctx.touched, bucketID)
	}
}

func (ctx *g1GLVStripeBatchAffine) resetStripe() {
	for _, bucketID := range ctx.touched {
		ctx.buckets[bucketID].SetInfinity()
		ctx.bucketsJE[bucketID] = g1JacExtended{}
		ctx.touchedMarks[bucketID] = false
	}
	ctx.touched = ctx.touched[:0]
	ctx.qID = 0
	ctx.resetBatch()
}

func (ctx *g1GLVStripeBatchAffine) executeAndReset() {
	if ctx.cptAdd == 0 {
		return
	}
	batchAddG1AffineDynamic(ctx.R, ctx.P, ctx.lambda, ctx.lambdain, ctx.cptAdd)
	ctx.resetBatch()
}

func (ctx *g1GLVStripeBatchAffine) addFromQueue(op batchOpG1AffineWide) {
	BK := &ctx.buckets[op.bucketID]
	ctx.touch(op.bucketID)
	if BK.IsInfinity() {
		BK.Set(&op.point)
		return
	}
	if BK.X.Equal(&op.point.X) {
		if BK.Y.Equal(&op.point.Y) {
			ctx.bucketsJE[op.bucketID].addMixed(&op.point)
			return
		}
		BK.SetInfinity()
		return
	}

	ctx.bucketStamps[op.bucketID] = ctx.stamp
	ctx.R[ctx.cptAdd] = BK
	ctx.P[ctx.cptAdd] = op.point
	ctx.cptAdd++
}

func (ctx *g1GLVStripeBatchAffine) add(bucketID uint32, point *G1Affine, isAdd bool) {
	BK := &ctx.buckets[bucketID]
	ctx.touch(bucketID)
	if BK.IsInfinity() {
		if isAdd {
			BK.Set(point)
		} else {
			BK.Neg(point)
		}
		return
	}
	if BK.X.Equal(&point.X) {
		if BK.Y.Equal(&point.Y) {
			if isAdd {
				ctx.bucketsJE[bucketID].addMixed(point)
			} else {
				BK.SetInfinity()
			}
			return
		}
		if isAdd {
			BK.SetInfinity()
		} else {
			ctx.bucketsJE[bucketID].subMixed(point)
		}
		return
	}

	ctx.bucketStamps[bucketID] = ctx.stamp
	ctx.R[ctx.cptAdd] = BK
	if isAdd {
		ctx.P[ctx.cptAdd].Set(point)
	} else {
		ctx.P[ctx.cptAdd].Neg(point)
	}
	ctx.cptAdd++
}

func (ctx *g1GLVStripeBatchAffine) flushQueue() {
	for i := range ctx.qID {
		ctx.touch(ctx.queue[i].bucketID)
		ctx.bucketsJE[ctx.queue[i].bucketID].addMixed(&ctx.queue[i].point)
	}
	ctx.qID = 0
}

func (ctx *g1GLVStripeBatchAffine) processTopQueue() {
	for i := ctx.qID - 1; i >= 0; i-- {
		if ctx.bucketStamps[ctx.queue[i].bucketID] == ctx.stamp {
			return
		}
		ctx.addFromQueue(ctx.queue[i])
		ctx.qID--
	}
}

func (ctx *g1GLVStripeBatchAffine) addOp(op glvMSMG1WideOp, points []G1Affine) {
	point := &points[op.index]
	if point.IsInfinity() {
		return
	}

	if ctx.bucketStamps[op.bucket] == ctx.stamp {
		ctx.queue[ctx.qID].bucketID = op.bucket
		if op.neg {
			ctx.queue[ctx.qID].point.Neg(point)
		} else {
			ctx.queue[ctx.qID].point.Set(point)
		}
		ctx.qID++
		if ctx.qID == len(ctx.queue)-1 {
			ctx.flushQueue()
		}
		return
	}

	ctx.add(op.bucket, point, !op.neg)
	if ctx.cptAdd == len(ctx.P) {
		ctx.executeAndReset()
		ctx.processTopQueue()
	}
}

func (ctx *g1GLVStripeBatchAffine) reduceInto(runningSum, total *g1JacExtended) {
	ctx.executeAndReset()
	ctx.flushQueue()
	for k := len(ctx.buckets) - 1; k >= 0; k-- {
		runningSum.addMixed(&ctx.buckets[k])
		if !ctx.bucketsJE[k].IsInfinity() {
			runningSum.add(&ctx.bucketsJE[k])
		}
		total.add(runningSum)
	}
	ctx.resetStripe()
}

func (ctx *g1GLVStripeBatchAffine) reduce(total *g1JacExtended) {
	var runningSum g1JacExtended
	runningSum.SetInfinity()
	total.SetInfinity()
	ctx.reduceInto(&runningSum, total)
}

func processChunkG1GLVWideStriped(chunk int, chRes chan<- g1JacExtended, points []G1Affine, digits []glvMSMG1WideScalar) {
	var counts0, counts1 [glvMSMG1WideStripeCount]int
	for i := range points {
		if points[i].IsInfinity() {
			continue
		}
		if d := digits[i].k0[chunk]; d != 0 {
			counts0[stripeIDG1Wide(d)]++
		}
		if d := digits[i].k1[chunk]; d != 0 {
			counts1[stripeIDG1Wide(d)]++
		}
	}

	offsets0, totalOps0 := prefixWideCountsG1(counts0)
	offsets1, totalOps1 := prefixWideCountsG1(counts1)
	ops0 := make([]glvMSMG1WideOp, totalOps0)
	ops1 := make([]glvMSMG1WideOp, totalOps1)
	cursor0, cursor1 := offsets0, offsets1

	for i := range points {
		if points[i].IsInfinity() {
			continue
		}
		if d := digits[i].k0[chunk]; d != 0 {
			stripe, bucket, neg := splitWideDigitG1(d)
			ops0[cursor0[stripe]] = glvMSMG1WideOp{index: uint32(i), bucket: bucket, neg: neg}
			cursor0[stripe]++
		}
		if d := digits[i].k1[chunk]; d != 0 {
			stripe, bucket, neg := splitWideDigitG1(d)
			ops1[cursor1[stripe]] = glvMSMG1WideOp{index: uint32(i), bucket: bucket, neg: neg}
			cursor1[stripe]++
		}
	}

	total0 := reduceWideOpsG1(points, ops0, offsets0, counts0)
	total1 := reduceWideOpsG1(points, ops1, offsets1, counts1)
	phiG1JacExtended(&total1)
	total0.add(&total1)
	chRes <- total0
}

func processChunkG1GLVWideDense(chunk int, chRes chan<- g1JacExtended, points []G1Affine, digits []glvMSMG1WideScalar) {
	bucket0 := newG1GLVBatchAffineDynamic(1 << (glvMSMG1WideWindow - 1))
	bucket1 := newG1GLVBatchAffineDynamic(1 << (glvMSMG1WideWindow - 1))

	for i := range points {
		if points[i].IsInfinity() {
			continue
		}
		bucket0.addWideDigit(digits[i].k0[chunk], &points[i])
		bucket1.addWideDigit(digits[i].k1[chunk], &points[i])
	}

	var total0, total1 g1JacExtended
	bucket0.reduce(&total0)
	bucket1.reduce(&total1)
	phiG1JacExtended(&total1)
	total0.add(&total1)
	chRes <- total0
}

func (ctx *g1GLVStripeBatchAffine) addWideDigit(digit int32, point *G1Affine) {
	if digit == 0 {
		return
	}
	neg := digit < 0
	if neg {
		digit = -digit
	}
	bucketID := uint32(digit) - 1

	if ctx.bucketStamps[bucketID] == ctx.stamp {
		ctx.queue[ctx.qID].bucketID = bucketID
		if neg {
			ctx.queue[ctx.qID].point.Neg(point)
		} else {
			ctx.queue[ctx.qID].point.Set(point)
		}
		ctx.qID++
		if ctx.qID == len(ctx.queue)-1 {
			ctx.flushQueue()
		}
		return
	}

	ctx.add(bucketID, point, !neg)
	if ctx.cptAdd == len(ctx.P) {
		ctx.executeAndReset()
		ctx.processTopQueue()
	}
}

func stripeIDG1Wide(digit int32) int {
	if digit < 0 {
		digit = -digit
	}
	return int((uint32(digit) - 1) >> glvMSMG1WideStripeBits)
}

func splitWideDigitG1(digit int32) (int, uint32, bool) {
	neg := digit < 0
	if neg {
		digit = -digit
	}
	bucketID := uint32(digit) - 1
	stripe := int(bucketID >> glvMSMG1WideStripeBits)
	bucket := bucketID & (glvMSMG1WideStripeSize - 1)
	return stripe, bucket, neg
}

func prefixWideCountsG1(counts [glvMSMG1WideStripeCount]int) ([glvMSMG1WideStripeCount]int, int) {
	var offsets [glvMSMG1WideStripeCount]int
	total := 0
	for i, count := range counts {
		offsets[i] = total
		total += count
	}
	return offsets, total
}

func reduceWideOpsG1(points []G1Affine, ops []glvMSMG1WideOp, offsets, counts [glvMSMG1WideStripeCount]int) g1JacExtended {
	ctx := newG1GLVStripeBatchAffine()
	var runningSum, total g1JacExtended
	runningSum.SetInfinity()
	total.SetInfinity()

	for stripe := glvMSMG1WideStripeCount - 1; stripe >= 0; stripe-- {
		if counts[stripe] == 0 {
			if !runningSum.IsInfinity() {
				for range glvMSMG1WideStripeSize {
					total.add(&runningSum)
				}
			}
			continue
		}

		start := offsets[stripe]
		end := start + counts[stripe]
		for _, op := range ops[start:end] {
			ctx.addOp(op, points)
		}
		ctx.reduceInto(&runningSum, &total)
	}

	return total
}

func processExtraChunkG1GLVWide(chRes chan<- g1JacExtended, points []G1Affine, digits []glvMSMG1WideScalar) {
	var total0, total1 g1JacExtended
	total0.SetInfinity()
	total1.SetInfinity()

	for i := range points {
		switch digits[i].extra0 {
		case 1:
			total0.addMixed(&points[i])
		case -1:
			total0.subMixed(&points[i])
		}
		switch digits[i].extra1 {
		case 1:
			total1.addMixed(&points[i])
		case -1:
			total1.subMixed(&points[i])
		}
	}

	phiG1JacExtended(&total1)
	total0.add(&total1)
	chRes <- total0
}

func batchAddG1AffineDynamic(R []*G1Affine, P []G1Affine, lambda, lambdain []fp.Element, batchSize int) {
	for j := range batchSize {
		lambdain[j].Sub(&P[j].X, &R[j].X)
	}

	var accumulator fp.Element
	lambda[0].SetOne()
	accumulator.Set(&lambdain[0])

	for i := 1; i < batchSize; i++ {
		lambda[i] = accumulator
		accumulator.Mul(&accumulator, &lambdain[i])
	}

	accumulator.Inverse(&accumulator)

	for i := batchSize - 1; i > 0; i-- {
		lambda[i].Mul(&lambda[i], &accumulator)
		accumulator.Mul(&accumulator, &lambdain[i])
	}
	lambda[0].Set(&accumulator)

	var t fp.Element
	var Q G1Affine
	for j := range batchSize {
		t.Sub(&P[j].Y, &R[j].Y)
		lambda[j].Mul(&lambda[j], &t)

		Q.X.Square(&lambda[j])
		Q.X.Sub(&Q.X, &R[j].X)
		Q.X.Sub(&Q.X, &P[j].X)

		t.Sub(&R[j].X, &Q.X)
		Q.Y.Mul(&lambda[j], &t)
		Q.Y.Sub(&Q.Y, &R[j].Y)

		R[j].Set(&Q)
	}
}
