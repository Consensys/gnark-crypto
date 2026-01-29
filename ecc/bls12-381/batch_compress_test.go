// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package bls12381

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
)

func randomG1() G1Affine {
	var p G1Affine
	var s fr.Element
	s.SetRandom()
	p.ScalarMultiplicationBase(s.BigInt(new(big.Int)))
	return p
}

func randomG2() G2Affine {
	var p G2Affine
	var s fr.Element
	s.SetRandom()
	p.ScalarMultiplicationBase(s.BigInt(new(big.Int)))
	return p
}

func TestBatchCompress2G1(t *testing.T) {
	// Test with random points
	for i := 0; i < 100; i++ {
		p0 := randomG1()
		p1 := randomG1()

		z0, z1, flags, err := BatchCompress2G1(&p0, &p1)
		if err != nil {
			t.Fatalf("compression failed: %v", err)
		}

		p0Dec, p1Dec, err := BatchDecompress2G1(z0, z1, flags)
		if err != nil {
			t.Fatalf("decompression failed: %v", err)
		}

		if !p0.Equal(&p0Dec) || !p1.Equal(&p1Dec) {
			t.Fatalf("round-trip failed: got (%v, %v), expected (%v, %v)", p0Dec, p1Dec, p0, p1)
		}
	}
}

func TestBatchCompress2G1Infinity(t *testing.T) {
	var inf G1Affine
	inf.SetInfinity()

	p := randomG1()

	// Both infinity
	z0, z1, flags, err := BatchCompress2G1(&inf, &inf)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	p0Dec, p1Dec, err := BatchDecompress2G1(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !p0Dec.IsInfinity() || !p1Dec.IsInfinity() {
		t.Fatal("expected both infinity")
	}

	// P0 infinity
	z0, z1, flags, err = BatchCompress2G1(&inf, &p)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	p0Dec, p1Dec, err = BatchDecompress2G1(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !p0Dec.IsInfinity() || !p1Dec.Equal(&p) {
		t.Fatal("round-trip failed for P0 infinity")
	}

	// P1 infinity
	z0, z1, flags, err = BatchCompress2G1(&p, &inf)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	p0Dec, p1Dec, err = BatchDecompress2G1(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !p0Dec.Equal(&p) || !p1Dec.IsInfinity() {
		t.Fatal("round-trip failed for P1 infinity")
	}
}

func TestBatchCompress2G1SamePoint(t *testing.T) {
	p := randomG1()

	// Same point
	z0, z1, flags, err := BatchCompress2G1(&p, &p)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	p0Dec, p1Dec, err := BatchDecompress2G1(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !p0Dec.Equal(&p) || !p1Dec.Equal(&p) {
		t.Fatal("round-trip failed for same point")
	}

	// Negation
	var negP G1Affine
	negP.Neg(&p)
	z0, z1, flags, err = BatchCompress2G1(&p, &negP)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	p0Dec, p1Dec, err = BatchDecompress2G1(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !p0Dec.Equal(&p) || !p1Dec.Equal(&negP) {
		t.Fatal("round-trip failed for negation")
	}
}

func TestBatchCompress2G1Automorphism(t *testing.T) {
	p := randomG1()

	// Test automorphism cases where P1 = [-ω]^k(P0)
	for k := 0; k < 6; k++ {
		var p1 G1Affine
		// Apply ω^(k%3) to x and (-1)^k to y
		switch k % 3 {
		case 0:
			p1.X = p.X
		case 1:
			p1.X.Mul(&p.X, &thirdRootOneG1)
		case 2:
			p1.X.Mul(&p.X, &thirdRootOneG2)
		}
		if k%2 == 0 {
			p1.Y = p.Y
		} else {
			p1.Y.Neg(&p.Y)
		}

		z0, z1, flags, err := BatchCompress2G1(&p, &p1)
		if err != nil {
			t.Fatalf("k=%d: compression failed: %v", k, err)
		}

		p0Dec, p1Dec, err := BatchDecompress2G1(z0, z1, flags)
		if err != nil {
			t.Fatalf("k=%d: decompression failed: %v", k, err)
		}

		if !p0Dec.Equal(&p) || !p1Dec.Equal(&p1) {
			t.Fatalf("k=%d: round-trip failed", k)
		}
	}
}

func TestBatchCompress2G2(t *testing.T) {
	// Test with random points
	for i := 0; i < 100; i++ {
		p0 := randomG2()
		p1 := randomG2()

		z0, z1, flags, err := BatchCompress2G2(&p0, &p1)
		if err != nil {
			t.Fatalf("compression failed: %v", err)
		}

		p0Dec, p1Dec, err := BatchDecompress2G2(z0, z1, flags)
		if err != nil {
			t.Fatalf("decompression failed: %v", err)
		}

		if !p0.Equal(&p0Dec) || !p1.Equal(&p1Dec) {
			t.Fatalf("round-trip failed: got (%v, %v), expected (%v, %v)", p0Dec, p1Dec, p0, p1)
		}
	}
}

func TestBatchCompress2G2Infinity(t *testing.T) {
	var inf G2Affine
	inf.SetInfinity()

	p := randomG2()

	// Both infinity
	z0, z1, flags, err := BatchCompress2G2(&inf, &inf)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	p0Dec, p1Dec, err := BatchDecompress2G2(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !p0Dec.IsInfinity() || !p1Dec.IsInfinity() {
		t.Fatal("expected both infinity")
	}

	// P0 infinity
	z0, z1, flags, err = BatchCompress2G2(&inf, &p)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	p0Dec, p1Dec, err = BatchDecompress2G2(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !p0Dec.IsInfinity() || !p1Dec.Equal(&p) {
		t.Fatal("round-trip failed for P0 infinity")
	}

	// P1 infinity
	z0, z1, flags, err = BatchCompress2G2(&p, &inf)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	p0Dec, p1Dec, err = BatchDecompress2G2(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !p0Dec.Equal(&p) || !p1Dec.IsInfinity() {
		t.Fatal("round-trip failed for P1 infinity")
	}
}

func TestBatchCompress2G2SamePoint(t *testing.T) {
	p := randomG2()

	// Same point
	z0, z1, flags, err := BatchCompress2G2(&p, &p)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	p0Dec, p1Dec, err := BatchDecompress2G2(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !p0Dec.Equal(&p) || !p1Dec.Equal(&p) {
		t.Fatal("round-trip failed for same point")
	}

	// Negation
	var negP G2Affine
	negP.Neg(&p)
	z0, z1, flags, err = BatchCompress2G2(&p, &negP)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}
	p0Dec, p1Dec, err = BatchDecompress2G2(z0, z1, flags)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}
	if !p0Dec.Equal(&p) || !p1Dec.Equal(&negP) {
		t.Fatal("round-trip failed for negation")
	}
}

func TestBatchCompressG1Slice(t *testing.T) {
	// Test with various sizes
	testCases := []int{0, 1, 2, 10, 11, 100, 101}

	for _, n := range testCases {
		points := make([]G1Affine, n)
		for i := range points {
			points[i] = randomG1()
		}

		// Compress
		compressed, err := BatchCompressG1Slice(points)
		if err != nil {
			t.Fatalf("n=%d: compression failed: %v", n, err)
		}

		// Decompress
		decompressed, err := BatchDecompressG1Slice(compressed, n)
		if err != nil {
			t.Fatalf("n=%d: decompression failed: %v", n, err)
		}

		// Verify
		if len(decompressed) != n {
			t.Fatalf("n=%d: expected %d points, got %d", n, n, len(decompressed))
		}
		for i := 0; i < n; i++ {
			if !points[i].Equal(&decompressed[i]) {
				t.Fatalf("n=%d, i=%d: point mismatch", n, i)
			}
		}
	}
}

func TestBatchCompressG2Slice(t *testing.T) {
	// Test with various sizes
	testCases := []int{0, 1, 2, 10, 11, 100, 101}

	for _, n := range testCases {
		points := make([]G2Affine, n)
		for i := range points {
			points[i] = randomG2()
		}

		// Compress
		compressed, err := BatchCompressG2Slice(points)
		if err != nil {
			t.Fatalf("n=%d: compression failed: %v", n, err)
		}

		// Decompress
		decompressed, err := BatchDecompressG2Slice(compressed, n)
		if err != nil {
			t.Fatalf("n=%d: decompression failed: %v", n, err)
		}

		// Verify
		if len(decompressed) != n {
			t.Fatalf("n=%d: expected %d points, got %d", n, n, len(decompressed))
		}
		for i := 0; i < n; i++ {
			if !points[i].Equal(&decompressed[i]) {
				t.Fatalf("n=%d, i=%d: point mismatch", n, i)
			}
		}
	}
}

func BenchmarkBatchCompress2G1(b *testing.B) {
	p0 := randomG1()
	p1 := randomG1()

	b.Run("Compress", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			BatchCompress2G1(&p0, &p1)
		}
	})

	z0, z1, flags, _ := BatchCompress2G1(&p0, &p1)

	b.Run("Decompress", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			BatchDecompress2G1(z0, z1, flags)
		}
	})

	// Compare with standard compression
	b.Run("StandardCompressPair", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p0.Bytes()
			p1.Bytes()
		}
	})

	p0Bytes := p0.Bytes()
	p1Bytes := p1.Bytes()

	b.Run("StandardDecompressPair", func(b *testing.B) {
		var q0, q1 G1Affine
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			q0.SetBytes(p0Bytes[:])
			q1.SetBytes(p1Bytes[:])
		}
	})
}

func BenchmarkBatchCompress2G2(b *testing.B) {
	p0 := randomG2()
	p1 := randomG2()

	b.Run("Compress", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			BatchCompress2G2(&p0, &p1)
		}
	})

	z0, z1, flags, _ := BatchCompress2G2(&p0, &p1)

	b.Run("Decompress", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			BatchDecompress2G2(z0, z1, flags)
		}
	})

	// Compare with standard compression
	b.Run("StandardCompressPair", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p0.Bytes()
			p1.Bytes()
		}
	})

	p0Bytes := p0.Bytes()
	p1Bytes := p1.Bytes()

	b.Run("StandardDecompressPair", func(b *testing.B) {
		var q0, q1 G2Affine
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			q0.SetBytes(p0Bytes[:])
			q1.SetBytes(p1Bytes[:])
		}
	})
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
