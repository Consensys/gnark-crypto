package hash_test

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/poseidon2"
)

func TestMerkleDamgardPadding(t *testing.T) {

	var (
		in       [fr.Bytes * 4]byte
		state, x fr.Element
	)
	for i := range len(in) / 2 {
		in[2*i] = byte(i / 256)
		in[2*i+1] = byte(i % 256)
	}

	const (
		pos1  = fr.Bytes
		pos2  = fr.Bytes + fr.Bytes/2 - 1
		pos3  = 2*fr.Bytes + fr.Bytes/2 + 5
		pos4  = 3*fr.Bytes + fr.Bytes/2 + 3
		pos54 = 4 * fr.Bytes
	)

	poseidon2.NewMerkleDamgardHasher()
}
