// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// FOO

package starkcurve

import (
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"sync/atomic"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fr"
	"github.com/consensys/gnark-crypto/internal/parallel"
)

// To encode G1Affine points, we mask the most significant bits with these bits to specify without ambiguity
// metadata needed for point (de)compression
// we have less than 3 bits available on the msw, so we can't follow BLS12-381 style encoding.
// the difference is the case where a point is infinity and uncompressed is not flagged
const (
	mMask               byte = 0b11 << 6
	mUncompressed       byte = 0b00 << 6
	mCompressedSmallest byte = 0b10 << 6
	mCompressedLargest  byte = 0b11 << 6
	mCompressedInfinity byte = 0b01 << 6
)

// Encoder writes stark-curve object values to an output stream
type Encoder struct {
	w   io.Writer
	n   int64 // written bytes
	raw bool  // raw vs compressed encoding
}

// Decoder reads stark-curve object values from an inbound stream
type Decoder struct {
	r             io.Reader
	n             int64 // read bytes
	subGroupCheck bool  // default to true
}

// NewDecoder returns a binary decoder supporting curve stark-curve objects in both
// compressed and uncompressed (raw) forms
func NewDecoder(r io.Reader, options ...func(*Decoder)) *Decoder {
	d := &Decoder{r: r, subGroupCheck: true}

	for _, o := range options {
		o(d)
	}

	return d
}

// Decode reads the binary encoding of v from the stream
// type must be *uint64, *fr.Element, *fp.Element, *G1Affine or *[]G1Affine
func (dec *Decoder) Decode(v interface{}) (err error) {
	rv := reflect.ValueOf(v)
	if v == nil || rv.Kind() != reflect.Ptr || rv.IsNil() || !rv.Elem().CanSet() {
		return errors.New("stark-curve decoder: unsupported type, need pointer")
	}

	// implementation note: code is a bit verbose (abusing code generation), but minimize allocations on the heap
	// in particular, careful attention must be given to usage of Bytes() method on Elements and Points
	// that return an array (not a slice) of bytes. Using this is beneficial to minimize memallocs
	// in very large (de)serialization upstream in gnark.
	// (but detrimental to code visibility here)

	var buf [SizeOfG1AffineUncompressed]byte
	var read int

	switch t := v.(type) {
	case *fr.Element:
		read, err = io.ReadFull(dec.r, buf[:fr.Bytes])
		dec.n += int64(read)
		if err != nil {
			return
		}
		err = t.SetBytesCanonical(buf[:fr.Bytes])
		return
	case *fp.Element:
		read, err = io.ReadFull(dec.r, buf[:fp.Bytes])
		dec.n += int64(read)
		if err != nil {
			return
		}
		err = t.SetBytesCanonical(buf[:fp.Bytes])
		return
	case *[]fr.Element:
		var sliceLen uint32
		sliceLen, err = dec.readUint32()
		if err != nil {
			return
		}
		if len(*t) != int(sliceLen) {
			*t = make([]fr.Element, sliceLen)
		}

		for i := 0; i < len(*t); i++ {
			read, err = io.ReadFull(dec.r, buf[:fr.Bytes])
			dec.n += int64(read)
			if err != nil {
				return
			}
			if err = (*t)[i].SetBytesCanonical(buf[:fr.Bytes]); err != nil {
				return
			}
		}
		return
	case *[]fp.Element:
		var sliceLen uint32
		sliceLen, err = dec.readUint32()
		if err != nil {
			return
		}
		if len(*t) != int(sliceLen) {
			*t = make([]fp.Element, sliceLen)
		}

		for i := 0; i < len(*t); i++ {
			read, err = io.ReadFull(dec.r, buf[:fp.Bytes])
			dec.n += int64(read)
			if err != nil {
				return
			}
			if err = (*t)[i].SetBytesCanonical(buf[:fp.Bytes]); err != nil {
				return
			}
		}
		return
	case *G1Affine:
		// we start by reading compressed point size, if metadata tells us it is uncompressed, we read more.
		read, err = io.ReadFull(dec.r, buf[:SizeOfG1AffineCompressed])
		dec.n += int64(read)
		if err != nil {
			return
		}
		nbBytes := SizeOfG1AffineCompressed
		// most significant byte contains metadata
		if !isCompressed(buf[0]) {
			nbBytes = SizeOfG1AffineUncompressed
			// we read more.
			read, err = io.ReadFull(dec.r, buf[SizeOfG1AffineCompressed:SizeOfG1AffineUncompressed])
			dec.n += int64(read)
			if err != nil {
				return
			}
		}
		_, err = t.setBytes(buf[:nbBytes], dec.subGroupCheck)
		return
	case *[]G1Affine:
		var sliceLen uint32
		sliceLen, err = dec.readUint32()
		if err != nil {
			return
		}
		if len(*t) != int(sliceLen) {
			*t = make([]G1Affine, sliceLen)
		}
		compressed := make([]bool, sliceLen)
		for i := 0; i < len(*t); i++ {

			// we start by reading compressed point size, if metadata tells us it is uncompressed, we read more.
			read, err = io.ReadFull(dec.r, buf[:SizeOfG1AffineCompressed])
			dec.n += int64(read)
			if err != nil {
				return
			}
			nbBytes := SizeOfG1AffineCompressed
			// most significant byte contains metadata
			if !isCompressed(buf[0]) {
				nbBytes = SizeOfG1AffineUncompressed
				// we read more.
				read, err = io.ReadFull(dec.r, buf[SizeOfG1AffineCompressed:SizeOfG1AffineUncompressed])
				dec.n += int64(read)
				if err != nil {
					return
				}
				_, err = (*t)[i].setBytes(buf[:nbBytes], false)
				if err != nil {
					return
				}
			} else {
				var r bool
				if r, err = ((*t)[i].unsafeSetCompressedBytes(buf[:nbBytes])); err != nil {
					return
				}
				compressed[i] = !r
			}
		}
		var nbErrs uint64
		parallel.Execute(len(compressed), func(start, end int) {
			for i := start; i < end; i++ {
				if compressed[i] {
					if err := (*t)[i].unsafeComputeY(dec.subGroupCheck); err != nil {
						atomic.AddUint64(&nbErrs, 1)
					}
				} else if dec.subGroupCheck {
					if !(*t)[i].IsInSubGroup() {
						atomic.AddUint64(&nbErrs, 1)
					}
				}
			}
		})
		if nbErrs != 0 {
			return errors.New("point decompression failed")
		}

		return nil
	default:
		n := binary.Size(t)
		if n == -1 {
			return errors.New("stark-curve encoder: unsupported type")
		}
		err = binary.Read(dec.r, binary.BigEndian, t)
		if err == nil {
			dec.n += int64(n)
		}
		return
	}
}

// BytesRead return total bytes read from reader
func (dec *Decoder) BytesRead() int64 {
	return dec.n
}

func (dec *Decoder) readUint32() (r uint32, err error) {
	var read int
	var buf [4]byte
	read, err = io.ReadFull(dec.r, buf[:4])
	dec.n += int64(read)
	if err != nil {
		return
	}
	r = binary.BigEndian.Uint32(buf[:4])
	return
}

func isCompressed(msb byte) bool {
	mData := msb & mMask
	return mData != mUncompressed
}

// NewEncoder returns a binary encoder supporting curve stark-curve objects
func NewEncoder(w io.Writer, options ...func(*Encoder)) *Encoder {
	// default settings
	enc := &Encoder{
		w:   w,
		n:   0,
		raw: false,
	}

	// handle options
	for _, option := range options {
		option(enc)
	}

	return enc
}

// Encode writes the binary encoding of v to the stream
// type must be uint64, *fr.Element, *fp.Element, *G1Affine, *G2Affine, []G1Affine or []G2Affine
func (enc *Encoder) Encode(v interface{}) (err error) {
	if enc.raw {
		return enc.encodeRaw(v)
	}
	return enc.encode(v)
}

// BytesWritten return total bytes written on writer
func (enc *Encoder) BytesWritten() int64 {
	return enc.n
}

// RawEncoding returns an option to use in NewEncoder(...) which sets raw encoding mode to true
// points will not be compressed using this option
func RawEncoding() func(*Encoder) {
	return func(enc *Encoder) {
		enc.raw = true
	}
}

// NoSubgroupChecks returns an option to use in NewDecoder(...) which disable subgroup checks on the points
// the decoder will read. Use with caution, as crafted points from an untrusted source can lead to crypto-attacks.
func NoSubgroupChecks() func(*Decoder) {
	return func(dec *Decoder) {
		dec.subGroupCheck = false
	}
}

func (enc *Encoder) encode(v interface{}) (err error) {
	rv := reflect.ValueOf(v)
	if v == nil || (rv.Kind() == reflect.Ptr && rv.IsNil()) {
		return errors.New("<no value> encoder: can't encode <nil>")
	}

	// implementation note: code is a bit verbose (abusing code generation), but minimize allocations on the heap

	var written int
	switch t := v.(type) {
	case *fr.Element:
		buf := t.Bytes()
		written, err = enc.w.Write(buf[:])
		enc.n += int64(written)
		return
	case *fp.Element:
		buf := t.Bytes()
		written, err = enc.w.Write(buf[:])
		enc.n += int64(written)
		return
	case *G1Affine:
		buf := t.Bytes()
		written, err = enc.w.Write(buf[:])
		enc.n += int64(written)
		return
	case []fr.Element:
		// write slice length
		err = binary.Write(enc.w, binary.BigEndian, uint32(len(t)))
		if err != nil {
			return
		}
		enc.n += 4
		var buf [fr.Bytes]byte
		for i := 0; i < len(t); i++ {
			buf = t[i].Bytes()
			written, err = enc.w.Write(buf[:])
			enc.n += int64(written)
			if err != nil {
				return
			}
		}
		return nil
	case []fp.Element:
		// write slice length
		err = binary.Write(enc.w, binary.BigEndian, uint32(len(t)))
		if err != nil {
			return
		}
		enc.n += 4
		var buf [fp.Bytes]byte
		for i := 0; i < len(t); i++ {
			buf = t[i].Bytes()
			written, err = enc.w.Write(buf[:])
			enc.n += int64(written)
			if err != nil {
				return
			}
		}
		return nil

	case []G1Affine:
		// write slice length
		err = binary.Write(enc.w, binary.BigEndian, uint32(len(t)))
		if err != nil {
			return
		}
		enc.n += 4

		var buf [SizeOfG1AffineCompressed]byte

		for i := 0; i < len(t); i++ {
			buf = t[i].Bytes()
			written, err = enc.w.Write(buf[:])
			enc.n += int64(written)
			if err != nil {
				return
			}
		}
		return nil
	default:
		n := binary.Size(t)
		if n == -1 {
			return errors.New("<no value> encoder: unsupported type")
		}
		err = binary.Write(enc.w, binary.BigEndian, t)
		enc.n += int64(n)
		return
	}
}

func (enc *Encoder) encodeRaw(v interface{}) (err error) {
	rv := reflect.ValueOf(v)
	if v == nil || (rv.Kind() == reflect.Ptr && rv.IsNil()) {
		return errors.New("<no value> encoder: can't encode <nil>")
	}

	// implementation note: code is a bit verbose (abusing code generation), but minimize allocations on the heap

	var written int
	switch t := v.(type) {
	case *fr.Element:
		buf := t.Bytes()
		written, err = enc.w.Write(buf[:])
		enc.n += int64(written)
		return
	case *fp.Element:
		buf := t.Bytes()
		written, err = enc.w.Write(buf[:])
		enc.n += int64(written)
		return
	case *G1Affine:
		buf := t.RawBytes()
		written, err = enc.w.Write(buf[:])
		enc.n += int64(written)
		return
	case []fr.Element:
		// write slice length
		err = binary.Write(enc.w, binary.BigEndian, uint32(len(t)))
		if err != nil {
			return
		}
		enc.n += 4
		var buf [fr.Bytes]byte
		for i := 0; i < len(t); i++ {
			buf = t[i].Bytes()
			written, err = enc.w.Write(buf[:])
			enc.n += int64(written)
			if err != nil {
				return
			}
		}
		return nil
	case []fp.Element:
		// write slice length
		err = binary.Write(enc.w, binary.BigEndian, uint32(len(t)))
		if err != nil {
			return
		}
		enc.n += 4
		var buf [fp.Bytes]byte
		for i := 0; i < len(t); i++ {
			buf = t[i].Bytes()
			written, err = enc.w.Write(buf[:])
			enc.n += int64(written)
			if err != nil {
				return
			}
		}
		return nil

	case []G1Affine:
		// write slice length
		err = binary.Write(enc.w, binary.BigEndian, uint32(len(t)))
		if err != nil {
			return
		}
		enc.n += 4

		var buf [SizeOfG1AffineUncompressed]byte

		for i := 0; i < len(t); i++ {
			buf = t[i].RawBytes()
			written, err = enc.w.Write(buf[:])
			enc.n += int64(written)
			if err != nil {
				return
			}
		}
		return nil
	default:
		n := binary.Size(t)
		if n == -1 {
			return errors.New("<no value> encoder: unsupported type")
		}
		err = binary.Write(enc.w, binary.BigEndian, t)
		enc.n += int64(n)
		return
	}
}

// SizeOfG1AffineCompressed represents the size in bytes that a G1Affine need in binary form, compressed
const SizeOfG1AffineCompressed = 32

// SizeOfG1AffineUncompressed represents the size in bytes that a G1Affine need in binary form, uncompressed
const SizeOfG1AffineUncompressed = SizeOfG1AffineCompressed * 2

// Marshal converts p to a byte slice (without point compression)
func (p *G1Affine) Marshal() []byte {
	b := p.RawBytes()
	return b[:]
}

// Unmarshal is an alias to SetBytes()
func (p *G1Affine) Unmarshal(buf []byte) error {
	_, err := p.SetBytes(buf)
	return err
}

// Bytes returns binary representation of p
// will store X coordinate in regular form and a parity bit
// as we have less than 3 bits available in our coordinate, we can't follow BLS12-381 style encoding (ZCash/IETF)
//
// we use the 2 most significant bits instead
//
//	00 -> uncompressed
//	10 -> compressed, use smallest lexicographically square root of Y^2
//	11 -> compressed, use largest lexicographically square root of Y^2
//	01 -> compressed infinity point
//	the "uncompressed infinity point" will just have 00 (uncompressed) followed by zeroes (infinity = 0,0 in affine coordinates)
func (p *G1Affine) Bytes() (res [SizeOfG1AffineCompressed]byte) {

	// check if p is infinity point
	if p.X.IsZero() && p.Y.IsZero() {
		res[0] = mCompressedInfinity
		return
	}

	msbMask := mCompressedSmallest
	// compressed, we need to know if Y is lexicographically bigger than -Y
	// if p.Y ">" -p.Y
	if p.Y.LexicographicallyLargest() {
		msbMask = mCompressedLargest
	}

	// we store X  and mask the most significant word with our metadata mask
	fp.BigEndian.PutElement((*[fp.Bytes]byte)(res[0:0+fp.Bytes]), p.X)

	res[0] |= msbMask

	return
}

// RawBytes returns binary representation of p (stores X and Y coordinate)
// see Bytes() for a compressed representation
func (p *G1Affine) RawBytes() (res [SizeOfG1AffineUncompressed]byte) {

	// check if p is infinity point
	if p.X.IsZero() && p.Y.IsZero() {

		res[0] = mUncompressed

		return
	}

	// not compressed
	// we store the Y coordinate
	fp.BigEndian.PutElement((*[fp.Bytes]byte)(res[32:32+fp.Bytes]), p.Y)

	// we store X  and mask the most significant word with our metadata mask
	fp.BigEndian.PutElement((*[fp.Bytes]byte)(res[0:0+fp.Bytes]), p.X)

	res[0] |= mUncompressed

	return
}

// SetBytes sets p from binary representation in buf and returns number of consumed bytes
//
// bytes in buf must match either RawBytes() or Bytes() output
//
// if buf is too short io.ErrShortBuffer is returned
//
// if buf contains compressed representation (output from Bytes()) and we're unable to compute
// the Y coordinate (i.e the square root doesn't exist) this function returns an error
//
// this check if the resulting point is on the curve and in the correct subgroup
func (p *G1Affine) SetBytes(buf []byte) (int, error) {
	return p.setBytes(buf, true)
}

func (p *G1Affine) setBytes(buf []byte, subGroupCheck bool) (int, error) {
	if len(buf) < SizeOfG1AffineCompressed {
		return 0, io.ErrShortBuffer
	}

	// most significant byte
	mData := buf[0] & mMask

	// check buffer size
	if mData == mUncompressed {
		if len(buf) < SizeOfG1AffineUncompressed {
			return 0, io.ErrShortBuffer
		}
	}

	// if infinity is encoded in the metadata, we don't need to read the buffer
	if mData == mCompressedInfinity {
		p.X.SetZero()
		p.Y.SetZero()
		return SizeOfG1AffineCompressed, nil
	}

	// uncompressed point
	if mData == mUncompressed {
		// read X and Y coordinates
		if err := p.X.SetBytesCanonical(buf[:fp.Bytes]); err != nil {
			return 0, err
		}
		if err := p.Y.SetBytesCanonical(buf[fp.Bytes : fp.Bytes*2]); err != nil {
			return 0, err
		}

		// subgroup check
		if subGroupCheck && !p.IsInSubGroup() {
			return 0, errors.New("invalid point: subgroup check failed")
		}

		return SizeOfG1AffineUncompressed, nil
	}

	// we have a compressed coordinate
	// we need to
	// 	1. copy the buffer (to keep this method thread safe)
	// 	2. we need to solve the curve equation to compute Y

	var bufX [fp.Bytes]byte
	copy(bufX[:fp.Bytes], buf[:fp.Bytes])
	bufX[0] &= ^mMask

	// read X coordinate
	if err := p.X.SetBytesCanonical(bufX[:fp.Bytes]); err != nil {
		return 0, err
	}

	var YSquared, Y fp.Element

	// y^2=x^3+x+b
	YSquared.Square(&p.X).Mul(&YSquared, &p.X)
	YSquared.Add(&YSquared, &p.X).
		Add(&YSquared, &bCurveCoeff)

	if Y.Sqrt(&YSquared) == nil {
		return 0, errors.New("invalid compressed coordinate: square root doesn't exist")
	}

	if Y.LexicographicallyLargest() {
		// Y ">" -Y
		if mData == mCompressedSmallest {
			Y.Neg(&Y)
		}
	} else {
		// Y "<=" -Y
		if mData == mCompressedLargest {
			Y.Neg(&Y)
		}
	}

	p.Y = Y

	// subgroup check
	if subGroupCheck && !p.IsInSubGroup() {
		return 0, errors.New("invalid point: subgroup check failed")
	}

	return SizeOfG1AffineCompressed, nil
}

// unsafeComputeY called by Decoder when processing slices of compressed point in parallel (step 2)
// it computes the Y coordinate from the already set X coordinate and is compute intensive
func (p *G1Affine) unsafeComputeY(subGroupCheck bool) error {
	// stored in unsafeSetCompressedBytes

	mData := byte(p.Y[0])

	// we have a compressed coordinate, we need to solve the curve equation to compute Y
	var YSquared, Y fp.Element

	// y^2=x^3+x+b
	YSquared.Square(&p.X).Mul(&YSquared, &p.X)
	YSquared.Add(&YSquared, &p.X).
		Add(&YSquared, &bCurveCoeff)

	if Y.Sqrt(&YSquared) == nil {
		return errors.New("invalid compressed coordinate: square root doesn't exist")
	}

	if Y.LexicographicallyLargest() {
		// Y ">" -Y
		if mData == mCompressedSmallest {
			Y.Neg(&Y)
		}
	} else {
		// Y "<=" -Y
		if mData == mCompressedLargest {
			Y.Neg(&Y)
		}
	}

	p.Y = Y

	// subgroup check
	if subGroupCheck && !p.IsInSubGroup() {
		return errors.New("invalid point: subgroup check failed")
	}

	return nil
}

// unsafeSetCompressedBytes is called by Decoder when processing slices of compressed point in parallel (step 1)
// assumes buf[:8] mask is set to compressed
// returns true if point is infinity and need no further processing
// it sets X coordinate and uses Y for scratch space to store decompression metadata
func (p *G1Affine) unsafeSetCompressedBytes(buf []byte) (isInfinity bool, err error) {

	// read the most significant byte
	mData := buf[0] & mMask

	if mData == mCompressedInfinity {
		p.X.SetZero()
		p.Y.SetZero()
		isInfinity = true
		return isInfinity, nil
	}

	// we need to copy the input buffer (to keep this method thread safe)
	var bufX [fp.Bytes]byte
	copy(bufX[:fp.Bytes], buf[:fp.Bytes])
	bufX[0] &= ^mMask

	// read X coordinate
	if err := p.X.SetBytesCanonical(bufX[:fp.Bytes]); err != nil {
		return false, err
	}
	// store mData in p.Y[0]
	p.Y[0] = uint64(mData)

	// recomputing Y will be done asynchronously
	return isInfinity, nil
}
