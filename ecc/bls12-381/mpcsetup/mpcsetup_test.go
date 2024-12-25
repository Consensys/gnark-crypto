// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package mpcsetup

import (
	"bytes"
	"github.com/consensys/gnark-crypto/ecc"
	curve "github.com/consensys/gnark-crypto/ecc/bls12-381"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/gnark-crypto/utils"
	"github.com/stretchr/testify/require"
	"slices"
	"testing"
)

func TestContributionPok(t *testing.T) {
	const (
		pokChallenge = "challenge"
		pokDst       = 1
	)
	x0, err := curve.HashToG1([]byte("contribution test"), nil)
	require.NoError(t, err)
	y0, err := curve.RandomOnG2()
	require.NoError(t, err)
	x1, y1 := x0, y0
	proof := UpdateValues(nil, []byte(pokChallenge), pokDst, &x1, &y1)

	representations := []ValueUpdate{
		{Previous: x0, Next: x1},
		{Previous: y0, Next: y1},
	}

	// verify proof - G1 only
	require.NoError(t, proof.Verify([]byte(pokChallenge), pokDst, representations[0]))

	// verify proof - G2 only
	require.NoError(t, proof.Verify([]byte(pokChallenge), pokDst, representations[1]))

	// verify proof - G1 and G2
	require.NoError(t, proof.Verify([]byte(pokChallenge), pokDst, representations...))

	// read/write round-trip
	var bb bytes.Buffer
	n0, err := proof.WriteTo(&bb)
	require.NoError(t, err)
	var proofBack UpdateProof
	n1, err := proofBack.ReadFrom(&bb)
	require.NoError(t, err)
	require.Equal(t, n0, n1)

	require.NoError(t, proofBack.Verify([]byte(pokChallenge), pokDst, representations[0]))
	require.NoError(t, proofBack.Verify([]byte(pokChallenge), pokDst, representations[1]))
	require.NoError(t, proofBack.Verify([]byte(pokChallenge), pokDst, representations...))
}

func TestSameRatioMany(t *testing.T) {
	_, _, g1, g2 := curve.Generators()
	g1Slice := []curve.G1Affine{g1, g1, g1}
	g2Slice := []curve.G2Affine{g2, g2}

	require.NoError(t, SameRatioMany(g1Slice, g2Slice, g1Slice, g1Slice))
	require.NoError(t, SameRatioMany(g1Slice, g1Slice[:len(g1Slice)-1], g2Slice))
}

func TestBivariateRandomMonomials(t *testing.T) {
	xDeg := []int{3, 2, 3}
	ends := utils.PartialSums(xDeg...)
	values := bivariateRandomMonomials(ends...)
	//extract the variables
	x := make([]fr.Element, slices.Max(xDeg))
	y := make([]fr.Element, len(ends))
	x[1].Div(&values[1], &values[0])
	y[1].Div(&values[xDeg[0]], &values[0])

	x[0].SetOne()
	y[0].SetOne()

	for i := range x[:len(x)-1] {
		x[i+1].Mul(&x[i], &x[1])
	}

	for i := range y[:len(x)-1] {
		y[i+1].Mul(&y[i], &y[1])
	}

	prevEnd := 0
	for i := range ends {
		for j := range xDeg[i] {
			var z fr.Element
			z.Mul(&y[i], &x[j])
			require.Equal(t, z.String(), values[prevEnd+j].String(), "X^%d Y^%d: expected %s, encountered %s", j, i)
		}
		prevEnd = ends[i]
	}
}

func TestLinearCombinationsG1(t *testing.T) {

	test := func(ends []int, powers, truncatedPowers, shiftedPowers []fr.Element, A ...curve.G1Affine) {

		multiExpConfig := ecc.MultiExpConfig{
			NbTasks: 1,
		}

		if len(A) == 0 {
			A = make([]curve.G1Affine, ends[len(ends)-1])
			var err error
			for i := range A {
				A[i], err = curve.HashToG1([]byte{byte(i)}, nil)
				require.NoError(t, err)
			}
		}

		truncated, shifted := linearCombinationsG1(slices.Clone(A), powers, ends)

		var res curve.G1Affine

		_, err := res.MultiExp(A, truncatedPowers, multiExpConfig)
		require.NoError(t, err)
		require.Equal(t, res, truncated, "truncated")

		_, err = res.MultiExp(A, shiftedPowers, multiExpConfig)
		require.NoError(t, err)
		require.Equal(t, res, shifted, "shifted")
	}

	_, _, g, _ := curve.Generators()
	var infty curve.G1Affine

	test(
		[]int{3},
		frs(1, -1, 1),
		frs(1, -1, 0),
		frs(0, 1, -1),
		infty, g, infty,
	)

	test(
		[]int{3},
		frs(1, 1, 1),
		frs(1, 1, 0),
		frs(0, 1, 1),
		infty, g, infty,
	)

	test(
		[]int{3},
		frs(1, 1, 1),
		frs(1, 1, 0),
		frs(0, 1, 1),
		infty, infty, g,
	)

	test(
		[]int{3},
		frs(1, 1, 1),
		frs(1, 1, 0),
		frs(0, 1, 1),
		g, infty, infty,
	)

	test(
		[]int{3},
		frs(1, 2, 4),
		frs(1, 2, 0),
		frs(0, 1, 2),
	)

	test(
		[]int{3, 6},
		frs(1, 1, 1, 1, 1, 1),
		frs(1, 1, 0, 1, 1, 0),
		frs(0, 1, 1, 0, 1, 1),
		g, infty, infty, infty, infty, infty,
	)

	test(
		[]int{3, 6},
		frs(1, -1, 1, 1, -1, 1),
		frs(1, -1, 0, 1, -1, 0),
		frs(0, 1, -1, 0, 1, -1),
		g, infty, infty, infty, infty, infty,
	)

	test(
		[]int{4, 7},
		frs(1, 2, 4, 8, 3, 6, 12),
		frs(1, 2, 4, 0, 3, 6, 0),
		frs(0, 1, 2, 4, 0, 3, 6),
	)
}

func TestLinearCombinationsG2(t *testing.T) {
	test := func(powers []fr.Element, A ...curve.G2Affine) {

		multiExpConfig := ecc.MultiExpConfig{
			NbTasks: 1,
		}

		if len(A) == 0 {
			A = make([]curve.G2Affine, len(powers))
			var err error
			for i := range A {
				A[i], err = curve.RandomOnG2()
				require.NoError(t, err)
			}
		}

		truncated, shifted := linearCombinationsG2(slices.Clone(A), slices.Clone(powers), []int{len(powers)})

		truncatedPowers := make([]fr.Element, len(powers))
		copy(truncatedPowers[:len(truncatedPowers)-1], powers)
		shiftedPowers := make([]fr.Element, len(powers))
		copy(shiftedPowers[1:], powers)

		var res curve.G2Affine

		_, err := res.MultiExp(A, truncatedPowers, multiExpConfig)
		require.NoError(t, err)
		require.Equal(t, res, truncated, "truncated")

		_, err = res.MultiExp(A, shiftedPowers, multiExpConfig)
		require.NoError(t, err)
		require.Equal(t, res, shifted, "shifted")
	}

	_, _, _, g := curve.Generators()
	var infty curve.G2Affine

	test(
		frs(1, 1, 1),
		g, infty, infty,
	)

	test(
		frs(1, 2, 4),
		infty, infty, g,
	)

	test(
		frs(1, -1, 1),
	)

	test(
		frs(1, 3, 9, 27, 81),
	)
}

func frs(x ...int) []fr.Element {
	res := make([]fr.Element, len(x))
	for i := range res {
		res[i].SetInt64(int64(x[i]))
	}
	return res
}
