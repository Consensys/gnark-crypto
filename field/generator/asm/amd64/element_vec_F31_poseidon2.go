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

package amd64

import (
	"fmt"
	"strconv"

	"github.com/consensys/bavard/amd64"
)

func (f *FFAmd64) generatePoseidon2_F31(params Poseidon2Parameters) {
	fullRounds := params.FullRounds
	partialRounds := params.PartialRounds
	width := params.Width
	rf := fullRounds / 2

	if width != 16 && width != 24 {
		panic("only width 16 and 24 are supported")
	}
	fnName := fmt.Sprintf("permutation%d_avx512", width)

	width24 := width == 24

	const argSize = 2 * 3 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 1, 0)
	registers := f.FnHeader(fnName, stackSize, argSize, amd64.AX, amd64.DX)

	addrInput := registers.Pop()
	addrRoundKeys := registers.Pop()
	addrDiagonal := registers.Pop()
	rKey := registers.Pop()

	// constants
	qd := registers.PopV()
	qInvNeg := registers.PopV()

	// state
	b0 := registers.PopV()
	b1 := registers.PopV()

	// temporary registers
	v0 := registers.PopV()
	v1 := registers.PopV()
	t0 := registers.PopV()
	t1 := registers.PopV()
	t2 := registers.PopV()
	t3 := registers.PopV()
	t4 := registers.PopV()
	t5 := registers.PopV()
	aOdd := registers.PopV()
	bOdd := registers.PopV()
	PL0 := registers.PopV()
	PL1 := registers.PopV()
	acc := registers.PopV().Y()
	accShuffled := registers.PopV().Y()

	// diagonal of the matrix
	d0 := registers.PopV()
	d0odd := registers.PopV()
	d1 := registers.PopV()
	d1odd := registers.PopV()

	// prepare the masks used for shuffling the vectors
	f.MOVQ(uint64(0b01_01_01_01_01_01_01_01), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	f.MOVQ(uint64(0x1), amd64.AX)
	f.KMOVQ(amd64.AX, amd64.K2)

	// load the constants
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qd)
	f.MOVD("$const_qInvNeg", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qInvNeg)

	f.MOVQ("input+0(FP)", addrInput)
	f.MOVQ("roundKeys+24(FP)", addrRoundKeys)
	// load the 3 * 8 uint32
	f.VMOVDQU32(addrInput.AtD(0), b0)
	if width24 {
		f.VMOVDQU32(addrInput.AtD(16), b1.Y())
		f.MOVQ("·diag24+0(SB)", addrDiagonal)
		f.VMOVDQU32(addrDiagonal.AtD(0), d0)
		f.VMOVDQU32(addrDiagonal.AtD(16), d1.Y())
		f.VPSRLQ("$32", d0, d0odd)
		f.VPSRLQ("$32", d1.Y(), d1odd.Y())
	} else {
		f.MOVQ("·diag16+0(SB)", addrDiagonal)
		f.VMOVDQU32(addrDiagonal.AtD(0), d0)
		f.VPSRLQ("$32", d0, d0odd)
	}

	add := f.Define("add", 5, func(args ...any) {
		a := args[0]
		b := args[1]
		qd := args[2]
		r0 := args[3]
		into := args[4]

		f.VPADDD(a, b, into)
		f.VPSUBD(qd, into, r0)
		f.VPMINUD(into, r0, into)
	}, true)

	matMulM4 := f.Define("mat_mul_m4", 6, func(args ...any) {
		block := args[0]
		t0 := args[1]
		t1 := args[2]
		t2 := args[3]
		qd := args[4]
		r0 := args[5]
		// We multiply by this matrix, each block of 4:
		// (2 3 1 1)
		// (1 2 3 1)
		// (1 1 2 3)
		// (3 1 1 2)
		// so we have
		// s0 = Σ + s0 + 2s1
		// s1 = Σ + s1 + 2s2
		// s2 = Σ + s2 + 2s3
		// s3 = Σ + s3 + 2s0
		// 1. we compute the sum
		// 2. we compute the shifted double(s)
		// 3. we add
		f.VPSHUFD(uint64(0x4e), block, t0)
		add(t0, block, qd, r0, t0)
		f.VPSHUFD(uint64(0xb1), t0, t1)
		add(t0, t1, qd, r0, t0)

		f.VPSHUFD(uint64(0x39), block, t2)
		f.VPSLLD("$1", t2, t2)
		f.VPSUBD(qd, t2, r0)
		f.VPMINUD(t2, r0, t2)

		// compute the sum
		add(block, t0, qd, r0, block)
		add(block, t2, qd, r0, block)
	}, true)

	matMulExternalInPlace := f.Define("mat_mul_external", 0, func(args ...any) {
		matMulM4(b0, t0, t1, t2, qd, t5)
		matMulM4(b1.Y(), t0.Y(), t1.Y(), t2.Y(), qd.Y(), t5.Y())

		// matMulExternalInPlace
		// we need to compute
		// acc[0] = Σ s[i%4]
		// acc[1] = Σ s[(i+1)%4]
		// acc[2] = Σ s[(i+2)%4]
		// acc[3] = Σ s[(i+3)%4]
		f.VEXTRACTI64X4(1, b0, acc)
		add(acc, b0.Y(), qd.Y(), t5.Y(), acc)
		add(acc, b1.Y(), qd.Y(), t5.Y(), acc)

		// we now have a Y register with the 8 elements
		// we permute to compute the desired result duplicated in acc[0..3] and acc[4..7]
		f.VSHUFF64X2(0b1, acc, acc, accShuffled)
		add(acc.Y(), accShuffled, qd.Y(), t5.Y(), acc.Y())

		f.VINSERTI64X4(1, acc.Y(), acc.Z(), acc.Z())

		add(b1.Y(), acc.Y(), qd.Y(), t3.Y(), b1.Y())
		add(b0, acc.Z(), qd, t5, b0)
	}, true)
	if !width24 {
		matMulExternalInPlace = f.Define("mat_mul_external_16", 0, func(args ...any) {
			matMulM4(b0, t0, t1, t2, qd, t5)
			f.VEXTRACTI64X4(1, b0, acc)
			add(acc, b0.Y(), qd.Y(), t5.Y(), acc)
			f.VSHUFF64X2(0b1, acc, acc, accShuffled)
			add(acc.Y(), accShuffled, qd.Y(), t5.Y(), acc.Y())
			f.VINSERTI64X4(1, acc.Y(), acc.Z(), acc.Z())
			add(b0, acc.Z(), qd, t5, b0)
		})
	}

	// computes c = a * b mod q
	// a and b can be in [0, 2q)
	_mulD := f.Define("mulD", 11, func(args ...any) {
		a := args[0]
		b := args[1]
		aOdd := args[2]
		bOdd := args[3]
		b0 := args[4]
		b1 := args[5]
		PL0 := args[6]
		PL1 := args[7]
		q := args[8]
		qInvNeg := args[9]
		c := args[10]

		f.VPSRLQ("$32", a, aOdd) // keep high 32 bits
		f.VPSRLQ("$32", b, bOdd) // keep high 32 bits

		// VPMULUDQ conveniently ignores the high 32 bits of each QWORD lane
		f.VPMULUDQ(a, b, b0)
		f.VPMULUDQ(aOdd, bOdd, b1)
		f.VPMULUDQ(b0, qInvNeg, PL0)
		f.VPMULUDQ(b1, qInvNeg, PL1)

		f.VPMULUDQ(PL0, q, PL0)
		f.VPADDQ(b0, PL0, b0)

		f.VPMULUDQ(PL1, q, PL1)
		f.VPADDQ(b1, PL1, c)

		f.VMOVSHDUPk(b0, amd64.K3, c)
	}, true)

	reduce1Q := f.Define("reduce1Q", 3, func(args ...any) {
		qd := args[0]
		c := args[1]
		r0 := args[2]

		f.VPSUBD(qd, c, r0)
		f.VPMINUD(c, r0, c)
	}, true)

	mulY := func(a, b, c amd64.VectorRegister, reduce bool) {
		mulInput := []any{a, b, aOdd, bOdd, t0, t1, PL0, PL1, qd, qInvNeg, c}
		for i := range mulInput {
			mulInput[i] = (mulInput[i]).(amd64.VectorRegister).Y()
		}
		_mulD(mulInput...)
		if reduce {
			reduce1Q(mulInput[8], mulInput[10], mulInput[7])
		}
	}

	mul := func(a, b, c amd64.VectorRegister, reduce bool) {
		mulInput := []any{a, b, aOdd, bOdd, t0, t1, PL0, PL1, qd, qInvNeg, c}
		for i := range mulInput {
			mulInput[i] = (mulInput[i]).(amd64.VectorRegister).Z()
		}
		_mulD(mulInput...)
		if reduce {
			reduce1Q(mulInput[8], mulInput[10], mulInput[7])
		}
	}

	var sbox, sboxPartial defineFn
	switch params.SBoxDegree {
	case 3:
		sbox = f.Define("sbox_full", 0, func(args ...any) {
			mul(b0, b0, t2, false)
			mul(b0, t2, b0, true)

			mulY(b1, b1, t1, false)
			mulY(b1, t1, b1, true)

		}, true)
		if !width24 {
			sbox = f.Define("sbox_full_16", 0, func(args ...any) {
				mul(b0, b0, t2, false)
				mul(b0, t2, b0, true)
			})
		}

		sboxPartial = f.Define("sbox_partial", 0, func(args ...any) {
			// t2.X() = b0 * b0
			// this is similar to the mulD macro
			// but since we only care about the mul result at [0],
			// we unroll and remove unnecessary code.
			f.VPMULUDQ(v1.X(), v1.X(), t0.X())
			f.VPMULUDQ(t0.X(), qInvNeg.X(), PL0.X())
			f.VPMULUDQ(PL0.X(), qd.X(), PL0.X())
			f.VPADDQ(t0.X(), PL0.X(), t0.X())
			f.VPSRLQ("$32", t0.X(), t2.X())

			// b0 = b0 * t2.X()
			f.VPMULUDQ(v1.X(), t2.X(), t0.X())
			f.VPMULUDQ(t0.X(), qInvNeg.X(), PL0.X())
			f.VPMULUDQ(PL0.X(), qd.X(), PL0.X())
			f.VPADDQ(t0.X(), PL0.X(), t0.X())
			f.VPSRLQ("$32", t0.X(), v1.X())
			f.VPSUBD(qd.X(), v1.X(), PL0.X())
			f.VPMINUD(v1.X(), PL0.X(), v1.X())
		}, true)
	case 7:
		sbox = f.Define("sbox_full", 0, func(args ...any) {
			mul(b0, b0, t2, true)
			mul(t2, t2, t3, false)
			mul(b0, t2, b0, false)
			mul(b0, t3, b0, true)

			mulY(b1, b1, t2, true)
			mulY(t2, t2, t3, false)
			mulY(b1, t2, b1, false)
			mulY(b1, t3, b1, true)
		}, true)
		if !width24 {
			sbox = f.Define("sbox_full_16", 0, func(args ...any) {
				mul(b0, b0, t2, true)
				mul(t2, t2, t3, false)
				mul(b0, t2, b0, false)
				mul(b0, t3, b0, true)
			})
		}

		sboxPartial = f.Define("sbox_partial", 0, func(args ...any) {
			mulY(v1, v1, t2, true)
			mulY(t2, t2, t3, false)
			mulY(v1, t2, v1, false)
			mulY(v1, t3, v1, true)
			// TODO: do it the following way.
			// // t2.X() = b0 * b0
			// // this is similar to the mulD macro
			// // but since we only care about the mul result at [0],
			// // we unroll and remove unnecessary code.
			// f.VPMULUDQ(v1.X(), v1.X(), t0.X())
			// f.VPMULUDQ(t0.X(), qInvNeg.X(), PL0.X())
			// f.VPMULUDQ(PL0.X(), qd.X(), PL0.X())
			// f.VPADDQ(t0.X(), PL0.X(), t0.X())
			// f.VPSRLQ("$32", t0.X(), t2.X())

			// // b0 = b0 * t2.X()
			// f.VPMULUDQ(v1.X(), t2.X(), t0.X())
			// f.VPMULUDQ(t0.X(), qInvNeg.X(), PL0.X())
			// f.VPMULUDQ(PL0.X(), qd.X(), PL0.X())
			// f.VPADDQ(t0.X(), PL0.X(), t0.X())
			// f.VPSRLQ("$32", t0.X(), v1.X())
			// f.VPSUBD(qd.X(), v1.X(), PL0.X())
			// f.VPMINUD(v1.X(), PL0.X(), v1.X())
		}, true)
	default:
		panic("only SBox degree 3 and 7 are supported")
	}

	sumState := f.Define("sum_state", 0, func(args ...any) {
		// first we compute the sum
		f.VEXTRACTI64X4(1, b0, acc) // TODO @gbotrel here
		add(acc, b1.Y(), qd.Y(), t5.Y(), acc)
		add(acc, t4.Y(), qd.Y(), t5.Y(), acc)

		// now we can work with acc.Y()
		f.VSHUFF64X2(0b1, acc, acc, accShuffled)
		add(acc, accShuffled, qd.Y(), t5.Y(), acc)

		f.VPSHUFD(uint64(0x4e), acc, accShuffled)
		add(acc, accShuffled, qd.Y(), t5.Y(), acc)
		f.VPSHUFD(uint64(0xb1), acc, accShuffled)
		add(acc, accShuffled, qd.Y(), t5.Y(), acc)

		f.VINSERTI64X4(1, acc, acc.Z(), acc.Z())
	}, true)
	if !width24 {
		sumState = f.Define("sum_state_16", 0, func(args ...any) {
			// first we compute the sum
			f.VEXTRACTI64X4(1, b0, acc)
			add(acc, t4.Y(), qd.Y(), t5.Y(), acc)

			// now we can work with acc.Y()
			f.VSHUFF64X2(0b1, acc, acc, accShuffled)
			add(acc, accShuffled, qd.Y(), t5.Y(), acc)

			f.VPSHUFD(uint64(0x4e), acc, accShuffled)
			add(acc, accShuffled, qd.Y(), t5.Y(), acc)
			f.VPSHUFD(uint64(0xb1), acc, accShuffled)
			add(acc, accShuffled, qd.Y(), t5.Y(), acc)

			f.VINSERTI64X4(1, acc, acc.Z(), acc.Z())
		})
	}

	fullRound := f.Define("full_round", 0, func(args ...any) {
		// load round keys
		f.VMOVDQU32(rKey.AtD(0), v0)
		f.VMOVDQU32(rKey.AtD(16), v1.Y())

		// add round keys
		add(b0, v0, qd, t5, b0)
		add(b1.Y(), v1.Y(), qd.Y(), t2.Y(), b1.Y())
		sbox()
		matMulExternalInPlace()
	}, true)
	if !width24 {
		fullRound = f.Define("full_round_16", 0, func(args ...any) {
			// load round keys
			f.VMOVDQU32(rKey.AtD(0), v0)

			// add round keys
			add(b0, v0, qd, t5, b0)
			sbox()
			matMulExternalInPlace()
		})
	}

	partialRound := func() {
		// load round keys
		f.VMOVD(rKey.At(0), v0.X())
		// copy b0 to break the dependency chain
		f.VMOVDQA32(b0, t4)

		add(t4.X(), v0.X(), qd.X(), PL0.X(), v1.X())

		// do the sbox
		sboxPartial()

		// merge the sbox at the first index of b0
		f.VPBLENDMD(v1, t4, t4, amd64.K2)

		// multiply b1 by diagonal[1] (diag24)
		// this is equivalent to mulY(b1, d1, t3, true)
		// but we already have d1Odd that don't change so we unroll and modify the code
		if width24 {
			f.VPSRLQ("$32", b1.Y(), aOdd.Y())
			f.VPMULUDQ(b1.Y(), d1.Y(), t0.Y())
			f.VPMULUDQ(aOdd.Y(), d1odd.Y(), t1.Y())
			f.VPMULUDQ(t0.Y(), qInvNeg.Y(), PL0.Y())
			f.VPMULUDQ(t1.Y(), qInvNeg.Y(), PL1.Y())

			f.VPMULUDQ(PL0.Y(), qd.Y(), PL0.Y())
			f.VPADDQ(t0.Y(), PL0.Y(), t0.Y())

			f.VPMULUDQ(PL1.Y(), qd.Y(), PL1.Y())
			f.VPADDQ(t1.Y(), PL1.Y(), t3.Y())

			f.VMOVSHDUPk(t0.Y(), amd64.K3, t3.Y())
			f.VPSUBD(qd.Y(), t3.Y(), t5.Y())
			f.VPMINUD(t3.Y(), t5.Y(), t3.Y())
		}

		// multiply the part of b0 that don't depend on b[0] (i.e. round keys + sbox)
		f.VPSRLQ("$32", b0, aOdd)
		f.VPMULUDQ(aOdd, d0odd, t2)
		f.VPMULUDQ(t2, qInvNeg, PL1)
		f.VPMULUDQ(PL1, qd, PL1)
		f.VPADDQ(t2, PL1, t2)

		// compute the sum: this depends on applying the sbox to b[0]
		sumState()

		// multiply the part of b0 that depends on b[0]
		f.VPMULUDQ(t4, d0, t0)
		f.VPMULUDQ(t0, qInvNeg, PL0)
		f.VPMULUDQ(PL0, qd, PL0)
		f.VPADDQ(t0, PL0, t0)

		f.VMOVSHDUPk(t0, amd64.K3, t2)
		f.VPSUBD(qd, t2, t5)
		f.VPMINUD(t2, t5, b0)

		// now we add the sum
		add(b0, acc.Z(), qd, t5, b0)
		if width24 {
			add(t3.Y(), acc.Y(), qd.Y(), v1.Y(), b1.Y())
		}
	}

	matMulExternalInPlace()

	for i := 0; i < rf; i++ {
		f.MOVQ(addrRoundKeys.At(i*3), rKey)
		fullRound()
	}

	f.Comment("loop over the partial rounds")
	{
		n := registers.Pop()
		addrRoundKeys2 := registers.Pop()
		f.MOVQ(partialRounds, n, fmt.Sprintf("nb partial rounds --> %d", partialRounds))
		lblLoop := f.NewLabel("loop")
		lblDone := f.NewLabel("done")
		f.MOVQ(addrRoundKeys, addrRoundKeys2)
		f.ADDQ(rf*24, addrRoundKeys2)

		f.LABEL(lblLoop)
		f.TESTQ(n, n)
		f.JEQ(lblDone)
		f.DECQ(n)

		f.MOVQ(addrRoundKeys2.At(0), rKey)
		partialRound()
		f.ADDQ("$24", addrRoundKeys2)

		f.JMP(lblLoop)

		f.LABEL(lblDone)
	}

	for i := rf + partialRounds; i < fullRounds+partialRounds; i++ {
		f.MOVQ(addrRoundKeys.At(i*3), rKey)
		fullRound()
	}

	f.VMOVDQU32(b0, addrInput.AtD(0))
	if width24 {
		f.VMOVDQU32(b1.Y(), addrInput.AtD(16))
	}

	f.RET()
}

func (f *FFAmd64) generatePoseidon2_F31_16x24(params Poseidon2Parameters) {
	width := params.Width
	fullRounds := params.FullRounds
	partialRounds := params.PartialRounds
	rf := fullRounds / 2

	_ = partialRounds
	_ = rf

	if width != 24 {
		panic("only width 24 is supported")
	}
	const fnName = "permutation16x24_avx512"
	// func permutation16x24_avx512(input *fr.Element, nbBlocks uint64, res *fr.Element, roundKeys [][]fr.Element)
	const argSize = 6 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 1, 0)
	registers := f.FnHeader(fnName, stackSize, argSize, amd64.AX, amd64.DX)

	addrInput := registers.Pop()
	maskFFFF := registers.Pop()
	nbBlocksExternal := registers.Pop()
	addrRes := registers.Pop()
	addrRoundKeys := registers.Pop()
	addrIndexScatter8 := registers.Pop()
	addrIndexGather512 := registers.Pop()
	rKey := registers.Pop()

	// input
	v := registers.PopVN(24)

	// constants
	qd := registers.PopV()
	qInvNeg := registers.PopV()

	// load the constants
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qd)
	f.MOVD("$const_qInvNeg", amd64.AX)
	f.VPBROADCASTD(amd64.AX, qInvNeg)

	f.MOVQ("$0xffffffffffffffff", maskFFFF)

	// prepare the masks used for shuffling the vectors
	f.MOVQ(uint64(0b01_01_01_01_01_01_01_01), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	f.MOVQ("·indexScatter8+0(SB)", addrIndexScatter8)
	f.MOVQ("·indexGather512+0(SB)", addrIndexGather512)

	f.MOVQ("input+0(FP)", addrInput)
	f.MOVQ("nbBlocks+8(FP)", nbBlocksExternal)
	f.MOVQ("res+16(FP)", addrRes)
	f.MOVQ("roundKeys+24(FP)", addrRoundKeys)

	const blockSize = 4
	const nbBlocks = 24 / blockSize

	for i := 0; i < 8; i++ {
		// zero first v[i]
		f.VXORPS(v[i], v[i], v[i])
	}

	// h.matMulExternalInPlace(input)
	// 		h.matMulM4InPlace(input)

	add, _ := f.DefineFn("add")
	reduce1Q, _ := f.DefineFn("reduce1Q")

	double := f.Define("double", 4, func(args ...any) {
		a := args[0]
		qd := args[1]
		r0 := args[2]
		into := args[3]

		f.VPSLLD("$1", a, into)
		f.VPSUBD(qd, into, r0)
		f.VPMINUD(into, r0, into)
	})

	sub := f.Define("sub", 5, func(args ...any) {
		a := args[0]
		b := args[1]
		qd := args[2]
		r0 := args[3]
		into := args[4]

		f.VPSUBD(b, a, into)
		f.VPADDD(qd, into, r0)
		f.VPMINUD(into, r0, into)
	})

	halve := func(a amd64.VectorRegister) {
		// TODO @gbotrel we can save the broadcasts
		t0 := registers.PopV()
		k4 := amd64.K4
		f.MOVQ(1, amd64.AX)
		f.VPBROADCASTD(amd64.AX, t0)

		f.VPTESTMD(a, t0, k4)
		// if a & 1 == 1 ; we add q;
		f.VPADDDk(a, qd, a, k4)
		// we shift right
		f.VPSRLD(1, a, a)

		registers.PushV(t0)
	}

	_mul := f.Define("mul_w", 8, func(args ...any) {
		a := args[0]
		b := args[1]
		aOdd := args[2]
		bOdd := args[3]
		t0 := args[4]
		t1 := args[5]
		PL0 := args[6]
		c := args[7]

		f.VPSRLQ("$32", a, aOdd) // keep high 32 bits
		f.VPSRLQ("$32", b, bOdd) // keep high 32 bits

		// VPMULUDQ conveniently ignores the high 32 bits of each QWORD lane
		f.VPMULUDQ(a, b, t0)
		f.VPMULUDQ(aOdd, bOdd, t1)

		f.VPMULUDQ(t0, qInvNeg, PL0)
		PL1 := bOdd
		f.VPMULUDQ(t1, qInvNeg, PL1)

		f.VPMULUDQ(PL0, qd, PL0)
		f.VPADDQ(t0, PL0, t0)

		f.VPMULUDQ(PL1, qd, PL1)
		f.VPADDQ(t1, PL1, c)

		f.VMOVSHDUPk(t0, amd64.K3, c)

	})

	mul := func(a, b, c amd64.VectorRegister, reduce bool) {
		aOdd := registers.PopV()
		bOdd := registers.PopV()
		t0 := registers.PopV()
		t1 := registers.PopV()
		PL0 := registers.PopV()

		_mul(a, b, aOdd, bOdd, t0, t1, PL0, c)

		if reduce {
			reduce1Q(qd, c, t0)
		}

		registers.PushV(aOdd, bOdd, t0, t1, PL0)
	}

	// Mul2ExpNegN multiplies x by -1/2^n
	//
	// Since the Montgomery constant is 2^32, the Montgomery form of 1/2^n is
	// 2^{32-n}. Montgomery reduction works provided the input is < 2^32 so this
	// works for 0 <= n <= 32.
	//
	// N.B. n must be < 33.
	// perf: see Plonky3 for a more optimized version
	_mul2ExpNegN := f.Define("mul2ExpNegN", 9, func(args ...any) {
		a := args[0]
		t0 := args[1]
		c := args[2]
		n := args[3]
		m := args[4]
		aOdd := args[5]
		t1 := args[6]
		PL0 := args[7]
		PL1 := args[8]

		f.VPSRLQ("$32", a, aOdd) // keep high 32 bits

		// VPMULUDQ conveniently ignores the high 32 bits of each QWORD lane
		// but now we must "zero out the high bits"
		// so we shift left;
		// then instead of shifting right by 32 and left by (32 - n)
		// we just shift right by n
		f.VPSLLQ("$32", a, a)
		f.VPSRLQ(n, a, t0)
		f.VPSLLQ(m, aOdd, t1)

		f.VPMULUDQ(t0, qInvNeg, PL0)
		f.VPMULUDQ(t1, qInvNeg, PL1)

		f.VPMULUDQ(PL0, qd, PL0)
		f.VPADDQ(t0, PL0, t0)

		f.VPMULUDQ(PL1, qd, PL1)
		f.VPADDQ(t1, PL1, c)

		f.VMOVSHDUPk(t0, amd64.K3, c)

		reduce1Q(qd, c, t0)
	})

	mul2ExpNegN := func(a, t0, c amd64.VectorRegister, n uint64) {
		aOdd := registers.PopV()
		t1 := registers.PopV()
		PL0 := registers.PopV()
		PL1 := registers.PopV()

		m := 32 - n

		_mul2ExpNegN(a, t0, c, "$"+strconv.Itoa(int(n)), "$"+strconv.Itoa(int(m)), aOdd, t1, PL0, PL1)

		registers.PushV(aOdd, t1, PL0, PL1)
	}

	var sbox func(a amd64.VectorRegister)
	switch params.SBoxDegree {
	case 3:
		sbox = func(a amd64.VectorRegister) {
			t0 := registers.PopV()
			mul(a, a, t0, false)
			mul(a, t0, a, true)
			registers.PushV(t0)
		}

	case 7:
		sbox = func(a amd64.VectorRegister) {
			// TODO @gbotrel not enough registers for 7 for now.
			t0 := registers.PopV()
			mul(a, a, t0, false)
			mul(a, t0, a, true)
			registers.PushV(t0)
		}
		// 	t1 := registers.PopV()
		// 	mul(a, a, t0, true)
		// 	mul(t0, t0, t1, false)
		// 	mul(a, t0, a, false)
		// 	mul(a, t1, a, true)
		// 	registers.PushV(t1)
		// }

	default:
		panic("only SBox degree 3 and 7 are supported")
	}

	matMul4 := f.Define("mat_mul_4_w", 6, func(args ...any) {
		sum := args[0]
		sd0 := args[1]
		sd1 := args[2]
		sd2 := args[3]
		sd3 := args[4]
		t0 := args[5]

		// for each block in v
		for i := 0; i < nbBlocks; i++ {
			s0 := v[4*i]
			s1 := v[4*i+1]
			s2 := v[4*i+2]
			s3 := v[4*i+3]

			// We multiply by this matrix, each block of 4:
			// (2 3 1 1)
			// (1 2 3 1)
			// (1 1 2 3)
			// (3 1 1 2)
			// so we have
			// s0 = Σ + s0 + 2s1
			// s1 = Σ + s1 + 2s2
			// s2 = Σ + s2 + 2s3
			// s3 = Σ + s3 + 2s0
			add(s0, s1, qd, t0, sum)
			add(sum, s2, qd, t0, sum)
			add(sum, s3, qd, t0, sum)

			double(s0, qd, t0, sd0)
			double(s1, qd, t0, sd1)
			double(s2, qd, t0, sd2)
			double(s3, qd, t0, sd3)

			add(s0, sum, qd, t0, s0)
			add(s0, sd1, qd, t0, s0)
			add(s1, sum, qd, t0, s1)
			add(s1, sd2, qd, t0, s1)
			add(s2, sum, qd, t0, s2)
			add(s2, sd3, qd, t0, s2)
			add(s3, sum, qd, t0, s3)
			add(s3, sd0, qd, t0, s3)
		}
	})

	_matMulExternal := f.Define("mat_mul_external_w", 6, func(args ...any) {
		sum := args[0]
		sd0 := args[1]
		sd1 := args[2]
		sd2 := args[3]
		sd3 := args[4]
		t0 := args[5]

		matMul4(sum, sd0, sd1, sd2, sd3, t0)

		tmp0 := sd0
		tmp1 := sd1
		tmp2 := sd2
		tmp3 := sd3

		add(v[0], v[4], qd, t0, tmp0)
		add(v[1], v[5], qd, t0, tmp1)
		add(v[2], v[6], qd, t0, tmp2)
		add(v[3], v[7], qd, t0, tmp3)

		for i := 2; i < nbBlocks; i++ {
			s0 := v[4*i]
			s1 := v[4*i+1]
			s2 := v[4*i+2]
			s3 := v[4*i+3]

			add(s0, tmp0, qd, t0, tmp0)
			add(s1, tmp1, qd, t0, tmp1)
			add(s2, tmp2, qd, t0, tmp2)
			add(s3, tmp3, qd, t0, tmp3)
		}

		for i := 0; i < nbBlocks; i++ {
			s0 := v[4*i]
			s1 := v[4*i+1]
			s2 := v[4*i+2]
			s3 := v[4*i+3]

			add(s0, tmp0, qd, t0, s0)
			add(s1, tmp1, qd, t0, s1)
			add(s2, tmp2, qd, t0, s2)
			add(s3, tmp3, qd, t0, s3)
		}
	})

	matMulExternal := func() {
		sum := registers.PopV()
		sd0 := registers.PopV()
		sd1 := registers.PopV()
		sd2 := registers.PopV()
		sd3 := registers.PopV()
		t0 := registers.PopV()

		_matMulExternal(sum, sd0, sd1, sd2, sd3, t0)

		registers.PushV(sum, sd0, sd1, sd2, sd3, t0)
	}

	// outer loop
	lblOuterLoop := f.NewLabel("outer_loop")
	lblOuterLoopEnd := f.NewLabel("outer_loop_end")

	f.LABEL(lblOuterLoop)
	f.TESTQ(nbBlocksExternal, nbBlocksExternal)
	f.JEQ(lblOuterLoopEnd)

	{
		vIndexGather := registers.PopV()
		f.VMOVDQU32(addrIndexGather512.At(0), vIndexGather)

		// copy (and transpose) input into v[8:24]
		for i := 8; i < 24; i++ {
			f.KMOVD(maskFFFF, amd64.K1)
			f.VPGATHERDD((i-8)*4, addrInput, vIndexGather, 4, amd64.K1, v[i])
		}
		// increment addrInput
		f.ADDQ(16*4, addrInput)

		registers.PushV(vIndexGather)
	}

	matMulExternal()

	addRoundKey := func(round, index int) {
		t0 := registers.PopV()
		f.MOVQ(addrRoundKeys.At(round*3), rKey)
		f.VPBROADCASTD(rKey.AtD(index), t0)

		add(v[index], t0, qd, t0, v[index])

		registers.PushV(t0)
	}

	for i := 0; i < rf; i++ {
		for j := range v {
			addRoundKey(i, j)
			sbox(v[j])
		}
		matMulExternal()
	}

	partialRound := func() {
		// add round key 0;
		{
			t0 := registers.PopV()
			f.VPBROADCASTD(rKey.AtD(0), t0)
			add(v[0], t0, qd, t0, v[0])
			registers.PushV(t0)
		}

		sbox(v[0])

		// h.matMulInternalInPlace(input)
		// let's do it for koalabear only for now.
		sum := registers.PopV()
		t0 := registers.PopV()
		t1 := registers.PopV()

		// TODO @gbotrel do less adds.
		add(v[0], v[1], qd, t0, sum)
		for i := 2; i < len(v); i++ {
			add(v[i], sum, qd, t0, sum)
		}

		// mul by diag24:
		// [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/4, 1/8, 1/16, 1/32, 1/64, 1/2^24, -1/2^8, -1/8, -1/16, -1/32, -1/64, -1/2^7, -1/2^9, -1/2^24]
		// var temp fr.Element
		// input[0].Sub(&sum, temp.Double(&input[0]))
		double(v[0], qd, t0, t1)
		sub(sum, t1, qd, t0, v[0])

		// input[1].Add(&sum, &input[1])
		add(sum, v[1], qd, t0, v[1])

		// input[2].Add(&sum, temp.Double(&input[2]))
		double(v[2], qd, t0, v[2])
		add(v[2], sum, qd, t0, v[2])

		// temp.Set(&input[3]).Halve()
		// input[3].Add(&sum, &temp)
		halve(v[3])
		add(sum, v[3], qd, t0, v[3])

		// input[4].Add(&sum, temp.Double(&input[4]).Add(&temp, &input[4]))
		double(v[4], qd, t0, t1)
		add(v[4], t1, qd, t0, v[4])
		add(v[4], sum, qd, t0, v[4])

		// input[5].Add(&sum, temp.Double(&input[5]).Double(&temp))
		double(v[5], qd, t0, v[5])
		add(v[5], v[5], qd, t0, v[5])
		add(v[5], sum, qd, t0, v[5])

		// temp.Set(&input[6]).Halve()
		// input[6].Sub(&sum, &temp)
		halve(v[6])
		sub(sum, v[6], qd, t0, v[6])

		// input[7].Sub(&sum, temp.Double(&input[7]).Add(&temp, &input[7]))
		double(v[7], qd, t0, t1)
		add(v[7], t1, qd, t0, v[7])
		sub(sum, v[7], qd, t0, v[7])

		// input[8].Sub(&sum, temp.Double(&input[8]).Double(&temp))
		double(v[8], qd, t0, v[8])
		double(v[8], qd, t0, v[8])
		sub(sum, v[8], qd, t0, v[8])

		registers.PushV(t1)

		ns := []uint64{8, 2, 3, 4, 5, 6, 24, 8, 3, 4, 5, 6, 7, 9, 24}

		for i := 9; i < len(v); i++ {
			mul2ExpNegN(v[i], t0, v[i], ns[i-9])
			if i <= 15 {
				add(v[i], sum, qd, t0, v[i])
			} else {
				sub(sum, v[i], qd, t0, v[i])
			}
		}

		registers.PushV(sum, t0)
	}

	f.Comment("loop over the partial rounds")
	{
		n := registers.Pop()
		addrRoundKeys2 := registers.Pop()
		f.MOVQ(partialRounds, n, fmt.Sprintf("nb partial rounds --> %d", partialRounds))
		lblLoop := f.NewLabel("loop")
		lblDone := f.NewLabel("done")
		f.MOVQ(addrRoundKeys, addrRoundKeys2)
		f.ADDQ(rf*24, addrRoundKeys2)

		f.LABEL(lblLoop)
		f.TESTQ(n, n)
		f.JEQ(lblDone)
		f.DECQ(n)

		f.MOVQ(addrRoundKeys2.At(0), rKey)

		partialRound()

		f.ADDQ("$24", addrRoundKeys2)

		f.JMP(lblLoop)

		f.LABEL(lblDone)
	}

	for i := rf + partialRounds; i < fullRounds+partialRounds; i++ {
		for j := range v {
			addRoundKey(i, j)
			sbox(v[j])
		}
		matMulExternal()
	}

	f.DECQ(nbBlocksExternal)
	f.JMP(lblOuterLoop)

	f.LABEL(lblOuterLoopEnd)

	// now we just copy the result
	// need to transpose 8x16 to 16x8
	{
		vIndexScatter := registers.PopV()
		f.VMOVDQU32(addrIndexScatter8.At(0), vIndexScatter)
		transposed := v[:8]
		for i := range transposed {
			f.KMOVD(maskFFFF, amd64.K1)
			f.VPSCATTERDD(i*4, addrRes, vIndexScatter, 4, amd64.K1, transposed[i])
		}
		registers.PushV(vIndexScatter)
	}

	f.RET()
}
