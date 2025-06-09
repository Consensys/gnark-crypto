package bls12377

import (
	"crypto/rand"
	"math/big"
	"sync/atomic"

	"github.com/consensys/gnark-crypto/internal/parallel"
)

// IsInSubGroupBatchNaive checks if a batch of points P_i are in G1.
// This is a naive method that checks each point individually using Scott test
// [Scott21].
//
// [Scott21]: https://eprint.iacr.org/2021/1130.pdf
func IsInSubGroupBatchNaive(points []G1Affine) bool {
	var nbErrors int64
	parallel.Execute(len(points), func(start, end int) {
		for i := start; i < end; i++ {
			if !points[i].IsInSubGroup() {
				atomic.AddInt64(&nbErrors, 1)
				return
			}
		}
	})
	return nbErrors == 0
}

// IsInSubGroupBatch checks if a batch of points P_i are in G1.
// It generates random scalars s_i in the range [0, bound), performs
// n=rounds multi-scalar-multiplication Sj=∑[s_i]P_i of sizes N=len(points) and
// checks if Sj are on E[r] using Scott test [Scott21].
//
// [Scott21]: https://eprint.iacr.org/2021/1130.pdf
func IsInSubGroupBatch(points []G1Affine, bound *big.Int, rounds int) bool {
	// ensure bound is 2 for now.
	if !bound.IsUint64() && bound.Uint64() != 2 {
		panic("IsInSubGroupBatch only supports bound=2 for now")
	}
	// ensure rounds == 64
	const nbRounds = 64
	if rounds != nbRounds {
		panic("IsInSubGroupBatch only supports rounds=64 for now")
	}

	var nbErrors int64
	parallel.Execute(rounds, func(start, end int) {
		var sum G1Jac

		const windowSize = 64
		var br [windowSize / 8]byte

		// Check Sj are on E[r]
		for i := start; i < end; i++ {
			for j := range len(points) {
				pos := j % windowSize
				if pos == 0 {
					// re sample the random bytes every windowSize points
					// as per the doc:
					// Read fills b with cryptographically secure random bytes. It never returns an error, and always fills b entirely.
					rand.Read(br[:])
				}
				// check if the bit is set
				if br[pos/8]&(1<<(pos%8)) != 0 {
					// add the point to the sum
					sum.AddMixed(&points[j])
				}
			}
		}
		if !sum.IsInSubGroup() {
			atomic.AddInt64(&nbErrors, 1)
		}
	})

	return nbErrors == 0

}
