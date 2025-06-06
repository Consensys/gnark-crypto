package bls12381

import (
	"crypto/rand"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
)

// IsInSubGroupBatchNaive checks if a batch of points P_i are in G1.
// This is a naive method that checks each point individually.
func IsInSubGroupBatchNaive(points []G1Affine) bool {
	for i := range points {
		if !points[i].IsInSubGroup() {
			return false
		}
	}
	return true
}

// isFirstTateOne...
func isFirstTateOne(point G1Affine) bool {
	var tate, two fp.Element
	two.SetInt64(2)
	tate.Sub(&point.Y, &two).Exp(tate, &exp1)
	return tate.IsOne()
}

// isSecondTateOne...

// IsInSubGroupBatch checks if a batch of points P_i are in G1.
// It generates random scalars s_i in the range [0, bound) and performs
// n=rounds multi-scalar-multiplication ∑[s_i]P_i of sizes N=len(points)
func IsInSubGroupBatch(points []G1Affine, bound *big.Int, rounds int) bool {

	// 1. Check points are on E[r*e']
	// 1.1. Tate_{3,P3}(Q) = (y-2)^((p-1)/3) == 1, with P3 = (0,2).
	for i := range points {
		if !isFirstTateOne(points[i]) {
			return false
		}
	}
	// 1.2. Tate_{11,P11}(Q) == 1

	// 2. Check Sj are on E[r]
	for i := 0; i < rounds; i++ {
		b, err := rand.Int(rand.Reader, bound)
		if err != nil {
			panic(err)
		}
		randoms := make([]fr.Element, len(points))
		for j := range randoms {
			randoms[j].SetBigInt(b)
		}
		var sum G1Jac
		sum.MultiExp(points[:], randoms[:], ecc.MultiExpConfig{})
		if !sum.IsInSubGroup() {
			return false
		}
	}
	return true
}
