// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package mpcsetup

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	curve "github.com/consensys/gnark-crypto/ecc/bw6-633"
	"github.com/consensys/gnark-crypto/ecc/bw6-633/fr"
	"github.com/consensys/gnark-crypto/utils"
	"io"
	"math/big"
	"runtime"
)

// Generate R∈𝔾₂ as Hash(gˢ, challenge, dst)
// it is to be used as a challenge for generating a proof of knowledge to x
// π ≔ x.r; e([1]₁, π) =﹖ e([x]₁, r)
func pokBase(xG curve.G1Affine, challenge []byte, dst byte) curve.G2Affine {
	var buf bytes.Buffer
	buf.Grow(len(challenge) + curve.SizeOfG1AffineUncompressed)
	buf.Write(xG.Marshal())
	buf.Write(challenge)
	xpG2, err := curve.HashToG2(buf.Bytes(), []byte{dst})
	if err != nil {
		panic(err)
	}
	return xpG2
}

type UpdateProof struct {
	contributionCommitment curve.G1Affine // x or [Xⱼ]₁
	contributionPok        curve.G2Affine // π ≔ x.r ∈ 𝔾₂
}

type ValueUpdate struct {
	Previous, Next any
}

// UpdateValues scales g1 and g2 representations by the given contribution value and provides a proof of update correctness.
// If the provided contribution value is zero, it will be randomized.
func UpdateValues(contributionValue *fr.Element, challenge []byte, dst byte, representations ...any) UpdateProof {
	if contributionValue == nil {
		contributionValue = new(fr.Element)
	}
	if contributionValue.IsZero() {
		if _, err := contributionValue.SetRandom(); err != nil {
			panic(err)
		}
	}

	var contributionValueI big.Int
	contributionValue.BigInt(&contributionValueI)

	var proof UpdateProof
	_, _, gen1, _ := curve.Generators()
	proof.contributionCommitment.ScalarMultiplication(&gen1, &contributionValueI)

	for _, repr := range representations {
		switch r := repr.(type) {
		case *curve.G1Affine:
			r.ScalarMultiplication(r, &contributionValueI)
		case *curve.G2Affine:
			r.ScalarMultiplication(r, &contributionValueI)
		case []curve.G1Affine:
			for i := range r {
				r[i].ScalarMultiplication(&r[i], &contributionValueI)
			}
		case []curve.G2Affine:
			for i := range r {
				r[i].ScalarMultiplication(&r[i], &contributionValueI)
			}
		default:
			panic("unsupported type")
		}
	}

	// proof of knowledge to commitment. Algorithm 3 from section 3.7
	pokBase := pokBase(proof.contributionCommitment, challenge, dst) // r
	proof.contributionPok.ScalarMultiplication(&pokBase, &contributionValueI)

	return proof
}

// Verify that the updates to representations are consistent with the contribution in x.
// Verify does not subgroup check the representations.
func (x *UpdateProof) Verify(challenge []byte, dst byte, representations ...ValueUpdate) error {
	if !x.contributionCommitment.IsInSubGroup() || !x.contributionPok.IsInSubGroup() {
		return errors.New("proof subgroup check failed")
	}
	if x.contributionCommitment.IsInfinity() {
		return errors.New("zero contribution not allowed")
	}

	var g1Len, g2Len int
	for i := range representations {
		switch r := representations[i].Previous.(type) {
		case curve.G1Affine:
			g1Len++
		case *curve.G1Affine:
			g1Len++
		case curve.G2Affine:
			g2Len++
		case *curve.G2Affine:
			g2Len++
		case []curve.G1Affine:
			g1Len += len(r)
		case []curve.G2Affine:
			g2Len += len(r)
		default:
			return errors.New("unsupported type")
		}
	}

	g1Prev := make([]curve.G1Affine, 0, g1Len)
	g2Prev := make([]curve.G2Affine, 0, g2Len)
	g1Next := make([]curve.G1Affine, 0, g1Len)
	g2Next := make([]curve.G2Affine, 0, g2Len)
	for i := range representations {
		switch r := representations[i].Previous.(type) {
		case curve.G1Affine:
			g1Prev = append(g1Prev, r)
			g1Next = append(g1Next, representations[i].Next.(curve.G1Affine))
		case *curve.G1Affine:
			g1Prev = append(g1Prev, *r)
			g1Next = append(g1Next, *representations[i].Next.(*curve.G1Affine))
		case curve.G2Affine:
			g2Prev = append(g2Prev, r)
			g2Next = append(g2Next, representations[i].Next.(curve.G2Affine))
		case *curve.G2Affine:
			g2Prev = append(g2Prev, *r)
			g2Next = append(g2Next, *representations[i].Next.(*curve.G2Affine))
		case []curve.G1Affine:
			g1Prev = append(g1Prev, r...)
			g1Next = append(g1Next, representations[i].Next.([]curve.G1Affine)...)
		case []curve.G2Affine:
			g2Prev = append(g2Prev, r...)
			g2Next = append(g2Next, representations[i].Next.([]curve.G2Affine)...)
		default:
			return errors.New("unsupported type")
		}

		if len(g1Prev) != len(g1Next) || len(g2Prev) != len(g2Next) {
			return errors.New("length mismatch")
		}
	}

	r := randomMonomials(max(g1Len, g2Len))

	pokBase := pokBase(x.contributionCommitment, challenge, dst)

	_, _, g1, _ := curve.Generators()
	if !sameRatio(x.contributionCommitment, g1, x.contributionPok, pokBase) { // π =? x.r i.e. x/g1 =? π/r
		return errors.New("contribution proof of knowledge verification failed")
	}

	if g1Len > 0 {
		// verify G1 representations update
		prev := linearCombinationG1(g1Prev, r)
		next := linearCombinationG1(g1Next, r)
		if !sameRatio(next, prev, x.contributionPok, pokBase) {
			return errors.New("g1 update inconsistent")
		}
	}

	if g2Len > 0 {
		// verify G2 representations update
		prev := linearCombinationG2(g2Prev, r)
		next := linearCombinationG2(g2Next, r)
		if !sameRatio(x.contributionCommitment, g1, next, prev) {
			return errors.New("g2 update inconsistent")
		}
	}

	return nil
}

// BeaconContributions provides a reproducible slice of field elements
// used for the final update in a multiparty setup ceremony.
// beaconChallenge is a publicly checkable value at time t, of
// moderate entropy to any party before time t.
func BeaconContributions(hash, dst, beaconChallenge []byte, n int) []fr.Element {
	var (
		bb  bytes.Buffer
		err error
	)
	bb.Grow(len(hash) + len(beaconChallenge))
	bb.Write(hash)
	bb.Write(beaconChallenge)

	res := make([]fr.Element, 1)

	allNonZero := func() bool {
		for i := range res {
			if res[i].IsZero() {
				return false
			}
		}
		return true
	}

	// cryptographically unlikely for this to be run more than once
	for !allNonZero() {
		if res, err = fr.Hash(bb.Bytes(), dst, n); err != nil {
			panic(err)
		}
		bb.WriteByte('=') // padding just so that the hash is different next time
	}

	return res
}

// bivariateRandomMonomials returns 1, x, ..., x^{ends[0]-1}; y, xy, ..., x^{ends[1]-ends[0]-1}y; ...
// all concatenated in the same slice
func bivariateRandomMonomials(ends ...int) []fr.Element {
	if len(ends) == 0 {
		return nil
	}

	res := make([]fr.Element, ends[len(ends)-1])
	if _, err := res[1].SetRandom(); err != nil {
		panic(err)
	}
	setPowers(res[:ends[0]])

	if len(ends) == 1 {
		return res
	}

	y := make([]fr.Element, len(ends))
	if _, err := y[1].SetRandom(); err != nil {
		panic(err)
	}
	setPowers(y)

	for d := 1; d < len(ends); d++ {
		xdeg := ends[d] - ends[d-1]
		if xdeg > ends[0] {
			panic("impl detail: first maximum degree for x must be the greatest")
		}

		for i := range xdeg {
			res[ends[d-1]+i].Mul(&res[i], &y[d])
		}
	}

	return res
}

// sets x[i] = x[1]ⁱ
func setPowers(x []fr.Element) {
	if len(x) == 0 {
		return
	}
	x[0].SetOne()
	for i := 2; i < len(x); i++ {
		x[i].Mul(&x[i-1], &x[1])
	}
}

// Returns [1, a, a², ..., aᴺ⁻¹ ] for random a
func randomMonomials(N int) []fr.Element {
	switch N {
	case 0:
		return nil
	case 1:
		return []fr.Element{fr.One()}
	}
	return bivariateRandomMonomials(N)
}

// Check n₁/d₁ = n₂/d₂ i.e. e(n₁, d₂) = e(d₁, n₂). No subgroup checks.
func sameRatio(n1, d1 curve.G1Affine, n2, d2 curve.G2Affine) bool {
	var nd1 curve.G1Affine
	nd1.Neg(&d1)
	res, err := curve.PairingCheck(
		[]curve.G1Affine{n1, nd1},
		[]curve.G2Affine{d2, n2})
	if err != nil {
		panic(err)
	}
	return res
}

// WriteTo implements io.WriterTo
func (x *UpdateProof) WriteTo(writer io.Writer) (n int64, err error) {
	enc := curve.NewEncoder(writer)
	if err = enc.Encode(&x.contributionCommitment); err != nil {
		return enc.BytesWritten(), err
	}
	err = enc.Encode(&x.contributionPok)
	return enc.BytesWritten(), err
}

// ReadFrom implements io.ReaderFrom
func (x *UpdateProof) ReadFrom(reader io.Reader) (n int64, err error) {
	dec := curve.NewDecoder(reader)
	if err = dec.Decode(&x.contributionCommitment); err != nil {
		return dec.BytesRead(), err
	}
	err = dec.Decode(&x.contributionPok)
	return dec.BytesRead(), err
}

// SameRatioMany proves that all input slices
// are geometric sequences with the same ratio.
func SameRatioMany(slices ...any) error {

	var longest1, longest2, longestLen1, longestLen2 int
	g1 := make([][]curve.G1Affine, 0, len(slices))
	g2 := make([][]curve.G2Affine, 0, len(slices))

	for _, s := range slices {
		switch r := s.(type) {
		case []curve.G1Affine:
			if len(r) > longestLen1 {
				longest1 = len(g1)
				longestLen1 = len(r)
			}
			g1 = append(g1, r)
		case []curve.G2Affine:
			if len(r) > longestLen2 {
				longest2 = len(g2)
				longestLen2 = len(r)
			}
			g2 = append(g2, r)
		default:
			return fmt.Errorf("unsupported type %T", s)
		}
	}

	if len(g1) == 0 || len(g2) == 0 {
		return errors.New("need both G1 and G2 representatives")
	}

	// make sure the longest progression is first
	g1[0], g1[longest1] = g1[longest1], g1[0]
	g2[0], g2[longest2] = g2[longest2], g2[0]

	ends1 := utils.PartialSumsF(len(g1), func(i int) int { return len(g1[i]) })
	ends2 := utils.PartialSumsF(len(g2), func(i int) int { return len(g2[i]) })

	r1 := bivariateRandomMonomials(ends1...)
	r2 := bivariateRandomMonomials(ends2...)

	g1Flat := make([]curve.G1Affine, 0, ends1[len(ends1)-1])
	for i := range g1 {
		g1Flat = append(g1Flat, g1[i]...)
	}

	g2Flat := make([]curve.G2Affine, 0, ends2[len(ends2)-1])
	for i := range g2 {
		g2Flat = append(g2Flat, g2[i]...)
	}

	truncated1, shifted1 := linearCombinationsG1(g1Flat, r1, ends1)
	truncated2, shifted2 := linearCombinationsG2(g2Flat, r2, ends2)

	if !sameRatio(truncated1, shifted1, truncated2, shifted2) {
		return errors.New("pairing mismatch")
	}
	return nil
}

// UpdateMonomialsG1 A[i] <- r^i.A[i]
func UpdateMonomialsG1(A []curve.G1Affine, r *fr.Element) {
	var (
		rExp fr.Element
		I    big.Int
	)
	r.BigInt(&I)
	A[1].ScalarMultiplication(&A[1], &I)
	rExp.Mul(r, r)
	for i := 2; i < len(A); i++ {
		rExp.BigInt(&I)
		if i+1 != len(A) {
			rExp.Mul(&rExp, r)
		}
		A[i].ScalarMultiplication(&A[i], &I)
	}
}

// linearCombinationsG1 returns
//
//		powers[0].A[0] + powers[1].A[1] + ... + powers[ends[0]-2].A[ends[0]-2]
//	  + powers[ends[0]].A[ends[0]] + ... + powers[ends[1]-2].A[ends[1]-2]
//	    ....       (truncated)
//
//		powers[0].A[1] + powers[1].A[2] + ... + powers[ends[0]-2].A[ends[0]-1]
//	  + powers[ends[0]].A[ends[0]+1]  + ... + powers[ends[1]-2].A[ends[1]-1]
//	    ....       (shifted)
//
// It is assumed without checking that powers[i+1] = powers[i]*powers[1] unless i+1 is a partial sum of sizes.
// Also assumed that powers[0] = 1.
// The slices powers and A will be modified
func linearCombinationsG1(A []curve.G1Affine, powers []fr.Element, ends []int) (truncated, shifted curve.G1Affine) {
	if ends[len(ends)-1] != len(A) || len(A) != len(powers) {
		panic("lengths mismatch")
	}

	if len(ends) == 1 && ends[0] == 2 {
		truncated, shifted = A[0], A[1]
		return
	}

	// zero out the large coefficients
	for i := range ends {
		powers[ends[i]-1].SetZero()
	}

	msmCfg := ecc.MultiExpConfig{NbTasks: runtime.NumCPU()}

	if _, err := truncated.MultiExp(A, powers, msmCfg); err != nil {
		panic(err)
	}

	var rInvNeg fr.Element
	rInvNeg.Inverse(&powers[1])
	rInvNeg.Neg(&rInvNeg)
	prevEnd := 0

	// r⁻¹.truncated =
	//		r⁻¹.powers[0].A[0] + powers[0].A[1] + ... + powers[ends[0]-3].A[ends[0]-2]
	//	  + r⁻¹.powers[ends[0]].A[ends[0]] + ... + powers[ends[1]-3].A[ends[1]-2]
	//	    ...
	//
	// compute shifted as
	//    - r⁻¹.powers[0].A[0] - r⁻¹.powers[ends[0]].A[ends[0]] - ...
	//    + powers[ends[0]-2].A[ends[0]-1] + powers[ends[1]-2].A[ends[1]-1] + ...
	//    + r⁻¹.truncated
	for i := range ends {
		powers[2*i].Mul(&powers[prevEnd], &rInvNeg)
		powers[2*i+1] = powers[ends[i]-2]
		A[2*i] = A[prevEnd]
		A[2*i+1] = A[ends[i]-1]
		prevEnd = ends[i]
	}
	powers[2*len(ends)].Neg(&rInvNeg) // r⁻¹: coefficient for truncated
	A[2*len(ends)] = truncated

	// TODO @Tabaie O(1) MSM worth it?
	if _, err := shifted.MultiExp(A[:2*len(ends)+1], powers[:2*len(ends)+1], msmCfg); err != nil {
		panic(err)
	}

	return
}

// linearCombinationG1 returns ∑ᵢ A[i].r[i]
func linearCombinationG1(A []curve.G1Affine, r []fr.Element) curve.G1Affine {
	var res curve.G1Affine
	if _, err := res.MultiExp(A, r[:len(A)], ecc.MultiExpConfig{NbTasks: runtime.NumCPU()}); err != nil {
		panic(err)
	}
	return res
}

// UpdateMonomialsG2 A[i] <- r^i.A[i]
func UpdateMonomialsG2(A []curve.G1Affine, r *fr.Element) {
	var (
		rExp fr.Element
		I    big.Int
	)
	r.BigInt(&I)
	A[1].ScalarMultiplication(&A[1], &I)
	rExp.Mul(r, r)
	for i := 2; i < len(A); i++ {
		rExp.BigInt(&I)
		if i+1 != len(A) {
			rExp.Mul(&rExp, r)
		}
		A[i].ScalarMultiplication(&A[i], &I)
	}
}

// linearCombinationsG2 returns
//
//		powers[0].A[0] + powers[1].A[1] + ... + powers[ends[0]-2].A[ends[0]-2]
//	  + powers[ends[0]].A[ends[0]] + ... + powers[ends[1]-2].A[ends[1]-2]
//	    ....       (truncated)
//
//		powers[0].A[1] + powers[1].A[2] + ... + powers[ends[0]-2].A[ends[0]-1]
//	  + powers[ends[0]].A[ends[0]+1]  + ... + powers[ends[1]-2].A[ends[1]-1]
//	    ....       (shifted)
//
// It is assumed without checking that powers[i+1] = powers[i]*powers[1] unless i+1 is a partial sum of sizes.
// Also assumed that powers[0] = 1.
// The slices powers and A will be modified
func linearCombinationsG2(A []curve.G2Affine, powers []fr.Element, ends []int) (truncated, shifted curve.G2Affine) {
	if ends[len(ends)-1] != len(A) || len(A) != len(powers) {
		panic("lengths mismatch")
	}

	if len(ends) == 1 && ends[0] == 2 {
		truncated, shifted = A[0], A[1]
		return
	}

	// zero out the large coefficients
	for i := range ends {
		powers[ends[i]-1].SetZero()
	}

	msmCfg := ecc.MultiExpConfig{NbTasks: runtime.NumCPU()}

	if _, err := truncated.MultiExp(A, powers, msmCfg); err != nil {
		panic(err)
	}

	var rInvNeg fr.Element
	rInvNeg.Inverse(&powers[1])
	rInvNeg.Neg(&rInvNeg)
	prevEnd := 0

	// r⁻¹.truncated =
	//		r⁻¹.powers[0].A[0] + powers[0].A[1] + ... + powers[ends[0]-3].A[ends[0]-2]
	//	  + r⁻¹.powers[ends[0]].A[ends[0]] + ... + powers[ends[1]-3].A[ends[1]-2]
	//	    ...
	//
	// compute shifted as
	//    - r⁻¹.powers[0].A[0] - r⁻¹.powers[ends[0]].A[ends[0]] - ...
	//    + powers[ends[0]-2].A[ends[0]-1] + powers[ends[1]-2].A[ends[1]-1] + ...
	//    + r⁻¹.truncated
	for i := range ends {
		powers[2*i].Mul(&powers[prevEnd], &rInvNeg)
		powers[2*i+1] = powers[ends[i]-2]
		A[2*i] = A[prevEnd]
		A[2*i+1] = A[ends[i]-1]
		prevEnd = ends[i]
	}
	powers[2*len(ends)].Neg(&rInvNeg) // r⁻¹: coefficient for truncated
	A[2*len(ends)] = truncated

	// TODO @Tabaie O(1) MSM worth it?
	if _, err := shifted.MultiExp(A[:2*len(ends)+1], powers[:2*len(ends)+1], msmCfg); err != nil {
		panic(err)
	}

	return
}

// linearCombinationG2 returns ∑ᵢ A[i].r[i]
func linearCombinationG2(A []curve.G2Affine, r []fr.Element) curve.G2Affine {
	var res curve.G2Affine
	if _, err := res.MultiExp(A, r[:len(A)], ecc.MultiExpConfig{NbTasks: runtime.NumCPU()}); err != nil {
		panic(err)
	}
	return res
}
