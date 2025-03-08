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

	"github.com/consensys/bavard/amd64"
)

func (f *FFAmd64) generatePoseidon2_24_F31(width int) {
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

	sbox := f.Define("sbox_full", 0, func(args ...any) {
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

	sboxPartial := f.Define("sbox_partial", 0, func(args ...any) {
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

	const fullRounds = 6
	const partialRounds = 21
	const rf = fullRounds / 2

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
