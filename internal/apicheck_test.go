package main

import (
	"github.com/consensys/gurvy/bls377"
	"github.com/consensys/gurvy/bls381"
	"github.com/consensys/gurvy/bn256"
	"github.com/consensys/gurvy/bw761"
)

// note: pairing API is not code generated, and don't use interfaces{} for performance reasons
// we end up having some API disparities -- this section ensures that we don't forget to update some APIs

var err error

var (
	gtbls377 bls377.GT
	gtbls381 bls381.GT
	gtbn256  bn256.GT
	gtbw761  bw761.GT
)

func init() {
	// Pair
	gtbls377, err = bls377.Pair([]bls377.G1Affine{}, []bls377.G2Affine{})
	gtbls381, err = bls381.Pair([]bls381.G1Affine{}, []bls381.G2Affine{})
	gtbn256, err = bn256.Pair([]bn256.G1Affine{}, []bn256.G2Affine{})
	gtbw761, err = bw761.Pair([]bw761.G1Affine{}, []bw761.G2Affine{})

	// MillerLoop
	gtbls377, err = bls377.MillerLoop([]bls377.G1Affine{}, []bls377.G2Affine{})
	gtbls381, err = bls381.MillerLoop([]bls381.G1Affine{}, []bls381.G2Affine{})
	gtbn256, err = bn256.MillerLoop([]bn256.G1Affine{}, []bn256.G2Affine{})
	gtbw761, err = bw761.MillerLoop([]bw761.G1Affine{}, []bw761.G2Affine{})

	// FinalExp
	gtbls377 = bls377.FinalExponentiation(&gtbls377)
	gtbls381 = bls381.FinalExponentiation(&gtbls381)
	gtbn256 = bn256.FinalExponentiation(&gtbn256)
	gtbw761 = bw761.FinalExponentiation(&gtbw761)
}
