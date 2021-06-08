package bls24315

import (
	"crypto/rand"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls24-315/fp"
	"github.com/leanovate/gopter"
)

// GenBigInt generates a big.Int
func GenBigInt() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var s big.Int
		var b [fp.Bytes]byte
		_, err := rand.Read(b[:])
		if err != nil {
			panic(err)
		}
		s.SetBytes(b[:])
		genResult := gopter.NewGenResult(s, gopter.NoShrinker)
		return genResult
	}
}
