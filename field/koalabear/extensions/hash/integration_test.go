package hash

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
)

// TestBaseFieldHashAdapterE6OverPoseidon2 checks that, over the Poseidon2 hasher,
// the adapter is transparent: its digest is that of the hasher fed the
// coordinates of the elements in a single call, whatever the split among the calls.
func TestBaseFieldHashAdapterE6OverPoseidon2(t *testing.T) {
	for n := range 12 {
		es := make([]extensions.E6, n)
		var coords []koalabear.Element
		for i := range es {
			es[i].MustSetRandom()
			coords = es[i].AppendBase(coords)
		}

		ref := poseidon2.NewMerkleDamgardHasher()
		ref.WriteField(coords...)
		var want extensions.E6
		want.SetBase(ref.SumSmallField()[:6])

		for split := range n + 1 {
			h := NewBaseFieldHashAdapterE6(poseidon2.NewMerkleDamgardHasher())
			h.WriteField(es[:split]...)
			h.WriteField(es[split:]...)
			got := h.SumField()
			require.True(t, got.Equal(&want), "%d elements split at %d", n, split)
		}
	}
}
