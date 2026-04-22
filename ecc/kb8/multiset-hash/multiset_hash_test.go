package multisethash

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/kb8"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/stretchr/testify/require"
)

func sampleMessages(n int) []uint16 {
	res := make([]uint16, n)
	var x uint32 = 1
	for i := range res {
		x = 1664525*x + 1013904223
		res[i] = uint16(x)
	}
	return res
}

func expectedY(msg uint16, offset uint8) extensions.E8 {
	var y extensions.E8
	y.C0.B0.A0.SetUint64(uint64(msg)*tweakBound + uint64(offset))
	return y
}

func TestMapDeterministic(t *testing.T) {
	for _, msg := range sampleMessages(32) {
		p1, o1, err := Map(msg)
		require.NoError(t, err)
		p2, o2, err := Map(msg)
		require.NoError(t, err)
		wantY := expectedY(msg, o1)
		require.Equal(t, o1, o2)
		require.True(t, p1.Equal(&p2))
		require.True(t, p1.IsOnCurve())
		require.True(t, p1.IsInSubGroup())
		require.False(t, p1.IsInfinity())
		require.True(t, p1.Y.Equal(&wantY))
	}
}

func TestHashPermutationInvariant(t *testing.T) {
	msgs := sampleMessages(8)
	got1, err := Hash(msgs)
	require.NoError(t, err)

	permuted := append([]uint16(nil), msgs...)
	permuted[0], permuted[5] = permuted[5], permuted[0]
	permuted[1], permuted[7] = permuted[7], permuted[1]

	got2, err := Hash(permuted)
	require.NoError(t, err)
	require.True(t, got1.Equal(&got2))
}

func TestInsertRemove(t *testing.T) {
	msgs := sampleMessages(6)
	acc := NewAccumulator()
	for _, msg := range msgs {
		require.NoError(t, acc.Insert(msg))
	}
	withAll := acc.Digest()
	require.True(t, withAll.IsOnCurve())

	require.NoError(t, acc.Remove(msgs[2]))
	require.NoError(t, acc.Remove(msgs[4]))

	acc2 := NewAccumulator()
	require.NoError(t, acc2.Insert(msgs[0]))
	require.NoError(t, acc2.Insert(msgs[1]))
	require.NoError(t, acc2.Insert(msgs[3]))
	require.NoError(t, acc2.Insert(msgs[5]))

	d1 := acc.Digest()
	d2 := acc2.Digest()
	require.True(t, d1.Equal(&d2))
}

func TestHashMatchesAccumulator(t *testing.T) {
	msgs := sampleMessages(10)
	got, err := Hash(msgs)
	require.NoError(t, err)

	acc := NewAccumulator()
	for _, msg := range msgs {
		require.NoError(t, acc.Insert(msg))
	}
	digest := acc.Digest()
	require.True(t, got.Equal(&digest))
}

func TestDuplicatesMatter(t *testing.T) {
	msg := sampleMessages(1)[0]
	single, err := Hash([]uint16{msg})
	require.NoError(t, err)

	double, err := Hash([]uint16{msg, msg})
	require.NoError(t, err)

	require.False(t, single.Equal(&double))
}

func TestEmptyHashIsInfinity(t *testing.T) {
	got, err := Hash(nil)
	require.NoError(t, err)
	require.True(t, got.IsInfinity())
}

func TestDigestRoundTrip(t *testing.T) {
	msgs := sampleMessages(4)
	got, err := Hash(msgs)
	require.NoError(t, err)
	buf := got.Bytes()
	var dec kb8.G1Affine
	_, err = dec.SetBytes(buf[:])
	require.NoError(t, err)
	require.True(t, dec.Equal(&got))
}

func TestMapSatisfiesYIncrementRelation(t *testing.T) {
	for _, msg := range sampleMessages(64) {
		p, offset, err := Map(msg)
		require.NoError(t, err)
		require.Less(t, int(offset), tweakBound)
		wantY := expectedY(msg, offset)
		require.True(t, p.Y.Equal(&wantY))
	}
}

func BenchmarkMap(b *testing.B) {
	msg := sampleMessages(1)[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := Map(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAccumulatorInsert(b *testing.B) {
	msgs := sampleMessages(256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		acc := NewAccumulator()
		for _, msg := range msgs {
			if err := acc.Insert(msg); err != nil {
				b.Fatal(err)
			}
		}
		_ = acc.Digest()
	}
}

func BenchmarkHash256(b *testing.B) {
	msgs := sampleMessages(256)
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
