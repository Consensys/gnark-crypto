// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package bn254

import (
	"encoding/binary"
	"errors"
	"io"
	"sync/atomic"

	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/consensys/gnark-crypto/ecc/bn254/internal/fptower"
	"github.com/consensys/gnark-crypto/internal/parallel"
)

// BatchEncoder writes bn254 slices using batch compression (2-by-2 point compression).
// This provides faster decompression at the cost of slightly larger compressed size per pair.
type BatchEncoder struct {
	w io.Writer
	n int64
}

// BatchDecoder reads bn254 slices that were encoded with batch compression.
type BatchDecoder struct {
	r             io.Reader
	n             int64
	subGroupCheck bool
}

// NewBatchEncoder returns a binary encoder that uses batch compression for point slices.
// Batch compression compresses pairs of points together, enabling faster decompression
// using cube roots instead of square roots.
func NewBatchEncoder(w io.Writer) *BatchEncoder {
	return &BatchEncoder{w: w}
}

// NewBatchDecoder returns a binary decoder for batch-compressed point slices.
func NewBatchDecoder(r io.Reader, options ...func(*BatchDecoder)) *BatchDecoder {
	d := &BatchDecoder{r: r, subGroupCheck: true}
	for _, o := range options {
		o(d)
	}
	return d
}

// NoBatchSubgroupChecks returns an option to disable subgroup checks during batch decoding.
func NoBatchSubgroupChecks() func(*BatchDecoder) {
	return func(dec *BatchDecoder) {
		dec.subGroupCheck = false
	}
}

// BytesWritten returns the total bytes written.
func (enc *BatchEncoder) BytesWritten() int64 {
	return enc.n
}

// BytesRead returns the total bytes read.
func (dec *BatchDecoder) BytesRead() int64 {
	return dec.n
}

// EncodeG1 writes a slice of G1Affine points using batch compression.
// Format: [length (4 bytes)][batch compressed pairs][optional last point if odd]
func (enc *BatchEncoder) EncodeG1(points []G1Affine) error {
	// Write slice length
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(len(points)))
	written, err := enc.w.Write(buf[:])
	enc.n += int64(written)
	if err != nil {
		return err
	}

	if len(points) == 0 {
		return nil
	}

	// Compress and write
	data, err := BatchCompressG1Slice(points)
	if err != nil {
		return err
	}

	written, err = enc.w.Write(data)
	enc.n += int64(written)
	return err
}

// DecodeG1 reads a slice of G1Affine points from batch-compressed form.
func (dec *BatchDecoder) DecodeG1(points *[]G1Affine) error {
	// Read slice length
	var buf [4]byte
	read, err := io.ReadFull(dec.r, buf[:])
	dec.n += int64(read)
	if err != nil {
		return err
	}

	n := int(binary.BigEndian.Uint32(buf[:]))
	if n == 0 {
		*points = nil
		return nil
	}

	// Calculate expected data size
	nPairs := n / 2
	hasOdd := n%2 == 1
	dataSize := nPairs * SizeOfBatchCompressedG1Pair
	if hasOdd {
		dataSize += SizeOfG1AffineCompressed
	}

	// Read compressed data
	data := make([]byte, dataSize)
	read, err = io.ReadFull(dec.r, data)
	dec.n += int64(read)
	if err != nil {
		return err
	}

	// Decompress
	result, err := BatchDecompressG1Slice(data, n)
	if err != nil {
		return err
	}

	// Subgroup check in parallel
	if dec.subGroupCheck {
		var nbErrs uint64
		parallel.Execute(len(result), func(start, end int) {
			for i := start; i < end; i++ {
				if !result[i].IsInSubGroup() {
					atomic.AddUint64(&nbErrs, 1)
				}
			}
		})
		if nbErrs != 0 {
			return errors.New("invalid point: subgroup check failed")
		}
	}

	*points = result
	return nil
}

// EncodeG2 writes a slice of G2Affine points using batch compression.
// Format: [length (4 bytes)][batch compressed pairs][optional last point if odd]
func (enc *BatchEncoder) EncodeG2(points []G2Affine) error {
	// Write slice length
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(len(points)))
	written, err := enc.w.Write(buf[:])
	enc.n += int64(written)
	if err != nil {
		return err
	}

	if len(points) == 0 {
		return nil
	}

	// Compress and write
	data, err := BatchCompressG2Slice(points)
	if err != nil {
		return err
	}

	written, err = enc.w.Write(data)
	enc.n += int64(written)
	return err
}

// DecodeG2 reads a slice of G2Affine points from batch-compressed form.
func (dec *BatchDecoder) DecodeG2(points *[]G2Affine) error {
	// Read slice length
	var buf [4]byte
	read, err := io.ReadFull(dec.r, buf[:])
	dec.n += int64(read)
	if err != nil {
		return err
	}

	n := int(binary.BigEndian.Uint32(buf[:]))
	if n == 0 {
		*points = nil
		return nil
	}

	// Calculate expected data size
	nPairs := n / 2
	hasOdd := n%2 == 1
	dataSize := nPairs * SizeOfBatchCompressedG2Pair
	if hasOdd {
		dataSize += SizeOfG2AffineCompressed
	}

	// Read compressed data
	data := make([]byte, dataSize)
	read, err = io.ReadFull(dec.r, data)
	dec.n += int64(read)
	if err != nil {
		return err
	}

	// Decompress
	result, err := BatchDecompressG2Slice(data, n)
	if err != nil {
		return err
	}

	// Subgroup check in parallel
	if dec.subGroupCheck {
		var nbErrs uint64
		parallel.Execute(len(result), func(start, end int) {
			for i := start; i < end; i++ {
				if !result[i].IsInSubGroup() {
					atomic.AddUint64(&nbErrs, 1)
				}
			}
		})
		if nbErrs != 0 {
			return errors.New("invalid point: subgroup check failed")
		}
	}

	*points = result
	return nil
}

// BatchCompressedSizeG1 returns the size in bytes needed to store n G1 points
// using batch compression.
func BatchCompressedSizeG1(n int) int {
	if n == 0 {
		return 4 // just the length
	}
	nPairs := n / 2
	size := 4 + nPairs*SizeOfBatchCompressedG1Pair // length + pairs
	if n%2 == 1 {
		size += SizeOfG1AffineCompressed // odd point
	}
	return size
}

// BatchCompressedSizeG2 returns the size in bytes needed to store n G2 points
// using batch compression.
func BatchCompressedSizeG2(n int) int {
	if n == 0 {
		return 4 // just the length
	}
	nPairs := n / 2
	size := 4 + nPairs*SizeOfBatchCompressedG2Pair // length + pairs
	if n%2 == 1 {
		size += SizeOfG2AffineCompressed // odd point
	}
	return size
}

// WriteBatchG1 writes G1 points using batch compression to the writer.
// This is a convenience function that creates a BatchEncoder internally.
func WriteBatchG1(w io.Writer, points []G1Affine) (int64, error) {
	enc := NewBatchEncoder(w)
	if err := enc.EncodeG1(points); err != nil {
		return enc.BytesWritten(), err
	}
	return enc.BytesWritten(), nil
}

// ReadBatchG1 reads G1 points from batch-compressed form.
// This is a convenience function that creates a BatchDecoder internally.
func ReadBatchG1(r io.Reader, subGroupCheck bool) ([]G1Affine, int64, error) {
	var opts []func(*BatchDecoder)
	if !subGroupCheck {
		opts = append(opts, NoBatchSubgroupChecks())
	}
	dec := NewBatchDecoder(r, opts...)

	var points []G1Affine
	if err := dec.DecodeG1(&points); err != nil {
		return nil, dec.BytesRead(), err
	}
	return points, dec.BytesRead(), nil
}

// WriteBatchG2 writes G2 points using batch compression to the writer.
// This is a convenience function that creates a BatchEncoder internally.
func WriteBatchG2(w io.Writer, points []G2Affine) (int64, error) {
	enc := NewBatchEncoder(w)
	if err := enc.EncodeG2(points); err != nil {
		return enc.BytesWritten(), err
	}
	return enc.BytesWritten(), nil
}

// ReadBatchG2 reads G2 points from batch-compressed form.
// This is a convenience function that creates a BatchDecoder internally.
func ReadBatchG2(r io.Reader, subGroupCheck bool) ([]G2Affine, int64, error) {
	var opts []func(*BatchDecoder)
	if !subGroupCheck {
		opts = append(opts, NoBatchSubgroupChecks())
	}
	dec := NewBatchDecoder(r, opts...)

	var points []G2Affine
	if err := dec.DecodeG2(&points); err != nil {
		return nil, dec.BytesRead(), err
	}
	return points, dec.BytesRead(), nil
}

// BatchEncoding returns an option to use with NewEncoder that enables batch compression.
// Note: This option is only effective for encoding []G1Affine and []G2Affine slices
// when using the BatchEncoder directly. For the standard Encoder, use BatchEncoder instead.
func BatchEncoding() func(*Encoder) {
	return func(enc *Encoder) {
		// This is a marker option - the actual batch encoding is done via BatchEncoder
		// Keeping this for API consistency
	}
}

// Ensure BatchEncoder/BatchDecoder implement the expected patterns
var (
	_ = (*BatchEncoder)(nil)
	_ = (*BatchDecoder)(nil)
)

// Helper to remove unused import warning
var _ = fptower.E2{}
var _ = fp.Element{}
