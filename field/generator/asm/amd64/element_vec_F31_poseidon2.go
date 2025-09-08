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
	stackSize := f.StackSize(f.NbWords*2+4, 2, 0)
	registers := f.FnHeader(fnName, stackSize, argSize, amd64.AX, amd64.DX)
	defer f.AssertCleanStack(stackSize, 0)

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
		}, true)
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
			}, true)
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
			}, true)
		}

		sboxPartial = f.Define("sbox_partial", 0, func(args ...any) {
			mulY(v1, v1, t2, true)
			mulY(t2, t2, t3, false)
			mulY(v1, t2, v1, false)
			mulY(v1, t3, v1, true)
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
		}, true)
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
		}, true)
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
		f.MOVQ(addrRoundKeys, addrRoundKeys2)
		f.ADDQ(rf*24, addrRoundKeys2)

		f.Loop(n, func() {
			f.MOVQ(addrRoundKeys2.At(0), rKey)
			partialRound()
			f.ADDQ("$24", addrRoundKeys2)
		})
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

func (_f *FFAmd64) generatePoseidon2_F31_16x24(params Poseidon2Parameters) {
	f := &fieldHelper{FFAmd64: _f}
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
	// func permutation16x24_avx512(input *[24][16]fr.Element, roundKeys [][]fr.Element)
	const argSize = 4 * 8
	stackSize := f.StackSize(f.NbWords*2+4, 2, 0)
	registers := f.FnHeader(fnName, stackSize, argSize, amd64.AX, amd64.DX)
	defer f.AssertCleanStack(stackSize, 0)
	f.registers = &registers

	// input
	v := registers.PopVN(24)

	// constants
	f.loadQ()
	f.loadQInvNeg()

	addrInput := registers.Pop()
	addrRoundKeys := registers.Pop()
	rKey := registers.Pop()

	// prepare the mask used for the merging mul results
	f.MOVQ(uint64(0b01_01_01_01_01_01_01_01), amd64.AX)
	f.KMOVD(amd64.AX, amd64.K3)

	f.MOVQ("input+0(FP)", addrInput)
	f.MOVQ("roundKeys+8(FP)", addrRoundKeys)

	const blockSize = 4
	const nbBlocks = 24 / blockSize

	// load input
	for i := range v {
		f.VMOVDQU32(addrInput.AtD(i*16), v[i])
	}

	matMul4 := func() {
		t01 := registers.PopV()
		t23 := registers.PopV()
		t0123 := registers.PopV()
		t01123 := registers.PopV()
		t01233 := registers.PopV()

		// for each block in v
		for i := 0; i < nbBlocks; i++ {
			s0 := v[4*i]
			s1 := v[4*i+1]
			s2 := v[4*i+2]
			s3 := v[4*i+3]

			// for the addition chain, see:
			// https://github.com/Plonky3/Plonky3/blob/f91c76545cf5c4ae9182897bcc557715817bcbdc/poseidon2/src/external.rs#L43
			// for i := 0; i < c; i++ {
			// 	var t01, t23, t0123, t01123, t01233 fr.Element
			// 	t01.Add(&s[4*i], &s[4*i+1])
			// 	t23.Add(&s[4*i+2], &s[4*i+3])
			// 	t0123.Add(&t01, &t23)
			// 	t01123.Add(&t0123, &s[4*i+1])
			// 	t01233.Add(&t0123, &s[4*i+3])
			// The order here is important. Need to overwrite x[0] and x[2] after x[1] and x[3].
			// 	s[4*i+3].Double(&s[4*i]).Add(&s[4*i+3], &t01233)
			// 	s[4*i+1].Double(&s[4*i+2]).Add(&s[4*i+1], &t01123)
			// 	s[4*i].Add(&t01, &t01123)
			// 	s[4*i+2].Add(&t23, &t01233)
			// }
			f.add(s0, s1, t01)
			f.add(s2, s3, t23)
			f.add(t01, t23, t0123)
			f.add(t0123, s1, t01123)
			f.add(t0123, s3, t01233)

			f.double(s0, s3)
			f.add(s3, t01233, s3)

			f.double(s2, s1)
			f.add(s1, t01123, s1)

			f.add(t01, t01123, s0)
			f.add(t23, t01233, s2)
		}

		registers.PushV(t01, t23, t0123, t01123, t01233)
	}

	matMulExternal := func() {

		matMul4()

		tmp0 := registers.PopV()
		tmp1 := registers.PopV()
		tmp2 := registers.PopV()
		tmp3 := registers.PopV()

		f.add(v[0], v[4], tmp0)
		f.add(v[1], v[5], tmp1)
		f.add(v[2], v[6], tmp2)
		f.add(v[3], v[7], tmp3)

		for i := 2; i < nbBlocks; i++ {
			s0 := v[4*i]
			s1 := v[4*i+1]
			s2 := v[4*i+2]
			s3 := v[4*i+3]

			f.add(s0, tmp0, tmp0)
			f.add(s1, tmp1, tmp1)
			f.add(s2, tmp2, tmp2)
			f.add(s3, tmp3, tmp3)
		}

		for i := 0; i < nbBlocks; i++ {
			s0 := v[4*i]
			s1 := v[4*i+1]
			s2 := v[4*i+2]
			s3 := v[4*i+3]

			f.add(s0, tmp0, s0)
			f.add(s1, tmp1, s1)
			f.add(s2, tmp2, s2)
			f.add(s3, tmp3, s3)
		}

		registers.PushV(tmp0, tmp1, tmp2, tmp3)
	}

	sbox := func(a, into amd64.VectorRegister) {
		t5 := registers.PopV()
		f.mul(a, a, t5, false)
		f.mul(a, t5, into, true)
		registers.PushV(t5)
	}

	if params.SBoxDegree == 7 {
		sbox = func(a, into amd64.VectorRegister) {
			t5 := registers.PopV()
			t6 := registers.PopV()
			f.mul(a, a, t5, true)
			f.mul(t5, t5, t6, false)
			f.mul(a, t6, a, false)
			f.mul(a, t5, into, true)
			registers.PushV(t5, t6)
		}
	}

	addRoundKeySbox := func(index int) {
		rc := registers.PopV()
		f.VPBROADCASTD(rKey.AtD(index), rc)
		f.add(v[index], rc, v[index])
		registers.PushV(rc)
		sbox(v[index], v[index])
	}

	fullRound := func() {
		for j := range v {
			addRoundKeySbox(j)
		}
		matMulExternal()
	}

	partialRound := func() {
		addRoundKeySbox(0)

		// h.matMulInternalInPlace(input)
		// let's do it for koalabear only for now.
		sum := registers.PopV()
		t1 := registers.PopV()
		t2 := registers.PopV()
		t3 := registers.PopV()
		t4 := registers.PopV()

		{
			// compute the sum of all v[i]
			// we do it that way rather than accumulate to break some
			// dependencies chains
			f.add(v[0], v[1], t2)
			f.add(v[2], v[3], t3)
			f.add(v[4], v[5], t4)
			f.add(v[6], v[7], sum)
			for i := 8; i < len(v); i += 4 {
				f.add(v[i], t2, t2)
				f.add(v[i+1], t3, t3)
				f.add(v[i+2], t4, t4)
				f.add(v[i+3], sum, sum)
			}
			f.add(t2, t3, t2)
			f.add(t4, sum, t4)
			f.add(t2, t4, sum)

		}

		// mul by diag24:
		// koalabear:
		// [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/4, 1/8, 1/16, 1/32, 1/64, 1/2^24, -1/2^8, -1/8, -1/16, -1/32, -1/64, -1/2^7, -1/2^9, -1/2^24]
		// babybear:
		// [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/4, 1/8, 1/16, 1/2^7, 1/2^9, 1/2^27, -1/2^8, -1/4, -1/8, -1/16, -1/32, -1/64, -1/2^7, -1/2^27]
		// var temp fr.Element
		// input[0].Sub(&sum, temp.Double(&input[0]))
		f.double(v[0], v[0])

		// input[1].Add(&sum, &input[1])

		// input[2].Add(&sum, temp.Double(&input[2]))
		f.double(v[2], v[2])

		// temp.Set(&input[3]).Halve()
		// input[3].Add(&sum, &temp)
		f.halve(v[3], v[3])
		// temp.Set(&input[6]).Halve()
		// input[6].Sub(&sum, &temp)
		f.halve(v[6], v[6])

		// input[4].Add(&sum, temp.Double(&input[4]).Add(&temp, &input[4]))
		f.double(v[4], t2)
		f.add(v[4], t2, v[4])

		// input[5].Add(&sum, temp.Double(&input[5]).Double(&temp))
		f.double(v[5], v[5])
		f.double(v[5], v[5])

		// input[7].Sub(&sum, temp.Double(&input[7]).Add(&temp, &input[7]))
		f.double(v[7], t1)
		f.add(v[7], t1, v[7])

		// input[8].Sub(&sum, temp.Double(&input[8]).Double(&temp))
		f.double(v[8], v[8])
		f.double(v[8], v[8])

		registers.PushV(t1, t2, t3, t4)

		var ns []int
		if params.SBoxDegree == 3 {
			// koalabear
			ns = []int{8, 2, 3, 4, 5, 6, 24, 8, 3, 4, 5, 6, 7, 9, 24}
		} else {
			// babybear
			ns = []int{8, 2, 3, 4, 7, 9, 27, 8, 2, 3, 4, 5, 6, 7, 27}
		}

		for i := 9; i < len(v); i++ {
			f.mul2ExpNegN(v[i], ns[i-9], v[i])
		}

		// Sum part.
		f.sub(sum, v[0], v[0])
		f.add(sum, v[1], v[1])
		f.add(v[2], sum, v[2])
		f.add(v[3], sum, v[3])
		f.add(v[4], sum, v[4])
		f.add(v[5], sum, v[5])
		f.sub(sum, v[6], v[6])
		f.sub(sum, v[7], v[7])
		f.sub(sum, v[8], v[8])
		for i := 9; i < len(v); i++ {
			if i <= 15 {
				f.add(v[i], sum, v[i])
			} else {
				f.sub(sum, v[i], v[i])
			}
		}

		registers.PushV(sum)
	}

	// private function to help write for loops with known bounds
	// for the rounds
	loop := func(nbRounds int, fn func()) {

		n := registers.Pop()
		f.MOVQ(nbRounds, n)

		f.Loop(n, func() {
			// move the current round key address into rKey
			f.MOVQ(addrRoundKeys.At(0), rKey)

			fn()

			// move to the next round key
			f.ADDQ("$24", addrRoundKeys)
		})

		registers.Push(n)
	}

	matMulExternal()

	f.Comment("loop over the first full rounds")
	loop(rf, fullRound)

	f.Comment("loop over the partial rounds")
	loop(partialRounds, partialRound)

	f.Comment("loop over the final full rounds")
	loop(rf, fullRound)

	// store the result back
	for i := range v {
		f.VMOVDQU32(v[i], addrInput.AtD(i*16))
	}

	f.RET()
}

type fieldHelper struct {
	*FFAmd64
	registers   *amd64.Registers
	qd, qInvNeg amd64.VectorRegister
}

func (f *fieldHelper) loadQ() {
	f.qd = f.registers.PopV()
	f.MOVD("$const_q", amd64.AX)
	f.VPBROADCASTD(amd64.AX, f.qd)
}

func (f *fieldHelper) loadQInvNeg() {
	f.qInvNeg = f.registers.PopV()
	f.MOVD("$const_qInvNeg", amd64.AX)
	f.VPBROADCASTD(amd64.AX, f.qInvNeg)
}

type width int

const (
	fX width = iota
	fY
	fZ
)

// add a and b and store the result in into
func (f *fieldHelper) add(a, b, into amd64.VectorRegister, width ...width) {
	qd := f.qd
	r0 := f.registers.PopV()

	if len(width) > 0 {
		switch width[0] {
		case fX:
			qd = qd.X()
			r0 = r0.X()
			a = a.X()
			b = b.X()
			into = into.X()
		case fY:
			qd = qd.Y()
			r0 = r0.Y()
			a = a.Y()
			b = b.Y()
			into = into.Y()
		case fZ:
			qd = qd.Z()
			r0 = r0.Z()
			a = a.Z()
			b = b.Z()
			into = into.Z()
		default:
			panic("unknown field width")
		}
	}

	f.Define("add", 5, func(args ...any) {
		a := args[0]
		b := args[1]
		qd := args[2]
		r0 := args[3]
		into := args[4]

		f.VPADDD(b, a, into)
		f.VPSUBD(qd, into, r0)
		f.VPMINUD(into, r0, into)
	}, true)(a, b, qd, r0, into)
	f.registers.PushV(r0)
}

func (f *fieldHelper) addNoReduce(a, b, into amd64.VectorRegister) {
	f.VPADDD(b, a, into)
}

// double a and store the result in into
func (f *fieldHelper) double(a, into amd64.VectorRegister) {
	r0 := f.registers.PopV()
	f.Define("double", 4, func(args ...any) {
		a := args[0]
		qd := args[1]
		r0 := args[2]
		into := args[3]

		f.VPSLLD("$1", a, into)
		f.VPSUBD(qd, into, r0)
		f.VPMINUD(into, r0, into)
	}, true)(a, f.qd, r0, into)
	f.registers.PushV(r0)
}

// sub a and b and store the result in into
func (f *fieldHelper) sub(a, b, into amd64.VectorRegister) {
	r0 := f.registers.PopV()
	f.Define("sub", 5, func(args ...any) {
		a := args[0]
		b := args[1]
		qd := args[2]
		r0 := args[3]
		into := args[4]

		f.VPSUBD(b, a, into)
		f.VPADDD(qd, into, r0)
		f.VPMINUD(into, r0, into)
	}, true)(a, b, f.qd, r0, into)
	f.registers.PushV(r0)
}

// halve a and store the result in into
func (f *fieldHelper) halve(a, into amd64.VectorRegister) {
	ones := f.registers.PopV()

	f.Define("halve", 2, func(args ...any) {
		a := args[0]
		ones := args[1]
		f.MOVD("$1", amd64.AX)
		f.VPBROADCASTD(amd64.AX, ones)

		f.VPTESTMD(a, ones, amd64.K4)
		// if a & 1 == 1 ; we add q;
		f.VPADDDk(a, f.qd, a, amd64.K4)
		// we shift right
		f.VPSRLD(1, a, a)
	}, true)(a, ones)
	f.registers.PushV(ones)
}

func (f *fieldHelper) mul(a, b, into amd64.VectorRegister, reduce bool) {
	if f.registers.AvailableV() >= 5 {
		f.mul_5(a, b, into, reduce)
	} else {
		f.mul_4(a, b, into, reduce)
	}
}

// mul_4 a and b and store the result in into
// this version uses only 4 temporary registers
// see mul_5 for the version with 5 temporary registers
// see mul_6 for the version with 6 temporary registers
func (f *fieldHelper) mul_4(a, b, into amd64.VectorRegister, reduce bool) {
	t := f.registers.PopVN(4)

	f.Define("mul_4w", 7, func(args ...any) {
		a := args[0]
		b := args[1]
		aOdd := args[2]
		bOdd := args[3]
		t0 := args[4]
		t1 := args[5]
		c := args[6]

		PL0 := aOdd
		PL1 := bOdd

		f.VPSRLQ("$32", a, aOdd) // keep high 32 bits
		f.VPSRLQ("$32", b, bOdd) // keep high 32 bits

		// VPMULUDQ conveniently ignores the high 32 bits of each QWORD lane
		f.VPMULUDQ(a, b, t0)
		f.VPMULUDQ(aOdd, bOdd, t1)
		f.VPMULUDQ(t0, f.qInvNeg, PL0)
		f.VPMULUDQ(t1, f.qInvNeg, PL1)

		f.VPMULUDQ(PL0, f.qd, PL0)
		f.VPADDQ(t0, PL0, t0)

		f.VPMULUDQ(PL1, f.qd, PL1)
		f.VPADDQ(t1, PL1, c)

		f.VMOVSHDUPk(t0, amd64.K3, c)
	}, true)(a, b, t[0], t[1], t[2], t[3], into)
	f.registers.PushV(t...)

	if reduce {
		f.reduce1Q(into)
	}
}

func (f *fieldHelper) mul_5(a, b, into amd64.VectorRegister, reduce bool) {
	t := f.registers.PopVN(5)
	// same as mul_4, except we don't reuse aOdd for PL0

	f.Define("mul_5w", 8, func(args ...any) {
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
		f.VPMULUDQ(t0, f.qInvNeg, PL0)
		PL1 := bOdd
		f.VPMULUDQ(t1, f.qInvNeg, PL1)

		f.VPMULUDQ(PL0, f.qd, PL0)
		f.VPADDQ(t0, PL0, t0)

		f.VPMULUDQ(PL1, f.qd, PL1)
		f.VPADDQ(t1, PL1, c)

		f.VMOVSHDUPk(t0, amd64.K3, c)
	}, true)(a, b, t[0], t[1], t[2], t[3], t[4], into)
	f.registers.PushV(t...)

	if reduce {
		f.reduce1Q(into)
	}
}

// mul2ExpNegN multiplies a by -1/2^n (and reduces mod q)
// uses 5 temporary registers
func (f *fieldHelper) mul2ExpNegN(a amd64.VectorRegister, N int, into amd64.VectorRegister) {
	t := f.registers.PopVN(5)

	// Since the Montgomery constant is 2^32, the Montgomery form of 1/2^n is
	// 2^{32-n}. Montgomery reduction works provided the input is < 2^32 so this
	// works for 0 <= n <= 32.
	//
	// N.B. n must be < 33.
	// perf: see Plonky3 impl for specific N values
	// gains are minimal so keeping this generic version for simplicity of the code.

	f.Define("mul_2_exp_neg_n", 9, func(args ...any) {
		a := args[0]
		c := args[1]
		n := args[2]
		m := args[3]
		t0 := args[4]
		t1 := args[5]
		t2 := args[6]
		t3 := args[7]
		t4 := args[8]

		f.VPSRLQ("$32", a, t2) // keep high 32 bits

		// VPMULUDQ conveniently ignores the high 32 bits of each QWORD lane
		// but now we must "zero out the high bits"
		// so we shift left;
		// then instead of shifting right by 32 and left by (32 - n)
		// we just shift right by n
		f.VPSLLQ("$32", a, a)
		f.VPSRLQ(n, a, t0)
		f.VPSLLQ(m, t2, t1)

		f.VPMULUDQ(t0, f.qInvNeg, t3)
		f.VPMULUDQ(t1, f.qInvNeg, t4)

		f.VPMULUDQ(t3, f.qd, t3)
		f.VPADDQ(t0, t3, t0)

		f.VPMULUDQ(t4, f.qd, t4)
		f.VPADDQ(t1, t4, c)

		f.VMOVSHDUPk(t0, amd64.K3, c)
		f.VPSUBD(f.qd, c, t4)
		f.VPMINUD(c, t4, c)
	}, true)(a, into, "$"+strconv.Itoa(N), "$"+strconv.Itoa(32-N), t[0], t[1], t[2], t[3], t[4])
	f.registers.PushV(t...)

}

// reduce1Q reduces a by q and stores the result in into
func (f *fieldHelper) reduce1Q(a amd64.VectorRegister) {
	r0 := f.registers.PopV()
	f.Define("reduce1Q", 3, func(args ...any) {
		qd := args[0]
		c := args[1]
		r0 := args[2]

		f.VPSUBD(qd, c, r0)
		f.VPMINUD(c, r0, c)
	}, true)(f.qd, a, r0)
	f.registers.PushV(r0)
}
