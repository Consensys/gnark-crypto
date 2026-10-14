package hash

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/consensys/gnark-crypto/field/mamabear"
	"github.com/consensys/gnark-crypto/field/mamabear/extensions"
	"github.com/consensys/gnark-crypto/field/mamabear/poseidon2"
)

// TestBaseFieldHashAdapterE3OverPoseidon2 checks that, over the Poseidon2 hasher,
// the adapter is transparent: its digest is that of the hasher fed the
// coordinates of the elements in a single call, whatever the split among the calls.
func TestBaseFieldHashAdapterE3OverPoseidon2(t *testing.T) {
	for n := range 12 {
		es := make([]extensions.E3, n)
		var coords []mamabear.Element
		for i := range es {
			es[i].MustSetRandom()
			coords = es[i].AppendBase(coords)
		}

		ref := poseidon2.NewMerkleDamgardHasher()
		ref.WriteField(coords...)
		var want extensions.E3
		want.SetBase(ref.SumSmallField()[:3])

		for split := range n + 1 {
			h := NewBaseFieldHashAdapterE3(poseidon2.NewMerkleDamgardHasher())
			h.WriteField(es[:split]...)
			h.WriteField(es[split:]...)
			got := h.SumField()
			require.True(t, got.Equal(&want), "%d elements split at %d", n, split)
		}
	}
}
