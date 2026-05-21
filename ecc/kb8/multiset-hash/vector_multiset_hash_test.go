package multisethash

import (
	"math"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/kb8"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/stretchr/testify/require"
)

// halfModulus returns floor(p/2), the threshold below which the encoded
// ordinate must remain to keep the map-to-curve image inverse-free.
func halfModulus() *big.Int {
	p := koalabear.Modulus()
	return new(big.Int).Rsh(p, 1)
}

func sampleLinearMessages(n int) []uint32 {
	res := make([]uint32, n)
	var x uint32 = 1
	for i := range res {
		x = 1664525*x + 1013904223
		res[i] = x % linearM
	}
	return res
}

func sampleVectorMessages64(n int) []uint64 {
	res := make([]uint64, n)
	var x uint64 = 1
	for i := range res {
		x = 6364136223846793005*x + 1442695040888963407
		res[i] = x
	}
	return res
}

// ----- Linear separator -----

func TestMapLinearDeterministic(t *testing.T) {
	for _, msg := range sampleLinearMessages(16) {
		p1, o1, err := MapLinear(msg)
		require.NoError(t, err)
		p2, o2, err := MapLinear(msg)
		require.NoError(t, err)
		require.Equal(t, o1, o2)
		for i := range p1 {
			require.True(t, p1[i].Equal(&p2[i]))
			require.True(t, p1[i].IsOnCurve())
			require.True(t, p1[i].IsInSubGroup())
			require.False(t, p1[i].IsInfinity())
			require.Less(t, int(o1[i]), linearT)
		}
	}
}

func TestMapLinearInverseFree(t *testing.T) {
	half := halfModulus()
	var y big.Int
	for _, msg := range sampleLinearMessages(32) {
		pts, _, err := MapLinear(msg)
		require.NoError(t, err)
		for i := range pts {
			pts[i].Y.C0.B0.A0.BigInt(&y)
			require.Negative(t, y.Cmp(half), "linear y_%d out of inverse-free domain for msg=%d", i, msg)
		}
	}
}

func TestLinearRejectsOutOfRange(t *testing.T) {
	_, _, err := MapLinear(linearM)
	require.Error(t, err)
	_, _, err = MapLinear(linearM - 1)
	require.NoError(t, err)
	acc := NewLinearAccumulator()
	require.Error(t, acc.Insert(linearM))
	require.NoError(t, acc.Insert(linearM-1))
}

func TestLinearHashPermutationInvariant(t *testing.T) {
	msgs := sampleLinearMessages(8)
	got1, err := HashLinear(msgs)
	require.NoError(t, err)

	permuted := append([]uint32(nil), msgs...)
	permuted[0], permuted[5] = permuted[5], permuted[0]
	permuted[1], permuted[7] = permuted[7], permuted[1]

	got2, err := HashLinear(permuted)
	require.NoError(t, err)
	for i := range got1 {
		require.True(t, got1[i].Equal(&got2[i]))
	}
}

func TestLinearInsertRemoveCancellation(t *testing.T) {
	msgs := sampleLinearMessages(6)
	acc := NewLinearAccumulator()
	for _, msg := range msgs {
		require.NoError(t, acc.Insert(msg))
	}
	require.NoError(t, acc.Remove(msgs[2]))
	require.NoError(t, acc.Remove(msgs[4]))

	acc2 := NewLinearAccumulator()
	require.NoError(t, acc2.Insert(msgs[0]))
	require.NoError(t, acc2.Insert(msgs[1]))
	require.NoError(t, acc2.Insert(msgs[3]))
	require.NoError(t, acc2.Insert(msgs[5]))

	d1 := acc.Digest()
	d2 := acc2.Digest()
	for i := range d1 {
		require.True(t, d1[i].Equal(&d2[i]))
	}
}

func TestLinearDuplicatesMatter(t *testing.T) {
	msg := sampleLinearMessages(1)[0]
	single, err := HashLinear([]uint32{msg})
	require.NoError(t, err)
	double, err := HashLinear([]uint32{msg, msg})
	require.NoError(t, err)
	var differs bool
	for i := range single {
		if !single[i].Equal(&double[i]) {
			differs = true
			break
		}
	}
	require.True(t, differs, "doubling msg must change the linear digest")
}

func TestEmptyLinearHashIsInfinity(t *testing.T) {
	got, err := HashLinear(nil)
	require.NoError(t, err)
	for i := range got {
		require.True(t, got[i].IsInfinity())
	}
}

// ----- Poseidon2 sponge separator -----

func TestMapPoseidon2Deterministic(t *testing.T) {
	for _, msg := range sampleVectorMessages64(16) {
		p1, o1, err := MapPoseidon2(msg)
		require.NoError(t, err)
		p2, o2, err := MapPoseidon2(msg)
		require.NoError(t, err)
		require.Equal(t, o1, o2)
		for i := range p1 {
			require.True(t, p1[i].Equal(&p2[i]))
			require.True(t, p1[i].IsOnCurve())
			require.True(t, p1[i].IsInSubGroup())
			require.False(t, p1[i].IsInfinity())
			require.Less(t, int(o1[i]), pqT)
		}
	}
}

func TestMapPoseidon2InverseFree(t *testing.T) {
	half := halfModulus()
	var y big.Int
	for _, msg := range sampleVectorMessages64(32) {
		pts, _, err := MapPoseidon2(msg)
		require.NoError(t, err)
		for i := range pts {
			pts[i].Y.C0.B0.A0.BigInt(&y)
			require.Negative(t, y.Cmp(half), "poseidon2 y_%d out of inverse-free domain for msg=%d", i, msg)
		}
	}
}

func TestPoseidon2HashPermutationInvariant(t *testing.T) {
	msgs := sampleVectorMessages64(8)
	got1, err := HashPoseidon2(msgs)
	require.NoError(t, err)

	permuted := append([]uint64(nil), msgs...)
	permuted[0], permuted[5] = permuted[5], permuted[0]
	permuted[1], permuted[7] = permuted[7], permuted[1]

	got2, err := HashPoseidon2(permuted)
	require.NoError(t, err)
	for i := range got1 {
		require.True(t, got1[i].Equal(&got2[i]))
	}
}

func TestPoseidon2InsertRemoveCancellation(t *testing.T) {
	msgs := sampleVectorMessages64(6)
	acc := NewPoseidon2Accumulator()
	for _, msg := range msgs {
		require.NoError(t, acc.Insert(msg))
	}
	require.NoError(t, acc.Remove(msgs[2]))
	require.NoError(t, acc.Remove(msgs[4]))

	acc2 := NewPoseidon2Accumulator()
	require.NoError(t, acc2.Insert(msgs[0]))
	require.NoError(t, acc2.Insert(msgs[1]))
	require.NoError(t, acc2.Insert(msgs[3]))
	require.NoError(t, acc2.Insert(msgs[5]))

	d1 := acc.Digest()
	d2 := acc2.Digest()
	for i := range d1 {
		require.True(t, d1[i].Equal(&d2[i]))
	}
}

func TestPoseidon2DuplicatesMatter(t *testing.T) {
	msg := sampleVectorMessages64(1)[0]
	single, err := HashPoseidon2([]uint64{msg})
	require.NoError(t, err)
	double, err := HashPoseidon2([]uint64{msg, msg})
	require.NoError(t, err)
	var differs bool
	for i := range single {
		if !single[i].Equal(&double[i]) {
			differs = true
			break
		}
	}
	require.True(t, differs, "doubling msg must change the poseidon2 digest")
}

func TestEmptyPoseidon2HashIsInfinity(t *testing.T) {
	got, err := HashPoseidon2(nil)
	require.NoError(t, err)
	for i := range got {
		require.True(t, got[i].IsInfinity())
	}
}

// ----- Hash <-> Accumulator equivalence -----

func TestLinearHashMatchesAccumulator(t *testing.T) {
	msgs := sampleLinearMessages(10)
	got, err := HashLinear(msgs)
	require.NoError(t, err)

	acc := NewLinearAccumulator()
	for _, msg := range msgs {
		require.NoError(t, acc.Insert(msg))
	}
	digest := acc.Digest()
	for i := range got {
		require.True(t, got[i].Equal(&digest[i]))
	}
}

func TestPoseidon2HashMatchesAccumulator(t *testing.T) {
	msgs := sampleVectorMessages64(10)
	got, err := HashPoseidon2(msgs)
	require.NoError(t, err)

	acc := NewPoseidon2Accumulator()
	for _, msg := range msgs {
		require.NoError(t, acc.Insert(msg))
	}
	digest := acc.Digest()
	for i := range got {
		require.True(t, got[i].Equal(&digest[i]))
	}
}

// ----- Homomorphic additivity: Hash(A ∪ B) = Hash(A) + Hash(B) -----

func TestLinearHomomorphicAdditivity(t *testing.T) {
	msgs := sampleLinearMessages(12)
	mid := len(msgs) / 2
	a, b := msgs[:mid], msgs[mid:]

	full, err := HashLinear(msgs)
	require.NoError(t, err)
	dA, err := HashLinear(a)
	require.NoError(t, err)
	dB, err := HashLinear(b)
	require.NoError(t, err)

	for i := range full {
		var sum kb8.G1Affine
		sum.Add(&dA[i], &dB[i])
		require.True(t, sum.Equal(&full[i]),
			"linear: Hash(A∪B)[%d] must equal Hash(A)+Hash(B)", i)
	}
}

func TestPoseidon2HomomorphicAdditivity(t *testing.T) {
	msgs := sampleVectorMessages64(12)
	mid := len(msgs) / 2
	a, b := msgs[:mid], msgs[mid:]

	full, err := HashPoseidon2(msgs)
	require.NoError(t, err)
	dA, err := HashPoseidon2(a)
	require.NoError(t, err)
	dB, err := HashPoseidon2(b)
	require.NoError(t, err)

	for i := range full {
		var sum kb8.G1Affine
		sum.Add(&dA[i], &dB[i])
		require.True(t, sum.Equal(&full[i]),
			"poseidon2: Hash(A∪B)[%d] must equal Hash(A)+Hash(B)", i)
	}
}

// ----- Distinct messages produce distinct digests -----

func TestLinearDistinctMessagesDiffer(t *testing.T) {
	msgs := sampleLinearMessages(8)
	seen := make(map[string]uint32)
	for _, msg := range msgs {
		digest, err := HashLinear([]uint32{msg})
		require.NoError(t, err)
		// digest the first coordinate's bytes as a fingerprint; full
		// equality is exercised by the deterministic test.
		buf := digest[0].Bytes()
		key := string(buf[:])
		if prev, ok := seen[key]; ok {
			require.Equal(t, prev, msg,
				"distinct linear msgs %d and %d collide on coord 0", prev, msg)
		}
		seen[key] = msg
	}
}

func TestPoseidon2DistinctMessagesDiffer(t *testing.T) {
	msgs := sampleVectorMessages64(8)
	seen := make(map[string]uint64)
	for _, msg := range msgs {
		digest, err := HashPoseidon2([]uint64{msg})
		require.NoError(t, err)
		buf := digest[0].Bytes()
		key := string(buf[:])
		if prev, ok := seen[key]; ok {
			require.Equal(t, prev, msg,
				"distinct poseidon2 msgs %d and %d collide on coord 0", prev, msg)
		}
		seen[key] = msg
	}
}

// ----- Boundary messages -----

func TestLinearBoundaryMessages(t *testing.T) {
	for _, msg := range []uint32{0, linearM - 1} {
		pts, _, err := MapLinear(msg)
		require.NoError(t, err, "MapLinear must succeed for boundary msg=%d", msg)
		for i := range pts {
			require.True(t, pts[i].IsOnCurve())
			require.True(t, pts[i].IsInSubGroup())
		}
	}
}

func TestPoseidon2BoundaryMessages(t *testing.T) {
	for _, msg := range []uint64{0, math.MaxUint64} {
		pts, _, err := MapPoseidon2(msg)
		require.NoError(t, err, "MapPoseidon2 must succeed for boundary msg=%d", msg)
		for i := range pts {
			require.True(t, pts[i].IsOnCurve())
			require.True(t, pts[i].IsInSubGroup())
		}
	}
}

func TestMapAtSlotRejectsOutOfRange(t *testing.T) {
	bound := pqReducerBound.Uint64()

	p, _, err := MapAtSlot(bound - 1)
	require.NoError(t, err)
	var y big.Int
	p.Y.C0.B0.A0.BigInt(&y)
	require.Negative(t, y.Cmp(halfModulus()))

	_, _, err = MapAtSlot(bound)
	require.ErrorIs(t, err, errPqSlotOutOfRange)

	_, _, err = MapAtSlot(math.MaxUint64/pqT + 1)
	require.ErrorIs(t, err, errPqSlotOutOfRange)
}

// ----- Linear per-coordinate slot range -----

func TestLinearPerCoordinateSlotRange(t *testing.T) {
	// Coordinate i must encode y in [T*i*M, T*(i+1)*M). This catches an
	// off-by-one or index swap in the linear separator.
	var y big.Int
	for _, msg := range sampleLinearMessages(8) {
		pts, _, err := MapLinear(msg)
		require.NoError(t, err)
		for i := range pts {
			pts[i].Y.C0.B0.A0.BigInt(&y)
			loBound := uint64(linearT) * uint64(i) * uint64(linearM)
			hiBound := uint64(linearT) * (uint64(i) + 1) * uint64(linearM)
			require.GreaterOrEqual(t, y.Cmp(new(big.Int).SetUint64(loBound)), 0,
				"coord %d y must be >= %d (msg=%d)", i, loBound, msg)
			require.Negative(t, y.Cmp(new(big.Int).SetUint64(hiBound)),
				"coord %d y must be < %d (msg=%d)", i, hiBound, msg)
		}
	}
}

// ----- Reset -----

func TestLinearReset(t *testing.T) {
	acc := NewLinearAccumulator()
	for _, msg := range sampleLinearMessages(4) {
		require.NoError(t, acc.Insert(msg))
	}
	acc.Reset()
	digest := acc.Digest()
	for i := range digest {
		require.True(t, digest[i].IsInfinity(),
			"linear Reset must clear coord %d", i)
	}
}

func TestPoseidon2Reset(t *testing.T) {
	acc := NewPoseidon2Accumulator()
	for _, msg := range sampleVectorMessages64(4) {
		require.NoError(t, acc.Insert(msg))
	}
	acc.Reset()
	digest := acc.Digest()
	for i := range digest {
		require.True(t, digest[i].IsInfinity(),
			"poseidon2 Reset must clear coord %d", i)
	}
}

// ----- Cross-variant sanity -----

func TestLinearAndPoseidon2DigestsDiffer(t *testing.T) {
	// The two variants build their digests over different domain separators,
	// so even at "msg=0" their per-coordinate accumulators should not match
	// point-by-point.
	linDigest, err := HashLinear([]uint32{0})
	require.NoError(t, err)
	pqDigest, err := HashPoseidon2([]uint64{0})
	require.NoError(t, err)

	var anyDiff bool
	for i := range linDigest {
		l, p := linDigest[i], pqDigest[i]
		if !l.Equal(&p) {
			anyDiff = true
			break
		}
	}
	require.True(t, anyDiff)
}

// ----- Benchmarks -----

func BenchmarkMapLinear(b *testing.B) {
	msg := sampleLinearMessages(1)[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := MapLinear(msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMapPoseidon2(b *testing.B) {
	msg := sampleVectorMessages64(1)[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := MapPoseidon2(msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHashLinear256(b *testing.B) {
	msgs := sampleLinearMessages(256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := HashLinear(msgs); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHashPoseidon2_256(b *testing.B) {
	msgs := sampleVectorMessages64(256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := HashPoseidon2(msgs); err != nil {
			b.Fatal(err)
		}
	}
}
