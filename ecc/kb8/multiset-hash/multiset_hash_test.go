package multisethash

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/kb8"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/stretchr/testify/require"
)

func randomMessages(t *testing.T, n int) []extensions.E8 {
	t.Helper()
	return mustRandomMessages(n)
}

func mustRandomMessages(n int) []extensions.E8 {
	res := make([]extensions.E8, n)
	for i := range res {
		res[i].MustSetRandom()
	}
	return res
}

func TestMapDeterministic(t *testing.T) {
	msgs := randomMessages(t, 16)
	for i := range msgs {
		p1, o1, err := Map(&msgs[i])
		require.NoError(t, err)
		p2, o2, err := Map(&msgs[i])
		require.NoError(t, err)
		require.Equal(t, o1, o2)
		require.True(t, p1.Equal(&p2))
		require.True(t, p1.IsOnCurve())
		require.True(t, p1.IsInSubGroup())
	}
}

func TestHashPermutationInvariant(t *testing.T) {
	msgs := randomMessages(t, 8)
	got1, err := Hash(msgs)
	require.NoError(t, err)

	permuted := append([]extensions.E8(nil), msgs...)
	permuted[0], permuted[5] = permuted[5], permuted[0]
	permuted[1], permuted[7] = permuted[7], permuted[1]

	got2, err := Hash(permuted)
	require.NoError(t, err)
	require.True(t, got1.Equal(&got2))
}

func TestAddRemove(t *testing.T) {
	msgs := randomMessages(t, 6)
	acc := NewAccumulator()
	for i := range msgs {
		require.NoError(t, acc.Insert(&msgs[i]))
	}
	withAll := acc.Digest()
	require.True(t, withAll.IsOnCurve())

	require.NoError(t, acc.Remove(&msgs[2]))
	require.NoError(t, acc.Remove(&msgs[4]))

	acc2 := NewAccumulator()
	require.NoError(t, acc2.Insert(&msgs[0]))
	require.NoError(t, acc2.Insert(&msgs[1]))
	require.NoError(t, acc2.Insert(&msgs[3]))
	require.NoError(t, acc2.Insert(&msgs[5]))

	d1 := acc.Digest()
	d2 := acc2.Digest()
	require.True(t, d1.Equal(&d2))
}

func TestHashMatchesAccumulator(t *testing.T) {
	msgs := randomMessages(t, 10)
	got, err := Hash(msgs)
	require.NoError(t, err)

	acc := NewAccumulator()
	for i := range msgs {
		require.NoError(t, acc.Insert(&msgs[i]))
	}
	digest := acc.Digest()
	require.True(t, got.Equal(&digest))
}

func TestDuplicatesMatter(t *testing.T) {
	msgs := randomMessages(t, 1)
	single, err := Hash(msgs)
	require.NoError(t, err)

	double, err := Hash([]extensions.E8{msgs[0], msgs[0]})
	require.NoError(t, err)

	require.False(t, single.Equal(&double))
}

func BenchmarkMap(b *testing.B) {
	var msg extensions.E8
	msg.MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := Map(&msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAccumulatorInsert(b *testing.B) {
	msgs := mustRandomMessages(256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		acc := NewAccumulator()
		for j := range msgs {
			if err := acc.Insert(&msgs[j]); err != nil {
				b.Fatal(err)
			}
		}
		_ = acc.Digest()
	}
}

func BenchmarkHash256(b *testing.B) {
	msgs := mustRandomMessages(256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got, err := Hash(msgs)
		if err != nil {
			b.Fatal(err)
		}
		if !got.IsOnCurve() {
			b.Fatal("invalid digest")
		}
	}
}

func TestEmptyHashIsInfinity(t *testing.T) {
	got, err := Hash(nil)
	require.NoError(t, err)
	require.True(t, got.IsInfinity())
}

func TestMappedPointNotInfinity(t *testing.T) {
	msgs := randomMessages(t, 16)
	for i := range msgs {
		p, _, err := Map(&msgs[i])
		require.NoError(t, err)
		require.False(t, p.IsInfinity())
	}
}

func TestDigestRoundTrip(t *testing.T) {
	msgs := randomMessages(t, 4)
	got, err := Hash(msgs)
	require.NoError(t, err)
	buf := got.Bytes()
	var dec kb8.G1Affine
	_, err = dec.SetBytes(buf[:])
	require.NoError(t, err)
	require.True(t, dec.Equal(&got))
}
