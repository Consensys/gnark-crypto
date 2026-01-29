// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package bn254

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/internal/fptower"
)

func TestBatchCompress2G1(t *testing.T) {
	// Test with random points
	for i := 0; i < 100; i++ {
		var p0, p1 G1Affine
		p0.X.SetRandom()
		for p0.Y.Square(&p0.X).Mul(&p0.Y, &p0.X).Add(&p0.Y, &bCurveCoeff); p0.Y.Sqrt(&p0.Y) == nil; {
			p0.X.SetRandom()
			p0.Y.Square(&p0.X).Mul(&p0.Y, &p0.X).Add(&p0.Y, &bCurveCoeff)
		}

		p1.X.SetRandom()
		for p1.Y.Square(&p1.X).Mul(&p1.Y, &p1.X).Add(&p1.Y, &bCurveCoeff); p1.Y.Sqrt(&p1.Y) == nil; {
			p1.X.SetRandom()
			p1.Y.Square(&p1.X).Mul(&p1.Y, &p1.X).Add(&p1.Y, &bCurveCoeff)
		}

		// Compress
		z0, z1, flags, err := BatchCompress2G1(&p0, &p1)
		if err != nil {
			t.Fatalf("iteration %d: compression failed: %v", i, err)
		}

		// Decompress
		q0, q1, err := BatchDecompress2G1(z0, z1, flags)
		if err != nil {
			t.Fatalf("iteration %d: decompression failed: %v", i, err)
		}

		// Verify (allow swapped order due to symmetry in the formulas)
		if p0.Equal(&q0) && p1.Equal(&q1) {
			continue // OK
		}
		if p0.Equal(&q1) && p1.Equal(&q0) {
			continue // OK (swapped)
		}
		t.Errorf("iteration %d: points don't match. p0=%v, p1=%v, q0=%v, q1=%v", i, p0, p1, q0, q1)
	}
}

func TestBatchCompress2G1Infinity(t *testing.T) {
	var p0, inf G1Affine
	inf.SetInfinity()

	// Generate a valid point
	p0.X.SetRandom()
	for p0.Y.Square(&p0.X).Mul(&p0.Y, &p0.X).Add(&p0.Y, &bCurveCoeff); p0.Y.Sqrt(&p0.Y) == nil; {
		p0.X.SetRandom()
		p0.Y.Square(&p0.X).Mul(&p0.Y, &p0.X).Add(&p0.Y, &bCurveCoeff)
	}

	// Test both infinity
	z0, z1, flags, err := BatchCompress2G1(&inf, &inf)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	q0, q1, err := BatchDecompress2G1(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !q0.IsInfinity() || !q1.IsInfinity() {
		t.Error("expected both points to be infinity")
	}

	// Test p0 infinity
	z0, z1, flags, err = BatchCompress2G1(&inf, &p0)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	q0, q1, err = BatchDecompress2G1(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !q0.IsInfinity() {
		t.Error("expected q0 to be infinity")
	}
	if !p0.Equal(&q1) {
		t.Error("p0 != q1")
	}

	// Test p1 infinity
	z0, z1, flags, err = BatchCompress2G1(&p0, &inf)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	q0, q1, err = BatchDecompress2G1(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !p0.Equal(&q0) {
		t.Error("p0 != q0")
	}
	if !q1.IsInfinity() {
		t.Error("expected q1 to be infinity")
	}
}

func TestBatchCompress2G1SamePoint(t *testing.T) {
	var p0 G1Affine
	p0.X.SetRandom()
	for p0.Y.Square(&p0.X).Mul(&p0.Y, &p0.X).Add(&p0.Y, &bCurveCoeff); p0.Y.Sqrt(&p0.Y) == nil; {
		p0.X.SetRandom()
		p0.Y.Square(&p0.X).Mul(&p0.Y, &p0.X).Add(&p0.Y, &bCurveCoeff)
	}

	// Test same point
	z0, z1, flags, err := BatchCompress2G1(&p0, &p0)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	q0, q1, err := BatchDecompress2G1(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !p0.Equal(&q0) || !p0.Equal(&q1) {
		t.Error("decompression of same point failed")
	}

	// Test negation
	var p1 G1Affine
	p1.Neg(&p0)
	z0, z1, flags, err = BatchCompress2G1(&p0, &p1)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	q0, q1, err = BatchDecompress2G1(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !p0.Equal(&q0) {
		t.Error("p0 != q0")
	}
	if !p1.Equal(&q1) {
		t.Error("p1 != q1")
	}
}

func TestBatchCompress2G1Automorphism(t *testing.T) {
	// Test points related by the automorphism [-ω]
	var p0 G1Affine
	p0.X.SetRandom()
	for p0.Y.Square(&p0.X).Mul(&p0.Y, &p0.X).Add(&p0.Y, &bCurveCoeff); p0.Y.Sqrt(&p0.Y) == nil; {
		p0.X.SetRandom()
		p0.Y.Square(&p0.X).Mul(&p0.Y, &p0.X).Add(&p0.Y, &bCurveCoeff)
	}

	// Test p1 = [-ω](p0) = (ω·x0, -y0)
	var p1 G1Affine
	p1.X.Mul(&p0.X, &thirdRootOneG1)
	p1.Y.Neg(&p0.Y)

	z0, z1, flags, err := BatchCompress2G1(&p0, &p1)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	q0, q1, err := BatchDecompress2G1(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !p0.Equal(&q0) {
		t.Error("p0 != q0")
	}
	if !p1.Equal(&q1) {
		t.Error("p1 != q1")
	}

	// Test p1 = [ω](p0) = (ω·x0, y0)
	p1.X.Mul(&p0.X, &thirdRootOneG1)
	p1.Y = p0.Y

	z0, z1, flags, err = BatchCompress2G1(&p0, &p1)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	q0, q1, err = BatchDecompress2G1(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !p0.Equal(&q0) {
		t.Error("p0 != q0 for [ω] case")
	}
	if !p1.Equal(&q1) {
		t.Error("p1 != q1 for [ω] case")
	}
}

func BenchmarkBatchCompress2G1(b *testing.B) {
	var points [100]G1Affine
	for i := range points {
		points[i].X.SetRandom()
		for points[i].Y.Square(&points[i].X).Mul(&points[i].Y, &points[i].X).Add(&points[i].Y, &bCurveCoeff); points[i].Y.Sqrt(&points[i].Y) == nil; {
			points[i].X.SetRandom()
			points[i].Y.Square(&points[i].X).Mul(&points[i].Y, &points[i].X).Add(&points[i].Y, &bCurveCoeff)
		}
	}

	b.Run("BatchCompress2G1", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := i % 50
			BatchCompress2G1(&points[idx*2], &points[idx*2+1])
		}
	})

	// Pre-compress for decompression benchmark
	var z0s, z1s [50]fp.Element
	var flagsArr [50]byte
	for i := 0; i < 50; i++ {
		z0s[i], z1s[i], flagsArr[i], _ = BatchCompress2G1(&points[i*2], &points[i*2+1])
	}

	b.Run("BatchDecompress2G1", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := i % 50
			BatchDecompress2G1(z0s[idx], z1s[idx], flagsArr[idx])
		}
	})

	// Compare with standard compression
	b.Run("StandardCompress2G1", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := i % 50
			points[idx*2].Bytes()
			points[idx*2+1].Bytes()
		}
	})

	// Pre-compress for standard decompression
	var compressed [100][SizeOfG1AffineCompressed]byte
	for i := range points {
		compressed[i] = points[i].Bytes()
	}

	b.Run("StandardDecompress2G1", func(b *testing.B) {
		var p G1Affine
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := i % 50
			p.SetBytes(compressed[idx*2][:])
			p.SetBytes(compressed[idx*2+1][:])
		}
	})
}

func BenchmarkBatchCompress2G2(b *testing.B) {
	var points [100]G2Affine
	for i := range points {
		points[i] = randomG2()
	}

	b.Run("BatchCompress2G2", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := i % 50
			BatchCompress2G2(&points[idx*2], &points[idx*2+1])
		}
	})

	// Pre-compress for decompression benchmark
	type e2Pair struct {
		z0, z1 fptower.E2
		flags  byte
	}
	compressed := make([]e2Pair, 50)
	for i := 0; i < 50; i++ {
		compressed[i].z0, compressed[i].z1, compressed[i].flags, _ = BatchCompress2G2(&points[i*2], &points[i*2+1])
	}

	b.Run("BatchDecompress2G2", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := i % 50
			BatchDecompress2G2(compressed[idx].z0, compressed[idx].z1, compressed[idx].flags)
		}
	})

	// Compare with standard compression
	b.Run("StandardCompress2G2", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := i % 50
			points[idx*2].Bytes()
			points[idx*2+1].Bytes()
		}
	})

	// Pre-compress for standard decompression
	var stdCompressed [100][SizeOfG2AffineCompressed]byte
	for i := range points {
		stdCompressed[i] = points[i].Bytes()
	}

	b.Run("StandardDecompress2G2", func(b *testing.B) {
		var p G2Affine
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := i % 50
			p.SetBytes(stdCompressed[idx*2][:])
			p.SetBytes(stdCompressed[idx*2+1][:])
		}
	})
}

// randomG1 generates a random G1 point on the curve
func randomG1() G1Affine {
	var p G1Affine
	var buf [32]byte
	rand.Read(buf[:])
	p.X.SetBytes(buf[:])
	for p.Y.Square(&p.X).Mul(&p.Y, &p.X).Add(&p.Y, &bCurveCoeff); p.Y.Sqrt(&p.Y) == nil; {
		rand.Read(buf[:])
		p.X.SetBytes(buf[:])
		p.Y.Square(&p.X).Mul(&p.Y, &p.X).Add(&p.Y, &bCurveCoeff)
	}
	return p
}

func TestBatchCompress2G2(t *testing.T) {
	// Test with random points
	for i := 0; i < 100; i++ {
		p0 := randomG2()
		p1 := randomG2()

		// Compress
		z0, z1, flags, err := BatchCompress2G2(&p0, &p1)
		if err != nil {
			t.Fatalf("iteration %d: compression failed: %v", i, err)
		}

		// Decompress
		q0, q1, err := BatchDecompress2G2(z0, z1, flags)
		if err != nil {
			t.Fatalf("iteration %d: decompression failed: %v", i, err)
		}

		// Verify (allow swapped order due to symmetry in the formulas)
		if p0.Equal(&q0) && p1.Equal(&q1) {
			continue // OK
		}
		if p0.Equal(&q1) && p1.Equal(&q0) {
			continue // OK (swapped)
		}
		t.Errorf("iteration %d: points don't match. p0=%v, p1=%v, q0=%v, q1=%v", i, p0, p1, q0, q1)
	}
}

func TestBatchCompress2G2Infinity(t *testing.T) {
	var p0, inf G2Affine
	inf.SetInfinity()

	// Generate a valid point
	p0 = randomG2()

	// Test both infinity
	z0, z1, flags, err := BatchCompress2G2(&inf, &inf)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	q0, q1, err := BatchDecompress2G2(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !q0.IsInfinity() || !q1.IsInfinity() {
		t.Error("expected both points to be infinity")
	}

	// Test p0 infinity
	z0, z1, flags, err = BatchCompress2G2(&inf, &p0)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	q0, q1, err = BatchDecompress2G2(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !q0.IsInfinity() {
		t.Error("expected q0 to be infinity")
	}
	if !p0.Equal(&q1) {
		t.Error("p0 != q1")
	}

	// Test p1 infinity
	z0, z1, flags, err = BatchCompress2G2(&p0, &inf)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	q0, q1, err = BatchDecompress2G2(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !p0.Equal(&q0) {
		t.Error("p0 != q0")
	}
	if !q1.IsInfinity() {
		t.Error("expected q1 to be infinity")
	}
}

func TestBatchCompress2G2SamePoint(t *testing.T) {
	p0 := randomG2()

	// Test same point
	z0, z1, flags, err := BatchCompress2G2(&p0, &p0)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	q0, q1, err := BatchDecompress2G2(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !p0.Equal(&q0) || !p0.Equal(&q1) {
		t.Error("decompression of same point failed")
	}

	// Test negation
	var p1 G2Affine
	p1.Neg(&p0)
	z0, z1, flags, err = BatchCompress2G2(&p0, &p1)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	q0, q1, err = BatchDecompress2G2(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !p0.Equal(&q0) {
		t.Error("p0 != q0")
	}
	if !p1.Equal(&q1) {
		t.Error("p1 != q1")
	}
}

// randomG2 generates a random G2 point on the curve
func randomG2() G2Affine {
	_, _, _, g2 := Generators()
	var s fr.Element
	s.SetRandom()
	var p G2Affine
	p.ScalarMultiplication(&g2, s.BigInt(new(big.Int)))
	return p
}

func TestBatchCompressG1Slice(t *testing.T) {
	// Test empty slice
	data, err := BatchCompressG1Slice(nil)
	if err != nil {
		t.Fatalf("empty slice: %v", err)
	}
	if data != nil {
		t.Error("expected nil for empty slice")
	}

	points, err := BatchDecompressG1Slice(nil, 0)
	if err != nil {
		t.Fatalf("decompress empty: %v", err)
	}
	if points != nil {
		t.Error("expected nil for empty decompress")
	}

	// Test single point (odd)
	single := []G1Affine{randomG1()}
	data, err = BatchCompressG1Slice(single)
	if err != nil {
		t.Fatalf("single point compress: %v", err)
	}
	if len(data) != SizeOfG1AffineCompressed {
		t.Errorf("expected %d bytes, got %d", SizeOfG1AffineCompressed, len(data))
	}

	restored, err := BatchDecompressG1Slice(data, 1)
	if err != nil {
		t.Fatalf("single point decompress: %v", err)
	}
	if !single[0].Equal(&restored[0]) {
		t.Error("single point mismatch")
	}

	// Test even number of points
	evenPoints := make([]G1Affine, 10)
	for i := range evenPoints {
		evenPoints[i] = randomG1()
	}

	data, err = BatchCompressG1Slice(evenPoints)
	if err != nil {
		t.Fatalf("even slice compress: %v", err)
	}

	expectedSize := 5 * SizeOfBatchCompressedG1Pair
	if len(data) != expectedSize {
		t.Errorf("expected %d bytes, got %d", expectedSize, len(data))
	}

	restored, err = BatchDecompressG1Slice(data, 10)
	if err != nil {
		t.Fatalf("even slice decompress: %v", err)
	}

	for i := 0; i < 5; i++ {
		p0, p1 := evenPoints[i*2], evenPoints[i*2+1]
		q0, q1 := restored[i*2], restored[i*2+1]
		// Allow swapped order within pairs
		if !((p0.Equal(&q0) && p1.Equal(&q1)) || (p0.Equal(&q1) && p1.Equal(&q0))) {
			t.Errorf("pair %d mismatch", i)
		}
	}

	// Test odd number of points
	oddPoints := make([]G1Affine, 11)
	for i := range oddPoints {
		oddPoints[i] = randomG1()
	}

	data, err = BatchCompressG1Slice(oddPoints)
	if err != nil {
		t.Fatalf("odd slice compress: %v", err)
	}

	expectedSize = 5*SizeOfBatchCompressedG1Pair + SizeOfG1AffineCompressed
	if len(data) != expectedSize {
		t.Errorf("expected %d bytes, got %d", expectedSize, len(data))
	}

	restored, err = BatchDecompressG1Slice(data, 11)
	if err != nil {
		t.Fatalf("odd slice decompress: %v", err)
	}

	// Check pairs
	for i := 0; i < 5; i++ {
		p0, p1 := oddPoints[i*2], oddPoints[i*2+1]
		q0, q1 := restored[i*2], restored[i*2+1]
		if !((p0.Equal(&q0) && p1.Equal(&q1)) || (p0.Equal(&q1) && p1.Equal(&q0))) {
			t.Errorf("pair %d mismatch", i)
		}
	}
	// Check last point
	if !oddPoints[10].Equal(&restored[10]) {
		t.Error("last odd point mismatch")
	}
}

func TestBatchCompressG2Slice(t *testing.T) {
	// Test empty slice
	data, err := BatchCompressG2Slice(nil)
	if err != nil {
		t.Fatalf("empty slice: %v", err)
	}
	if data != nil {
		t.Error("expected nil for empty slice")
	}

	points, err := BatchDecompressG2Slice(nil, 0)
	if err != nil {
		t.Fatalf("decompress empty: %v", err)
	}
	if points != nil {
		t.Error("expected nil for empty decompress")
	}

	// Test single point (odd)
	single := []G2Affine{randomG2()}
	data, err = BatchCompressG2Slice(single)
	if err != nil {
		t.Fatalf("single point compress: %v", err)
	}
	if len(data) != SizeOfG2AffineCompressed {
		t.Errorf("expected %d bytes, got %d", SizeOfG2AffineCompressed, len(data))
	}

	restored, err := BatchDecompressG2Slice(data, 1)
	if err != nil {
		t.Fatalf("single point decompress: %v", err)
	}
	if !single[0].Equal(&restored[0]) {
		t.Error("single point mismatch")
	}

	// Test even number of points
	evenPoints := make([]G2Affine, 10)
	for i := range evenPoints {
		evenPoints[i] = randomG2()
	}

	data, err = BatchCompressG2Slice(evenPoints)
	if err != nil {
		t.Fatalf("even slice compress: %v", err)
	}

	expectedSize := 5 * SizeOfBatchCompressedG2Pair
	if len(data) != expectedSize {
		t.Errorf("expected %d bytes, got %d", expectedSize, len(data))
	}

	restored, err = BatchDecompressG2Slice(data, 10)
	if err != nil {
		t.Fatalf("even slice decompress: %v", err)
	}

	for i := 0; i < 5; i++ {
		p0, p1 := evenPoints[i*2], evenPoints[i*2+1]
		q0, q1 := restored[i*2], restored[i*2+1]
		// Allow swapped order within pairs
		if !((p0.Equal(&q0) && p1.Equal(&q1)) || (p0.Equal(&q1) && p1.Equal(&q0))) {
			t.Errorf("pair %d mismatch", i)
		}
	}

	// Test odd number of points
	oddPoints := make([]G2Affine, 11)
	for i := range oddPoints {
		oddPoints[i] = randomG2()
	}

	data, err = BatchCompressG2Slice(oddPoints)
	if err != nil {
		t.Fatalf("odd slice compress: %v", err)
	}

	expectedSize = 5*SizeOfBatchCompressedG2Pair + SizeOfG2AffineCompressed
	if len(data) != expectedSize {
		t.Errorf("expected %d bytes, got %d", expectedSize, len(data))
	}

	restored, err = BatchDecompressG2Slice(data, 11)
	if err != nil {
		t.Fatalf("odd slice decompress: %v", err)
	}

	// Check pairs
	for i := 0; i < 5; i++ {
		p0, p1 := oddPoints[i*2], oddPoints[i*2+1]
		q0, q1 := restored[i*2], restored[i*2+1]
		if !((p0.Equal(&q0) && p1.Equal(&q1)) || (p0.Equal(&q1) && p1.Equal(&q0))) {
			t.Errorf("pair %d mismatch", i)
		}
	}
	// Check last point
	if !oddPoints[10].Equal(&restored[10]) {
		t.Error("last odd point mismatch")
	}
}

func BenchmarkBatchCompressG1Slice(b *testing.B) {
	const n = 1000
	points := make([]G1Affine, n)
	for i := range points {
		points[i] = randomG1()
	}

	b.Run("BatchCompress", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			BatchCompressG1Slice(points)
		}
	})

	compressedBatch, _ := BatchCompressG1Slice(points)

	b.Run("BatchDecompress", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			BatchDecompressG1Slice(compressedBatch, n)
		}
	})

	// Compare with standard parallel compression/decompression using Encoder/Decoder
	b.Run("StandardCompressParallel", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.Encode(points)
		}
	})

	var stdBuf bytes.Buffer
	enc := NewEncoder(&stdBuf)
	enc.Encode(points)
	stdData := stdBuf.Bytes()

	b.Run("StandardDecompressParallel", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewReader(stdData)
			dec := NewDecoder(buf, NoSubgroupChecks())
			var restored []G1Affine
			dec.Decode(&restored)
		}
	})
}

func BenchmarkBatchCompressG2Slice(b *testing.B) {
	const n = 1000
	points := make([]G2Affine, n)
	for i := range points {
		points[i] = randomG2()
	}

	b.Run("BatchCompress", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			BatchCompressG2Slice(points)
		}
	})

	compressedBatch, _ := BatchCompressG2Slice(points)

	b.Run("BatchDecompress", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			BatchDecompressG2Slice(compressedBatch, n)
		}
	})

	// Compare with standard parallel compression/decompression using Encoder/Decoder
	b.Run("StandardCompressParallel", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.Encode(points)
		}
	})

	var stdBuf bytes.Buffer
	enc := NewEncoder(&stdBuf)
	enc.Encode(points)
	stdData := stdBuf.Bytes()

	b.Run("StandardDecompressParallel", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewReader(stdData)
			dec := NewDecoder(buf, NoSubgroupChecks())
			var restored []G2Affine
			dec.Decode(&restored)
		}
	})
}

func TestBatchEncoderDecoderG1(t *testing.T) {
	// Test with various sizes
	testCases := []int{0, 1, 2, 10, 11, 100, 101}

	for _, n := range testCases {
		points := make([]G1Affine, n)
		for i := range points {
			points[i] = randomG1()
		}

		// Encode
		var buf bytes.Buffer
		enc := NewBatchEncoder(&buf)
		if err := enc.EncodeG1(points); err != nil {
			t.Fatalf("n=%d: encode failed: %v", n, err)
		}

		// Check size
		expectedSize := BatchCompressedSizeG1(n)
		if buf.Len() != expectedSize {
			t.Errorf("n=%d: expected size %d, got %d", n, expectedSize, buf.Len())
		}

		// Decode
		dec := NewBatchDecoder(&buf)
		var restored []G1Affine
		if err := dec.DecodeG1(&restored); err != nil {
			t.Fatalf("n=%d: decode failed: %v", n, err)
		}

		// Verify
		if len(restored) != n {
			t.Fatalf("n=%d: expected %d points, got %d", n, n, len(restored))
		}

		// Check pairs (allow swap within pairs)
		nPairs := n / 2
		for i := 0; i < nPairs; i++ {
			p0, p1 := points[i*2], points[i*2+1]
			q0, q1 := restored[i*2], restored[i*2+1]
			if !((p0.Equal(&q0) && p1.Equal(&q1)) || (p0.Equal(&q1) && p1.Equal(&q0))) {
				t.Errorf("n=%d, pair %d: mismatch", n, i)
			}
		}
		// Check odd point
		if n%2 == 1 {
			if !points[n-1].Equal(&restored[n-1]) {
				t.Errorf("n=%d: last point mismatch", n)
			}
		}
	}
}

func TestBatchEncoderDecoderG2(t *testing.T) {
	// Test with various sizes
	testCases := []int{0, 1, 2, 10, 11}

	for _, n := range testCases {
		points := make([]G2Affine, n)
		for i := range points {
			points[i] = randomG2()
		}

		// Encode
		var buf bytes.Buffer
		enc := NewBatchEncoder(&buf)
		if err := enc.EncodeG2(points); err != nil {
			t.Fatalf("n=%d: encode failed: %v", n, err)
		}

		// Check size
		expectedSize := BatchCompressedSizeG2(n)
		if buf.Len() != expectedSize {
			t.Errorf("n=%d: expected size %d, got %d", n, expectedSize, buf.Len())
		}

		// Decode
		dec := NewBatchDecoder(&buf)
		var restored []G2Affine
		if err := dec.DecodeG2(&restored); err != nil {
			t.Fatalf("n=%d: decode failed: %v", n, err)
		}

		// Verify
		if len(restored) != n {
			t.Fatalf("n=%d: expected %d points, got %d", n, n, len(restored))
		}

		// Check pairs (allow swap within pairs)
		nPairs := n / 2
		for i := 0; i < nPairs; i++ {
			p0, p1 := points[i*2], points[i*2+1]
			q0, q1 := restored[i*2], restored[i*2+1]
			if !((p0.Equal(&q0) && p1.Equal(&q1)) || (p0.Equal(&q1) && p1.Equal(&q0))) {
				t.Errorf("n=%d, pair %d: mismatch", n, i)
			}
		}
		// Check odd point
		if n%2 == 1 {
			if !points[n-1].Equal(&restored[n-1]) {
				t.Errorf("n=%d: last point mismatch", n)
			}
		}
	}
}

func TestWriteReadBatchG1(t *testing.T) {
	points := make([]G1Affine, 20)
	for i := range points {
		points[i] = randomG1()
	}

	var buf bytes.Buffer
	written, err := WriteBatchG1(&buf, points)
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	restored, read, err := ReadBatchG1(&buf, true)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if written != read {
		t.Errorf("written %d != read %d", written, read)
	}

	if len(restored) != len(points) {
		t.Fatalf("expected %d points, got %d", len(points), len(restored))
	}
}

func TestWriteReadBatchG2(t *testing.T) {
	points := make([]G2Affine, 20)
	for i := range points {
		points[i] = randomG2()
	}

	var buf bytes.Buffer
	written, err := WriteBatchG2(&buf, points)
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	restored, read, err := ReadBatchG2(&buf, true)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if written != read {
		t.Errorf("written %d != read %d", written, read)
	}

	if len(restored) != len(points) {
		t.Fatalf("expected %d points, got %d", len(points), len(restored))
	}
}

func BenchmarkBatchEncoderDecoderG1(b *testing.B) {
	const n = 100
	points := make([]G1Affine, n)
	for i := range points {
		points[i] = randomG1()
	}

	b.Run("BatchEncode", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			enc := NewBatchEncoder(&buf)
			enc.EncodeG1(points)
		}
	})

	var encodedBuf bytes.Buffer
	enc := NewBatchEncoder(&encodedBuf)
	enc.EncodeG1(points)
	encodedData := encodedBuf.Bytes()

	b.Run("BatchDecode", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewReader(encodedData)
			dec := NewBatchDecoder(buf, NoBatchSubgroupChecks())
			var restored []G1Affine
			dec.DecodeG1(&restored)
		}
	})

	// Compare with standard encoder
	b.Run("StandardEncode", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.Encode(points)
		}
	})

	var stdEncodedBuf bytes.Buffer
	stdEnc := NewEncoder(&stdEncodedBuf)
	stdEnc.Encode(points)
	stdEncodedData := stdEncodedBuf.Bytes()

	b.Run("StandardDecode", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewReader(stdEncodedData)
			dec := NewDecoder(buf, NoSubgroupChecks())
			var restored []G1Affine
			dec.Decode(&restored)
		}
	})
}

func BenchmarkBatchEncoderDecoderG2(b *testing.B) {
	const n = 100
	points := make([]G2Affine, n)
	for i := range points {
		points[i] = randomG2()
	}

	b.Run("BatchEncode", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			enc := NewBatchEncoder(&buf)
			enc.EncodeG2(points)
		}
	})

	var encodedBuf bytes.Buffer
	enc := NewBatchEncoder(&encodedBuf)
	enc.EncodeG2(points)
	encodedData := encodedBuf.Bytes()

	b.Run("BatchDecode", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewReader(encodedData)
			dec := NewBatchDecoder(buf, NoBatchSubgroupChecks())
			var restored []G2Affine
			dec.DecodeG2(&restored)
		}
	})

	// Compare with standard encoder
	b.Run("StandardEncode", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.Encode(points)
		}
	})

	var stdEncodedBuf bytes.Buffer
	stdEnc := NewEncoder(&stdEncodedBuf)
	stdEnc.Encode(points)
	stdEncodedData := stdEncodedBuf.Bytes()

	b.Run("StandardDecode", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewReader(stdEncodedData)
			dec := NewDecoder(buf, NoSubgroupChecks())
			var restored []G2Affine
			dec.Decode(&restored)
		}
	})
}
