// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package mpcsetup

import (
	"bytes"
	"errors"
	"github.com/consensys/gnark-crypto/ecc"
	curve "github.com/consensys/gnark-crypto/ecc/bls24-315"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr"
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
	if !x.contributionCommitment.IsInfinity() {
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

// BeaconContributions provides the last
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

func partialSums(s ...int) []int {
	if len(s) == 0 {
		return nil
	}
	sums := make([]int, len(s))
	sums[0] = s[0]
	for i := 1; i < len(s); i++ {
		sums[i] = sums[i-1] + s[i]
	}
	return sums
}

// Returns [1, a, a², ..., aᴺ⁻¹ ] for random a
func randomMonomials(N int) []fr.Element {
	return bivariateRandomMonomials(N)
}

// Returns [1, a, a², ..., aᴺ⁻¹ ]
func powers(a *fr.Element, N int) []fr.Element {

	result := make([]fr.Element, N)
	if N >= 1 {
		result[0].SetOne()
	}
	if N >= 2 {
		result[1].Set(a)
	}
	for i := 2; i < N; i++ {
		result[i].Mul(&result[i-1], a)
	}
	return result
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

func linearCombinationG1(A []curve.G1Affine, r []fr.Element) curve.G1Affine {
	var res curve.G1Affine
	if _, err := res.MultiExp(A, r[:len(A)], ecc.MultiExpConfig{NbTasks: runtime.NumCPU()}); err != nil {
		panic(err)
	}
	return res
}

func linearCombinationG2(A []curve.G2Affine, r []fr.Element) curve.G2Affine {
	var res curve.G2Affine
	if _, err := res.MultiExp(A, r[:len(A)], ecc.MultiExpConfig{NbTasks: runtime.NumCPU()}); err != nil {
		panic(err)
	}
	return res
}

func (x *UpdateProof) WriteTo(writer io.Writer) (n int64, err error) {
	enc := curve.NewEncoder(writer)
	if err = enc.Encode(&x.contributionCommitment); err != nil {
		return enc.BytesWritten(), err
	}
	err = enc.Encode(&x.contributionPok)
	return enc.BytesWritten(), err
}

func (x *UpdateProof) ReadFrom(reader io.Reader) (n int64, err error) {
	dec := curve.NewDecoder(reader)
	if err = dec.Decode(&x.contributionCommitment); err != nil {
		return dec.BytesRead(), err
	}
	err = dec.Decode(&x.contributionPok)
	return dec.BytesRead(), err
}
