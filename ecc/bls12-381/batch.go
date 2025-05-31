package bls12381

import (
	"crypto/rand"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
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

// IsInSubGroupBatch checks if a batch of points P_i are in G1.
// It generates random scalars s_i in the range [0, bound) and performs
// n=rounds multi-scalar-multiplication ∑[s_i]P_i of sizes N=len(points)
func IsInSubGroupBatch(points []G1Affine, bound *big.Int, rounds int) bool {
	if len(points) == 0 {
		return true
	}
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
